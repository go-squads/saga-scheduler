package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pborman/uuid"
	log "github.com/sirupsen/logrus"
)

type scheduler struct {
	Router    *mux.Router
	DB        PostgresQL
	client    client
	metricsDB metricsDB
}

type PostgresQL interface {
	Queryx(query string, args ...interface{}) (*sqlx.Rows, error)
	NamedExec(query string, arg interface{}) (sql.Result, error)
	Exec(query string, args ...interface{}) (sql.Result, error)
}

type client interface {
	executeOperationRequest(req *http.Request) (*operation, error)
}

type agentClient struct{}

func (a agentClient) executeOperationRequest(req *http.Request) (*operation, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	response, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	var op *operation

	err = json.Unmarshal(body, &op)
	if err != nil {
		return nil, err
	}

	return op, nil
}

func (s *scheduler) initialize(user, password, dbname, host, port, sslmode string) error {
	connectionString := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode=%s", user, password, dbname, host, port, sslmode)
	var err error
	s.DB, err = sqlx.Connect("postgres", connectionString)
	if err != nil {
		return err
	}

	s.Router = mux.NewRouter()
	s.Router.HandleFunc("/api/v1/lxc", s.createNewLxcHandler).Methods("POST")
	s.Router.HandleFunc("/api/v1/lxc", s.getContainerHandler).Methods("GET")
	s.Router.HandleFunc("/api/v1/lxc", s.updateLxcStatusByIDHandler).Methods("PUT")
	s.Router.HandleFunc("/api/v1/lxc", s.deleteLxcHandler).Methods("DELETE")
	s.Router.HandleFunc("/api/v1/lxd/{lxdName}/lxc", s.getLxcListByLxdNameHandler).Methods("GET")
	s.Router.HandleFunc("/api/v1/lxc-services", s.createNewLxcServiceHandler).Methods("POST")
	s.Router.HandleFunc("/api/v1/lxc-services/{lxdName}", s.getLxcServicesListHandler).Methods("GET")
	s.Router.HandleFunc("/api/v1/lxc-services", s.updateLxcServicesStatusByIDHandler).Methods("PUT")
	s.Router.HandleFunc("/api/v1/lxc-services/lxc/{lxcId}", s.getLxcServiceListByLxcIDHandler).Methods("GET")
	s.client = agentClient{}
	s.metricsDB = prometheusMetricsDB{}

	return nil
}

func (s *scheduler) run(port string) {
	log.Fatal(http.ListenAndServe(port, s.Router))
}

func (s *scheduler) getContainerHandler(w http.ResponseWriter, r *http.Request) {
	type resp struct {
		ID      string `json:"id" db:"id"`
		LxdID   string `json:"lxd_id" db:"lxd_id"`
		LXDName string `json:"lxd_name" db:"lxd_name"`
		LXCName string `json:"lxc_name" db:"lxc_name"`
		Image   string `json:"image" db:"image"`
		Status  string `json:"status" db:"status"`
	}

	var result []resp
	rows, err := s.DB.Queryx(`SELECT c.id as "id", c.name as "lxc_name", d.id as "lxd_id", d.name as "lxd_name", c.alias as "image", c.status as "status" FROM lxc c JOIN lxd d ON c.lxd_id = d.id`)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	for rows.Next() {
		var temp resp
		err = rows.StructScan(&temp)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, err.Error())
		}
		result = append(result, temp)
	}

	respondWithJSON(w, http.StatusOK, result)
}

// BUG IF LXC ALREADY EXIST
func (s *scheduler) createNewLxcHandler(w http.ResponseWriter, r *http.Request) {
	log.Info("-- Got new create lxc request --")
	newLxc := lxc{}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&newLxc); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	exist := newLxc.checkIfLxcExist(s.DB)
	if !exist {
		lxdInstance, err := s.metricsDB.getLowestLoadLxdInstance(newLxc.WeightScore, newLxc.WeightValue, s.DB)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		err = lxdInstance.getLxdByIP(s.DB)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		newLxc.ID = uuid.New()
		newLxc.Status = "creating"
		newLxc.LxdID = lxdInstance.ID

		err = newLxc.insertLxc(s.DB)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		respondWithJSON(w, http.StatusOK, newLxc)
	} else {
		errStr := fmt.Sprintf("Lxc with name %s already exists", newLxc.Name)
		respondWithError(w, http.StatusBadRequest, errStr)
	}
}

func (s *scheduler) deleteLxcHandler(w http.ResponseWriter, r *http.Request) {
	log.Info("-- Got delete lxc request --")
	lxcDeleteData := lxc{}

	if err := json.NewDecoder(r.Body).Decode(&lxcDeleteData); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := lxcDeleteData.deleteLxc(s.DB); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "delete lxc success"})
}

func (s *scheduler) getLxcListByLxdNameHandler(w http.ResponseWriter, r *http.Request) {
	log.Info("-- Got get lxc by lxd name request --")
	vars := mux.Vars(r)
	lxdName := vars["lxdName"]
	lxdSearch := lxd{Name: lxdName}

	if err := lxdSearch.getLxdIDByName(s.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	lxcSearch := lxc{}
	lxcList, err := lxcSearch.getLxcListByLxdID(s.DB, lxdSearch.ID)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, lxcList)
}

func (s *scheduler) updateLxcStatusByIDHandler(w http.ResponseWriter, r *http.Request) {
	log.Info("-- Got update lxc status by id request --")

	lxcUpdateData := lxc{}

	if err := json.NewDecoder(r.Body).Decode(&lxcUpdateData); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := lxcUpdateData.updateStatusByID(s.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "success updating lxc state"})
}

func (s *scheduler) createNewLxcServiceHandler(w http.ResponseWriter, r *http.Request) {
	log.Info("-- Got new create lxc services request --")
	newLxcService := lxcService{}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&newLxcService); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	log.Infof("LXC ID %s", newLxcService.LxcID)
	lxc := lxc{ID: newLxcService.LxcID}
	lxd := lxd{ID: newLxcService.LxdID}
	lxcName := lxc.getLxcNameByID(s.DB)
	lxdAddress, lxdName := lxd.getLxdNameAndAddressByID(s.DB)

	newLxcService.LxcName = lxcName
	newLxcService.LxdName = lxdName
	newLxcService.LxdAddress = lxdAddress
	exist := newLxcService.checkIfLxcServiceExist(s.DB)
	if !exist {
		newLxcService.ID = uuid.New()

		if err := newLxcService.insertLxcService(s.DB); err != nil {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		respondWithJSON(w, http.StatusOK, newLxcService)
	} else {
		errMsg := fmt.Sprintf("Lxc Service: %s for lxc: %s on %s:%s already exists", newLxcService.Service, newLxcService.LxcName, newLxcService.LxdPort, newLxcService.LxcPort)
		respondWithError(w, http.StatusBadRequest, errMsg)
	}
}

func (s *scheduler) getLxcServicesListHandler(w http.ResponseWriter, r *http.Request) {
	log.Info("==> Received Request for Lxc Services List <==")
	vars := mux.Vars(r)
	lxdName := vars["lxdName"]
	lxcServicesList, err := getLxcServicesListByLxdName(s.DB, lxdName)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, lxcServicesList)
}

func (s *scheduler) updateLxcServicesStatusByIDHandler(w http.ResponseWriter, r *http.Request) {
	l := lxcService{}
	if err := json.NewDecoder(r.Body).Decode(&l); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := l.updateStatusByID(s.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "success updating lxc service status"})
}

func (s *scheduler) getLxcServiceListByLxcIDHandler(w http.ResponseWriter, r *http.Request) {
	log.Info("Get lxc services list by lxc id request")
	vars := mux.Vars(r)
	lxcId := vars["lxcId"]
	l := lxcService{LxcID: lxcId}
	lxcServicesList, err := l.getLxcServicesListByLxcID(s.DB)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, lxcServicesList)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(code)
	w.Write(response)
}

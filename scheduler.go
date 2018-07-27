package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pborman/uuid"
)

type scheduler struct {
	Router *mux.Router
	DB     *sqlx.DB
	client client
}

type createContainerRequestData struct {
	Name     string `json:"name,omitempty"`
	Type     string `json:"type,omitempty"`
	Protocol string `json:"protocol,omitempty"`
	Server   string `json:"server,omitempty"`
	Alias    string `json:"alias,omitempty"`
}

type client interface {
	executeRequest(req *http.Request) (*http.Response, error)
}

type agentClient struct{}

func (a agentClient) executeRequest(req *http.Request) (*http.Response, error) {
	client := &http.Client{
		Timeout: 10,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *scheduler) initialize(user, password, dbname, host, port, sslmode string) error {
	connectionString := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode=%s", user, password, dbname, host, port, sslmode)
	var err error
	s.DB, err = sqlx.Connect("postgres", connectionString)
	if err != nil {
		return err
	}
	s.Router = mux.NewRouter()
	s.Router.HandleFunc("/api/v1/container", s.createNewLxcHandler).Methods("POST")

	s.client = agentClient{}

	return nil
}

func (s *scheduler) run(port string) {
	log.Fatal(http.ListenAndServe(port, s.Router))
}

func (s *scheduler) createNewLxcHandler(w http.ResponseWriter, r *http.Request) {
	var data createContainerRequestData
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&data); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	defer r.Body.Close()

	//some lxd chooser algorithm ...
	lxdIPAddress := "172.28.128.6"
	lxdID := "very-unique-lxd-uuid"

	newLxc := lxc{
		ID:         uuid.New(),
		LxdID:      lxdID,
		Name:       data.Name,
		Type:       data.Type,
		Alias:      data.Alias,
		IsDeployed: 1,
	}

	err := newLxc.insertLxc(s.DB)
	if err != nil {
		fmt.Println(err.Error())
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	op, err := s.createNewLxc(data, lxdIPAddress)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	err = op.insertOperation(s.DB)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, nil)
	return
}

func (s *scheduler) createNewLxc(data createContainerRequestData, lxdIPAddress string) (op *operation, err error) {
	url := fmt.Sprintf("%s/api/v1/container", lxdIPAddress)

	payload, err := json.Marshal(data)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}

	response, err := s.client.executeRequest(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(body, &op)
	if err != nil {
		return nil, err
	}

	return op, nil
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

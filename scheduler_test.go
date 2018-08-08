package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/pborman/uuid"
	"github.com/stretchr/testify/suite"
)

var schedulerSuites scheduler

type SchedulerSuite struct {
	suite.Suite
}

type testAgentClient struct{}

func (t testAgentClient) executeOperationRequest(req *http.Request) (*operation, error) {
	op := operation{
		ID:         uuid.New(),
		LxcID:      "very-unique-lxc-uuid",
		Status:     "Running",
		StatusCode: 200,
	}

	return &op, nil
}

func TestSchedulerSuite(t *testing.T) {
	suite.Run(t, new(SchedulerSuite))
}

func (suite *SchedulerSuite) SetupSuite() {
	schedulerSuites = scheduler{}
	err := schedulerSuites.initialize("postgres", "postgres", "saga", "localhost", "5432", "disable")

	suite.NoError(err, "They should be no error")

	schedulerSuites.client = testAgentClient{}
	schedulerSuites.metricsDB = testPrometheusMetricsDB{}

	clearQuery := `DELETE FROM operation;
	DELETE FROM lxc;
	DELETE FROM lxd;`

	_, err = schedulerSuites.DB.Exec(clearQuery)
	suite.NoError(err, "They should be no error")

	_, err = operationScheduler.DB.Exec("INSERT INTO lxd (id, name, address) VALUES ('very-unique-lxd-uuid', 'test-lxd', 'test.gojek.com');")
	suite.NoError(err, "They should be no error")

	_, err = operationScheduler.DB.Exec("INSERT INTO lxc (id, lxd_id, name, type, alias, is_deployed) VALUES ('very-unique-lxc-uuid','very-unique-lxd-uuid', 'test-lxc-1', 'image', '16.04', 0);")
	suite.NoError(err, "They should be no error")

	valueSmall = append(valueSmall, 123.456)
	valueSmall = append(valueSmall, "30.0")

	valueLarge = append(valueLarge, 123.456)
	valueLarge = append(valueLarge, "70.0")

	testMemResponse = promResponse{
		Status: "success",
		Data: promResponseData{
			Result: []result{
				result{
					Metric: metric{
						Instance: "192.168.0.1:9090",
					},
					Value: valueLarge,
				},
				result{
					Metric: metric{
						Instance: "Bar",
					},
					Value: valueSmall,
				},
			},
		},
	}

}

func (suite *SchedulerSuite) TearDownSuite() {
	clearQuery := `DELETE FROM operation;
	DELETE FROM lxc;
	DELETE FROM lxd;`

	_, err := schedulerSuites.DB.Exec(clearQuery)
	suite.NoError(err, "They should be no error")
}

func (suite *SchedulerSuite) TestCreateNewLxcHandlerSuccessful() {
	payload := []byte(`{"name":"test-container-12","type":"none"}`)
	req, err := http.NewRequest("POST", "/api/v1/container", bytes.NewBuffer(payload))
	if err != nil {
		suite.Fail(err.Error())
	}

	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/api/v1/container", schedulerSuites.createNewLxcHandler)
	router.ServeHTTP(rr, req)
	suite.Equal(http.StatusOK, rr.Code, "They should be equal")
}

func (suite *SchedulerSuite) TestUpdateStateLxcHandlerSuccessful() {
	payload := []byte(`{
		"id": "very-unique-lxc-uuid",
		"name":"container-new", 
		"state": {
			"action":"stop", 
			"timeout":60
		}
	}`)
	req, err := http.NewRequest("POST", "/api/v1/container/updatestate", bytes.NewBuffer(payload))
	if err != nil {
		suite.Fail(err.Error())
	}

	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/api/v1/container/updatestate", schedulerSuites.updateStateLxcHandler)
	router.ServeHTTP(rr, req)
	suite.Equal(http.StatusOK, rr.Code, "They should be equal")
}

func (suite *SchedulerSuite) TestDeleteLxcHandler() {
	payload := []byte(`{
		"id": "very-unique-lxc-uuid",
	}`)
	req, err := http.NewRequest("DELETE", "/api/v1/container", bytes.NewBuffer(payload))
	if err != nil {
		suite.Fail(err.Error())
	}

	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/api/v1/container", schedulerSuites.deleteLxcHandler)
	router.ServeHTTP(rr, req)
	suite.Equal(http.StatusOK, rr.Code, "They should be equal")
}

package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
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

func (t testAgentClient) executeRequest(req *http.Request) (*http.Response, error) {
	op := operation{
		ID:         uuid.New(),
		LxcID:      "very-unique-lxc-uuid",
		Status:     "Running",
		StatusCode: 200,
	}

	payload, _ := json.Marshal(&op)

	resp := http.Response{
		Body: ioutil.NopCloser(bytes.NewReader(payload)),
	}

	return &resp, nil
}

func TestSchedulerSuite(t *testing.T) {
	suite.Run(t, new(SchedulerSuite))
}

func (suite *SchedulerSuite) SetupSuite() {
	schedulerSuites = scheduler{}
	err := schedulerSuites.initialize("postgres", "postgres", "saga", "localhost", "5432", "disable")
	suite.NoError(err, "They should be no error")

	schedulerSuites.client = testAgentClient{}

	clearQuery := `DELETE FROM operation;
	DELETE FROM lxc;
	DELETE FROM lxd;`

	_, err = schedulerSuites.DB.Exec(clearQuery)
	suite.NoError(err, "They should be no error")

	_, err = operationScheduler.DB.Exec("INSERT INTO lxd (id, name, address) VALUES ('very-unique-lxd-uuid', 'test-lxd', 'test.gojek.com');")
	suite.NoError(err, "They should be no error")

	_, err = operationScheduler.DB.Exec("INSERT INTO lxc (id, lxd_id, name, type, alias, is_deployed) VALUES ('very-unique-lxc-uuid','very-unique-lxd-uuid', 'test-lxc-1', 'image', '16.04', 0);")
	suite.NoError(err, "They should be no error")

}

func (suite *SchedulerSuite) TearDownSuite() {
	// s.DB.Close()
}

// func (suite *SchedulerSuite) TestInitializeSuccessful() {
// 	err := s.initialize("postgres", "postgres", "saga", "localhost", "5432", "disable")
// 	suite.NoError(err, "They should be no error")
// }

// func (suite *SchedulerSuite) TestInitializeFailed() {
// 	err := s.initialize("postgres", "postgres", "saga", "localhost", "80", "disable")
// 	suite.Error(err, "They should be error")
// }

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

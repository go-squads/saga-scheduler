package main

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

var s scheduler

type SchedulerSuite struct {
	suite.Suite
}

func TestSchedulerSuite(t *testing.T) {
	suite.Run(t, new(SchedulerSuite))
}

func (suite *SchedulerSuite) SetupSuite() {

}

func (suite *SchedulerSuite) TearDownSuite() {
	s.DB.Close()
}

func (suite *SchedulerSuite) TestInitializeSuccessful() {
	err := s.initialize("postgres", "postgres", "saga", "localhost", "5432", "disable")
	suite.NoError(err, "They should be no error")
}

func (suite *SchedulerSuite) TestInitializeFailed() {
	err := s.initialize("postgres", "postgres", "saga", "localhost", "80", "disable")
	suite.Error(err, "They should be error")
}

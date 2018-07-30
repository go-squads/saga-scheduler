package main

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

var lxdScheduler scheduler

type LxdSuite struct {
	suite.Suite
}

func TestLxdSuite(t *testing.T) {
	suite.Run(t, new(LxdSuite))
}

func (suite *LxdSuite) SetupSuite() {
	lxdScheduler = scheduler{}
	err := lxdScheduler.initialize("postgres", "postgres", "saga", "localhost", "5432", "disable")
	suite.NoError(err, "They should be no error")

	clearQuery := `DELETE FROM lxd;`

	_, err = lxdScheduler.DB.Exec(clearQuery)
	suite.NoError(err, "They should be no error")
}

func (suite *LxdSuite) TearDownSuite() {
	clearQuery := `DELETE FROM lxd;`

	_, err := lxdScheduler.DB.Exec(clearQuery)
	suite.NoError(err, "They should be no error")
}

func (suite *LxdSuite) TestGetLxdSuccessful() {
	testLxd := lxd{
		ID: "1",
	}

	err := testLxd.getLxd(lxdScheduler.DB)
	suite.NoError(err, "They should be no error")

}

func (suite *LxdSuite) TestGetLxdByIPSuccessful() {
	testLxd := lxd{
		Address: "1",
	}

	err := testLxd.getLxd(lxdScheduler.DB)
	suite.NoError(err, "They should be no error")

}

func (suite *LxdSuite) TestInsertLxdSuccessful() {
	testLxd := lxd{
		ID:          "very-unique-lxc-uuid",
		Name:        "test-lxc-1",
		Address:     "192.0.0.1",
		Description: "",
	}

	err := testLxd.insertLxd(lxdScheduler.DB)
	suite.NoError(err, "They should be no error")
}

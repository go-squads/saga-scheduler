package main

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

var lxcScheduler scheduler

type LxcSuite struct {
	suite.Suite
}

func TestLxcSuite(t *testing.T) {
	suite.Run(t, new(LxcSuite))
}

func (suite *LxcSuite) SetupSuite() {
	lxcScheduler = scheduler{}
	err := lxcScheduler.initialize("postgres", "postgres", "saga", "localhost", "5432", "disable")
	suite.NoError(err, "They should be no error")

	clearQuery := `DELETE FROM operation;
	DELETE FROM lxc;
	DELETE FROM lxd;`

	_, err = lxcScheduler.DB.Exec(clearQuery)
	suite.NoError(err, "They should be no error")

	_, err = lxcScheduler.DB.Exec("INSERT INTO lxd (id, name, address) VALUES ('very-unique-lxd-uuid', 'test-lxd', 'test.gojek.com');")
	suite.NoError(err, "They should be no error")
}

func (suite *LxcSuite) TearDownSuite() {
	clearQuery := `DELETE FROM operation;
	DELETE FROM lxc;
	DELETE FROM lxd;`

	_, err := lxcScheduler.DB.Exec(clearQuery)
	suite.NoError(err, "They should be no error")
}

func (suite *LxcSuite) TestGetLxcSuccessful() {
	testLxc := lxc{
		ID: "1",
	}

	err := testLxc.getLxc(lxcScheduler.DB)
	suite.NoError(err, "They should be no error")

}

func (suite *LxcSuite) TestInsertLxcSuccessful() {
	testLxc := lxc{
		ID:          "very-unique-lxc-uuid",
		LxdID:       "very-unique-lxd-uuid",
		Name:        "test-lxc-1",
		Type:        "image",
		Alias:       "16.04",
		Address:     "",
		Description: "",
		IsDeployed:  1,
	}

	err := testLxc.insertLxc(lxcScheduler.DB)
	suite.NoError(err, "They should be no error")
}

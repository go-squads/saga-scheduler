package main

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

var operationScheduler scheduler

type OperationSuite struct {
	suite.Suite
}

func TestOperationSuite(t *testing.T) {
	suite.Run(t, new(OperationSuite))
}

func (suite *OperationSuite) SetupSuite() {
	operationScheduler = scheduler{}
	err := operationScheduler.initialize("postgres", "postgres", "saga", "localhost", "5432", "disable")
	suite.NoError(err, "They should be no error")

	clearQuery := `DELETE FROM operation;
	DELETE FROM lxc;
	DELETE FROM lxd;`

	_, err = operationScheduler.DB.Exec(clearQuery)
	suite.NoError(err, "They should be no error")

	_, err = operationScheduler.DB.Exec("INSERT INTO lxd (id, name, address) VALUES ('very-unique-lxd-uuid', 'test-lxd', 'test.gojek.com');")
	suite.NoError(err, "They should be no error")

	_, err = operationScheduler.DB.Exec("INSERT INTO lxc (id, lxd_id, name, type, alias, is_deployed) VALUES ('very-unique-lxc-uuid','very-unique-lxd-uuid', 'test-lxc-1', 'image', '16.04', 0);")
	suite.NoError(err, "They should be no error")
}

func (suite *OperationSuite) TearDownSuite() {
	clearQuery := `DELETE FROM operation;
	DELETE FROM lxc;
	DELETE FROM lxd;`

	_, err := lxcScheduler.DB.Exec(clearQuery)
	suite.NoError(err, "They should be no error")
}

func (suite *OperationSuite) TestGetOperationSuccessful() {
	testLxc := lxc{
		ID: "1",
	}

	err := testLxc.getLxc(lxcScheduler.DB)
	suite.NoError(err, "They should be no error")

}

func (suite *OperationSuite) TestInsertOperationSuccessful() {
	testOperation := operation{
		ID:          "very-unique-operation-uuid",
		LxcID:       "very-unique-lxc-uuid",
		Status:      "OK",
		StatusCode:  200,
		Description: "",
	}

	err := testOperation.insertOperation(operationScheduler.DB)
	suite.NoError(err, "They should be no error")
}

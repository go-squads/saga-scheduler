package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

var s *scheduler

type LxcSuite struct {
	suite.Suite
}

func TestLxcSuite(t *testing.T) {
	suite.Run(t, new(LxcSuite))
}

func (suite *LxcSuite) SetupSuite() {
	err := s.initialize("postgres", "", "test", "127.0.0.1", "5432", "require")
	if err != nil {
		fmt.Println(err.Error())
	}
}

func (suite *LxcSuite) TearDownSuite() {

}

func (suite *LxcSuite) TestGetLxcSuccessful() {
	var testLxc lxc
	testLxc.ID = "1"

	err := testLxc.getLxc(s.DB)
	suite.NoError(err, "They should be no error")
}

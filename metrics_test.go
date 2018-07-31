package main

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type MetricsSuite struct {
	suite.Suite
}

var valueSmall []interface{}
var valueLarge []interface{}
var testMemResponse promResponse

var textLXD lxd

var metricsScheduler scheduler

type testPrometheusMetricsDB struct{}

func (p testPrometheusMetricsDB) callMetricAPI(query string) (*promResponse, error) {
	return &testMemResponse, nil
}

func (p testPrometheusMetricsDB) getLowestLoadLxdInstance() (*lxd, error) {
	return &lxd{
		ID:          "very-unique-lxd-uuid",
		Name:        "test-lxc-1",
		Address:     "192.168.0.1",
		Description: "",
	}, nil
}

func (p testPrometheusMetricsDB) getFreeMemory() (*promResponse, error) {
	return &testMemResponse, nil
}

func TestMetricsSuite(t *testing.T) {
	suite.Run(t, new(MetricsSuite))
}

func (suite *MetricsSuite) SetupSuite() {
	metricsScheduler = scheduler{}
	metricsScheduler.initialize("postgres", "postgres", "saga", "localhost", "5432", "disable")
	cleanupDB()
	valueSmall = append(valueSmall, 123.456)
	valueSmall = append(valueSmall, "30.0")

	valueLarge = append(valueLarge, 123.456)
	valueLarge = append(valueLarge, "70.0")

	metricsScheduler.metricsDB = testPrometheusMetricsDB{}

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

func (suite *MetricsSuite) TearDownSuite() {
	cleanupDB()
}

func cleanupDB() {
	clearQuery := `DELETE FROM operation;
	DELETE FROM lxc;
	DELETE FROM lxd;`

	metricsScheduler.DB.Exec(clearQuery)
}

func (suite *MetricsSuite) TestCalculateMetricsSuccessful() {
	actual := calculateMetrics(testMemResponse, promResponse{}, promResponse{})
	suite.Equal("192.168.0.1", actual, "They should be equal")
}

func (suite *MetricsSuite) TestCalculateMetricsFailed() {
	actual := calculateMetrics(testMemResponse, promResponse{}, promResponse{})
	suite.NotEqual("Bar", actual, "They should be not equal")
}

func (suite *MetricsSuite) TestGetFreeMemorySuccessful() {
	actual, err := metricsScheduler.metricsDB.getFreeMemory()
	suite.NoError(err, "They should be no error")
	suite.Equal("success", actual.Status, "They should be equal")
}

func (suite *MetricsSuite) TestGetLowestLoadLxdInstanceSuccessful() {
	actual, err := metricsScheduler.metricsDB.getLowestLoadLxdInstance()
	suite.NoError(err, "They should be no error")
	suite.Equal("192.168.0.1", actual.Address, "They should be equal")
}

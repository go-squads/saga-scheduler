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

var metricsScheduler scheduler

type testPrometheusMetricsDB struct{}

func (p testPrometheusMetricsDB) callMetricAPI(query string) (*promResponse, error) {
	return &testMemResponse, nil
}

func TestMetricsSuite(t *testing.T) {
	suite.Run(t, new(MetricsSuite))
}

func (suite *MetricsSuite) SetupSuite() {
	metricsScheduler = scheduler{}
	metricsScheduler.metricsDB = testPrometheusMetricsDB{}
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
						Instance: "Foo",
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

}

func (suite *MetricsSuite) TestCalculateMetricsSuccessful() {
	actual := calculateMetrics(testMemResponse, promResponse{}, promResponse{})
	suite.Equal("Foo", actual, "They should be equal")
}

func (suite *MetricsSuite) TestCalculateMetricsFailed() {
	actual := calculateMetrics(testMemResponse, promResponse{}, promResponse{})
	suite.NotEqual("Bar", actual, "They should be not equal")
}

func (suite *MetricsSuite) TestGetMemoryMetricsSuccessful() {
	actual, err := metricsScheduler.getMemoryMetrics()
	suite.NoError(err, "They should be no error")
	suite.Equal("success", actual.Status, "They should be equal")
}

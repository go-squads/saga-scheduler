package main

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type MetricsSuite struct {
	suite.Suite
}

func TestMetricsSuite(t *testing.T) {
	suite.Run(t, new(MetricsSuite))
}

func (suite *MetricsSuite) SetupSuite() {

}

func (suite *MetricsSuite) TearDownSuite() {

}

func (suite *MetricsSuite) TestCalculateMetricsSuccessful() {
	var valueSmall []interface{}
	valueSmall = append(valueSmall, 123.456)
	valueSmall = append(valueSmall, "30.0")

	var valueLarge []interface{}
	valueLarge = append(valueLarge, 123.456)
	valueLarge = append(valueLarge, "70.0")

	memResponse := promResponse{
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
	actual := calculateMetrics(memResponse, promResponse{}, promResponse{})
	suite.Equal("Foo", actual, "They should be equal")

}

func (suite *MetricsSuite) TestCalculateMetricsFailed() {
	var valueSmall []interface{}
	valueSmall = append(valueSmall, 123.456)
	valueSmall = append(valueSmall, "30.0")

	var valueLarge []interface{}
	valueLarge = append(valueLarge, 123.456)
	valueLarge = append(valueLarge, "70.0")

	memResponse := promResponse{
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
	actual := calculateMetrics(memResponse, promResponse{}, promResponse{})
	suite.NotEqual("Bar", actual, "They should be not equal")

}

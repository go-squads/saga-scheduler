package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strconv"
)

type promResponse struct {
	Status string           `json:"status"`
	Data   promResponseData `json:"data"`
}

type promResponseData struct {
	ResultType string   `json:"resultType"`
	Result     []result `json:"result"`
}

type result struct {
	Metric metric        `json:"metric"`
	Value  []interface{} `json:"value"`
}

type metric struct {
	Instance string `json:"instance"`
	Job      string `json:"job"`
}

type scoring struct {
	address      string
	totalScore   float64
	memScore     float64
	cpuScore     float64
	storageScore float64
}

type metricsDB interface {
	callMetricAPI(query string) (*promResponse, error)
}

type prometheusMetricsDB struct{}

func (p prometheusMetricsDB) callMetricAPI(query string) (*promResponse, error) {
	prometheusAddress := "http://172.28.128.5:9090/api/v1/query"
	url := fmt.Sprintf("%s/api/v1/container", prometheusAddress)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("query", query)
	req.URL.RawQuery = q.Encode()
	client := &http.Client{
		Timeout: 10,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result promResponse

	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (s *scheduler) getMetrics() (*lxd, error) {
	// TODO :
	// 1. Call prometheus API
	// 2. sort by load
	// 3. return lxd
	// memResult, err := getMemoryMetrics()
	return nil, nil
}

func (s *scheduler) getMemoryMetrics() (*promResponse, error) {
	return s.metricsDB.callMetricAPI("100 * (((avg_over_time(node_memory_MemFree_bytes[24h]) + avg_over_time(node_memory_Cached_bytes[24h]) + avg_over_time(node_memory_Buffers_bytes[24h])) / avg_over_time(node_memory_MemTotal_bytes[24h])))")
}

func calculateMetrics(memResult, cpuResult, storageResult promResponse) string {
	var scores []scoring

	memoryScores := memResult.Data.Result

	for i := 0; i < len(memoryScores); i++ {
		address := memoryScores[i].Metric.Instance
		strScore := memoryScores[i].Value[1].(string)
		memScore, err := strconv.ParseFloat(strScore, 64)
		if err != nil {
			memScore = 0
		}
		scores = append(scores, scoring{
			address:  address,
			memScore: memScore,
		})
	}

	for i := 0; i < len(scores); i++ {
		scores[i].totalScore = calculateScore(scores[i])
	}

	sort.Slice(scores, func(i, j int) bool {
		return scores[i].totalScore > scores[j].totalScore
	})
	return scores[0].address
}

func calculateScore(s scoring) float64 {
	return s.cpuScore + s.memScore + s.storageScore
}

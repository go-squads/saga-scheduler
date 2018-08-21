package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
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

type scoringWeight struct {
	memScore float64
	cpuScore float64
}

type metricsDB interface {
	callMetricAPI(query string) (*promResponse, error)
	getLowestLoadLxdInstance(string, float64) (*lxd, error)
	getFreeMemory() (*promResponse, error)
}

type prometheusMetricsDB struct{}

func (p prometheusMetricsDB) callMetricAPI(query string) (*promResponse, error) {
	prometheusAddress := "172.28.128.5"
	url := fmt.Sprintf("http://%s:9090/api/v1/query", prometheusAddress)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("query", query)
	req.URL.RawQuery = q.Encode()
	client := &http.Client{
		Timeout: 10 * time.Second,
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

func (p prometheusMetricsDB) getLowestLoadLxdInstance(weight string, weightValue float64) (*lxd, error) {
	scoreWeight := setWeightValues(weight, weightValue)
	freeMemory, err := p.getFreeMemory()
	if err != nil {
		return nil, err
	}

	freeCpu, err := p.getCpuUsage()
	if err != nil {
		return nil, err
	}

	lowestLoadLxdIPAddress := calculateMetrics(*freeMemory, *freeCpu, promResponse{}, scoreWeight)
	lxdInstance := lxd{
		Address: lowestLoadLxdIPAddress,
	}

	return &lxdInstance, nil
}

func (p prometheusMetricsDB) getFreeMemory() (*promResponse, error) {
	return p.callMetricAPI("100 * (((avg_over_time(node_memory_MemFree_bytes[24h]) + avg_over_time(node_memory_Cached_bytes[24h]) + avg_over_time(node_memory_Buffers_bytes[24h])) / avg_over_time(node_memory_MemTotal_bytes[24h])))")
}

func (p prometheusMetricsDB) getCpuUsage() (*promResponse, error) {
	return p.callMetricAPI(`(avg by (instance) (irate(node_cpu_seconds_total{job="prometheus",mode="idle"}[1h])) * 100)`)
}

// Insert in the payload
func calculateMetrics(memResult, cpuResult, storageResult promResponse, scoringWeight scoringWeight) string {
	var scores []scoring

	memoryScores := memResult.Data.Result
	cpuScores := cpuResult.Data.Result

	for i := 0; i < len(memoryScores); i++ {
		address := memoryScores[i].Metric.Instance
		strMemScore := memoryScores[i].Value[1].(string)
		strCpuScore := cpuScores[i].Value[1].(string)
		memScore := convertAndCalcNewScoreWeight(strMemScore, scoringWeight.memScore)
		cpuScore := convertAndCalcNewScoreWeight(strCpuScore, scoringWeight.cpuScore)
		log.Infof("Previous Free CPU for %s : %v", address, strCpuScore)
		log.Infof("Modified Free CPU for %s : %v", address, cpuScore)
		log.Infof("Previous Free Memory for %s : %v", address, strMemScore)
		log.Infof("Modified Free Memory for %s : %v", address, memScore)

		scores = append(scores, scoring{
			address:  address,
			memScore: memScore,
			cpuScore: cpuScore,
		})
	}

	for i := 0; i < len(scores); i++ {
		scores[i].totalScore = calculateTotalScore(scores[i])
	}

	sort.Slice(scores, func(i, j int) bool {
		return scores[i].totalScore > scores[j].totalScore
	})

	ipAddress := strings.Split(scores[0].address, ":")
	return ipAddress[0]
}

func convertAndCalcNewScoreWeight(strScore string, weightValue float64) float64 {
	score, err := strconv.ParseFloat(strScore, 64)
	if err != nil {
		log.Error("Failed to convert str to float, setting it to 0")
		return 0
	}
	return score * weightValue
}

func calculateTotalScore(s scoring) float64 {
	return s.cpuScore + s.memScore + s.storageScore
}

func setWeightValues(weight string, weightValue float64) scoringWeight {
	s := scoringWeight{
		cpuScore: 1.0,
		memScore: 1.0,
	}
	switch w := weight; w {
	case "cpu-centric":
		s.cpuScore = 1.0 + (weightValue / 100)
	case "memory-centric":
		s.memScore = 1.0 + (weightValue / 100)
	}

	return s
}

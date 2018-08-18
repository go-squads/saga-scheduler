package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/jasonlvhit/gocron"
	"github.com/jmoiron/sqlx"
)

// DatabaseWorker ...
type DatabaseWorker struct {
	DB   *sqlx.DB
	Cron *gocron.Scheduler
}

func (d *DatabaseWorker) initialize(user, password, dbname, host, port, sslmode string) error {
	connectionString := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode=%s", user, password, dbname, host, port, sslmode)
	var err error
	d.DB, err = sqlx.Connect("postgres", connectionString)
	if err != nil {
		return err
	}
	d.Cron = gocron.NewScheduler()
	return nil
}

func (d *DatabaseWorker) startCronJob() {
	d.Cron.Every(1).Minute().Do(d.doCron)
	<-d.Cron.Start()
}

func (d *DatabaseWorker) doCron() {
	fmt.Println("--- running the cron job ---")
	lxds, err := getLxds(d.DB)
	if err != nil {
		panic(err)
	}
	for i := 0; i < len(lxds); i++ {
		url := fmt.Sprintf("http://%s:9200/api/v1/containers", lxds[i].Address)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			panic(err)
		}

		client := &http.Client{
			Timeout: 10 * time.Second,
		}
		response, err := client.Do(req)
		if err != nil {
			panic(err)
		}

		defer response.Body.Close()

		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			panic(err)
		}

		type lxdResponse struct {
			Name   string `json:"name"`
			Status string `json:"status"`
		}
		var resp []lxdResponse

		err = json.Unmarshal(body, &resp)
		if err != nil {
			panic(err)
		}

		// for j := 0; j < len(resp); j++ {
		// 	err = updateStatusByName(d.DB, resp[j].Name, resp[j].Status)
		// 	if err != nil {
		// 		panic(err)
		// 	}
		// }
	}
}

package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type scheduler struct {
	Router *mux.Router
	DB     *sqlx.DB
}

func (s *scheduler) initialize(user, password, dbname, host, port, sslmode string) error {
	connectionString := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode=%s", user, password, dbname, host, port, sslmode)
	var err error
	s.DB, err = sqlx.Connect("postgres", connectionString)
	if err != nil {
		return err
	}
	return nil
}

func (s *scheduler) run(addr string) {

}

func (s *scheduler) createNewLxcHandler(w http.ResponseWriter, r *http.Request) {
	// type createContainerRequestData struct {
	// 	Name     string `json:"name,omitempty"`
	// 	Type     string `json:"type,omitempty"`
	// 	Protocol string `json:"protocol,omitempty"`
	// 	Server   string `json:"server,omitempty"`
	// 	Alias    string `json:"alias,omitempty"`
	// }

	// vars := mux.Vars(r)
	// container := lxc{
	// 	LxdID:       vars["lxd_id"],
	// 	Name:        vars["name"],
	// 	Type:        vars["type"],
	// 	Alias:       vars["alias"],
	// 	Address:     vars["address"],
	// 	Description: vars["description"],
	// 	IsDeployed:  0,
	// }

	// data := createContainerRequestData{
	// 	Name:     container.Name,
	// 	Type:     container.Type,
	// 	Protocol: "simplestreams",
	// 	Server:   "https://cloud-images.ubuntu.com/daily",
	// 	Alias:    container.Alias,
	// }

	//container, err := (vars["name"])
}

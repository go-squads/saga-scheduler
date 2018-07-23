package main

import (
	"fmt"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
)

type scheduler struct {
	Router *mux.Router
	DB     *sqlx.DB
}

func (s *scheduler) initialize(user, password, dbname, host, port, sslmode string) error {
	connectionString := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode=%s", user, password, dbname, host, port, sslmode)

	fmt.Println(connectionString)
	var err error
	s.DB, err = sqlx.Connect("postgres", connectionString)
	if err != nil {
		return err
	}
	return nil
}

func (s *scheduler) run(addr string) {}

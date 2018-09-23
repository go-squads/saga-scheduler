package main

import (
	"os"
)

func main() {
	saga := scheduler{}
	saga.initialize(os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"), os.Getenv("DB_URL"), os.Getenv("DB_PORT"))
	saga.run(":9300")
}

package main

func main() {
	worker := DatabaseWorker{}
	worker.initialize("postgres", "postgres", "saga", "localhost", "5432", "disable")
	go worker.startCronJob()

	saga := scheduler{}
	saga.initialize("postgres", "postgres", "saga", "localhost", "5432", "disable")
	saga.run(":9300")
}

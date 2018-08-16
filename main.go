package main

func main() {
	saga := scheduler{}
	saga.initialize("postgres", "postgres", "saga", "localhost", "5432", "disable")
	saga.run(":9300")
}

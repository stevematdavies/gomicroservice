package main

import (
	"fmt"
	"log"
	"net/http"
)

const webPort = "8081"

type Config struct{}

func main() {
	a := Config{}
	log.Printf("Starting Broker Service on port: %s\n", webPort)

	s := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: a.routes(),
	}

	err := s.ListenAndServe()
	if err != nil {
		log.Panic(err)
	}

}

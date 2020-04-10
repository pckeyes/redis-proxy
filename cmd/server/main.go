package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"redis-proxy/pkg/config"
	"redis-proxy/pkg/server"

	"github.com/gorilla/mux"
)

const (
	defaultFilePath = "src/redis-proxy/config/config.json"
)

func main() {

	config, err := config.ReadConfig(defaultFilePath)
	if err != nil {
		log.Panic("Unable to read config file with error: ", err)
	}

	app, err := server.NewApp(config)
	if err != nil {
		log.Panic("Unable to instantiate new App with error: ", err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/{key}", app.GetValue).Methods("GET")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go app.Cache.ExpiryChecker(ctx)

	port := fmt.Sprintf(":%d", config.Port)
	if err := http.ListenAndServe(port, r); err != nil {
		log.Fatal(err)
	}
}

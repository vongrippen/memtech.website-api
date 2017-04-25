package main

import (
	"fmt"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/kelseyhightower/envconfig"
	"github.com/vongrippen/memtech.website-api/users"
	"log"
	"net/http"
	"os"
)

type Config struct {
	Host string `default:"0.0.0.0" envconfig:"HOST"`
	Port string `default:"8080" envconfig:"PORT"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

var (
	config Config
)

func main() {
	err := envconfig.Process("API", &config)
	if err != nil {
		log.Printf("Error processing config: %v\n", err.Error())
	}
	log.Println("Listening on port ", config.Port)

	r := mux.NewRouter()
	r.HandleFunc("/api/users.json", users.UserList)
	loggedRouter := handlers.LoggingHandler(os.Stdout, r)
	http.Handle("/", loggedRouter)
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%s", config.Host, config.Port), nil))
}

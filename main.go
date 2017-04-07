package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/kelseyhightower/envconfig"
	"io/ioutil"
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
	r.HandleFunc("/api/users.json", userList)
	http.Handle("/", r)

	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%s", config.Host, config.Port), nil))
}

type User struct {
	Username   string `json:"username"`
	Webring    bool   `json:"webring"`
	PublicHtml bool   `json:"-"`
}

func userList(w http.ResponseWriter, r *http.Request) {
	files, err := ioutil.ReadDir("/home")
	if err != nil {
		marsh, _ := json.Marshal(ErrorResponse{err.Error()})
		w.Write(marsh)
		return
	}

	var users []User
	msgchan := make(chan User)
	num_responded := 0

	for _, file := range files {
		go checkUser(file, msgchan)
	}

	for num_responded < len(files) {
		user := <-msgchan

		users = append(users, user)
	}

	marsh, _ := json.Marshal(users)
	w.Write(marsh)
}

func checkUser(file os.FileInfo, ret chan User) {
	user := User{}
	if file.IsDir() {
		user.Username = file.Name()
		sub, _ := ioutil.ReadDir(fmt.Sprintf("/home/%s", user.Username))
		for _, subfile := range sub {
			if subfile.Name() == "public_html" {
				user.PublicHtml = true
			} else if subfile.Name() == "webring" {
				user.Webring = true
			}
		}
	}
	ret <- user
}

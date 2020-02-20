package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

var port string

func init() {
	port = os.Getenv("APP_PORT")
	if port == "" {
		log.Fatalf("missing environment variable %s", "APP_PORT")
	}
}
func main() {
	http.HandleFunc("/consumer", func(rw http.ResponseWriter, req *http.Request) {
		var msg Message

		err := json.NewDecoder(req.Body).Decode(&msg)
		if err != nil {
			fmt.Println("error reading message from event hub binding", err)
			rw.WriteHeader(500)
			return
		}
		fmt.Printf("data from Event Hubs '%s'\n", msg)
		rw.WriteHeader(200)
	})
	http.ListenAndServe(":"+port, nil)
}

type Message struct {
	Msg string `json:"message"`
}

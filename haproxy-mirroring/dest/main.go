package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/sirupsen/logrus"
)

func main() {
	address := os.Getenv("ADDRESS")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		logrus.Infof("request received: method => %s; headers => %+v; host => %s", r.Method, r.Header, r.Host)
		fmt.Fprint(w, "")
	})

	logrus.Infof("listening on: %s", address)

	if err := http.ListenAndServe(address, nil); err != nil {
		log.Fatal(err)
	}
}

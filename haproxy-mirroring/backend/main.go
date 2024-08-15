package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/sirupsen/logrus"
)

func main() {
	socketAddress := os.Getenv("SOCKET_ADDRESS")
	os.Remove(socketAddress)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "hello world")
	})

	listener, err := net.Listen("unix", socketAddress)
	if err != nil {
		log.Fatal(err)
	}

	logrus.Infof("listening on: %s", socketAddress)

	if err := http.Serve(listener, nil); err != nil {
		log.Fatal(err)
	}
}

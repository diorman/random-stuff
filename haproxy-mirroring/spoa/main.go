package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"time"

	spoe "github.com/criteo/haproxy-spoe-go"
	"github.com/sirupsen/logrus"
)

const (
	queueSize = 10
	nWorkers  = 5
)

type message map[string]interface{}

func main() {
	destAddress := os.Getenv("DEST_ADDRESS")
	socketAddress := os.Getenv("SOCKET_ADDRESS")
	os.Remove(socketAddress)

	queue := make(chan message, queueSize)
	agent := newAgent(queue)
	runWorkers(context.TODO(), queue, nWorkers, destAddress)

	listener, err := net.Listen("unix", socketAddress)
	if err != nil {
		logrus.Fatal(err)
	}

	if err := agent.Serve(listener); err != nil {
		logrus.Fatal(err)
	}
}

func newAgent(queue chan<- message) *spoe.Agent {
	return spoe.New(func(messageIterator *spoe.MessageIterator) ([]spoe.Action, error) {
		for messageIterator.Next() {
			if !enqueMessage(queue, messageIterator.Message.Args.Map()) {
				logrus.Warn("event dropped!")
			}
		}

		return nil, nil
	})
}

func enqueMessage(queue chan<- message, message message) bool {
	select {
	case queue <- message:
		return true
	default:
		return false
	}
}

func runWorkers(ctx context.Context, queue <-chan message, nWorkers int, destAddress string) {
	for i := 0; i < nWorkers; i++ {
		go func() {
			client := &http.Client{
				Timeout: 1 * time.Second,
				// disable following redirects
				CheckRedirect: func(req *http.Request, via []*http.Request) error {
					return http.ErrUseLastResponse
				},
			}

			for message := range queue {
				handleMessage(ctx, client, destAddress, message)
			}
		}()
	}
}

func handleMessage(ctx context.Context, client *http.Client, destAddress string, message message) {
	request, err := newRequest(ctx, message, destAddress)
	if err != nil {
		logrus.Errorf("failed to create HTTP request: %v", err)
		return
	}

	response, err := client.Do(request)
	if err != nil {
		logrus.Errorf("failed to make HTTP request: %v", err)
		return
	}

	io.Copy(io.Discard, response.Body)
	response.Body.Close()
}

func newRequest(ctx context.Context, message message, destAddress string) (*http.Request, error) {
	method, err := validateArgString(message, "method")
	if err != nil {
		return nil, err
	}

	path, err := validateArgString(message, "path")
	if err != nil {
		return nil, err
	}

	// there's no way to set this on the http client so we should get rid of it
	if _, err := validateArgString(message, "version"); err != nil {
		return nil, err
	}

	scheme, err := validateArgString(message, "scheme")
	if err != nil {
		return nil, err
	}

	rawHeaders, err := validateArgBytes(message, "headers")
	if err != nil {
		return nil, err
	}

	body, err := validateArgBytes(message, "body")
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s://%s%s", scheme, destAddress, path)

	decodedHeaders, err := spoe.DecodeHeaders(rawHeaders)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequestWithContext(context.TODO(), method, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	request.Header = decodedHeaders
	request.Host = decodedHeaders.Get("Host")

	return request, nil
}

func validateArgString(args map[string]interface{}, name string) (string, error) {
	if err := validateArgPresent(args, name); err != nil {
		return "", err
	}

	value, ok := args[name].(string)
	if !ok {
		return "", fmt.Errorf("invalid arg: %s = %v (%T)", name, args[name], args[name])
	}

	return value, nil
}

func validateArgBytes(args map[string]interface{}, name string) ([]byte, error) {
	if err := validateArgPresent(args, name); err != nil {
		return nil, err
	}

	value, ok := args[name].([]byte)
	if !ok {
		return nil, fmt.Errorf("invalid arg: %s = %v (%T)", name, args[name], args[name])
	}

	return value, nil
}

func validateArgPresent(args map[string]interface{}, name string) error {
	if _, ok := args[name]; !ok {
		return fmt.Errorf("missing arg: %s", name)
	}
	return nil
}

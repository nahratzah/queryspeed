package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
	"time"
)

func measureGet(client *http.Client, url string) (statusCode int, deltaTime, readDeltaTime time.Duration, err error) {
	if client == nil {
		client = http.DefaultClient
	}

	t0 := time.Now()
	resp, err := http.Get(url)
	deltaTime = time.Since(t0)
	if err != nil {
		return 0, 0, 0, err
	}
	defer resp.Body.Close()

	if _, err := io.Copy(ioutil.Discard, resp.Body); err != nil {
		return 0, 0, 0, err
	}
	readDeltaTime = time.Since(t0)

	return resp.StatusCode, deltaTime, readDeltaTime, nil
}

func main() {
	type measurement struct {
		deltaTime, readDeltaTime time.Duration
		url                      string
		err                      error
		statusCode               int
	}
	var client http.Client

	var waiter sync.WaitGroup
	results := make([]measurement, len(os.Args)-1)
	for i := range results {
		results[i].url = os.Args[1+i]
		go func(i int, url string) {
			defer waiter.Done()
			results[i].statusCode, results[i].deltaTime, results[i].readDeltaTime, results[i].err = measureGet(&client, url)
		}(i, os.Args[1+i])
		waiter.Add(1)
	}
	waiter.Wait()

	for _, r := range results {
		if r.err != nil {
			fmt.Fprintf(os.Stderr, "Error for %s: %v\n", r.url, r.err)
		} else {
			fmt.Printf("%s: \u0394T(request) = %v; \u0394T(request+read body) = %v\n", r.url, r.deltaTime, r.readDeltaTime)
		}
	}
}

package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	n := 100
	for i := 0; i < n; i++ {
		go func() {
			for {
				makeRequest()
				time.Sleep(time.Millisecond * 100)
			}
		}()
		time.Sleep(time.Millisecond * 100)
	}
	time.Sleep(time.Second * 10)
}

func makeRequest() {
	start := time.Now()
	req, err := http.NewRequest("get", "http://localhost:5000/live/09248ef6-c401-4601-8928-5964d61f2c61", nil)
	if err != nil {
		log.Fatal(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		log.Fatal("status", resp.StatusCode)
	}
	fmt.Println(time.Since(start))
}

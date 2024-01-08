package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

func main() {
	for {
		makeRequest()
		time.Sleep(time.Millisecond * 50)
	}
}

func makeRequest() {
	req, err := http.NewRequest("get", "http://localhost:5000/live/09248ef6-c401-4601-8928-5964d61f2c61", nil)
	if err != nil {
		log.Fatal(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	fmt.Println(string(b))
}

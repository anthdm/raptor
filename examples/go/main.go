package main

import (
	"fmt"
	"io"
	"log"
	"net/http"

	ffaas "github.com/anthdm/ffaas/sdk"
)

func handleChatGPTWrapper(w http.ResponseWriter, r *http.Request) {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(b))
	w.Write([]byte("from my ffaas application"))
}

func main() {
	ffaas.HandleFunc(handleChatGPTWrapper)
}

package main

import (
	"net/http"

	ffaas "github.com/anthdm/ffaas/sdk"
)

func handleChatGPTWrapper(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("from my ffaas application"))
}

func main() {
	ffaas.HandleFunc(handleChatGPTWrapper)
}

package main

import (
	"fmt"
	"net/http"
	"os"

	ffaas "github.com/anthdm/ffaas/sdk"
)

func myHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("the env:", os.Getenv("FOO"))
	w.Write([]byte("from tinder swiper"))
}

func main() {
	ffaas.HandleFunc(myHandler)
}

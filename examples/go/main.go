package main

import (
	"fmt"
	"math/rand"
	"net/http"
)

func myHandler(w http.ResponseWriter, r *http.Request) {
	num := rand.Intn(100)
	w.Write([]byte(fmt.Sprintf("from my application: %d", num)))
}

func main() {
	fmt.Println("from the wasm guest")
	//ffaas.HandleFunc(myHandler)
}

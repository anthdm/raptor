package main

import (
	"fmt"
	"math/rand"
	"net/http"

	_ "github.com/stealthrocket/net/http"
	//_ "github.com/stealthrocket/net/wasip1"
)

func handleChatGPTWrapper(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("from my ffaas application"))
}

func main() {
	fmt.Println(rand.Intn(10))
	// ip, err := net.LookupIP("google.com")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// req, err := http.Get("http://google.com")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// b, err := io.ReadAll(req.Body)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Println(string(b))

	//
	// conn, err := net.Dial("tcp", "google.com:80")
	// fmt.Println(err)
	// fmt.Println(conn)
	// ffaas.HandleFunc(handleChatGPTWrapper)
}

package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/anthdm/ffaas/pkg/runtime"
	"github.com/tetratelabs/wazero"
)

func main() {
	b, err := os.ReadFile("examples/go/app.wasm")
	if err != nil {
		log.Fatal(err)
	}
	cache := wazero.NewCompilationCache()

	for {
		start := time.Now()
		wruntime, compiledMod, err := runtime.NewWASIRuntime(cache, b)
		if err != nil {
			log.Fatal(err)
		}

		r, err := runtime.New(wruntime, compiledMod)
		if err != nil {
			log.Fatal(err)
		}
		req, err := http.NewRequest("GET", "/", bytes.NewReader([]byte("foo")))
		if err != nil {
			log.Fatal(err)
		}
		if err := r.Exec(context.Background(), req); err != nil {
			log.Fatal(err)
		}
		fmt.Println(time.Since(start))
		// fmt.Println(string(r.Response()))
		time.Sleep(time.Second * 2)
	}
}

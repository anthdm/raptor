package run

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/anthdm/raptor/proto"
	// _ "github.com/stealthrocket/net/http"
	prot "google.golang.org/protobuf/proto"
)

var (
	requestBuffer  []byte
	responseBuffer []byte
)

type request struct {
	Body   []byte
	Method string
	Header http.Header
	URL    string
}

func Handle(h http.Handler) {
	b, err := io.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}
	var req proto.HTTPRequest
	if err := prot.Unmarshal(b, &req); err != nil {
		log.Fatal(err)
	}

	w := &ResponseWriter{}
	r, err := http.NewRequest(req.Method, req.URL, bytes.NewReader(req.Body))
	if err != nil {
		log.Fatal(err)
	}
	for k, v := range req.Header {
		r.Header[k] = v.Fields
	}
	h.ServeHTTP(w, r) // execute the user's handler
	out := strings.TrimRight(w.buffer.String(), "\n")
	fmt.Println(out)
	fmt.Println(w.statusCode)
}

type ResponseWriter struct {
	buffer     bytes.Buffer
	statusCode int
}

func (w *ResponseWriter) Header() http.Header {
	return http.Header{}
}

func (w *ResponseWriter) Write(b []byte) (n int, err error) {
	return w.buffer.Write(b)
}

func (w *ResponseWriter) WriteHeader(status int) {
	w.statusCode = status
}

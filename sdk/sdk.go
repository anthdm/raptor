package ffaas

import (
	"bytes"
	"log"
	"net/http"
	"unsafe"

	"github.com/anthdm/ffaas/proto"
	_ "github.com/stealthrocket/net/http"
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

//go:wasmimport env malloc
//go:noescape
func malloc() uint32

//go:wasmimport env write_request
//go:noescape
func writeRequest(ptr uint32)

//go:wasmimport env write_response
//go:noescape
func writeResponse(ptr uint32, size uint32)

func Handle(h http.Handler) {
	requestSize := malloc()
	requestBuffer = make([]byte, requestSize)

	ptr := &requestBuffer[0]
	unsafePtr := uint32(uintptr(unsafe.Pointer(ptr)))

	writeRequest(unsafePtr)

	var req proto.HTTPRequest
	if err := prot.Unmarshal(requestBuffer, &req); err != nil {
		log.Fatal(err)
	}

	w := &ResponseWriter{}
	r, _ := http.NewRequest(req.Method, req.URL, bytes.NewReader(req.Body))
	for k, v := range req.Header {
		r.Header[k] = v.Fields
	}
	h.ServeHTTP(w, r) // execute the user's handler

	if w.buffer.Len() > 0 {
		responseBuffer = w.buffer.Bytes()
	} else {
		responseBuffer = []byte("Hailstorm application. Coming soon...")
	}

	ptr = &responseBuffer[0]
	unsafePtr = uint32(uintptr(unsafe.Pointer(ptr)))
	writeResponse(unsafePtr, uint32(len(responseBuffer)))
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

package ffaas

import (
	"bytes"
	"net/http"
	"os"
	"unsafe"

	_ "github.com/stealthrocket/net/http"
	"github.com/vmihailenco/msgpack/v5"
)

var (
	requestBuffer  []byte
	responseBuffer []byte
)

type request struct {
	Body   []byte
	Method string
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

func HandleFunc(h http.HandlerFunc) {
	requestSize := malloc()
	requestBuffer = make([]byte, requestSize)

	ptr := &requestBuffer[0]
	unsafePtr := uint32(uintptr(unsafe.Pointer(ptr)))

	writeRequest(unsafePtr)

	var req request
	if err := msgpack.Unmarshal(requestBuffer, &req); err != nil {
		// todo
		os.Exit(1)
	}

	// execute the handler of the caller
	w := &ResponseWriter{}
	r, _ := http.NewRequest(req.Method, req.URL, bytes.NewReader(req.Body))
	h(w, r)

	responseBuffer = w.buffer.Bytes()
	ptr = &responseBuffer[0]
	unsafePtr = uint32(uintptr(unsafe.Pointer(ptr)))

	writeResponse(unsafePtr, uint32(w.buffer.Len()))
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

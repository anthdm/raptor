package shared

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/anthdm/raptor/proto"
	"github.com/google/uuid"
)

const (
	magicLen = 8
	UUIDZERO = "00000000-0000-0000-0000-000000000000"
)

var errInvalidHTTPResponse = errors.New("invalid HTTP response")

func ParseStdout(runtime string, stdout io.Reader) (logs []byte, resp []byte, status int, err error) {
	stdoutb, err := io.ReadAll(GetDecodedStdout(runtime, stdout))
	if err != nil {
		return
	}
	outLen := len(stdoutb)
	if outLen < magicLen {
		err = fmt.Errorf("mallformed HTTP response missing last %d bytes", magicLen)
		return
	}
	magicStart := outLen - magicLen
	status = int(binary.LittleEndian.Uint32(stdoutb[magicStart : magicStart+4]))
	respLen := binary.LittleEndian.Uint32(stdoutb[magicStart+4:])
	if int(respLen) > outLen-magicLen {
		err = fmt.Errorf("response length exceeds available data")
		return
	}
	respStart := outLen - magicLen - int(respLen)
	resp = stdoutb[respStart : respStart+int(respLen)]
	logs = stdoutb[:respStart]
	return
}

func ParseRuntimeHTTPResponse(in string) (resp string, status int, err error) {
	if len(in) < 16 {
		err = fmt.Errorf("misformed HTTP response missing last 16 bytes")
		return
	}
	var b []byte
	b, err = hex.DecodeString(in[len(in)-16:])
	if err != nil {
		err = errInvalidHTTPResponse
	}
	status = int(binary.LittleEndian.Uint32(b[0:4]))
	respLen := binary.LittleEndian.Uint32(b[4:8])
	resp = in[len(in)-16-int(respLen) : len(in)-16]
	return
}

func MakeProtoRequest(id string, r *http.Request) (*proto.HTTPRequest, error) {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	return &proto.HTTPRequest{
		Header: makeProtoHeader(r.Header),
		ID:     id,
		Body:   b,
		Method: r.Method,
		URL:    trimmedEndpointFromURL(r.URL),
	}, nil
}

func trimmedEndpointFromURL(url *url.URL) string {
	path := strings.TrimPrefix(url.Path, "/")
	pathParts := strings.Split(path, "/")
	if len(pathParts) == 0 {
		return "/"
	}
	return "/" + strings.Join(pathParts[2:], "/")
}

func makeProtoHeader(header http.Header) map[string]*proto.HeaderFields {
	m := make(map[string]*proto.HeaderFields, len(header))
	for k, v := range header {
		m[k] = &proto.HeaderFields{
			Fields: v,
		}
	}
	return m
}

func GetDecodedStdout(runtime string, stdout io.Reader) io.Reader {
	switch runtime {
	case "go":
		return stdout
	case "js":
		return hex.NewDecoder(stdout)
	default:
		return stdout
	}
}

func IsZeroUUID(id uuid.UUID) bool {
	return id.String() == UUIDZERO
}

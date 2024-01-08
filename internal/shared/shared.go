package shared

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/anthdm/raptor/proto"
)

func ParseRuntimeHTTPResponse(in string) (resp string, status int, err error) {
	lines := strings.Split(in, "\n")
	if len(lines) < 3 {
		err = fmt.Errorf("invalid response")
		return
	}
	resp = lines[len(lines)-3]
	status, err = strconv.Atoi(lines[len(lines)-2])
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

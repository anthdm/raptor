package runtime

import (
	"bytes"
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/anthdm/raptor/internal/shared"
	"github.com/anthdm/raptor/internal/spidermonkey"
	"github.com/anthdm/raptor/proto"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/tetratelabs/wazero"
	pb "google.golang.org/protobuf/proto"
)

func TestRuntimeInvokeJSCode(t *testing.T) {
	b, err := os.ReadFile("../_testdata/helloworld.js")
	require.Nil(t, err)

	req := &proto.HTTPRequest{
		Method: "get",
		URL:    "/",
		Body:   nil,
	}
	breq, err := pb.Marshal(req)
	require.Nil(t, err)

	out := &bytes.Buffer{}
	args := Args{
		Stdout:       out,
		DeploymentID: uuid.New(),
		Blob:         spidermonkey.WasmBlob,
		Engine:       "js",
		Cache:        wazero.NewCompilationCache(),
	}
	r, err := New(context.Background(), args)
	require.Nil(t, err)

	scriptArgs := []string{"", "-e", string(b)}
	require.Nil(t, r.Invoke(bytes.NewReader(breq), nil, scriptArgs...))

	_, res, status, err := shared.ParseStdout("js", out)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, "Hello world!", string(res))
	require.Nil(t, r.Close())
}

func TestRuntimeInvokeGoCode(t *testing.T) {
	b, err := os.ReadFile("../_testdata/helloworld.wasm")
	require.Nil(t, err)

	req := &proto.HTTPRequest{
		Method: "get",
		URL:    "/",
		Body:   nil,
	}
	breq, err := pb.Marshal(req)
	require.Nil(t, err)

	out := &bytes.Buffer{}
	args := Args{
		Stdout:       out,
		DeploymentID: uuid.New(),
		Blob:         b,
		Engine:       "go",
		Cache:        wazero.NewCompilationCache(),
	}
	r, err := New(context.Background(), args)
	require.Nil(t, err)
	require.Nil(t, r.Invoke(bytes.NewReader(breq), nil))
	_, res, status, err := shared.ParseStdout("go", out)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, "Hello world!", string(res))
	require.Nil(t, r.Close())
}

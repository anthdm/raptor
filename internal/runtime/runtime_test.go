package runtime

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/anthdm/raptor/internal/shared"
	"github.com/anthdm/raptor/proto"
	"github.com/stretchr/testify/require"
	"github.com/tetratelabs/wazero"
	pb "google.golang.org/protobuf/proto"
)

func TestRuntimeInvoke(t *testing.T) {
	b, err := os.ReadFile("../../examples/go/app.wasm")
	require.Nil(t, err)
	ctx := context.Background()

	out := &bytes.Buffer{}
	req := &proto.HTTPRequest{
		Method: "get",
		URL:    "/",
	}
	reqb, err := pb.Marshal(req)
	require.Nil(t, err)

	args := InvokeArgs{
		Cache: wazero.NewCompilationCache(),
		Env:   map[string]string{},
		Blob:  b,
		Out:   out,
		In:    bytes.NewReader(reqb),
	}
	err = Invoke(ctx, args)
	require.Nil(t, err)
	resp, status, err := shared.ParseRuntimeHTTPResponse(out.String())
	fmt.Println(resp)
	fmt.Println(status)
}

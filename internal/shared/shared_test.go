package shared

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log"
	"testing"

	"github.com/stretchr/testify/require"
)

func BenchmarkParseStdout(b *testing.B) {
	userResp := "<h1>This is the actual response</h1>"
	statusCode := uint32(200)
	userLogs := `
the big brown fox
the big brown fox
the big brown fox
the big brown fox
the big brown fox
`
	builder := &bytes.Buffer{}

	buf := make([]byte, 8)
	binary.LittleEndian.PutUint32(buf[0:4], statusCode)
	binary.LittleEndian.PutUint32(buf[4:8], uint32(len(userResp)))

	b.ResetTimer()
	b.StartTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		builder.WriteString(userLogs)
		builder.WriteString(userResp)
		builder.Write(buf)
		if _, _, _, err := ParseStdout(builder); err != nil {
			log.Fatal(err)
		}
		builder.Reset()
	}
	b.StopTimer()
}

func TestParseWithoutUserLogs(t *testing.T) {
	userResp := "<h1>This is the actual response</h1>"
	statusCode := uint32(200)
	builder := &bytes.Buffer{}
	builder.WriteString(userResp)

	// += userResp
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint32(buf[0:4], statusCode)
	binary.LittleEndian.PutUint32(buf[4:8], uint32(len(userResp)))
	builder.Write(buf)

	logs, resp, status, err := ParseStdout(builder)
	require.Nil(t, err)
	require.Equal(t, int(statusCode), status)
	require.Equal(t, userResp, string(resp))
	require.Equal(t, []byte{}, logs)
}

func TestParseWithUserLogs(t *testing.T) {
	userResp := "<h1>This is the actual response</h1>"
	statusCode := uint32(200)
	userLogs := `
the big brown fox
the big brown fox
the big brown fox
the big brown fox
the big brown fox
`
	builder := &bytes.Buffer{}
	builder.WriteString(userLogs)
	builder.WriteString(userResp)

	// += userResp
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint32(buf[0:4], statusCode)
	binary.LittleEndian.PutUint32(buf[4:8], uint32(len(userResp)))
	builder.Write(buf)

	logs, resp, status, err := ParseStdout(builder)
	require.Nil(t, err)
	require.Equal(t, int(statusCode), status)
	require.Equal(t, userResp, string(resp))
	require.Equal(t, userLogs, string(logs))
}

func TestParseRuntimeHTTPResponse(t *testing.T) {
	text := "This is the best.\nBut not always correct.\nThe big brown fox."
	statusCode := uint32(500)

	buf := make([]byte, 8)
	binary.LittleEndian.PutUint32(buf[0:4], statusCode)
	binary.LittleEndian.PutUint32(buf[4:8], uint32(len(text)))

	in := fmt.Sprintf("%s%s", text, hex.EncodeToString(buf))
	resp, status, err := ParseRuntimeHTTPResponse(in)
	require.Nil(t, err)
	require.Equal(t, int(statusCode), status)
	require.Equal(t, text, resp)
}

package shared

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseRuntimeHTTPResponse(t *testing.T) {
	t.Run("multiline response", func(t *testing.T) {
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
	})
}

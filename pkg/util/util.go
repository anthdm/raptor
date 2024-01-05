package util

import (
	"fmt"
	"strconv"
	"strings"
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

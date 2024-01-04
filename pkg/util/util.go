package util

import (
	"fmt"
	"strconv"
	"strings"
)

func ParseRuntimeHTTPResponse(in string) (resp string, status int, err error) {
	lines := strings.Split(in, "\n")
	if len(lines) < 2 {
		err = fmt.Errorf("invalid response")
		return
	}
	last := lines[len(lines)-2]
	parts := strings.Split(last, "|")
	if len(parts) != 2 {
		err = fmt.Errorf("invalid response")
		return
	}
	resp = parts[0]
	status, err = strconv.Atoi(parts[1])
	return
}

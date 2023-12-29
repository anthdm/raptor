package config

import "testing"

func TestMakeURL(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{":5000", "http://0.0.0.0:5000"},
		{"127.0.0.1:5000", "http://127.0.0.1:5000"},
		{":http", "http://0.0.0.0:80"},
		{"123.456.789.123:http", "http://123.456.789.123:80"},
	}

	for _, tc := range testCases {
		actual := makeURL(tc.input)
		if actual != tc.expected {
			t.Errorf("Expected %s, got %s", tc.expected, actual)
		}
	}

}

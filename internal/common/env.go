package common

import (
	"io"
	"net/http"
	"os"
	"regexp"
)

func Env(name string) string {
	value := os.Getenv(name)

	if value == "" {
		return ""
	}

	ok, err := regexp.MatchString("https?://", value)

	if err != nil {
		return value
	}

	if !ok {
		return value
	}

	resp, err := http.Get(value)

	if err != nil {
		return value
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return value
	}

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return value
	}

	return string(body)
}

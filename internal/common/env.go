package common

import (
	"github.com/patrickmn/go-cache"
	"io"
	"net/http"
	"os"
	"regexp"
	"time"
)

var configCache = cache.New(time.Minute*10, time.Second*1)

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

	if v, ok := configCache.Get(value); ok {
		return v.(string)
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

	configCache.Set(value, string(body), cache.DefaultExpiration)

	return string(body)
}

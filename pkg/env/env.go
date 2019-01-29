package env

import (
	"encoding/json"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/luxeria/doorbell/pkg/recaptcha"
	"github.com/luxeria/doorbell/pkg/openinghours"
	"github.com/luxeria/doorbell/pkg/ratelimit"
)

func String(key string, fallback ...string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		if len(fallback) == 0 {
			log.Fatalf("missing required environment variable: %s", key)
		}
		return fallback[0]
	}
	return value
}

func Bytes(key string, fallback ...string) []byte {
	return []byte(String(key, fallback...))
}

func Float(key string, fallback ...string) float64 {
	value, err := strconv.ParseFloat(String(key, fallback...), 64)
	if err != nil {
		log.Fatalf("failed to parse environment variable %s as float: %s", key, err)
	}
	return value
}

func Duration(key string, fallback ...string) time.Duration {
	value, err := time.ParseDuration(String(key, fallback...))
	if err != nil {
		log.Fatalf("failed to parse environment variable %s as duration: %s", key, err)
	}
	return value
}

func OpeningHours(key string, fallback ...string) openinghours.OpeningHours {
	value, err := openinghours.Parse(String(key, fallback...))
	if err != nil {
		log.Fatalf("failed to parse environment variable %s as opening hours: %s", key, err)
	}
	return value
}

func RateLimit(key string, fallback ...string) *ratelimit.Bucket {
	value, err := ratelimit.Parse(String(key, fallback...))
	if err != nil {
		log.Fatalf("failed to parse environment variable %s as rate limit: %s", key, err)
	}
	return value
}

func Recaptcha(key string, fallback ...string) *recaptcha.Recaptcha {
	return recaptcha.New(String(key, fallback...))
}

func StringSlice(key string, fallback ...string) []string {
	var value []string
	err := json.Unmarshal(Bytes(key, fallback...), &value)
	if err != nil {
		log.Fatalf("failed to parse environment variable %s as json string list: %s", key, err)
	}
	return value
}

func Addr(key string, fallback ...string) string {
	addr := String(key, fallback...)
	if len(addr) > 0 && !strings.Contains(addr, ":") {
		// treat it as port
		addr = ":" + addr
	}

	return addr
}
package main

import (
	"log"
	"os"
	"strconv"

	"github.com/eriktate/skribe/http"
)

func main() {
	server := http.NewServer(getEnvString("SKRIBE_HOST", "localhost"), getEnvUint("SKRIBE_PORT", 1337))

	if err := server.Start(); err != nil {
		log.Fatal(err)
	}
}

func getEnvString(key, def string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}

	return def
}

func getEnvUint(key string, def uint) uint {
	val, err := strconv.ParseUint(os.Getenv(key), 10, 32)
	if err != nil {
		return def
	}

	return uint(val)
}

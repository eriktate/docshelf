package main

import (
	"log"
	"os"
	"strconv"

	"github.com/docshelf/docshelf/bolt"
	"github.com/docshelf/docshelf/disk"
	"github.com/docshelf/docshelf/http"
	"github.com/sirupsen/logrus"
)

func main() {
	logger := logrus.New()
	server := http.NewServer(getEnvString("DS_HOST", "localhost"), getEnvUint("DS_PORT", 1337))

	fs, err := disk.New("documents")
	if err != nil {
		log.Fatal(err)
	}

	db, err := bolt.New("docshelf.db", fs)
	if err != nil {
		log.Fatal(err)
	}

	server.UserHandler = http.NewUserHandler(db, logger)
	server.DocHandler = http.NewDocHandler(db, logger)

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

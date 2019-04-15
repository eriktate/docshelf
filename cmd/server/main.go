package main

import (
	"log"
	"os"
	"strconv"

	"github.com/docshelf/docshelf"
	"github.com/docshelf/docshelf/bolt"
	"github.com/docshelf/docshelf/disk"
	"github.com/docshelf/docshelf/dynamo"
	"github.com/docshelf/docshelf/http"
	"github.com/docshelf/docshelf/s3"
	"github.com/sirupsen/logrus"
)

// A Config contains all of the values docshelf might need to start up.
type Config struct {
	Backend     string
	FileBackend string
	S3Bucket    string
	FilePrefix  string
	BoltPath    string
	Host        string
	Port        uint
}

func fromEnv() Config {
	return Config{
		Backend:     getEnvString("DS_BACKEND", "bolt"),
		FileBackend: getEnvString("DS_FILE_BACKEND", "disk"),
		S3Bucket:    getEnvString("DS_S3_BUCKET", ""),
		FilePrefix:  getEnvString("DS_FILE_PREFIX", "documents"),
		BoltPath:    getEnvString("DS_BOLTDB_PATH", "docshelf.db"),
		Host:        getEnvString("DS_HOST", "localhost"),
		Port:        getEnvUint("DS_PORT", 1337),
	}
}

func main() {
	logger := logrus.New()
	cfg := fromEnv()
	server := http.NewServer(cfg.Host, cfg.Port)

	fs, err := getFileStore(cfg)
	if err != nil {
		log.Fatal(err)
	}

	backend, err := getBackend(cfg, fs)
	if err != nil {
		log.Fatal(err)
	}

	server.UserHandler = http.NewUserHandler(backend, logger)
	server.DocHandler = http.NewDocHandler(backend, logger)

	if err := server.Start(); err != nil {
		log.Fatal(err)
	}
}

func getFileStore(cfg Config) (docshelf.FileStore, error) {
	switch cfg.FileBackend {
	case "s3":
		fs, err := s3.New(cfg.S3Bucket, cfg.FilePrefix)
		if err != nil {
			return nil, err
		}

		return fs, nil
	default:
		fs, err := disk.New(cfg.FilePrefix)
		if err != nil {
			return nil, err
		}

		return fs, nil
	}
}

func getBackend(cfg Config, fs docshelf.FileStore) (docshelf.Backend, error) {
	switch cfg.Backend {
	case "dynamo":
		backend, err := dynamo.New(fs)
		if err != nil {
			return nil, err
		}

		return backend, nil
	default:
		backend, err := bolt.New(cfg.BoltPath, fs)
		if err != nil {
			return nil, err
		}

		return backend, nil
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

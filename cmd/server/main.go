package main

import (
	"os"
	"strconv"

	"github.com/docshelf/docshelf"
	"github.com/docshelf/docshelf/bleve"
	"github.com/docshelf/docshelf/bolt"
	"github.com/docshelf/docshelf/disk"
	"github.com/docshelf/docshelf/dynamo"
	"github.com/docshelf/docshelf/http"
	"github.com/docshelf/docshelf/s3"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

// A Config contains all of the values docshelf might need to start up.
type Config struct {
	Backend     string
	FileBackend string
	TextIndex   string
	S3Bucket    string
	FilePrefix  string
	BoltPath    string
	Host        string
	Port        uint
}

func configFromEnv() Config {
	return Config{
		Backend:     getEnvString("DS_BACKEND", "bolt"),
		FileBackend: getEnvString("DS_FILE_BACKEND", "disk"),
		TextIndex:   getEnvString("DS_TEXT_INDEX", "bleve"),
		S3Bucket:    getEnvString("DS_S3_BUCKET", ""),
		FilePrefix:  getEnvString("DS_FILE_PREFIX", "documents"),
		BoltPath:    getEnvString("DS_BOLTDB_PATH", "docshelf.db"),
		Host:        getEnvString("DS_HOST", "localhost"),
		Port:        getEnvUint("DS_PORT", 1337),
	}
}

func main() {
	log = logrus.New()
	cfg := configFromEnv()
	server := http.NewServer(cfg.Host, cfg.Port, log)

	fs, err := getFileStore(cfg)
	if err != nil {
		log.Fatal(err)
	}

	ti, err := getTextIndex(cfg)
	if err != nil {
		log.Fatal(err)
	}

	backend, err := getBackend(cfg, fs, ti)
	if err != nil {
		log.Fatal(err)
	}

	server.UserHandler = http.NewUserHandler(backend, log)
	server.DocHandler = http.NewDocHandler(backend, log)

	if err := server.Start(); err != nil {
		log.Fatal(err)
	}
}

func getFileStore(cfg Config) (docshelf.FileStore, error) {
	switch cfg.FileBackend {
	case "s3":
		fs, err := s3.New(cfg.S3Bucket, cfg.FilePrefix)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create s3 file store")
		}

		return fs, nil
	default:
		fs, err := disk.New(cfg.FilePrefix)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create disk file store")
		}

		return fs, nil
	}
}

func getBackend(cfg Config, fs docshelf.FileStore, ti docshelf.TextIndex) (docshelf.Backend, error) {
	logrus.WithField("backend", cfg.Backend).Info("doc backend")
	switch cfg.Backend {
	case "dynamo":
		log.Info("initializing dynamo backend")
		backend, err := dynamo.New(fs, log)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create dynamo backend")
		}

		return backend, nil
	default:
		log.Info("initializing bolt backend")
		backend, err := bolt.New(cfg.BoltPath, fs, ti)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create bolt backend")
		}

		return backend, nil
	}
}

func getTextIndex(cfg Config) (docshelf.TextIndex, error) {
	switch cfg.TextIndex {
	default:
		ti, err := bleve.New()
		if err != nil {
			return nil, errors.Wrap(err, "failed to create bleve text index")
		}

		return ti, nil
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

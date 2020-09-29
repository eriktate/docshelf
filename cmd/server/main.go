package main

import (
	"context"
	"os"
	"strconv"

	"github.com/docshelf/docshelf"
	"github.com/docshelf/docshelf/bleve"
	"github.com/docshelf/docshelf/bolt"
	"github.com/docshelf/docshelf/disk"
	"github.com/docshelf/docshelf/dynamo"
	"github.com/docshelf/docshelf/http"
	"github.com/docshelf/docshelf/s3"
	"github.com/joho/godotenv"
	"github.com/pkg/errors"
	"github.com/rs/xid"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
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
	godotenv.Load()
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

	// make sure there's a root user
	if err := ensureRoot(backend, log); err != nil {
		log.Fatal(err)
	}

	server.UserStore = backend
	server.DocHandler = http.NewDocHandler(backend, log)
	server.Auth = http.NewAuth(backend)

	if err := server.Start(); err != nil {
		log.Fatal(err)
	}
}

func ensureRoot(us docshelf.UserStore, log *logrus.Logger) error {
	token := xid.New().String()
	// TODO (erik): Adjust the cost parameter once we can benchmark the time spent hashing the password.
	hashed, err := bcrypt.GenerateFromPassword([]byte(token), 12)
	if err != nil {
		return err
	}

	root := docshelf.User{
		Email: "root@docshelf.io",
		Token: string(hashed),
	}

	if _, err := us.GetUser(context.Background(), "root@docshelf.io"); err != nil {
		if docshelf.CheckNotFound(err) {
			if _, err := us.PutUser(context.Background(), root); err != nil {
				return err
			}

			log.WithField("email", root.Email).WithField("password", token).Info("root user created. Keep these credentials secret!")
			return nil
		}

		return errors.Wrap(err, "failed to identify root user")
	}

	return nil
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
		backend, err := dynamo.New(fs, ti, log)
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

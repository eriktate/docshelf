package s3

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/endpoints"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/s3iface"
	"github.com/docshelf/docshelf"
)

// A Store that can write and read documents from S3.
type Store struct {
	client s3iface.S3API
	root   string
	bucket string
}

// New returns a new S3 Store and checks that the given bucket exists. It will use the root as a prefix for
// all objects created.
func New(bucket, root string) (Store, error) {
	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		return Store{}, err
	}

	cfg.Region = endpoints.UsEast1RegionID
	svc := s3.New(cfg)

	// need to make sure the bucket exists
	if _, err := svc.HeadBucketRequest(&s3.HeadBucketInput{Bucket: aws.String(bucket)}).Send(); err != nil {
		return Store{}, err
	}

	return Store{
		client: svc,
		root:   root,
		bucket: bucket,
	}, nil
}

// ReadFile reads the content from an existing s3 object.
func (s Store) ReadFile(path string) ([]byte, error) {
	input := s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(fmt.Sprintf("%s/%s", s.root, path)),
	}

	res, err := s.client.GetObjectRequest(&input).Send()
	if err != nil {
		if strings.Contains(err.Error(), "NoSuchKey") {
			return nil, docshelf.NewErrDoesNotExist(fmt.Sprintf("could not find %s/%s/%s", s.bucket, s.root, path))
		}
		return nil, err
	}

	return ioutil.ReadAll(res.Body)
}

// WriteFile creates or overwrites an object in s3 at the given path with the given content.
func (s Store) WriteFile(path string, content []byte) error {
	input := s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(fmt.Sprintf("%s/%s", s.root, path)),
		Body:   bytes.NewReader(content),
	}

	if _, err := s.client.PutObjectRequest(&input).Send(); err != nil {
		return err
	}

	return nil
}

// RemoveFile removes an existing object from s3.
func (s Store) RemoveFile(path string) error {
	input := s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(fmt.Sprintf("%s/%s", s.root, path)),
	}

	if _, err := s.client.DeleteObjectRequest(&input).Send(); err != nil {
		return err
	}

	return nil
}

// ListDir returns a listing of all objects that exist with a given prefix.
func (s Store) ListDir(path string) ([]string, error) {
	input := s3.ListObjectsInput{
		Bucket: aws.String(s.bucket),
		Prefix: aws.String(fmt.Sprintf("%s/%s", s.root, path)),
	}

	res, err := s.client.ListObjectsRequest(&input).Send()
	if err != nil {
		return nil, err
	}

	listing := make([]string, len(res.Contents))
	for i, obj := range res.Contents {
		// TODO (erik): Maybe strip the prefix from these results.
		listing[i] = *obj.Key
	}

	return listing, nil
}

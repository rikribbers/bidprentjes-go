package cloud

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"cloud.google.com/go/storage"
)

type StorageClient struct {
	client     *storage.Client
	bucketName string
}

func NewStorageClient(ctx context.Context, bucketName string) (*StorageClient, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage client: %v", err)
	}

	return &StorageClient{
		client:     client,
		bucketName: bucketName,
	}, nil
}

func (s *StorageClient) DownloadFile(ctx context.Context, filename string) (io.Reader, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Minute*5)
	defer cancel()

	bucket := s.client.Bucket(s.bucketName)
	obj := bucket.Object(filename)

	reader, err := obj.NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create reader: %v", err)
	}

	content, err := io.ReadAll(reader)
	if err != nil {
		reader.Close()
		return nil, fmt.Errorf("failed to read file content: %v", err)
	}

	reader.Close()

	return bytes.NewReader(content), nil
}

func (s *StorageClient) MoveFile(ctx context.Context, srcPath, dstPath string) error {
	ctx, cancel := context.WithTimeout(ctx, time.Minute*5)
	defer cancel()

	bucket := s.client.Bucket(s.bucketName)
	src := bucket.Object(srcPath)
	dst := bucket.Object(dstPath)

	// Copy the object to the new location
	if _, err := dst.CopierFrom(src).Run(ctx); err != nil {
		return fmt.Errorf("failed to copy file: %v", err)
	}

	// Delete the original object
	if err := src.Delete(ctx); err != nil {
		return fmt.Errorf("failed to delete original file: %v", err)
	}

	return nil
}

func (s *StorageClient) UploadFile(ctx context.Context, filename string, reader io.Reader) error {
	ctx, cancel := context.WithTimeout(ctx, time.Minute*5)
	defer cancel()

	bucket := s.client.Bucket(s.bucketName)
	obj := bucket.Object(filename)

	writer := obj.NewWriter(ctx)
	if _, err := io.Copy(writer, reader); err != nil {
		writer.Close()
		return fmt.Errorf("failed to copy data to GCS: %v", err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close writer: %v", err)
	}

	return nil
}

func (s *StorageClient) Close() error {
	return s.client.Close()
}

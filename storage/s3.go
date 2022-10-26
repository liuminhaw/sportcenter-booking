package storage

import (
	"encoding/hex"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"golang.org/x/crypto/nacl/secretbox"
)

// RoleSession generate and return session pointer
func RoleSession() (*session.Session, error) {
	sess, err := session.NewSession()
	if err != nil {
		return nil, fmt.Errorf("S3RoleSession(): %w", err)
	}

	return sess, nil
}

func DownloadEncObj(sess *session.Session, bucket, key, encKey string) ([]byte, error) {
	buf := aws.NewWriteAtBuffer([]byte{})

	downloader := s3manager.NewDownloader(sess)
	_, err := downloader.Download(buf, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("DownloadEncObj(sess, %s, %s, encKey): %w", bucket, key, err)
	}
	decrypted, err := openObj(buf.Bytes(), encKey)
	if err != nil {
		return nil, fmt.Errorf("DownloadEncObj(sess, %s, %s, encKey): %w", bucket, key, err)
	}
	return decrypted, nil
}

// openObj decrypt input content with 32 bytes hex key and return decrypted result
func openObj(encContent []byte, key string) ([]byte, error) {
	secretKeyBytes, err := hex.DecodeString(key)
	if err != nil {
		return nil, fmt.Errorf("openObj(encContent, key): %w", err)
	}

	var secretKey [32]byte
	copy(secretKey[:], secretKeyBytes)

	var decryptNonce [24]byte
	copy(decryptNonce[:], encContent[:24])
	decrypted, ok := secretbox.Open(nil, encContent[24:], &decryptNonce, &secretKey)
	if !ok {
		return nil, fmt.Errorf("openObj(encContent, key): %w", err)
	}
	return decrypted, nil
}

package registry

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"golang.org/x/crypto/nacl/secretbox"
)

type Reservation struct {
	Username     string    `json:"username"`
	Password     string    `json:"password"`
	ReserveDate  time.Time `json:"reserveDate"`
	ReserveCourt string    `json:"reserveCourt"`
	ReserveTime  string    `json:"reserveTime"`
}

type Registry struct {
	Bucket   string
	Dirname  string
	Filename string
	Content  []byte
}

// CreateRegistryFile create temporary file in /tmp directory
// with content of Registry.Content
func (r Registry) CreateRegistryFile(key string) {
	f, err := os.Create(fmt.Sprintf("/tmp/%s", r.Filename))
	if err != nil {
		panic(err)
	}
	defer f.Close()

	encryptedContent := r.sealFile(key)
	f.Write(encryptedContent)
}

// UploadRegistryFile upload the file (created by CreateRegistryFile method)
// to S3 at path s3://Bucket/Dirname/Filename
func (r Registry) UploadRegistryFile(sess *session.Session) {
	f, err := os.Open(fmt.Sprintf("/tmp/%s", r.Filename))
	if err != nil {
		panic(err)
	}
	defer f.Close()

	uploader := s3manager.NewUploader(sess)
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket:      aws.String(r.Bucket),
		Key:         aws.String(fmt.Sprintf("%s/%s", r.Dirname, r.Filename)),
		Body:        f,
		ContentType: aws.String("application/json"),
	})
	if err != nil {
		fmt.Printf("Unable to upload %v to %q, %v", f, r.Bucket, err)
		return
	}
	fmt.Printf("Successfully uploaded %v to %q\n", f, r.Bucket)
}

// sealFile encrypt Registry.Content with provided 32 bytes hexKey
// and return the encrypted result
func (r Registry) sealFile(hexKey string) []byte {
	secretKeyBytes, err := hex.DecodeString(hexKey)
	if err != nil {
		panic(err)
	}

	var secretKey [32]byte
	copy(secretKey[:], secretKeyBytes)

	var nonce [24]byte
	if _, err := io.ReadFull(rand.Reader, nonce[:]); err != nil {
		panic(err)
	}

	return secretbox.Seal(nonce[:], r.Content, &nonce, &secretKey)
}

// openFile decrypt given encrypted input with 32 byte hexKey and assigned decrypted
// value to Registry.Content
func (r Registry) openFile(encrypted []byte, hexKey string) {
	secretKeyBytes, err := hex.DecodeString(hexKey)
	if err != nil {
		panic(err)
	}

	var secretKey [32]byte
	copy(secretKey[:], secretKeyBytes)

	var decryptNonce [24]byte
	copy(decryptNonce[:], encrypted[:24])
	decrypted, ok := secretbox.Open(nil, encrypted[24:], &decryptNonce, &secretKey)
	if !ok {
		panic("decryption error")
	}

	r.Content = decrypted
}

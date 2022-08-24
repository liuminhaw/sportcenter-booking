package registry

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
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

type S3Obj struct {
	reg     *S3Register
	Name    string
	Content []byte
}

type S3Register struct {
	Bucket  string
	Prefix  string
	Objects []S3Obj
}

// ListObjects read all objects from s3://r.Bucket/r.Prefix/ and store
// objects to r.Objects. Objects is listed without fetching their content.
//
// TODO: ListObjectsV2 is paginated, listing objects need to check if there
// are more objects not send in single request
func (r *S3Register) ListObjects(sess *session.Session) error {
	svc := s3.New(sess)
	objects, err := svc.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket:     aws.String(r.Bucket),
		Prefix:     aws.String(fmt.Sprintf("%s/", r.Prefix)),
		StartAfter: aws.String(fmt.Sprintf("%s/", r.Prefix)),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeNoSuchBucket:
				fmt.Println(s3.ErrCodeNoSuchBucket, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			fmt.Println(err.Error())
		}
		return err
	}

	for _, obj := range objects.Contents {
		s3obj := S3Obj{
			reg:  r,
			Name: filepath.Base(*obj.Key),
		}
		r.Objects = append(r.Objects, s3obj)
	}
	return nil
}

// GetRegister return parent S3Register of S3Obj
func (obj *S3Obj) GetRegister() S3Register {
	return *obj.reg
}

// UploadEncObj first encrypt file content then upload it to S3 Bucket
// at s3://S3Register.Bucket/S3Register.Prefix/S3Obj.Name
func (obj *S3Obj) UploadEncObj(sess *session.Session, key string) {
	encryptedContent := obj.sealObj(key)

	r := obj.GetRegister()
	uploader := s3manager.NewUploader(sess)
	_, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(r.Bucket),
		Key:    aws.String(fmt.Sprintf("%s/%s", r.Prefix, obj.Name)),
		Body:   bytes.NewReader(encryptedContent),
		// ContentType: aws.String("application/json"),
	})
	if err != nil {
		fmt.Printf("Unable to upload %s to %s: %v", obj.Name, r.Bucket, err)
		return
	}
	fmt.Printf("Successfully upload %s to %s\n", obj.Name, r.Bucket)
}

// DownloadEncObj download object from S3 then decrypt object content and store
// result into obj.Content
func (obj *S3Obj) DownloadEncObj(sess *session.Session, key string) error {
	buf := aws.NewWriteAtBuffer([]byte{})

	r := obj.GetRegister()
	downloader := s3manager.NewDownloader(sess)
	filepath := fmt.Sprintf("%s/%s", r.Prefix, obj.Name)
	_, err := downloader.Download(buf, &s3.GetObjectInput{
		Bucket: aws.String(r.Bucket),
		Key:    aws.String(filepath),
	})
	if err != nil {
		fmt.Printf("Unable to fetch file %s from bucket %s\n", filepath, r.Bucket)
		return err
	}
	fmt.Printf("Successfully fetch file %s from bucket %s\n", filepath, r.Bucket)

	obj.openObj(buf.Bytes(), key)
	return nil
}

// sealObj encrypt object content with provided 32 bytes hex key
// and return result in bytes
//
// TODO: error handling
func (obj *S3Obj) sealObj(key string) []byte {
	secretKeyBytes, err := hex.DecodeString(key)
	if err != nil {
		panic(err)
	}

	var secretKey [32]byte
	copy(secretKey[:], secretKeyBytes)

	var nonce [24]byte
	if _, err := io.ReadFull(rand.Reader, nonce[:]); err != nil {
		panic(err)
	}

	return secretbox.Seal(nonce[:], obj.Content, &nonce, &secretKey)
}

// openObj decrypt input content with 32 bytes hex key and store the
// result to obj.Content
func (obj *S3Obj) openObj(encContent []byte, key string) {
	secretKeyBytes, err := hex.DecodeString(key)
	if err != nil {
		panic(err)
	}

	var secretKey [32]byte
	copy(secretKey[:], secretKeyBytes)

	var decryptNonce [24]byte
	copy(decryptNonce[:], encContent[:24])
	decrypted, ok := secretbox.Open(nil, encContent[24:], &decryptNonce, &secretKey)
	if !ok {
		panic("decryption error")
	}

	obj.Content = decrypted
}

// ============================================================
// Deprecated below
// Plan to be replace by structures and methods defined above
type Registry struct {
	Bucket   string
	Dirname  string
	Filename string
	Content  []byte
}

// CreateRegistryFile create temporary file in /tmp directory
// with content of Registry.Content
func (r *Registry) CreateRegistryFile(key string) {
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
func (r *Registry) UploadRegistryFile(sess *session.Session) {
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

// UploadRegistry upload encrypted Registry.Content to S3 at path
// s3://Bucket/Dirname/Filename
func (r *Registry) UploadRegistry(sess *session.Session, key string) {
	encryptedContent := r.sealFile(key)

	uploader := s3manager.NewUploader(sess)
	_, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(r.Bucket),
		Key:    aws.String(fmt.Sprintf("%s/%s", r.Dirname, r.Filename)),
		Body:   bytes.NewReader(encryptedContent),
		// ContentType: aws.String("application/json"),
	})
	if err != nil {
		fmt.Printf("Unable to upload to %q: %v", r.Bucket, err)
		return
	}
	fmt.Printf("Successfully uploaded to %q\n", r.Bucket)
}

// FetchRegistryFile read content from s3 file and store content into Registry r
func (r *Registry) FetchRegistryFile(sess *session.Session, key string) error {
	buf := aws.NewWriteAtBuffer([]byte{})

	downloader := s3manager.NewDownloader(sess)
	filepath := fmt.Sprintf("%s/%s", r.Dirname, r.Filename)
	_, err := downloader.Download(buf, &s3.GetObjectInput{
		Bucket: aws.String(r.Bucket),
		Key:    aws.String(filepath),
	})
	if err != nil {
		fmt.Printf("Unable to fetch file %s from bucket %s\n", filepath, r.Bucket)
		return errors.New("s3 fetch file error")
	}
	fmt.Printf("Successfully fetch file %s from bucket %s\n", filepath, r.Bucket)

	r.openFile(buf.Bytes(), key)
	return nil
}

// GetS3Objects returns list of objects within s3://Bucket/Dirname/ except for
// s3://Bucket/Dirname/
func (r *Registry) GetS3Objects(sess *session.Session) (*s3.ListObjectsV2Output, error) {
	svc := s3.New(sess)

	objects, err := svc.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket:     aws.String(r.Bucket),
		Prefix:     aws.String(fmt.Sprintf("%s/", r.Dirname)),
		StartAfter: aws.String(fmt.Sprintf("%s/", r.Dirname)),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeNoSuchBucket:
				fmt.Println(s3.ErrCodeNoSuchBucket, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			fmt.Println(err.Error())
		}
		return nil, err
	}
	return objects, nil
}

// sealFile encrypt Registry.Content with provided 32 bytes hexKey
// and return the encrypted result
func (r *Registry) sealFile(hexKey string) []byte {
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
func (r *Registry) openFile(encrypted []byte, hexKey string) {
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

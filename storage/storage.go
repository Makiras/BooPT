package storage

import (
	. "BooPT/config"
	"context"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/sirupsen/logrus"
	"net/url"
	"time"
)

var minioClient *minio.Client

func InitClient() error {
	endpoint := CONFIG.S3.Endpoint
	accessKey := CONFIG.S3.AccessKey
	secretKey := CONFIG.S3.SecretKey
	useSSL := false

	// Initialize minio client object.
	_minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	minioClient = _minioClient

	if err != nil {
		logrus.Errorf("Failed to initialize minio client. Error: %#v", err)
		return err
	}
	logrus.Infof("minioClient has been setup: %v ", minioClient) // minioClient is now setup
	return nil
}

func GetPresignedURL(objectName string, expires int64, filename string, preview bool) (*url.URL, error) {
	// objectName is stored in the DownloadLink field of the database, such as: "/path/to/file.pdf" or "path/to/file.epub"
	// filename := book.Title + book.Version + book.Publisher + book.Extension

	reqParams := make(url.Values)
	if preview {
		reqParams.Set("response-content-disposition", "inline")
	} else {
		reqParams.Set("response-content-disposition", "attachment; filename=\""+filename+"\"")
	}
	return minioClient.PresignedGetObject(
		context.Background(), CONFIG.S3.Bucket, objectName, time.Second*time.Duration(expires), reqParams)
}

func PostFilePresignedURL(objectPath string, objectType string, expires int64) (*url.URL, map[string]string, error) {
	// file extension is contained in the DownloadLink field of the database. Check GetPresignedURL() for more info.

	policy := minio.NewPostPolicy()
	if err := policy.SetBucket(CONFIG.S3.Bucket); err != nil {
		logrus.Errorf("Failed to set bucket name. Error: %#v", policy.SetBucket(CONFIG.S3.Bucket))
		return nil, nil, err
	}
	if err := policy.SetExpires(time.Now().UTC().Add(time.Duration(expires) * time.Second)); err != nil {
		logrus.Errorf("Failed to set expires. Error: %#v", err)
		return nil, nil, err
	}

	// Only allow 'pdf','epub','mobi','iba','awz3' upload.
	if objectType != "pdf" && objectType != "epub" && objectType != "mobi" && objectType != "iba" && objectType != "awz3" {
		return nil, nil, minio.ErrorResponse{
			Code:      "InvalidArgument",
			Message:   "Invalid file extension",
			Resource:  objectPath + "file." + objectType,
			RequestID: "",
		}
	}

	// I don't care any error reason here, if failed just burn it.
	errct := policy.SetContentType("application/" + objectType)
	errk := policy.SetKey(objectPath + "file." + objectType)
	errclr := policy.SetContentLengthRange(1024, 2*1024*1024*1024) // Only allow content size in range 1KB to 2GB.
	if errct != nil || errk != nil || errclr != nil {
		logrus.Errorf("Failed to set policy. Error: %v,%v,%v ", errk, errct, errclr)
		return nil, nil, minio.ErrorResponse{
			Code:      "InvalidArgument",
			Message:   "Unexpected error",
			Resource:  "file." + objectType,
			RequestID: "",
		}
	}

	// Get the POST form key/value object:
	uploadLink, formData, err := minioClient.PresignedPostPolicy(context.Background(), policy)
	if err != nil {
		logrus.Errorf("Failed to get presigned post policy. Error: %#v", err)
		return nil, nil, err
	}
	return uploadLink, formData, nil
}

func CheckObjectExistence(objectName string) error {
	_, err := minioClient.StatObject(context.Background(), CONFIG.S3.Bucket, objectName, minio.StatObjectOptions{})
	return err
}

func DownloadFile(objectName string, filePath string) error {
	return minioClient.FGetObject(context.Background(), CONFIG.S3.Bucket, objectName, filePath, minio.GetObjectOptions{})
}

func UploadFile(objectName string, filePath string) error {
	_, err := minioClient.FPutObject(context.Background(), CONFIG.S3.Bucket, objectName, filePath, minio.PutObjectOptions{})
	return err
}

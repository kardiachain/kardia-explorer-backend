package aws

import (
	"bytes"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/kardiachain/kardia-explorer-backend/utils"
	"strings"
)

type Config struct {
	KeyID, AccessKey, Region string
}

type ConfigUploader struct {
	Bucket, ACL, Key, PathAvatar string
}

type FileStorage interface {
	UploadLogo(rawString string, fileName string, configUploader ConfigUploader) (string, error)
}

type S3 struct {
	Config
	session *session.Session
}

func (s *S3) UploadLogo(rawString string, fileName string, configUploader ConfigUploader) (string, error) {
	uploader := s3manager.NewUploader(s.session)

	if strings.Contains(rawString, "https") && (strings.Contains(rawString, "png") || strings.Contains(rawString, "jpeg") || strings.Contains(rawString, "webp")) {
		return rawString, nil
	}

	fileUpload, err := utils.Base64ToImage(rawString)
	if err != nil {
		return "", err
	}

	sendS3, uploadedFileName := utils.EncodeImage(fileUpload, rawString, fileName)
	if _, errUploader := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(configUploader.Bucket),
		ACL:    aws.String(configUploader.ACL),
		Key:    aws.String(configUploader.Key),
		Body:   bytes.NewReader(sendS3),
	}); errUploader != nil {
		return "", errUploader
	}

	return configUploader.PathAvatar + uploadedFileName, nil
}

func ConnectAws(config Config) (FileStorage, error) {
	sess, err := session.NewSession(
		&aws.Config{
			Region: aws.String(config.Region),
			Credentials: credentials.NewStaticCredentials(
				config.KeyID,
				config.AccessKey,
				"",
			),
		})

	if err != nil {
		return nil, err
	}

	s3Aws := &S3{
		session: sess,
	}

	return s3Aws, nil
}

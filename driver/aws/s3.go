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
	KeyID     string
	KeyAccess string
	Region    *string
}

type FileStorage interface {
	UploadLogo(rawString string, fileName string) (string, error)
}

type S3 struct {
	Config
	Session *session.Session
}

func (s *S3) UploadLogo(rawString string, fileName string) (string, error) {
	uploader := s3manager.NewUploader(s.Session)

	if strings.Contains(rawString, "https") && (strings.Contains(rawString, "png") || strings.Contains(rawString, "jpeg") || strings.Contains(rawString, "webp")) {
		return rawString, nil
	}

	fileUpload, err := utils.Base64ToImage(rawString)
	if err != nil {
		return "", err
	}

	sendS3, uploadedFileName := utils.EncodeImage(fileUpload, rawString, fileName)

	_, errUploader := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String("cdn1.bcms.tech"),
		ACL:    aws.String("public-read"),
		Key:    aws.String("/kai-explorer-backend/logo/" + uploadedFileName),
		Body:   bytes.NewReader(sendS3),
	})

	if errUploader != nil {
		return "", errUploader
	}
	pathAvatar := "https://s3-ap-southeast-1.amazonaws.com/cdn1.bcms.tech/kai-explorer-backend/logo/"
	filepath := pathAvatar + uploadedFileName

	return filepath, nil
}

func ConnectAws() (FileStorage, error) {
	KeyID := "AKIAJI3Y5XWKQTDRL5HQ"
	KeyAccess := "GWGuKvvVnUAQCGAmY937QcKkX//0RR2SPrdh+F3w"
	Region := aws.String("ap-southeast-1")
	sess, err := session.NewSession(
		&aws.Config{
			Region: Region,
			Credentials: credentials.NewStaticCredentials(
				KeyID,
				KeyAccess,
				"",
			),
		})

	if err != nil {
		return nil, err
	}

	s3Aws := &S3{
		Session: sess,
	}

	return s3Aws, nil
}

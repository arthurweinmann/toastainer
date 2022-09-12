package s3

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/rs/xid"
	"github.com/toastate/toastcloud/internal/config"
	"github.com/toastate/toastcloud/internal/db/objectstorage/objectstoragerror"
	"github.com/toastate/toastcloud/internal/utils"
)

const (
	S3KeyPrefix string = "toastcloud/"
)

type s3Handler struct {
	s3svc  *s3.S3
	s3up   *s3manager.Uploader
	s3down *s3manager.Downloader
}

func NewHandler() (*s3Handler, error) {
	h := &s3Handler{}
	err := h.setup()
	if err != nil {
		return nil, err
	}
	return h, nil
}

func (h *s3Handler) setup() error {
	err := os.MkdirAll(filepath.Join(config.Home, "tmp"), 0777)
	if err != nil {
		return err
	}

	awsCfg := aws.NewConfig().WithRegion(config.ObjectStorage.AWSS3.Region)
	awsCfg = awsCfg.WithCredentials(credentials.NewStaticCredentials(config.ObjectStorage.AWSS3.PubKey, config.ObjectStorage.AWSS3.PrivKey, ""))

	h.s3svc = s3.New(session.New(), awsCfg)
	h.s3up = s3manager.NewUploader(session.New(awsCfg))

	h.s3down = s3manager.NewDownloader(session.New(awsCfg))
	h.s3down.Concurrency = 1 // force sequential downloads to stream them into the false io.WriterAt which is the http.ResponseWriter

	return nil
}

func (h *s3Handler) DownloadFileInto(w http.ResponseWriter, remotePath string) error {
	dest := filepath.Join(S3KeyPrefix, remotePath)

	tmppath := filepath.Join(config.Home, "tmp", xid.New().String())
	f, err := os.OpenFile(tmppath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer func() {
		f.Close()
		os.Remove(tmppath)
	}()

	_, err = h.s3down.Download(f, &s3.GetObjectInput{
		Bucket: &config.ObjectStorage.AWSS3.Bucket,
		Key:    &dest,
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeNoSuchBucket, s3.ErrCodeNoSuchKey:
				return objectstoragerror.ErrNotFound
			}
		}

		return err
	}

	_, err = f.Seek(0, 0)
	if err != nil {
		return err
	}

	b := make([]byte, 512)
	var n, nn int
F:
	for nn < len(b) {
		n, err = f.Read(b[nn:])
		nn += n
		if err != nil {
			if err == io.EOF {
				break F
			}

			return err
		}
		if n == 0 {
			break F
		}
	}
	b = b[:nn]

	mt := http.DetectContentType(b)
	if mt == "application/octet-stream" {
		tmp := utils.MimetypeFromFileExtension(dest)
		if tmp != "" {
			mt = tmp
		}
	}

	w.Header().Set("Content-Type", mt)

	_, err = f.Seek(0, 0)
	if err != nil {
		return err
	}

	_, err = io.Copy(w, f)
	if err != nil && err != io.EOF {
		return err
	}

	return nil
}

func (h *s3Handler) PushReader(f io.Reader, destPath string) error {
	dest := filepath.Join(S3KeyPrefix, destPath)
	_, err := h.s3up.Upload(&s3manager.UploadInput{
		Bucket: &config.ObjectStorage.AWSS3.Bucket,
		Key:    &dest,
		Body:   f,
	})
	return err
}

func (h *s3Handler) PushFolderTar(folder, destPath string) error {
	cmd := exec.Command("tar", "cz", "./")
	cmd.Dir = folder
	dest := filepath.Join(S3KeyPrefix, destPath)
	buf := bytes.NewBuffer(nil)
	bufferr := bytes.NewBuffer(nil)
	cmd.Stdout = buf
	cmd.Stderr = bufferr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("%v: %v", err, bufferr.String())
	}
	_, err = h.s3up.Upload(&s3manager.UploadInput{
		Bucket: &config.ObjectStorage.AWSS3.Bucket,
		Key:    &dest,
		Body:   buf,
	})
	if err != nil {
		return fmt.Errorf("%v: %v", err, bufferr.String())
	}
	return nil
}

func (h *s3Handler) PullFolderTar(remotePath, destination string) error {
	dest := filepath.Join(S3KeyPrefix, remotePath)
	obj, err := h.s3svc.GetObject(&s3.GetObjectInput{
		Bucket: &config.ObjectStorage.AWSS3.Bucket,
		Key:    &dest,
	})
	if err != nil {
		return err
	}
	cmd := exec.Command("tar", "-C", destination, "-xz")
	cmd.Stdin = obj.Body
	buf := bytes.NewBuffer(nil)
	cmd.Stdout = buf
	cmd.Stderr = buf
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("%v: %v", err, buf.String())
	}
	return nil
}

func (h *s3Handler) DeleteFolder(remotePath string) error {
	dest := filepath.Join(S3KeyPrefix, remotePath)

	iter := s3manager.NewDeleteListIterator(h.s3svc, &s3.ListObjectsInput{
		Bucket: &config.ObjectStorage.AWSS3.Bucket,
		Prefix: &dest,
	})

	err := s3manager.NewBatchDeleteWithClient(h.s3svc).Delete(context.Background(), iter)
	if err != nil {
		return err
	}

	return nil
}

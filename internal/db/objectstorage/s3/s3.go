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
	"github.com/schollz/progressbar/v3"
	"github.com/toastate/toastainer/internal/config"
	"github.com/toastate/toastainer/internal/db/objectstorage/objectstoragerror"
	"github.com/toastate/toastainer/internal/utils"
)

const (
	S3KeyPrefix string = "toastainer/"
)

type s3Handler struct {
	s3svc  *s3.S3
	s3up   *s3manager.Uploader
	s3down *s3manager.Downloader
	tmpdir string
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
	var err error

	h.tmpdir, err = os.MkdirTemp("", "toastainer_s3handler")
	if err != nil {
		return err
	}

	awsCfg := aws.NewConfig().WithRegion(config.ObjectStorage.AWSS3.Region)
	awsCfg = awsCfg.WithCredentials(credentials.NewStaticCredentials(config.ObjectStorage.AWSS3.PubKey, config.ObjectStorage.AWSS3.PrivKey, ""))

	sess1, err := session.NewSession(awsCfg)
	if err != nil {
		return err
	}
	h.s3svc = s3.New(sess1, awsCfg)

	sess2, err := session.NewSession(awsCfg)
	if err != nil {
		return err
	}
	h.s3up = s3manager.NewUploader(sess2)

	sess3, err := session.NewSession(awsCfg)
	if err != nil {
		return err
	}
	h.s3down = s3manager.NewDownloader(sess3)

	return nil
}

func (h *s3Handler) Get(remotePath string) ([]byte, error) {
	dest := filepath.Join(S3KeyPrefix, remotePath)

	buff := &aws.WriteAtBuffer{}

	_, err := h.s3down.Download(buff, &s3.GetObjectInput{
		Bucket: &config.ObjectStorage.AWSS3.Bucket,
		Key:    &dest,
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case "NotFound", s3.ErrCodeNoSuchBucket, s3.ErrCodeNoSuchKey: // s3.ErrCodeNoSuchKey does not work, aws is missing this error code so we hardwire a string
				return nil, objectstoragerror.ErrNotFound
			}
		}

		return nil, err
	}

	return buff.Bytes(), nil
}

func (h *s3Handler) DownloadFileInto(w http.ResponseWriter, remotePath string) error {
	dest := filepath.Join(S3KeyPrefix, remotePath)

	tmppath := filepath.Join(h.tmpdir, xid.New().String())
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
			case "NotFound", s3.ErrCodeNoSuchBucket, s3.ErrCodeNoSuchKey: // s3.ErrCodeNoSuchKey does not work, aws is missing this error code so we hardwire a string
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

func (h *s3Handler) PullFolderTarWithProgressBar(remotePath, destination string) error {
	return h.pullFolderTar(remotePath, destination, true)
}

func (h *s3Handler) PullFolderTar(remotePath, destination string) error {
	return h.pullFolderTar(remotePath, destination, false)
}

func (h *s3Handler) pullFolderTar(remotePath, destination string, withProgressBar bool) error {
	dest := filepath.Join(S3KeyPrefix, remotePath)
	obj, err := h.s3svc.GetObject(&s3.GetObjectInput{
		Bucket: &config.ObjectStorage.AWSS3.Bucket,
		Key:    &dest,
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case "NotFound", s3.ErrCodeNoSuchBucket, s3.ErrCodeNoSuchKey: // s3.ErrCodeNoSuchKey does not work, aws is missing this error code so we hardwire a string
				return objectstoragerror.ErrNotFound
			}
		}

		return err
	}
	defer obj.Body.Close()

	cmd := exec.Command("tar", "-C", destination, "-xz")

	if withProgressBar {
		progresschan := make(chan int, 1)
		defer func() {
			progresschan <- -1
		}()
		proxyr := utils.NewProxyReader(obj.Body, progresschan)
		bar := progressbar.DefaultBytes(*obj.ContentLength, "downloading "+remotePath)
		go func() {
			prevp := 0
			for {
				p := <-progresschan
				if p < 0 {
					return
				}
				bar.Add(p - prevp)
				prevp = p
			}
		}()

		cmd.Stdin = proxyr
	} else {
		cmd.Stdin = obj.Body
	}

	buf := bytes.NewBuffer(nil)
	cmd.Stdout = buf
	cmd.Stderr = buf
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("%v: %v", err, buf.String())
	}
	return nil
}

func (h *s3Handler) Exists(remotePath string) (bool, error) {
	dest := filepath.Join(S3KeyPrefix, remotePath)
	_, err := h.s3svc.HeadObject(&s3.HeadObjectInput{
		Bucket: &config.ObjectStorage.AWSS3.Bucket,
		Key:    &dest,
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case "NotFound", s3.ErrCodeNoSuchBucket, s3.ErrCodeNoSuchKey: // s3.ErrCodeNoSuchKey does not work, aws is missing this error code so we hardwire a string
				return false, nil
			default:
				return false, err
			}
		}
		return false, err
	}
	return true, nil
}

func (h *s3Handler) DeleteFolder(remotePath string) error {
	dest := filepath.Join(S3KeyPrefix, remotePath)

	iter := s3manager.NewDeleteListIterator(h.s3svc, &s3.ListObjectsInput{
		Bucket: &config.ObjectStorage.AWSS3.Bucket,
		Prefix: &dest,
	})

	err := s3manager.NewBatchDeleteWithClient(h.s3svc).Delete(context.Background(), iter)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case "NotFound", s3.ErrCodeNoSuchBucket, s3.ErrCodeNoSuchKey: // s3.ErrCodeNoSuchKey does not work, aws is missing this error code so we hardwire a string
				return objectstoragerror.ErrNotFound
			}
		}

		return err
	}

	return nil
}

package s3

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/toastate/toastcloud/internal/config"
)

func (h *S3Handler) UploadFolder(folder, dest string) error {
	dest = filepath.Join(S3KeyPrefix, dest)

	var rels []string

	err := filepath.Walk(folder, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			rel, err := filepath.Rel(folder, path)
			if err != nil {
				return fmt.Errorf("Unable to get relative path: %v %v", path, err)
			}
			file, err := os.Open(path)
			if err != nil {
				return fmt.Errorf("Failed opening file: %v %v", path, err)
			}
			defer file.Close()
			_, err = h.s3up.Upload(&s3manager.UploadInput{
				Bucket: &config.ObjectStorage.AWSS3.Bucket,
				Key:    aws.String(filepath.Join(dest, rel)),
				Body:   file,
			})
			if err != nil {
				return fmt.Errorf("Failed to upload: %v %v", path, err)
			}
			rels = append(rels, rel)
		}

		return nil
	})

	if err != nil {
		for k := 0; k < len(rels)/1000+1; k++ {
			param := &s3.DeleteObjectsInput{
				Delete: &s3.Delete{},
			}
			for i := k * 1000; i < k*1000+1000 && i < len(rels); i++ {
				param.Delete.Objects = append(param.Delete.Objects, &s3.ObjectIdentifier{
					Key: aws.String(rels[i]),
				})
			}
			if len(param.Delete.Objects) > 0 {
				_, e := h.s3svc.DeleteObjects(param)
				if e != nil {
					break
				}
			}
		}

		return err
	}

	return nil
}

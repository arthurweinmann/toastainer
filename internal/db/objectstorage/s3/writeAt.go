package s3

import (
	"fmt"
	"net/http"

	"github.com/toastate/toastcloud/internal/utils"
)

// FakeWriterAt handles setting the content type header
type FakeWriterAt struct {
	buf         []byte
	mimetypeset bool
	path        string
	w           http.ResponseWriter
}

func NewFakeWriterAt(w http.ResponseWriter, path string) *FakeWriterAt {
	return &FakeWriterAt{
		w:    w,
		path: path,
		buf:  make([]byte, 0, 512),
	}
}

func (fw FakeWriterAt) WriteAt(p []byte, offset int64) (n int, err error) {
	// ignore 'offset' because we forced sequential downloads

	var poff, nn int

	if !fw.mimetypeset {
		if len(fw.buf)+len(p) >= 512 {
			fw.buf = append(fw.buf, p[:512-len(fw.buf)]...)
			poff = 512 - len(fw.buf)

			mt := http.DetectContentType(fw.buf)
			if mt == "application/octet-stream" {
				tmp := utils.MimetypeFromFileExtension(fw.path)
				if tmp != "" {
					mt = tmp
				}
			}

			fw.w.Header().Set("Content-Type", mt)
			fw.mimetypeset = true

			nn, err = fw.w.Write(fw.buf)
			n += nn
			if err != nil {
				return
			}
			if nn != len(fw.buf) {
				err = fmt.Errorf("broken pipe")
				return
			}
			fw.buf = fw.buf[:0]
		} else {
			fw.buf = append(fw.buf, p...)
			return len(p), nil
		}
	}

	if poff < len(p) {
		nn, err = fw.w.Write(p[poff:])
		n += nn
	}

	return
}

func (fw FakeWriterAt) Flush() error {
	if len(fw.buf) > 0 {
		n, err := fw.w.Write(fw.buf)
		if err != nil {
			return err
		}
		if n != len(fw.buf) {
			return fmt.Errorf("broken pipe")
		}
		fw.buf = fw.buf[:0]
	}

	return nil
}

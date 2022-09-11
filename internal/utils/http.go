package utils

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

func IsWebSocketRequest(req *http.Request) bool {
	return websocket.IsWebSocketUpgrade(req)
}

func Redirect2HTTPS(w http.ResponseWriter, req *http.Request) {
	// remove/add not default ports from req.Host
	target := "https://" + req.Host + req.URL.Path
	if len(req.URL.RawQuery) > 0 {
		target += "?" + req.URL.RawQuery
	}
	if len(req.URL.RawFragment) > 0 {
		target += "#" + req.URL.RawFragment
	}
	http.Redirect(w, req, target,
		// consider the codes 308, 302, or 301
		http.StatusTemporaryRedirect)
}

//IsMultipart returns true if the given request is multipart forrm
func IsMultipart(r *http.Request) bool {
	return strings.Index(r.Header.Get("Content-Type"), "multipart/form-data") > -1
}

type jsonErr struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Code    string `json:"code,omitempty"`
}

func SendInternalError(w http.ResponseWriter, origin string, err error) {
	Error("msg", origin, err)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(501)
	b, _ := json.Marshal(&jsonErr{
		Success: false,
		Message: "it looks like we have an issue on our side, please retry later",
		Code:    "internal",
	})
	w.Write(b)
}

func SendError(w http.ResponseWriter, message, code string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	b, _ := json.Marshal(&jsonErr{
		Success: false,
		Message: message,
		Code:    code,
	})
	w.Write(b)
}

func SendErrorAndLog(w http.ResponseWriter, message, code string, statusCode int, origin string, err error) {
	Error("msg", origin, message, code, err)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	b, _ := json.Marshal(&jsonErr{
		Success: false,
		Message: message,
		Code:    code,
	})
	w.Write(b)
}

func SendSuccess(w http.ResponseWriter, resp interface{}) {
	w.WriteHeader(200)
	w.Header().Add("Content-Type", "application/json")

	if resp == nil {
		w.Write([]byte(`{"success": true}`))
		return
	}

	b, err := JSONMarshal(resp)
	if err != nil {
		SendError(w, err.Error(), "invalidresponse", 501)
		return
	}

	w.Write(b)
}

// JSONMarshal does not escape HTML
func JSONMarshal(t interface{}) ([]byte, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(t)
	return buffer.Bytes(), err
}

func UploadFileMultipart(url string, path string) (*http.Response, error) {
	f, err := os.OpenFile(path, os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}

	// Reduce number of syscalls when reading from disk.
	bufferedFileReader := bufio.NewReader(f)
	defer f.Close()

	// Create a pipe for writing from the file and reading to
	// the request concurrently.
	bodyReader, bodyWriter := io.Pipe()
	formWriter := multipart.NewWriter(bodyWriter)

	// Store the first write error in writeErr.
	var (
		writeErr error
		errOnce  sync.Once
	)
	setErr := func(err error) {
		if err != nil {
			errOnce.Do(func() { writeErr = err })
		}
	}
	go func() {
		partWriter, err := formWriter.CreateFormFile("file", filepath.Base(path))
		setErr(err)
		if err != nil {
			fmt.Println("formWriter.CreateFormFile", err)
		} else {
			for {
				n, err := io.CopyN(partWriter, bufferedFileReader, 1024*1024*5)
				if err != nil {
					if err != io.EOF {
						setErr(err)
					}
					break
				}
				if n == 0 {
					break
				}
			}
		}

		// _, err = io.Copy(partWriter, bufferedFileReader)
		setErr(err)
		setErr(formWriter.Close())
		setErr(bodyWriter.Close())
	}()

	req, err := http.NewRequest(http.MethodPut, url, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", formWriter.FormDataContentType())

	// This operation will block until both the formWriter
	// and bodyWriter have been closed by the goroutine,
	// or in the event of a HTTP error.
	resp, err := (&http.Client{
		Timeout: 3600 * time.Hour,
	}).Do(req)
	if err != nil {
		fmt.Println("3", err)
		return nil, err
	}

	if writeErr != nil {
		return nil, writeErr
	}

	return resp, err
}

func UploadFolderMultipart(url, method string, folder string, fields ...string) (*http.Response, error) {

	// Create a pipe for writing from the file and reading to
	// the request concurrently.
	bodyReader, bodyWriter := io.Pipe()
	formWriter := multipart.NewWriter(bodyWriter)

	// Store the first write error in writeErr.
	var (
		writeErr error
		errOnce  sync.Once
	)
	setErr := func(err error) {
		if err != nil {
			errOnce.Do(func() { writeErr = err })
		}
	}
	go func() {
		err := filepath.Walk(folder, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			rel, err := filepath.Rel(folder, path)
			if err != nil {
				return err
			}

			partWriter, err := formWriter.CreateFormFile("file", rel)
			if err != nil {
				return err
			}

			f, err := os.OpenFile(path, os.O_RDONLY, 0644)
			if err != nil {
				return err
			}

			// Reduce number of syscalls when reading from disk.
			bufferedFileReader := bufio.NewReader(f)
			defer f.Close()

			for {
				n, err := io.CopyN(partWriter, bufferedFileReader, 1024*1024*5)
				if err != nil {
					if err != io.EOF {
						return err
					}
					break
				}
				if n == 0 {
					break
				}
			}

			return nil
		})

		// _, err = io.Copy(partWriter, bufferedFileReader)
		setErr(err)

		if err == nil {
			for i := 0; i < len(fields); i += 2 {
				err := formWriter.WriteField(fields[i], fields[i+1])
				if err != nil {
					setErr(err)
					break
				}
			}
		}

		setErr(formWriter.Close())
		setErr(bodyWriter.Close())
	}()

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", formWriter.FormDataContentType())

	// This operation will block until both the formWriter
	// and bodyWriter have been closed by the goroutine,
	// or in the event of a HTTP error.
	resp, err := (&http.Client{
		Timeout: 3600 * time.Hour,
	}).Do(req)
	if err != nil {
		fmt.Println("3", err)
		return nil, err
	}

	if writeErr != nil {
		return nil, writeErr
	}

	return resp, err
}

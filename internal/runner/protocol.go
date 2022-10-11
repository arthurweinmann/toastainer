package runner

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"io"

	"github.com/gorilla/websocket"
	"github.com/toastate/toastainer/internal/utils"
)

const (
	streamChunkSize = 128 * 1024
)

type MessageKind byte

const (
	ProxyKind   MessageKind = 1
	ExecuteKind MessageKind = 2
	BuildKind   MessageKind = 3
	GCExeKind   MessageKind = 4
	LogKind     MessageKind = 5
)

type RequestType uint8

const (
	HTTPRequest      RequestType = 1
	WebsocketRequest RequestType = 2
	SSHRequest       RequestType = 3
)

func readLength(connR *bufio.Reader) (int, error) {
	var nn int

	l := make([]byte, 8)

	for nn < 8 {
		n, err := connR.Read(l[nn:])
		if err != nil {
			return 0, err
		}
		nn += n
	}

	return int(binary.BigEndian.Uint64(l)), nil
}

func readCommand(connR *bufio.Reader) ([]byte, error) {
	l, err := readLength(connR)
	if err != nil {
		return nil, err
	}

	if l == 0 {
		return nil, nil
	}

	b := make([]byte, l)

	var nn int
	for nn < len(b) {
		n, err := connR.Read(b[nn:])
		if err != nil {
			return nil, err
		}
		nn += n
	}

	return b, nil
}

func WriteCommand(connW *bufio.Writer, cmd []byte) error {
	b := utils.BigEndianUint64(uint64(len(cmd)))
	var nn int
	for nn < len(b) {
		n, err := connW.Write(b[nn:])
		if err != nil {
			return err
		}
		nn += n
	}

	nn = 0
	for nn < len(cmd) {
		n, err := connW.Write(cmd[nn:])
		if err != nil {
			return err
		}
		nn += n
	}

	err := connW.Flush()
	if err != nil {
		return err
	}

	return nil
}

func writeSuccess(connW *bufio.Writer, body interface{}) error {
	_, e := connW.Write([]byte{1})
	if e == nil && body != nil {
		var b []byte
		b, e = json.Marshal(body)
		if e == nil {
			_, e = connW.Write(utils.BigEndianUint64(uint64(len(b))))
			if e == nil {
				connW.Write(b)
			}
		}
	} else {
		_, e = connW.Write(utils.BigEndianUint64(0))
	}

	if e == nil {
		e = connW.Flush()
	} else {
		connW.Flush()
	}

	return e
}

func writeSuccessRaw(connW *bufio.Writer, body []byte) error {
	_, e := connW.Write([]byte{1})
	if e == nil && body != nil {
		_, e = connW.Write(utils.BigEndianUint64(uint64(len(body))))
		if e == nil {
			connW.Write(body)
		}
	} else {
		_, e = connW.Write(utils.BigEndianUint64(0))
	}

	if e == nil {
		e = connW.Flush()
	} else {
		connW.Flush()
	}

	return e
}

func writeError(connW *bufio.Writer, err error) error {
	_, e := connW.Write([]byte{0})
	if e == nil {
		er := []byte(err.Error())
		_, e = connW.Write(utils.BigEndianUint64(uint64(len(er))))
		if e == nil {
			_, e = connW.Write(er)
		}
	}

	if e == nil {
		e = connW.Flush()
	} else {
		connW.Flush()
	}

	return e
}

func ReadResponse(connR *bufio.Reader) (success bool, payload []byte, err error) {
	var b byte
	b, err = connR.ReadByte()
	if err != nil {
		return
	}

	if b == 1 {
		success = true
	}

	payload, err = readCommand(connR)
	return
}

func StreamReader(connW *bufio.Writer, r io.Reader) error {
	b := make([]byte, streamChunkSize)

	var n, n2, nn int
	var err, err2 error
	for {
		n, err = r.Read(b)

		if n > 0 {
			_, err2 = connW.Write(utils.BigEndianUint64(uint64(n)))
			if err2 != nil {
				return err2
			}

			nn = 0
			for nn < n {
				n2, err2 = connW.Write(b[nn:n])
				if err2 != nil {
					return err2
				}
				nn += n2
			}
		}

		if err != nil {
			_, err2 = connW.Write(utils.BigEndianUint64(0))
			if err2 != nil {
				return err2
			}

			if err != io.EOF {
				return err
			}

			return connW.Flush()
		}
	}
}

func PipeReadStream(connR *bufio.Reader, w io.Writer) error {
	b := make([]byte, streamChunkSize)

	var l, n, nn int
	var err error

	for {
		l, err = readLength(connR)
		if err != nil {
			return err
		}

		if l > 0 {
			nn = 0
			for nn < l {
				n, err = connR.Read(b[nn:l])
				if err != nil {
					return err
				}
				nn += n
			}

			nn = 0
			for nn < l {
				n, err = w.Write(b[nn:l])
				if err != nil {
					return err
				}
				nn += n
			}

			continue
		}

		return nil
	}
}

func ReadAllStream(connR *bufio.Reader) ([]byte, error) {
	b := new(bytes.Buffer)

	err := PipeReadStream(connR, b)
	if err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

func readStreamChunk(connR *bufio.Reader, b []byte, remaining int) (int, int, error) {
	var n, nn, l int
	var err error

	for {
		if remaining == 0 {
			l, err = readLength(connR)
			if err != nil {
				return nn, remaining, err
			}

			if l == 0 {
				return nn, remaining, io.EOF
			}

			remaining = l
		}

		if remaining < (len(b) - nn) {
			n, err = connR.Read(b[nn : nn+remaining])
		} else {
			n, err = connR.Read(b[nn:])
		}
		nn += n

		remaining -= n
		if err != nil {
			return nn, remaining, err
		}

		if nn == len(b) {
			return nn, remaining, err
		}
	}
}

func StreamReadWebsocket(r *websocket.Conn, connW *bufio.Writer) error {
	defer connW.Flush()

	b := make([]byte, streamChunkSize)

	var l, n, nn int
	var sockMessType int
	var rd io.Reader
	var err, err2 error

	defer func() {
		connW.Write(utils.BigEndianUint64(0))
	}()

	for {
		sockMessType, rd, err = r.NextReader()
		if err != nil {
			return err
		}

		_, err = connW.Write([]byte{uint8(sockMessType)})
		if err != nil {
			return err
		}

		for {
			l, err = rd.Read(b)
			if err != nil && err != io.EOF {
				return err
			}

			if l > 0 {
				_, err2 = connW.Write(utils.BigEndianUint64(uint64(l)))
				if err2 != nil {
					return err2
				}

				nn = 0
				for nn < l {
					n, err2 = connW.Write(b[nn:l])
					if err2 != nil {
						return err2
					}
					nn += n
				}

				err2 = connW.Flush()
				if err2 != nil {
					return err2
				}
			}

			if err == io.EOF {
				_, err2 = connW.Write(utils.BigEndianUint64(0))
				if err2 != nil {
					return err2
				}
				err2 = connW.Flush()
				if err2 != nil {
					return err2
				}
				break
			}
		}
	}
}

// sockMessType is either websocket.TextMessage or websocket.BinaryMessage
func StreamWriteWebsocket(wr *websocket.Conn, connR *bufio.Reader) error {
	b := make([]byte, streamChunkSize)

	var l, n, n2, nn, nnnn int
	var sockMessType byte
	var w io.WriteCloser
	var err error

	for {
		sockMessType, err = connR.ReadByte()
		if err != nil {
			return err
		}
		if sockMessType == 0 {
			return nil
		}

		w, err = wr.NextWriter(int(sockMessType))
		if err != nil {
			return err
		}

		for {
			l, err = readLength(connR)
			if err != nil {
				return err
			}

			if l > 0 {
				nn = 0
				for nn < l {
					n, err = connR.Read(b[:utils.Min(streamChunkSize, l-nn)])
					if err != nil {
						return err
					}
					nn += n

					nnnn = 0
					for nnnn < n {
						n2, err = w.Write(b[nnnn:n])
						if err != nil {
							return err
						}
						nnnn += n2
					}
				}
			} else {
				w.Close()
				break
			}
		}
	}
}

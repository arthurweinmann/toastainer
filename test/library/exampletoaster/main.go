package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync/atomic"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}

var count uint64

func GetExeIDHandler(w http.ResponseWriter, r *http.Request) {
	exeid := os.Getenv("TOASTCLOUD_EXE_ID")

	if websocket.IsWebSocketUpgrade(r) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			panic(err)
		}
		defer c.Close()
		for {
			mt, _, err := c.ReadMessage()
			if err != nil {
				log.Println("websocket read:", err)
				break
			}
			err = c.WriteMessage(mt, []byte(exeid))
			if err != nil {
				log.Println("websocket write:", err)
				break
			}
		}
	} else {
		w.WriteHeader(200)
		w.Write([]byte(exeid))
	}
}

func CountHandler(w http.ResponseWriter, r *http.Request) {
	countvalue := atomic.AddUint64(&count, 1)

	if websocket.IsWebSocketUpgrade(r) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			panic(err)
		}
		defer c.Close()
		for {
			mt, _, err := c.ReadMessage()
			if err != nil {
				log.Println("websocket read:", err)
				break
			}
			err = c.WriteMessage(mt, []byte(strconv.Itoa(int(countvalue))))
			if err != nil {
				log.Println("websocket write:", err)
				break
			}
		}
	} else {
		w.WriteHeader(200)
		w.Write([]byte(strconv.Itoa(int(countvalue))))
	}
}

func EchoHandler(w http.ResponseWriter, r *http.Request) {
	if websocket.IsWebSocketUpgrade(r) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			panic(err)
		}
		defer c.Close()
		for {
			mt, message, err := c.ReadMessage()
			if err != nil {
				log.Println("websocket read:", err)
				break
			}
			err = c.WriteMessage(mt, message)
			if err != nil {
				log.Println("websocket write:", err)
				break
			}
		}
	} else {
		b, err := io.ReadAll(r.Body)
		if err != nil {
			SendError(w, err.Error(), "invalidBody", 400)
			return
		}

		for k, m := range r.Header {
			for i := 0; i < len(m); i++ {
				w.Header().Add(k, m[i])
			}
		}
		w.WriteHeader(200)
		w.Write(b)
	}
}

func main() {
	http.HandleFunc("/echo", EchoHandler)
	http.HandleFunc("/counter", CountHandler)
	http.HandleFunc("/exeid", GetExeIDHandler)

	fmt.Println("starting http server on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

type JSONErr struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Code    string `json:"code,omitempty"`
}

func SendError(w http.ResponseWriter, message, code string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	b, _ := json.Marshal(&JSONErr{
		Success: false,
		Message: message,
		Code:    code,
	})
	w.Write(b)
}

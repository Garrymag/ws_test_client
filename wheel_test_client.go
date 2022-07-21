// Copyright 2015 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

////go:build ignore
//// +build ignore

package main

import (
	"flag"
	"log"
	"math/rand"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

var addr = flag.String("addr", "192.168.8.200:81", "http service address")
var payload []byte

func main() {
	flag.Parse()
	log.SetFlags(0)
	var dialer websocket.Dialer = *websocket.DefaultDialer
	dialer.ReadBufferSize = 4096
	dialer.WriteBufferSize = 65536
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/"}
	log.Printf("connecting to %s", u.String())
	payload = append(payload, 0)
	for num := 1; num < 9588; num++ {
		payload = append(payload, 16 /* byte(num) */)
	}
	c, _, err := dialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			log.Printf("recv: %s", message)
		}
	}()

	ticker := time.NewTicker(time.Millisecond * 50)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case _ = <-ticker.C:
			for num := 31; num < 9588; num++ {
				payload[num] = byte(rand.Intn(128))
			}
			err := c.WriteMessage(websocket.BinaryMessage, payload)
			if err != nil {
				log.Println("write:", err)
				return
			}
		//	}

		case <-interrupt:
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}

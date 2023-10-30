// program that serves a websocket interface for talking
// to a k3 instance on the local machine.
package main

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/gorilla/websocket"
	"tangentcode.com/k3ipc-go/k3ipc"
)

func quitOn(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

func runIpcServer() {
	l, err := net.Listen("tcp", ":1024")
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		return
	}
	defer l.Close()

	fmt.Println("IPC server listening on port 1024")

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			return
		}
		go handleK3Request(conn)
	}
}

func handleK3Request(conn net.Conn) {
	// keep open until the k3 side closes it
	for {
		buf := make([]byte, 2048)
		_, err := conn.Read(buf[:8])
		quitOn(err)
		r := bytes.NewReader(buf)
		h := k3ipc.ParseMessageHeader(r)
		_, err = conn.Read(buf[8 : 8+h.MsgLen])
		quitOn(err)
		msg := k3ipc.Db(buf)
		fmt.Printf("k3: %v\n", msg)
	}
}

func webSocketHandler(w http.ResponseWriter, r *http.Request) {
	up := websocket.Upgrader{}
	ws, err := up.Upgrade(w, r, nil)
	quitOn(err)
	for {
		msgType, msg, err := ws.ReadMessage()
		quitOn(err)
		println("ws: ", string(msg))
		// echo the message back to the socket:
		err = ws.WriteMessage(msgType, msg)
		quitOn(err)
	}
}

func runHttpServer() {
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir("./web")))
	mux.HandleFunc("/ws", webSocketHandler)
	println("http listening on port 3000")
	http.ListenAndServe(":3000", mux)
}

func main() {
	go runIpcServer()
	go runHttpServer()
	select {}
}

// program that serves a websocket interface for talking
// to a k3 instance on the local machine.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/gorilla/websocket"
	"tangentcode.com/k3ipc-go/k3ipc"
)

const K3PORT = 5000

func quitOn(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

func sendToK3(msg string, c chan string) {
	conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%v", K3PORT))
	quitOn(err)
	defer conn.Close()

	println("send to k3: ", msg)

	_, err = conn.Write(k3ipc.K3Msg(msg, k3ipc.GET_MSG))
	quitOn(err)

	// wait for response:
	buf := make([]byte, 2048)
	_, err = conn.Read(buf[:8])
	quitOn(err)
	r := bytes.NewReader(buf)
	h := k3ipc.ParseMessageHeader(r)
	_, err = conn.Read(buf[8 : 8+h.MsgLen])
	quitOn(err)
	m0 := k3ipc.Db(buf)
	fmt.Printf("k3 responded: %v\n", m0)

	m := m0.([]any)
	switch m[0].(int32) {
	case 0:
		c <- m[1].(string)
	case 1:
		c <- fmt.Sprintf("error: %v", m[1].(string))
	default:
		c <- fmt.Sprintf("unknown response code: %v", m)
	}

}

func runIpcServer() {
	l, err := net.Listen("tcp", ":1024")
	quitOn(err)
	defer l.Close()
	fmt.Println("IPC server listening on port 1024")
	for {
		conn, err := l.Accept()
		quitOn(err)
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
		fmt.Printf("k3 connected and sent: %v\n", msg)
	}
}

func webSocketHandler(w http.ResponseWriter, r *http.Request) {
	up := websocket.Upgrader{}
	ws, err := up.Upgrade(w, r, nil)
	quitOn(err)
	for {
		msgType, jsMsg, err := ws.ReadMessage()
		quitOn(err)

		var msg [3]any
		json.Unmarshal(jsMsg, &msg)
		fmt.Printf("[wsmsg id: %v type: %v | %v]\n", msg[0], msg[1], msg[2])

		switch byte(msg[1].(float64)) {
		case 0:
			c := make(chan string)
			go sendToK3(msg[2].(string), c)
			v := <-c
			j, _ := json.Marshal([]any{msg[0], 0, v})
			ws.WriteMessage(msgType, []byte(j))
		case 1:
			println("todo: magic async messages?")
		default:
			println("unknown msg type: ", msg[1])
		}
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

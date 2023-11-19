package main

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var connections = make(map[string]bool)
var generatedNumbers = sync.Map{}

func main() {
	router()
	http.ListenAndServe(":8080", nil)
}

func router() {
	http.HandleFunc("/ws", wsHandler)
}

func wsHandler(w http.ResponseWriter, r *http.Request) {

	ws, err := upgrader.Upgrade(w, r, nil)

	ipAddress := getIpAddress(r)

	if ok := connections[ipAddress]; ok {
		fmt.Println("Connection already established")
		rejectWsConnection(ws)
		return

	}

	connections[ipAddress] = true

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Client Connected", ipAddress)

	go handleWsMessages(ws, ipAddress)
}

func rejectWsConnection(conn *websocket.Conn) {
	conn.WriteMessage(websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure,
			"Ip address is already connected"))
	conn.Close()
}

func handleWsMessages(conn *websocket.Conn, ipAddress string) {
	for {
		_, _, err := conn.ReadMessage()

		if err != nil {
			if websocket.IsUnexpectedCloseError(
				err,
				websocket.CloseGoingAway,
				websocket.CloseAbnormalClosure) {
				fmt.Println("Client disconnected", ipAddress)
				delete(connections, ipAddress)
				return
			}
		}

		randomBigInt := getRandomBigInt()
		m := msg{
			Number: randomBigInt.String(),
		}

		fmt.Println("Generated number: ", randomBigInt)

		if err = conn.WriteJSON(m); err != nil {
			fmt.Println(err)
		}
	}
}

func getRandomBigInt() *big.Int {
	for {
		max := new(big.Int)
		max.Exp(big.NewInt(2), big.NewInt(256), nil).Sub(max, big.NewInt(1))
		n, _ := rand.Int(rand.Reader, max)

		_, ok := generatedNumbers.Load(n.String())

		if !ok {
			generatedNumbers.Store(n.String(), true)
		} else {
			fmt.Println("Already have", n)
		}

		return n
	}

}

func getIpAddress(r *http.Request) string {
	IPAddress := r.Header.Get("X-Real-Ip")
	if IPAddress == "" {
		IPAddress = r.Header.Get("X-Forwarded-For")
	}
	if IPAddress == "" {
		IPAddress = r.RemoteAddr
	}

	return strings.Split(IPAddress, ":")[0]
}

type msg struct {
	Number string
}

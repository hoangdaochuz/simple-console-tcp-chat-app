package main

import (
	"bufio"
	"fmt"
	"net"
	"sync"
)

var (
	clients   = make(map[net.Conn]string) // map to store client connections and their usernames
	broadcast = make(chan string)         // channel to broadcast messages to all clients
	mutex     = &sync.Mutex{}             // mutex to protect access to the clients map
)

func handleClientConnection(conn net.Conn) {
	defer func() {
		mutex.Lock()
		delete(clients, conn)
		mutex.Unlock()
		conn.Close()
	}()
	reader := bufio.NewReader(conn)
	_, err := conn.Write([]byte("Enter your username: "))
	if err != nil {
		return
	}
	username, err := reader.ReadString('\n')
	if err != nil {
		return
	}

	mutex.Lock()
	clients[conn] = username
	mutex.Unlock()

	// broadcast the new user to all clients
	broadcast <- fmt.Sprintf("%s has joined the chat\n", username)

	// start get messages from the client and broadcast them to all other clients
	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			broadcast <- fmt.Sprintf("%s has left the chat\n", username)
			return
		}
		broadcast <- fmt.Sprintf("%s: %s", username, msg)
	}
}

func handleBroadcastMsg() {
	defer close(broadcast)

	for {
		msg := <-broadcast
		mutex.Lock()
		for client := range clients {
			_, err := client.Write([]byte(msg))
			if err != nil {
				delete(clients, client)
				client.Close()
				fmt.Println("Error writing to client:", err)
			}
		}
		mutex.Unlock()
	}

}

func main() {
	server, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}
	fmt.Println("Chat server started on :8080")
	defer server.Close()

	go handleBroadcastMsg()

	for {
		conn, err := server.Accept()
		if err != nil {
			panic(err)
		}
		// start a goroutine to handle the connection
		go handleClientConnection(conn)
	}
}

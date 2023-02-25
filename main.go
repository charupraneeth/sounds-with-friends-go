// websockets.go
package main

import (
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	// A mutex to synchronize access to connections.
	// This is needed because the main goroutine and
	// the goroutines that handle connections run in
	// parallel.
	connectionsLock sync.Mutex

	// A slice to store all connections.
	connections []*websocket.Conn
)

func main() {

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default to port 8080 if PORT is not set.
	}

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, _ := upgrader.Upgrade(w, r, nil) // error ignored for sake of simplicity

		// Add the new connection to the list.
		connectionsLock.Lock()
		connections = append(connections, conn)
		connectionsLock.Unlock()

		for {
			// Read message from browser
			msgType, msg, err := conn.ReadMessage()
			if err != nil {
				// Remove the connection from the list if there was an error.
				connectionsLock.Lock()
				for i, c := range connections {
					if c == conn {
						connections = append(connections[:i], connections[i+1:]...)
						break
					}
				}
				connectionsLock.Unlock()

				return
			}

			// Print the message to the console
			fmt.Printf("%s sent: %s\n", conn.RemoteAddr(), string(msg))

			// Write message back to browser
			connectionsLock.Lock()
			for _, c := range connections {
				// Skip the sender.
				if c == conn {
					continue
				}

				if err = c.WriteMessage(msgType, msg); err != nil {
					// Remove the connection from the list if there was an error.
					for i, cc := range connections {
						if cc == c {
							connections = append(connections[:i], connections[i+1:]...)
							break
						}
					}
				}
			}
			connectionsLock.Unlock()
		}
	})

	fs := http.FileServer(http.Dir("static/"))
	http.Handle("/", http.StripPrefix("/", fs))

	fmt.Printf("Listening on port %s...\n", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		panic(err)
	}
}

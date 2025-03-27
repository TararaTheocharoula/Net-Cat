// client.go
package server

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

// Client represents a connected user
// It holds the network connection, chosen name, and current chat room
type Client struct {
	conn net.Conn
	name string
	room string
}

// acceptConnections waits for new incoming TCP connections
// If less than 10 users are connected, it handles them in separate goroutines
func (s *Server) acceptConnections() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			continue // Skip failed connections
		}

		s.mutex.Lock()
		// Limit to 10 concurrent clients
		if len(s.clients) >= 10 {
			conn.Write([]byte("Chat room full. Try again later.\n"))
			conn.Close()
			s.mutex.Unlock()
			continue
		}
		s.mutex.Unlock()

		// Handle the client in a new goroutine (thread-like)
		go s.handleClient(conn)
	}
}

// handleClient manages a single client's lifecycle
// Including welcome, username input, joining a room, and message listening
func (s *Server) handleClient(conn net.Conn) {
	defer conn.Close()            // Ensure connection closes when done
	client := &Client{conn: conn} // Create new client instance

	// Send welcome message and ASCII logo
	conn.Write([]byte("\nWelcome to TCP-Chat!\n"))
	s.printLogo(conn)

	// Loop until the user enters a valid, unique name
	for {
		conn.Write([]byte("[ENTER YOUR NAME]: "))
		scanner := bufio.NewScanner(conn)
		if !scanner.Scan() {
			return // Client disconnected during input
		}

		name := strings.TrimSpace(scanner.Text()) // Clean name input
		s.mutex.Lock()
		// Validate name format and check uniqueness
		if s.validateName(name) && !s.usernames[name] {
			client.name = name
			s.usernames[name] = true // Reserve the name
			s.mutex.Unlock()
			break
		}
		s.mutex.Unlock()
		// Notify client that name is not valid
		conn.Write([]byte("Invalid or already taken name. Use a unique name (3-20 alphanumeric characters).\n"))
	}

	s.mutex.Lock()
	s.clients[conn] = client      // Add client to global map
	s.joinRoom(client, "general") // Put them in default "general" room
	s.mutex.Unlock()

	s.handleMessages(client) // Begin chat loop for this client

	// When client leaves, clean up
	s.mutex.Lock()
	delete(s.clients, conn)
	delete(s.usernames, client.name)
	leaveMsg := fmt.Sprintf("\n%s has left the chat...\n", client.name)
	s.broadcast(client.room, leaveMsg, nil) // Notify others
	s.logMessage(leaveMsg)                  // Save to server log
	s.mutex.Unlock()
}

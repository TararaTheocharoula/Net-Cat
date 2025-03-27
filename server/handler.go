// handler.go
package server

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"
)

// handleMessages continuously listens for messages from a connected client
// It handles commands (starting with '/') and normal chat messages
func (s *Server) handleMessages(client *Client) {

	scanner := bufio.NewScanner(client.conn)
	for scanner.Scan() {
		msg := strings.TrimSpace(scanner.Text()) // Read the message
		if msg == "" {
			continue // Ignore empty messages
		}

		// If message starts with '/', treat as a command
		if strings.HasPrefix(msg, "/") {
			s.handleCommand(client, msg)
			continue
		}

		// Otherwise, broadcast it to the room
		s.mutex.Lock()
		timestamp := time.Now().Format("2006-01-02 15:04:05")

		// Assign color based on username length (for fun)
		userColor := Green
		if len(client.name)%2 == 0 {
			userColor = Blue
		}

		// Format message with timestamp, name, and content
		formattedMsg := fmt.Sprintf("[%s]%s: %s\n", colorize(timestamp, White), colorize(client.name, userColor), msg)

		// Save message in room history and server log
		s.history[client.room] = append(s.history[client.room], formattedMsg)
		s.logMessage(formattedMsg)

		// Send message to all users in the room except the sender
		s.broadcast(client.room, formattedMsg, client.conn)
		s.mutex.Unlock()
	}
}

// broadcast sends a message to all users in a chat room
// except for the sender (if provided)
func (s *Server) broadcast(room, msg string, sender net.Conn) {
	coloredMsg := colorize(msg, Yellow) // Color all messages yellow for consistency
	for conn := range s.rooms[room] {
		if conn != sender {
			conn.Write([]byte(coloredMsg))
		}
	}
}

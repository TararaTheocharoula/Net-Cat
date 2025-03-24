// commands.go
package server

import (
	"fmt"
	"net"
	"strings"
)

// handleCommand processes special slash (/) commands from a client
// Supported commands: /exit, /help, /join
func (s *Server) handleCommand(client *Client, command string) {
	// Split the command into parts: [command, argument]
	parts := strings.SplitN(command, " ", 2)
	switch parts[0] {
	case "/exit":
		// Exit the chat - send goodbye message and close connection
		client.conn.Write([]byte("Goodbye!\n"))
		client.conn.Close()
	case "/help":
		// Show list of available commands
		client.conn.Write([]byte("Available commands:\n/exit - Leave the chat\n/join [room] - Switch chat rooms\n"))
	case "/join":
		// Join another room
		if len(parts) < 2 {
			client.conn.Write([]byte("Usage: /join [room]\n"))
			return
		}
		newRoom := strings.TrimSpace(parts[1])
		s.joinRoom(client, newRoom)
	}
}

// joinRoom moves a client from their current room to a new one
// Notifies both rooms and loads chat history
func (s *Server) joinRoom(client *Client, room string) {
	// Notify the current room that the user left (if already in one)
	if client.room != "" {
		leaveMsg := fmt.Sprintf("\n%s has left the room %s...\n", client.name, client.room)
		s.broadcast(client.room, leaveMsg, nil)
	}

	// Update client's room
	client.room = room

	// Create the room map if it doesn't exist
	if s.rooms[room] == nil {
		s.rooms[room] = make(map[net.Conn]*Client)
	}

	// Add client to the room
	s.rooms[room][client.conn] = client

	// Notify the room that the user joined
	joinMsg := fmt.Sprintf("\n%s has joined the room %s...\n", client.name, client.room)
	s.broadcast(client.room, joinMsg, nil)

	// Show previous messages from the room
	s.sendHistory(client)
}

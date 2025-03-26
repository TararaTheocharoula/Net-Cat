// commands.go
package server

import (
	"fmt"
	"net"
	"strings"
)

// handleCommand processes special slash (/) commands from a client
// Supported commands: /exit, /help, /join, /name
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
		helpText := `
Available commands:
/exit         - Leave the chat
/join [room]  - Switch to another chat room
/name [new]   - Change your display name
/help         - Show this help message
`
		client.conn.Write([]byte(helpText))

	case "/join":
		// Join another room
		if len(parts) < 2 {
			client.conn.Write([]byte("Usage: /join [room]\n"))
			return
		}
		newRoom := strings.TrimSpace(parts[1])
		s.joinRoom(client, newRoom)

	case "/name":
		// Change username
		if len(parts) < 2 {
			client.conn.Write([]byte("Usage: /name [new_name]\n"))
			return
		}
		newName := strings.TrimSpace(parts[1])

		s.mutex.Lock()
		if !s.validateName(newName) || s.usernames[newName] {
			s.mutex.Unlock()
			client.conn.Write([]byte("Invalid or already taken name. Use 3–20 alphanumeric characters.\n"))
			return
		}

		oldName := client.name
		delete(s.usernames, oldName)
		client.name = newName
		s.usernames[newName] = true
		s.mutex.Unlock()

		// Notify the room of the name change
		notification := fmt.Sprintf("\n%s is now known as %s\n", oldName, newName)
		s.broadcast(client.room, notification, nil)
		s.logMessage(notification)

	default:
		client.conn.Write([]byte("Unknown command. Type /help for a list of commands.\n"))
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

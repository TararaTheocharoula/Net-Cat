package server

import (
	"fmt"
	"net"
	"strings"
)

func (s *Server) handleCommand(client *Client, command string) {
	parts := strings.SplitN(command, " ", 2)
	switch parts[0] {
	case "/exit":
		client.conn.Write([]byte("Goodbye!\n"))
		client.conn.Close()
	case "/help":
		client.conn.Write([]byte("Available commands:\n/exit - Leave the chat\n/join [room] - Switch chat rooms\n"))
	case "/join":
		if len(parts) < 2 {
			client.conn.Write([]byte("Usage: /join [room]\n"))
			return
		}
		newRoom := strings.TrimSpace(parts[1])
		s.joinRoom(client, newRoom)
	}
}

func (s *Server) joinRoom(client *Client, room string) {
	if client.room != "" {
		leaveMsg := fmt.Sprintf("\n%s has left the room %s...\n", client.name, client.room)
		s.broadcast(client.room, leaveMsg, nil)
	}

	client.room = room
	if s.rooms[room] == nil {
		s.rooms[room] = make(map[net.Conn]*Client)
	}
	s.rooms[room][client.conn] = client

	joinMsg := fmt.Sprintf("\n%s has joined the room %s...\n", client.name, client.room)
	s.broadcast(client.room, joinMsg, nil)
	s.sendHistory(client)
}

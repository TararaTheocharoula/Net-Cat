package server

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"
)

func (s *Server) handleMessages(client *Client) {
	s.sendHistory(client)

	scanner := bufio.NewScanner(client.conn)
	for scanner.Scan() {
		msg := strings.TrimSpace(scanner.Text())
		if msg == "" {
			continue
		}

		if strings.HasPrefix(msg, "/") {
			s.handleCommand(client, msg)
			continue
		}

		s.mutex.Lock()
		timestamp := time.Now().Format("2006-01-02 15:04:05")

		// Different colors per user
		userColor := Green
		if len(client.name)%2 == 0 {
			userColor = Blue
		}

		formattedMsg := fmt.Sprintf("[%s]%s: %s\n", colorize(timestamp, White), colorize(client.name, userColor), msg)
		s.history[client.room] = append(s.history[client.room], formattedMsg)
		s.logMessage(formattedMsg)
		s.broadcast(client.room, formattedMsg, client.conn)
		s.mutex.Unlock()
	}
}

func (s *Server) broadcast(room, msg string, sender net.Conn) {
	coloredMsg := colorize(msg, Yellow) // Messages in Yellow
	for conn := range s.rooms[room] {
		if conn != sender {
			conn.Write([]byte(coloredMsg))
		}
	}
}

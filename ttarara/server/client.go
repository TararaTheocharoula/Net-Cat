package server

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

type Client struct {
	conn net.Conn
	name string
	room string
}

func (s *Server) acceptConnections() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			continue
		}

		s.mutex.Lock()
		if len(s.clients) >= 10 {
			conn.Write([]byte("Chat room full. Try again later.\n"))
			conn.Close()
			s.mutex.Unlock()
			continue
		}
		s.mutex.Unlock()

		go s.handleClient(conn)
	}
}

func (s *Server) handleClient(conn net.Conn) {
	defer conn.Close()
	client := &Client{conn: conn}

	conn.Write([]byte("\nWelcome to TCP-Chat!\n"))
	s.printLogo(conn)

	for {
		conn.Write([]byte("[ENTER YOUR NAME]: "))
		scanner := bufio.NewScanner(conn)
		if !scanner.Scan() {
			return
		}

		name := strings.TrimSpace(scanner.Text())
		s.mutex.Lock()
		if s.validateName(name) && !s.usernames[name] {
			client.name = name
			s.usernames[name] = true
			s.mutex.Unlock()
			break
		}
		s.mutex.Unlock()
		conn.Write([]byte("Invalid or already taken name. Use a unique name (3-20 alphanumeric characters).\n"))
	}

	s.mutex.Lock()
	s.clients[conn] = client
	s.joinRoom(client, "general")
	s.mutex.Unlock()

	s.handleMessages(client)

	s.mutex.Lock()
	delete(s.clients, conn)
	delete(s.usernames, client.name)
	leaveMsg := fmt.Sprintf("\n%s has left the chat...\n", client.name)
	s.broadcast(client.room, leaveMsg, nil)
	s.logMessage(leaveMsg)
	s.mutex.Unlock()
}

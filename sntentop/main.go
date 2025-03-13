package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

type Client struct {
	conn net.Conn
	name string
	room string
}

type Server struct {
	port      string
	listener  net.Listener
	clients   map[net.Conn]*Client
	rooms     map[string]map[net.Conn]*Client
	mutex     sync.Mutex
	history   map[string][]string
	logFile   *os.File
	usernames map[string]bool // New field to track active usernames
}

func main() {
	if len(os.Args) > 2 {
		fmt.Println("[USAGE]: ./TCPChat $port")
		return
	}

	port := "8989"
	if len(os.Args) == 2 {
		port = os.Args[1]
	}

	server := NewServer(port)
	defer server.Close()

	if err := server.Start(); err != nil {
		log.Fatal("Server error:", err)
	}
}

func NewServer(port string) *Server {
	return &Server{
		port:      ":" + port,
		clients:   make(map[net.Conn]*Client),
		rooms:     make(map[string]map[net.Conn]*Client),
		history:   make(map[string][]string),
		usernames: make(map[string]bool), // Initialize usernames map
	}
}

func (s *Server) Start() error {
	var err error
	s.logFile, err = os.OpenFile("chat.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	s.listener, err = net.Listen("tcp", s.port)
	if err != nil {
		return err
	}

	fmt.Printf("Listening on port %s\n", s.port)

	go s.handleSignals()
	go s.acceptConnections()

	select {} // Block main goroutine
}

func (s *Server) logMessage(msg string) {
	logEntry, _ := json.Marshal(map[string]string{"timestamp": time.Now().Format("2006-01-02 15:04:05"), "message": msg})
	s.logFile.WriteString(string(logEntry) + "\n")
	fmt.Println(msg) // Print message to server terminal
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

	// Get a unique username
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
			s.usernames[name] = true // Mark username as taken
			s.mutex.Unlock()
			break
		}
		s.mutex.Unlock()
		conn.Write([]byte("Invalid or already taken name. Use a unique name (3-20 alphanumeric characters).\n"))
	}

	s.mutex.Lock()
	s.clients[conn] = client
	s.joinRoom(client, "general") // Default room
	s.mutex.Unlock()

	s.handleMessages(client)

	// Cleanup on disconnect
	s.mutex.Lock()
	delete(s.clients, conn)
	delete(s.usernames, client.name) // Free up username
	leaveMsg := fmt.Sprintf("\n%s has left the chat...\n", client.name)
	s.broadcast(client.room, leaveMsg, nil)
	s.logMessage(leaveMsg)
	s.mutex.Unlock()

	s.handleMessages(client)
}

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
		formattedMsg := fmt.Sprintf("[%s][%s]: %s\n", timestamp, client.name, msg)
		s.history[client.room] = append(s.history[client.room], formattedMsg)
		s.logMessage(formattedMsg)
		s.broadcast(client.room, formattedMsg, client.conn)
		s.mutex.Unlock()
	}

	s.mutex.Lock()
	delete(s.clients, client.conn)
	leaveMsg := fmt.Sprintf("\n%s has left the chat...\n", client.name)
	s.broadcast(client.room, leaveMsg, nil)
	s.logMessage(leaveMsg)
	s.mutex.Unlock()
}

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

func (s *Server) sendHistory(client *Client) {
	if len(s.history[client.room]) > 0 {
		client.conn.Write([]byte("\nChat history:\n"))
		for _, msg := range s.history[client.room] {
			client.conn.Write([]byte(msg))
		}
	}
}

func (s *Server) broadcast(room, msg string, sender net.Conn) {
	for conn := range s.rooms[room] {
		if conn != sender {
			conn.Write([]byte(msg))
		}
	}
}

func (s *Server) Close() {
	if s.listener != nil {
		s.listener.Close()
	}
	if s.logFile != nil {
		s.logFile.Close()
	}
}

func (s *Server) handleSignals() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	fmt.Println("\nShutting down server...")
	s.Close()
	os.Exit(0)
}

func (s *Server) printLogo(conn net.Conn) {
	logo := []string{
		"         _nnnn_        ",
		"        dGGGGMMb       ",
		"       @p~qp~~qMb      ",
		"       M|@||@) M|      ",
		"       @,----.JM|      ",
		"      JS^\\__/  qKL     ",
		"     dZP        qKRb   ",
		"    dZP          qKKb  ",
		"   fZP            SMMb ",
		"   HZM            MMMM ",
		"   FqM            MMMM ",
		" __| \".        |\\dS\"qML ",
		" |    .       | ' \\Zq ",
		"_)      \\.___.,|     .' ",
		"\\____   )MMMMMP|   .'   ",
		"     -'       --'     ",
	}
	conn.Write([]byte("\n"))
	for _, line := range logo {
		conn.Write([]byte(line + "\n"))
	}
	conn.Write([]byte("\n"))
}

func (s *Server) validateName(name string) bool {
	if len(name) < 3 || len(name) > 20 {
		return false
	}
	for _, c := range name {
		if !((c >= 'a' && c <= 'z') ||
			(c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9')) {
			return false
		}
	}
	return true
}

package main

import (
	"bufio"
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
}

type Server struct {
	port     string
	listener net.Listener
	clients  map[net.Conn]*Client
	mutex    sync.Mutex
	history  []string
	logFile  *os.File
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
		port:    ":" + port,
		clients: make(map[net.Conn]*Client),
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

func (s *Server) acceptConnections() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			return
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

	// Send welcome message and logo
	conn.Write([]byte("\nWelcome to TCP-Chat!\n"))
	s.printLogo(conn)

	// Get valid username
	for {
		conn.Write([]byte("[ENTER YOUR NAME]: "))
		scanner := bufio.NewScanner(conn)
		if !scanner.Scan() {
			return
		}

		name := strings.TrimSpace(scanner.Text())
		if s.validateName(name) {
			client.name = name
			break
		}
		conn.Write([]byte("Invalid name. Use 3-20 alphanumeric characters.\n"))
	}

	// Register client
	s.mutex.Lock()
	s.clients[conn] = client
	joinMsg := fmt.Sprintf("\n%s has joined our chat...\n", client.name)
	s.broadcast(joinMsg, conn)
	s.logMessage(joinMsg)
	s.sendHistory(conn)
	conn.Write([]byte(fmt.Sprintf("[%s] Enter message: ", time.Now().Format("15:04:05"))))
	s.mutex.Unlock()

	// Handle messages
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		msg := strings.TrimSpace(scanner.Text())
		if msg == "" {
			conn.Write([]byte(fmt.Sprintf("[%s] Enter message: ", time.Now().Format("15:04:05"))))
			continue
		}

		timestamp := time.Now().Format("2006-01-02 15:04:05")
		fullMsg := fmt.Sprintf("\n[%s][%s]: %s\n", timestamp, client.name, msg)

		s.mutex.Lock()
		s.history = append(s.history, fullMsg)
		s.logMessage(fullMsg)
		if len(s.history) > 100 {
			s.history = s.history[1:]
		}
		s.broadcast(fullMsg, conn)
		conn.Write([]byte(fmt.Sprintf("[%s] Enter message: ", time.Now().Format("15:04:05"))))
		s.mutex.Unlock()
	}

	// Handle disconnect
	s.mutex.Lock()
	delete(s.clients, conn)
	leaveMsg := fmt.Sprintf("\n%s has left our chat...\n", client.name)
	s.broadcast(leaveMsg, nil)
	s.logMessage(leaveMsg)
	s.mutex.Unlock()
}

func (s *Server) broadcast(msg string, sender net.Conn) {
	fmt.Print(msg) // Server console log
	for conn := range s.clients {
		if conn != sender {
			conn.Write([]byte(msg))
		}
	}
}

func (s *Server) sendHistory(conn net.Conn) {
	if len(s.history) > 0 {
		conn.Write([]byte("\nChat history:\n"))
		for _, msg := range s.history {
			conn.Write([]byte(msg))
		}
	}
	conn.Write([]byte(fmt.Sprintf("[%s] Enter message: ", time.Now().Format("15:04:05"))))
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

func (s *Server) handleSignals() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	fmt.Println("\nShutting down server...")
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for conn := range s.clients {
		conn.Close()
	}
	s.listener.Close()
	s.logFile.Close()
	os.Exit(0)
}

func (s *Server) Close() {
	s.listener.Close()
	s.logFile.Close()
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
		" |    `.       | `' \\Zq ",
		"_)      \\.___.,|     .' ",
		"\\____   )MMMMMP|   .'   ",
		"     `-'       `--'     ",
	}

	conn.Write([]byte("\n"))
	for _, line := range logo {
		conn.Write([]byte(line + "\n"))
	}
	conn.Write([]byte("\n"))
}

func (s *Server) logMessage(msg string) {
	logEntry := fmt.Sprintf("[%s] %s", time.Now().Format("2006-01-02 15:04:05"), strings.TrimSpace(msg))
	s.logFile.WriteString(logEntry + "\n")
}

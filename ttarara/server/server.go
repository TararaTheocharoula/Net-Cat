package server

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type Server struct {
	port      string
	listener  net.Listener
	clients   map[net.Conn]*Client
	rooms     map[string]map[net.Conn]*Client
	mutex     sync.Mutex
	history   map[string][]string
	logFile   *os.File
	usernames map[string]bool
}

func NewServer(port string) *Server {
	return &Server{
		port:      ":" + port,
		clients:   make(map[net.Conn]*Client),
		rooms:     make(map[string]map[net.Conn]*Client),
		history:   make(map[string][]string),
		usernames: make(map[string]bool),
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

func (s *Server) logMessage(msg string) {
	logEntry, _ := json.Marshal(map[string]string{
		"timestamp": time.Now().Format("2006-01-02 15:04:05"),
		"message":   msg,
	})
	s.logFile.WriteString(string(logEntry) + "\n")

	// Colorize message for the server terminal
	fmt.Println(colorize(msg, Cyan))
}

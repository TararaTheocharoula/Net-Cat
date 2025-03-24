// server.go
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

// Server represents the main chat server
// It manages clients, rooms, logs, and handles all network operations
type Server struct {
	port      string                          // Port to listen on (e.g., :8989)
	listener  net.Listener                    // TCP listener
	clients   map[net.Conn]*Client            // All connected clients
	rooms     map[string]map[net.Conn]*Client // Rooms and their participants
	mutex     sync.Mutex                      // Synchronization lock
	history   map[string][]string             // Message history per room
	logFile   *os.File                        // File for logging chat messages
	usernames map[string]bool                 // Map of taken usernames
}

// NewServer initializes a new Server instance with maps prepared
func NewServer(port string) *Server {
	return &Server{
		port:      ":" + port,
		clients:   make(map[net.Conn]*Client),
		rooms:     make(map[string]map[net.Conn]*Client),
		history:   make(map[string][]string),
		usernames: make(map[string]bool),
	}
}

// Start sets up the TCP listener, log file, and begins accepting connections
func (s *Server) Start() error {
	var err error

	// Open or create the chat log file
	s.logFile, err = os.OpenFile("chat.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	// Start listening on the specified port
	s.listener, err = net.Listen("tcp", s.port)
	if err != nil {
		return err
	}

	fmt.Printf("Listening on port %s\n", s.port)

	// Start background routines
	go s.handleSignals()     // Handles CTRL+C or termination
	go s.acceptConnections() // Waits for client connections

	select {} // Block forever (until server is shut down)
}

// Close shuts down the listener and closes the log file
func (s *Server) Close() {
	if s.listener != nil {
		s.listener.Close()
	}
	if s.logFile != nil {
		s.logFile.Close()
	}
}

// handleSignals waits for OS termination signals and gracefully shuts down
func (s *Server) handleSignals() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM) // Catch CTRL+C or kill
	<-sigCh                                               // Block until a signal is received

	fmt.Println("\nShutting down server...")
	s.Close()
	os.Exit(0)
}

// logMessage saves a formatted message to the log file with a timestamp
// Also prints it to the terminal in color
func (s *Server) logMessage(msg string) {
	logEntry, _ := json.Marshal(map[string]string{
		"timestamp": time.Now().Format("2006-01-02 15:04:05"),
		"message":   msg,
	})
	s.logFile.WriteString(string(logEntry) + "\n")

	// Colorize and show message in terminal
	fmt.Println(colorize(msg, Cyan))
}

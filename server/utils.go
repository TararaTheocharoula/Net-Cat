// utils.go
package server

import (
	"fmt"
	"net"
)

// ANSI Color Codes for terminal styling
const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Purple = "\033[35m"
	Cyan   = "\033[36m"
	White  = "\033[37m"
	Bold   = "\033[1m"
)

// colorize wraps text with ANSI color codes
// Used to style messages with color
func colorize(text, color string) string {
	return fmt.Sprintf("%s%s%s", color, text, Reset)
}

// validateName ensures a username is between 3-20 characters
// and only contains alphanumeric characters
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

// sendHistory sends stored messages from the current room to the client
// This gives new users a sense of chat context
func (s *Server) sendHistory(client *Client) {
	if len(s.history[client.room]) > 0 {
		client.conn.Write([]byte("\nChat history:\n"))
		for _, msg := range s.history[client.room] {
			client.conn.Write([]byte(msg))
		}
	}
}

// printLogo sends a multi-line ASCII logo to the client with colors
// Fun visual shown upon connecting
func (s *Server) printLogo(conn net.Conn) {
	logo := []string{
		colorize("         _nnnn_        ", Cyan),
		colorize("        dGGGGMMb       ", Green),
		colorize("       @p~qp~~qMb      ", Yellow),
		colorize("       M|@||@) M|      ", Blue),
		colorize("       @,----.JM|      ", Purple),
		colorize("      JS^\\__/  qKL     ", Red),
		colorize("     dZP        qKRb   ", Cyan),
		colorize("    dZP          qKKb  ", Green),
		colorize("   fZP            SMMb ", Yellow),
		colorize("   HZM            MMMM ", Blue),
		colorize("   FqM            MMMM ", Purple),
		colorize("__| \".        |\\dS\"qML ", Red),
		colorize("|    .       | ' \\Zq  ", Cyan),
		colorize("_)      \\.___.,|     .' ", Green),
		colorize("\\____   )MMMMMP|   .'   ", Yellow),
		colorize("     -'       --'     ", Red),
	}

	conn.Write([]byte("\n"))
	for _, line := range logo {
		conn.Write([]byte(line + "\n"))
	}
	conn.Write([]byte("\n"))
}

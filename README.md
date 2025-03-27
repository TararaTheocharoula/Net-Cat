# ⚡ TCPChat
A NetCat-inspired, multi-client chat server built in Go 🧑‍💻

---

## 🎯 Project Objective

Recreate the functionality of the `NetCat` (`nc`) tool using Go in a **Client–Server TCP architecture**.  
This application supports a real-time group chat system over TCP connections, mimicking the core behaviors of `nc` while adding features like user validation, room switching, message history, and more.

---

## 🛠️ Features

✅ Multi-client TCP connection (1 server : up to 10 clients)  
✅ Real-time message broadcasting  
✅ Unique usernames (validated)  
✅ Custom slash commands: `/name`, `/join`, `/exit`, `/help`  
✅ Message timestamps and formatting  
✅ Room-based message history  
✅ Join/leave notifications  
✅ Chat logs saved to `chat.log`  
✅ Graceful shutdown with `CTRL+C`  
✅ ANSI-colored output and ASCII welcome logo  

---

## 📦 Installation & Usage

### ⚙️ Requirements
- Go 1.18 or later
- Terminal `nc` (Netcat)

### 🚀 Run the Server

**Default Port (8989):**
```bash
go run .
```

**Custom Port:**
```bash
go run . 2525
```

**Invalid Usage:**
```bash
go run . 2525 localhost
# Output:
[USAGE]: ./TCPChat $port
```

---

## 💻 Connecting as a Client

Use Netcat to connect:
```bash
nc localhost 8989
```

You'll see:
```
Welcome to TCP-Chat!
         _nnnn_
        dGGGGMMb
       @p~qp~~qMb
       M|@||@) M|
       @,----.JM|
      JS^\__/  qKL
     dZP        qKRb
    dZP          qKKb
   fZP            SMMb
   HZM            MMMM
   FqM            MMMM
 __| ".        |\dS"qML
 |    `.       | `' \Zq
_)      \.___.,|     .'
\____   )MMMMMP|   .'
     `-'       `--'
[ENTER YOUR NAME]:
```

---

## 💬 Supported Commands

| Command            | Description                                |
|--------------------|--------------------------------------------|
| `/name [new_name]` | Change your username                       |
| `/join [room]`     | Join or create a chat room                 |
| `/exit`            | Leave the chat gracefully                  |
| `/help`            | Display the help menu                      |

---

## ✅ Testing Scenarios

| Scenario                                                                      | Result |
|-------------------------------------------------------------------------------|--------|
| Run `./TCPChat` → Listens on port 8989                                       | ✅     |
| Run `./TCPChat 2525 localhost` → Shows `[USAGE]: ./TCPChat $port`            | ✅     |
| Run `./TCPChat 2525` → Listens on port 2525                                  | ✅     |
| Connect 2 clients via `nc`                                                   | ✅     |
| All clients receive Linux logo and name prompt                               | ✅     |
| All clients receive join notification when a new user connects               | ✅     |
| All clients receive messages from other clients                              | ✅     |
| New clients receive message history upon join                                | ✅     |
| Remaining clients stay connected if one disconnects                          | ✅     |
| Remaining clients are notified if someone leaves                             | ✅     |
| Messages include `[timestamp][name]: message` format                         | ✅     |
| Empty messages are ignored                                                   | ✅     |
| Clients can change name and switch rooms                                     | ✅     |
| Chat logs saved to `chat.log`                                                | ✅     |
| Graceful shutdown with `CTRL+C`                                              | ✅     |
| Only allowed packages used                                                   | ✅     |

---

## 🐧 Multi-PC Usage (Linux)

Each client can connect to the TCPChat server from a **different computer on the same network** using Linux.

###  How to test:

1. **Start the server** on one machine:
```bash
$ go run .
Listening on the port :8989
```

2. **Find the server local IP**:
```bash
$ hostname -I
# Example output: 192.168.1.10
```

3. **On other machines**, use Netcat to connect using the IP:
```bash
$ nc 192.168.1.10 8989
```

4. Each client will receive the ASCII art and be prompted for a username.
5. All connected clients will be able to send/receive messages in real-time across multiple Linux machines.

> Ensure all devices are on the same local network and the server's firewall allows inbound connections to the chosen port.

---

## 📁 Project Structure

| File           | Description                               |
|----------------|-------------------------------------------|
| `main.go`      | Entry point; initializes server            |
| `server.go`    | Manages connections, logs, shutdown        |
| `client.go`    | Handles new client connections             |
| `commands.go`  | Parses and processes slash commands        |
| `handler.go`   | Manages message input/output               |
| `utils.go`     | Utility functions: validation, logo, color |

---

## 👥 Authors

- **Theocharoula Tarara**  ✨  [ttarara](https://platform.zone01.gr/git/ttarara)
- **Dionysis Pappas**  ✨  [dpappas](https://platform.zone01.gr/git/dpappas)
- **Stefanos Ntentopoulos** ✨  [sntentop](https://platform.zone01.gr/git/sntentop)

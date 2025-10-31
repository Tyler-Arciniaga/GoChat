# Go Chat Project 💬

A **fast, concurrent, TCP-based chat system** built in Go. This project was a personal playground for practicing network programming, and Go concurrency patterns.

---

## Features 🌟

### Real-Time Chat
- Users connect via TCP/IP with **unique, case-insensitive usernames**.
- Broadcast messages to entire rooms or whisper privately to other users with `/whisper`.
- Graceful leave handling via `/leave` or `Ctrl+C`.
- Handles abrupt disconnects safely.

### Room-Based Architecture
- Central **hub** manages multiple rooms and routes messages efficiently.
- Rooms have dedicated channels for broadcasts, whispers, and file transfers.
- Dynamic routing: rooms and hub process metadata and messages cleanly.

### File Transfer
- Send **files of any size** over TCP with chunking and reliable delivery with `/sendfile`.
- **Acknowledgment system** ensures every chunk is received correctly.
- Hub mediates transfers to avoid blocking slow clients.
- Supports simultaneous chat and file streaming without hiccups.

### Networking & Concurrency
- Go routines and channels everywhere: fully concurrent and **non-blocking**.
- Hub handles multiple clients simultaneously with minimal overhead.
- Robust TCP networking: connection handling, read/write streams, length-prefixed messages.

### Client & Server Architecture
- Custom client/server communication.
- JSON-based messages for structured communication.
- Decoupled design: clients send/receive, hub and rooms handle routing.
- Multi-machine support for LAN testing.

---

## What I Learned 💡

- **Concurrency**: Goroutines, channels, safe shutdowns, non-blocking communication.
- **Networking Basics + Beyond**: TCP sockets, length-prefixed protocols, reliable multi-client handling.
- **Data Serialization**: JSON marshaling/unmarshaling for structured messaging.
- **File Streaming**: Chunked byte transfer over TCP with proper acknowledgment handling.
- **Clean Architecture**: Hub-room-client separation for scalability and maintainability.
- **Cross-Machine Networking**: Configuring LAN IP servers, managing firewall rules safely.

---

## Next Steps (Optional, Future Upgrades) 🔮
- Persist chat logs and message history for full-stack functionality.
- Encrypted communication (TLS/SSL) for production-grade security.
- WebSocket or gRPC clients for browser/mobile access.
- Room management features: admins, moderators, permissions.
- Experimentation with different chunk sizes for file transfer.

---

**Summary:**  
Go Chat is my hands-on deep dive into **concurrent backend programming, TCP networking, and Go architecture**.

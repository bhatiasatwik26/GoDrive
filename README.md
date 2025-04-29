# 🚀 GoDrive: A Distributed File System in Go

GoDrive is a lightweight, fault-tolerant distributed file system built with Go. It supports chunk-based file storage, replication, real-time consistency checks, quorum-based conflict resolution, and intelligent caching. This project was designed to explore core concepts in distributed systems and file storage reliability.

---

## ✨ Features

- ⚙️ **File Chunking**: Files are split into fixed-size chunks for distribution.
- 🔁 **Replication**: Each chunk is replicated across multiple slave nodes for fault tolerance.
- 🧠 **Central Metadata Server**: Maintains mappings of file -> chunk -> node.
- ✅ **Checksum & Integrity**: Each chunk uses a hash for integrity validation.
- 📥 **Write & Update Support**: Efficient handling of new uploads and delta updates.
- 📤 **Chunk Deletion**: Removes unused or invalid chunks across slave nodes.
- ⚖️ **Quorum Resolution**: Uses Moore’s Voting Algorithm to determine the majority chunk version in case of conflicts.
- 🧊 **LRU Cache**: Speeds up read performance by caching frequently accessed chunks.
- 💓 **Node Health Checks**: Pulse detection ensures replicas are live and triggers automatic re-replication if a node is down.
- 🧵 **Concurrency**: Uses goroutines and sync primitives for parallel chunk distribution and updates.

---

## 📦 Architecture

![image](https://github.com/user-attachments/assets/0f244ec2-4f27-4efd-bbc6-7a9ec5e10320)

- Files are split and distributed to slave nodes.
- The master handles metadata, node selection, consistency, and coordination.
- Chunk replication ensures fault tolerance.

---

## 🧠 Technologies Used

- **Go** – core programming language for concurrency and networking
- **LRU Cache** – in-memory cache for optimized reads
- **Moore's Voting Algorithm** – quorum consensus for chunk consistency
- **Custom Node Selector Interface** – to decide which node receives which chunk
- **Standard Libraries** – `net/http`, `sync`, `os`, `log`

---

## 📁 Project Structure

```
GODRIVE/
├── config/                    # Configuration files and loader
│   ├── config.go
│   └── config.yaml
├── master/                    # Master server code
│   ├── master.file.go         # File chunk distribution
│   ├── master.http.go         # HTTP endpoints (if any)
│   ├── master.metadata.go     # Metadata management
│   ├── master.nodeManager.go  # Pulse check and node handling
│   ├── master.RoundRobin.go   # Round-robin node selector
│   └── master.tcp.go          # TCP listener for master
├── slave/                     # Slave node code
│   ├── storage/               # Chunk storage directory
│   └── slave.tcp.go           # TCP listener for slave
├── tmp/                       # Temporary files and logs
│   ├── build-errors.log
│   └── main.exe
├── .gitignore
├── go.mod
├── go.sum
├── main.go                    # Entry point
└── master.metadata.json       # Central metadata file
```

---

## ⚠️ Error Handling

- Upload fails if **any chunk isn't replicated to the minimum required number of nodes**.
- Logs all failed chunk operations.
- Metadata is updated **only after** successful replication.

---

## 🏁 Getting Started

1. Clone the repo:
   ```bash
   git clone https://github.com/yourusername/godrive.git
   cd godrive
   ```

2. Configure `config/config.json` with the desired replication factor and ports.

3. Run the master server:
   ```bash
   go run master/main.go
   ```

4. Run slave nodes:
   ```bash
   go run slave/main.go --port=8001
   ```

5. Upload files via client scripts or REST APIs.

---

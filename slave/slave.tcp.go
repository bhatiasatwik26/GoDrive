package slave

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"godrive/config"
	"godrive/master"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

func StartSlaveNodes() {
	slaveList := config.ReadConfig.SlaveNodes
	backupList := config.ReadConfig.BackupNodes
	for _, node := range slaveList {
		go startSlaveTcp(node, "slave")
	}
	for _, node := range backupList {
		go startSlaveTcp(node, "backup")
	}
}

func startSlaveTcp(node config.Node, role string) {
	fullAddress := fmt.Sprintf(":%s", node.Port)
	listener, err := net.Listen("tcp", fullAddress)

	if err != nil {
		log.Println("🔴   ", node.Port)
		return
	}

	defer listener.Close()

	if role == "slave" {
		log.Printf(`
╭────────────────────────────╮
│ Slave active on port %s  │
╰────────────────────────────╯`, node.Port)
	} else {
		log.Printf("┌─ Backup active on port %s ─┐", node.Port)
	}
	err = os.MkdirAll(fmt.Sprintf("slave/storage/Port_%v", node.Port), os.ModePerm)
	if err != nil {
		log.Println("🔴 Couldn't create file storage:", node.Port, err)
		return
	}

	for {
		connection, err := listener.Accept()
		if err != nil {
			log.Println("🔴 Error in connection, dropping now", node.Port)
			continue
		}
		go handleIncomingMasterRequest(node, connection)
	}
}

func handleIncomingMasterRequest(node config.Node, connection net.Conn) {
	defer connection.Close()
	buffer := make([]byte, 1024)
	n, err := connection.Read(buffer)
	if err != nil {
		log.Println("🔴 Error reading connection payload", node.Port)
		return
	}
	var incomingPayload master.TcpPayload
	err = json.Unmarshal(buffer[:n], &incomingPayload)
	if err != nil {
		log.Println("🔴 Error unmarshaling json", node.Port)
	}
	log.Printf("|TCP|: Master ---> Slave[%v]:::{ %v }", node.Port, incomingPayload.Type)
	if incomingPayload.Type == "chunk" {
		msg, err := handleIncomingChunk(incomingPayload, node.Port)
		if err != "" {
			connection.Write([]byte(err))
			return
		} else {
			connection.Write([]byte(msg))
		}
	} else if incomingPayload.Type == "req" {
		res, err := json.Marshal(handleChunkRequest(incomingPayload.Key, node.Port))
		if err != nil {
			log.Println("🔴 Error marshaling payload:", node.Port, err)
		}
		connection.Write([]byte(res))
	} else if incomingPayload.Type == "del" {
		chunkKey := incomingPayload.Key
		err := handleChunkDelete(chunkKey, node.Port)
		if err != nil {
			log.Printf("🔴 Failed to delete chunk %s: %v\n", chunkKey, err)
			connection.Write([]byte(err.Error()))
		} else {
			connection.Write([]byte("ACK"))
		}
	} else if incomingPayload.Type == "heartbeat" {
		if node.Port == "6001" {
			connection.Write([]byte(""))
		} else {
			connection.Write([]byte("ACK"))
		}
	} else if strings.HasPrefix(incomingPayload.Type, "transfer") {
		target := strings.TrimPrefix(incomingPayload.Type, "transfer@")
		if handleInternodeDataTransfer(target, incomingPayload.Key, node.Port) {
			connection.Write([]byte("ACK"))
		} else {
			connection.Write([]byte("NOACK"))
		}

	} else {
		log.Println("🔴 Invalid request by master:", incomingPayload.Type, node.Port)
	}
}

func handleIncomingChunk(incomingPayload master.TcpPayload, port string) (string, string) {
	newHash := sha256.Sum256([]byte(incomingPayload.FileChunk.Data))
	hashStr := fmt.Sprintf("%x", newHash)
	incomingHash := incomingPayload.FileChunk.Hash
	if hashStr != incomingHash {
		return "", "HashMismatch"
	}

	folder1 := hashStr[:2]
	folder2 := hashStr[2:4]
	filename := hashStr

	storagePath := fmt.Sprintf("slave/storage/Port_%s/%s/%s/%s", port, folder1, folder2, filename)

	err := os.MkdirAll(fmt.Sprintf("slave/storage/Port_%s/%s/%s", port, folder1, folder2), os.ModePerm)
	if err != nil {
		log.Println("🔴 Couldn't create directories:", err)
		return "", "DirectoryCreationError"
	}
	err = os.WriteFile(storagePath, []byte(incomingPayload.FileChunk.Data), os.ModePerm)
	if err != nil {
		log.Println("🔴 Error writing file:", err)
		return "", "FileWriteError"
	}

	return "ACK", ""
}

func handleChunkRequest(key string, port string) master.FileChunk {
	folder1 := key[:2]
	folder2 := key[2:4]
	filename := key
	var res = master.FileChunk{}

	storagePath := fmt.Sprintf("slave/storage/Port_%s/%s/%s/%s", port, folder1, folder2, filename)

	data, err := os.ReadFile(storagePath)
	if err != nil {
		log.Println("🔴 Error finding chunk:", port)
		res.Index = -1
		return res
	}
	log.Println("🟢 Chunk found in slave:", port)
	res.Index = 0
	res.Data = data
	return res
}

func handleChunkDelete(key string, port string) error {
	folder1 := key[:2]
	folder2 := key[2:4]
	filename := key
	path := fmt.Sprintf("slave/storage/Port_%s/%s/%s/%s", port, folder1, folder2, filename)

	err := os.RemoveAll(path)
	if err != nil {
		log.Println("🔴 Error deleting file:", err)
		return err
	}
	return nil
}

func handleInternodeDataTransfer(target, hash string, port string) bool {

	fileChunk := handleChunkRequest(hash, port)
	fileChunk.Hash = hash

	connection, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%s", target))
	if err != nil {
		log.Printf("🔴 [%v]: Could not connect to %v for internode transfer:", port, target)
		return false
	}
	defer connection.Close()

	payload := master.TcpPayload{Type: "chunk", FileChunk: fileChunk}
	jsonData, _ := json.Marshal(payload)
	_, err = connection.Write(jsonData)
	if err != nil {
		log.Printf("🔴 [%v]: Could not send to %v for internode transfer:", port, target)
		return false
	}

	connection.SetReadDeadline(time.Now().Add(10 * time.Second))

	buffer := make([]byte, 256)
	n, err := connection.Read(buffer)
	if err != nil {
		log.Printf("🔴 [%v]: Error reading from internode connection from %v", port, target)
		return false
	}

	ack := string(buffer[:n])
	if ack == "ACK" {
		log.Printf("🟢 [%v]: Positive ACK received from %v for internode transfer:", port, target)
		return true
	} else {
		log.Printf("🔴 [%v]: No ACK received from %v for internode transfer:", port, target)
		return false
	}
}

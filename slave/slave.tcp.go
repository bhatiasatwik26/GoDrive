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
)

func StartSlaveNodes() {
	slaveList := config.ReadConfig.SlaveNodes
	for _, node := range slaveList {
		go startSlaveTcp(node)
	}
}

func startSlaveTcp(node config.Node) {
	fullAddress := fmt.Sprintf(":%s", node.Port)
	listener, err := net.Listen("tcp", fullAddress)

	if err != nil {
		log.Fatal("Cant boot tcp server:", node.Port)
		return
	}

	defer listener.Close()
	log.Printf(`
 ╭······························ 
 │ Slave active on port: %s    
 ╰······························`, node.Port)

	err = os.MkdirAll(fmt.Sprintf("slave/storage/Port_%v", node.Port), os.ModePerm)
	if err != nil {
		log.Println("Couldn't create file storage:", node.Port, err)
		return
	}

	for {
		connection, err := listener.Accept()
		if err != nil {
			log.Println("Error in connection, dropping now", node.Port)
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
		log.Println("Error reading connection payload", node.Port)
		return
	}
	var incomingPayload master.TcpPayload
	err = json.Unmarshal(buffer[:n], &incomingPayload)
	if err != nil {
		log.Println("Error unmarshaling json", node.Port)
	}
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
			log.Println("Error marshelling payload:", node.Port, err)
		}
		connection.Write([]byte(res))
	} else if incomingPayload.Type == "del" {
		chunkKey := incomingPayload.Key
		err := handleChunkDelete(chunkKey, node.Port)
		if err != nil {
			log.Printf("Failed to delete chunk %s: %v\n", chunkKey, err)
			connection.Write([]byte(err.Error()))
		} else {
			connection.Write([]byte("ACK"))
		}
	} else {
		log.Println("Invalid request by master:", incomingPayload.Type, node.Port)
	}
}

func handleIncomingChunk(incomingPayload master.TcpPayload, port string) (string, string) {
	if port == "6001" || port == "6002" {
		return "", "HashMismatch"
	}
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
		log.Println("Couldn't create directories:", err)
		return "", "DirectoryCreationError"
	}
	err = os.WriteFile(storagePath, []byte(incomingPayload.FileChunk.Data), os.ModePerm)
	if err != nil {
		log.Println("Error writing file:", err)
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
		log.Println("Error reading file:", err, port)
		res.Index = -1
		return res
	}
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
		return err
	}
	return nil
}

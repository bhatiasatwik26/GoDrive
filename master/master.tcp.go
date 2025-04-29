package master

import (
	"encoding/json"
	"errors"
	"fmt"
	"godrive/config"
	"log"
	"net"
	"time"
)

type TcpPayload struct {
	Type      string    `json:"type"`
	FileChunk FileChunk `json:"fileChunk"`
	Key       string    `json:"key"`
}

func ConfigureMasterTcpServices() {
	loadMetaDataFromFile()
	log.Println("Metadata Loaded Successfully!")
}

func SendDataToSlave(slaveNode config.Node, chunk FileChunk) (bool, error) {
	connection, err := net.Dial("tcp", fmt.Sprintf("%s:%s", slaveNode.Host, slaveNode.Port))
	if err != nil {
		log.Println("Could not connect to slave to send data:", err)
		return false, err
	}
	defer connection.Close()

	payload := TcpPayload{Type: "chunk", FileChunk: chunk}
	jsonData, _ := json.Marshal(payload)
	_, err = connection.Write(jsonData)
	if err != nil {
		log.Println("🔴 Error sending chunk:", err)
		return false, err
	}

	connection.SetReadDeadline(time.Now().Add(10 * time.Second))

	buffer := make([]byte, 256)
	n, err := connection.Read(buffer)
	if err != nil {
		log.Println("🔴 No ACK received, Connection Timed Out:", slaveNode.Port)
		return false, err
	}

	ack := string(buffer[:n])
	if ack == "ACK" {
		log.Printf("🟢 Received positive ACK from %s for chunk index [%d]", slaveNode.Port, chunk.Index)
		return true, nil
	} else if ack == "HashMismatch" {
		log.Printf("🔴 Chunk corrupted during transfer to %s for chunk index [%d]", slaveNode.Port, chunk.Index)
		return false, errors.New("HashMismatch")
	} else {
		log.Printf("🔴 Unrecognized ACK from %s for chunk index [%d]", slaveNode.Port, chunk.Index)
		return false, errors.New("UnrecognizedAck")
	}
}

func RequestChunkFromSlave(slavePort string, key string) (FileChunk, error) {
	host := "127.0.0.1"
	connection, err := net.Dial("tcp", fmt.Sprintf("%s:%s", host, slavePort))
	if err != nil {
		log.Println("🔴 Could not connect to slave to send data:", err)
		return FileChunk{}, errors.New("Couldn't connect to slave")
	}
	defer connection.Close()

	payload := TcpPayload{Type: "req", Key: key}
	jsonData, _ := json.Marshal(payload)
	_, err = connection.Write(jsonData)
	if err != nil {
		return FileChunk{}, errors.New("Couldn't connect to slave")
	}

	connection.SetReadDeadline(time.Now().Add(10 * time.Second))

	buffer := make([]byte, 256)
	n, err := connection.Read(buffer)
	if err != nil {
		log.Println("🔴 No response received, Connection Timed Out:", slavePort)
		return FileChunk{}, errors.New("No response from slave")
	}
	var incomingFileChunk FileChunk
	err = json.Unmarshal(buffer[:n], &incomingFileChunk)
	if err != nil {
		log.Println("🔴 Error unmarshaling data from", slavePort)
	}
	if incomingFileChunk.Index == -1 {
		return incomingFileChunk, errors.New("Chunk not found")
	}
	return incomingFileChunk, nil
}

func RequestDeleteFromSlave(slaveNode config.Node, chunkHash string) error {

	connection, err := net.Dial("tcp", fmt.Sprintf("%s:%s", slaveNode.Host, slaveNode.Port))
	if err != nil {
		log.Println("🔴 Could not connect to slave to delete chunk:", err)
		return err
	}
	defer connection.Close()

	payload := TcpPayload{
		Type: "del",
		Key:  chunkHash,
	}
	jsonData, _ := json.Marshal(payload)

	_, err = connection.Write(jsonData)
	if err != nil {
		log.Println("🔴 Error sending delete request to slave:", err)
		return err
	}

	connection.SetReadDeadline(time.Now().Add(10 * time.Second))

	buffer := make([]byte, 256)
	n, err := connection.Read(buffer)
	if err != nil {
		log.Println("🔴 No ACK received, Connection Timed Out:", slaveNode.Port)
		return err
	}

	ack := string(buffer[:n])
	if ack == "ACK" {
		log.Println("🟢 Received DELETE_ACK from slave:", slaveNode.Port)
		return nil
	} else {
		log.Println("🔴 Unrecognized ACK from slave during delete:", ack)
		return errors.New("Unrecognized delete ACK")
	}
}

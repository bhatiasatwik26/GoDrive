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
	log.Println("Metadata Loaded Sucessfully!!!")
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
		log.Println("Error sending chunk:", err)
		return false, err
	}

	connection.SetReadDeadline(time.Now().Add(10 * time.Second))

	buffer := make([]byte, 256)
	n, err := connection.Read(buffer)
	if err != nil {
		log.Println("No ACK recieved, Connection Timed Out:", slaveNode.Port)
		return false, err
	}

	ack := string(buffer[:n])
	if ack == "ACK" {
		fmt.Println("Recieved positive ACK!", slaveNode.Port)
		return true, nil
	} else if ack == "HashMismatch" {
		fmt.Println("Chunk corrupted during transfer")
		return false, errors.New("HashMismatch")
	} else {
		fmt.Println("Unrecognised ACK")
		return false, errors.New("UnrecognisedAck")
	}
}

func RequestDataFromSlave(slavePort string, key string) {
	host := "127.0.0.1"
	connection, err := net.Dial("tcp", fmt.Sprintf("%s:%s", host, slavePort))
	if err != nil {
		log.Println("Could not connect to slave to send data:", err)
		return
	}
	defer connection.Close()

	payload := TcpPayload{Type: "req", Key: key}
	jsonData, _ := json.Marshal(payload)
	_, err = connection.Write(jsonData)
	if err != nil {
		log.Println("Error sending chunk:", err)
		return
	}
	fmt.Println("Request sent sucessfully to:", slavePort)

	connection.SetReadDeadline(time.Now().Add(10 * time.Second))

	buffer := make([]byte, 256)
	n, err := connection.Read(buffer)
	if err != nil {
		log.Println("No response recieved, Connection Timed Out:", slavePort)
		return
	}
	var incomingFileChunk FileChunk
	err = json.Unmarshal(buffer[:n], &incomingFileChunk)
	if err != nil {
		log.Println("Error unmarshaling data from", slavePort)
	}
	log.Println("incomingFileChunk:", incomingFileChunk)
}

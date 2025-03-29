package master

import (
	"encoding/json"
	"fmt"
	"godrive/config"
	"log"
	"net"
	"time"
)

type TcpPayload struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

func SendDataToSlave(slaveNode config.Node, data string) {

	connection, err := net.Dial("tcp", fmt.Sprintf("%s:%s", slaveNode.Host, slaveNode.Port))
	if err != nil {
		log.Println("Could not connect to slave to send data:", err)
		return
	}
	defer connection.Close()

	payload := TcpPayload{Type: "chunk", Value: data}
	jsonData, _ := json.Marshal(payload)
	_, err = connection.Write(jsonData)
	if err != nil {
		log.Println("Error sending chunk:", err)
		return
	}
	fmt.Println("Chunk sent sucessfully to:", slaveNode.Port)

	connection.SetReadDeadline(time.Now().Add(10 * time.Second))

	buffer := make([]byte, 256)
	n, err := connection.Read(buffer)
	if err != nil {
		log.Println("No ACK recieved, Connection Timed Out:", slaveNode.Port)
		return
	}

	ack := string(buffer[:n])
	if ack == "ACK" {
		fmt.Println("Recieved positive ACK!", slaveNode.Port)
	} else {
		fmt.Println("Unfamiliar formta:", slaveNode.Port)
	}
}
func RequestDataFromSlave(slaveNode config.Node, key string) {

	connection, err := net.Dial("tcp", fmt.Sprintf("%s:%s", slaveNode.Host, slaveNode.Port))
	if err != nil {
		log.Println("Could not connect to slave to send data:", err)
		return
	}
	defer connection.Close()

	payload := TcpPayload{Type: "req", Value: key}
	jsonData, _ := json.Marshal(payload)
	_, err = connection.Write(jsonData)
	if err != nil {
		log.Println("Error sending chunk:", err)
		return
	}
	fmt.Println("Request sent sucessfully to:", slaveNode.Port)

	connection.SetReadDeadline(time.Now().Add(10 * time.Second))

	buffer := make([]byte, 256)
	n, err := connection.Read(buffer)
	if err != nil {
		log.Println("No response recieved, Connection Timed Out:", slaveNode.Port)
		return
	}
	var incomingPayload TcpPayload
	err = json.Unmarshal(buffer[:n], &incomingPayload)
	if err != nil {
		log.Println("Error unmarshaling data from", slaveNode.Port)
	}
	log.Println(incomingPayload)
}

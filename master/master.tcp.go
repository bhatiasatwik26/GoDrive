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
	Type  string `json:"type"`
	Value []byte `json:"value"`
}

func SendDataToSlave(slaveNode config.Node, data []byte) error {

	connection, err := net.Dial("tcp", fmt.Sprintf("%s:%s", slaveNode.Host, slaveNode.Port))
	if err != nil {
		log.Println("Could not connect to slave to send data:", err)
		return err
	}
	defer connection.Close()

	payload := TcpPayload{Type: "chunk", Value: data}
	fmt.Println(payload.Value)
	jsonData, _ := json.Marshal(payload)
	_, err = connection.Write(jsonData)
	if err != nil {
		log.Println("Error sending chunk:", err)
		return err
	}

	connection.SetReadDeadline(time.Now().Add(10 * time.Second))

	buffer := make([]byte, 256)
	n, err := connection.Read(buffer)
	if err != nil {
		log.Println("No ACK recieved, Connection Timed Out:", slaveNode.Port)
		return err
	}

	ack := string(buffer[:n])
	if ack == "ACK" {
		fmt.Println("Recieved positive ACK!", slaveNode.Port)
		return errors.New("no ack")
	} else {
		fmt.Println("Unfamiliar formta:", slaveNode.Port)
		return errors.New("bad ack")
	}
}
func RequestDataFromSlave(slaveNode config.Node, key []byte) {

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

package slave

import (
	"encoding/json"
	"fmt"
	"godrive/config"
	"log"
	"net"
)

type TcpPayload struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

func StartSlaveNodes() {
	slaveList := config.ReadConfig.SlaveNodes
	go startSlaveTcp(slaveList[0])
	// for _, node := range slaveList {
	// 	go startSlaveTcp(node)
	// }
}

func startSlaveTcp(node config.Node) {
	fullAddress := fmt.Sprintf(":%s", node.Port)
	listener, err := net.Listen("tcp", fullAddress)
	if err != nil {
		log.Fatal("Cant boot tcp server:", node.Port)
		return
	}

	defer listener.Close()
	log.Println("Slave listening on port", fullAddress)

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
	var incomingPayload TcpPayload
	err = json.Unmarshal(buffer[:n], &incomingPayload)
	if err != nil {
		log.Println("Error unmarshaling json", node.Port)
	}
	if incomingPayload.Type == "chunk" {
		if _, err := connection.Write([]byte("ACK")); err != nil {
			log.Println("Error sending ACK", node.Port)
		} else {
			log.Println("ACK sent to master", node.Port)
		}
	} else if incomingPayload.Type == "req" {
		log.Println(incomingPayload.Value)
	} else {
		log.Println("Invalid request by master:", incomingPayload.Type, node.Port)
	}

}

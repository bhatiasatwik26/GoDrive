package master

import (
	"fmt"
	"godrive/config"
	"log"
	"time"
)

func StartHeartBeat() {
	log.Println("Heartbeat service initiated!")
	count := 1
	for {
		time.Sleep(3 * time.Second)
		log.Printf("---------------- Heartbeat: %v ----------------", count)
		for _, slaveNode := range config.ReadConfig.SlaveNodes {
			heartbeat := SendHeartBeatToSlave(slaveNode)
			if !heartbeat {
				log.Printf("🔴 Master <=xxxxxxxx= %s", slaveNode.Port)
				go removeDeadNodeFromSlaveList(slaveNode.Port)
			} else {
				log.Printf("🟢 Master <========== %s", slaveNode.Port)
			}
		}
		log.Println("-------------------------------------------------\n")
		count++
	}
}

func removeDeadNodeFromSlaveList(port string) {
	deleteIndex, deleteNode := -1, config.Node{}
	for index, node := range config.ReadConfig.SlaveNodes {
		if node.Port == port {
			deleteIndex = index
			deleteNode = node
			break
		}
	}
	config.ReadConfig.SlaveNodes = append(config.ReadConfig.SlaveNodes[:deleteIndex], config.ReadConfig.SlaveNodes[deleteIndex+1:]...)
	for _, node := range config.ReadConfig.BackupNodes {
		if heartbeat := SendHeartBeatToSlave(node); heartbeat {
			handleDataTransfer(port, node.Port)
			config.ReadConfig.BackupNodes = append(config.ReadConfig.BackupNodes, deleteNode)
			config.ReadConfig.SlaveNodes = append(config.ReadConfig.SlaveNodes, node)
			fmt.Println("\n-------------------------------------------------------------------------------------")
			fmt.Println("Slave Node List:\n", config.ReadConfig.SlaveNodes)
			fmt.Println("\nBackup Node List:", config.ReadConfig.BackupNodes, "")
			fmt.Print("-------------------------------------------------------------------------------------\n\n")
			break
		}
	}
}
func handleDataTransfer(from string, to string) {

	log.Printf("\n\n Transferring data: [🟥]%v ---> %v[🟩]\n\n", from, to)
	metadata.mu.Lock()
	defer metadata.mu.Unlock()

	success := true

	for _, chunkMap := range metadata.Chunks {
		for _, chunkInfo := range chunkMap {
			contains := false
			sourceNode := from
			for _, node := range chunkInfo.SlaveNodeList {
				if node == from {
					contains = true
				} else {
					sourceNode = node
				}
			}
			if contains && sourceNode != to && sourceNode != from {
				success = success && SendInterNodeTransferRequest(sourceNode, to, chunkInfo.ChunkHash)
			}
		}
	}
	if !success {
		log.Printf("🔴 InterNode chunk transfer failed from %s to %s\n", from, to)
	} else {
		log.Printf("🟢 InterNode chunk transfer successful from %s to %s\n", from, to)
	}

}

package master

import (
	"godrive/config"
	"log"
	"sync"
)

type NodeSelector interface {
	GiveNode() config.Node
}

func assignNode(chunk FileChunk) (bool, []string) {
	var wg sync.WaitGroup
	var mu sync.Mutex
	successCount := 0
	assignedPorts := make([]string, 0, config.ReadConfig.Master.ReplicationFactor)

	for replication := 0; replication < config.ReadConfig.Master.ReplicationFactor; replication++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			selectedNode := MyNodeSelector.GiveNode()
			success, _ := SendDataToSlave(selectedNode, chunk)
			if success {
				mu.Lock()
				successCount++
				assignedPorts = append(assignedPorts, selectedNode.Port)
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	if successCount < config.ReadConfig.Master.WriteQuorum {
		log.Printf("Write quorum not met for chunk index %d\n", chunk.Index)
		return false, nil
	}

	return true, assignedPorts
}

func DistriButeChunksToNode(file FileStruct) bool {
	var wg sync.WaitGroup
	chunkSuccessMap := make([]int, len(file.Chunks))
	var mu sync.Mutex

	for ind, chunk := range file.Chunks {
		for replication := 0; replication < config.ReadConfig.Master.ReplicationFactor; replication++ {
			wg.Add(1)
			go func(index int, chunk FileChunk) {
				defer wg.Done()
				selectedNode := MyNodeSelector.GiveNode()
				success, _ := SendDataToSlave(selectedNode, chunk)
				if success {
					mu.Lock()
					chunkSuccessMap[index] += 1
					mu.Unlock()
					addChunkInfoToMetaData(file.Name, chunk.Hash, chunk.Index, selectedNode.Port)
				}
			}(ind, chunk)
		}
	}
	wg.Wait()

	for index, count := range chunkSuccessMap {
		if count < config.ReadConfig.Master.WriteQuorum {
			log.Println("Write quorum not met for chunk:", index, "in file:", file.Name)
			DeleteFileIfQuoramFails(file.Name)
			return false
		}
	}
	return true
}

// func CompareChunksAndUpdate(file FileStruct) {
// 	var wg sync.WaitGroup
// 	currentFileMap := metadata.Chunks[file.Name] // ind -> fileChunk

// 	for index, chunk := range file.Chunks {
// 		if index < len(currentFileMap) {
// 			mapChunkInfo := currentFileMap[index]
// 			if chunk.Hash == mapChunkInfo.ChunkHash {
// 				continue
// 			} else {
// 				// update FileChunk On slaveNodes
// 			}
// 		} else {
// 			// save chunks to
// 		}

// 		for i := 0; i < len(mapChunkInfo.SlaveNodeList); i++ {
// 			port := mapChunkInfo.SlaveNodeList[i]
// 			wg.Add(1)
// 			go func(chunk FileChunk, port string) {
// 				defer wg.Done()
// 				selectedNode := config.Node{Host: "127.0.0.1", Port: port}

//					_, err := SendDataToSlave(selectedNode, chunk)
//					if err != nil {
//						log.Println("Error sending data to slave:", err)
//						return
//					}
//					updateChunkHashInMetaData(file.Name, chunk.Hash, chunk.Index, selectedNode.Port)
//				}(chunk, port)
//			}
//		}
//		wg.Wait()
//	}

func deleteChunkFromSlaves(chunkInfo *ChunkInfo, ackChan chan bool, wg *sync.WaitGroup) {
	defer wg.Done()
	chunkHash := chunkInfo.ChunkHash
	allDeleted := true

	for _, port := range chunkInfo.SlaveNodeList {
		slaveNode := config.Node{Host: "127.0.0.1", Port: port}
		err := RequestDeleteFromSlave(slaveNode, chunkHash)
		if err != nil {
			log.Printf("Failed to delete chunk %s from node %s: %v\n", chunkHash, port, err)
			allDeleted = false
		} else {
			log.Printf("Deleted chunk %s from node %s\n", chunkHash, port)
		}
	}

	ackChan <- allDeleted
}

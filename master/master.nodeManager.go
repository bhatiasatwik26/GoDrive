package master

import (
	"godrive/config"
	"log"
	"sync"
)

type NodeSelector interface {
	GiveNode() config.Node
}

func DistriButeChunksToNode(file FileStruct) {
	var wg sync.WaitGroup
	for _, chunk := range file.Chunks {
		for replication := 0; replication < 3; replication++ {
			wg.Add(1)
			go func(chunk FileChunk) {
				defer wg.Done()
				selectedNode := MyNodeSelector.GiveNode()
				_, err := SendDataToSlave(selectedNode, chunk)
				if err != nil {
					log.Println()
				}
				addChunkInfoToMetaData(file.Name, chunk.Hash, chunk.Index, selectedNode.Port)
			}(chunk)
		}
	}
	wg.Wait()
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

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
		for replication := 0; replication < 2; replication++ {
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

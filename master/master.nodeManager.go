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
		wg.Add(1)

		go func(chunk FileChunk) {
			defer wg.Done()
			selectedNode := MyNodeSelector.GiveNode()
			if err := SendDataToSlave(selectedNode, chunk.Data); err != nil {
				log.Println()
			}
		}(chunk)

	}
	wg.Wait()
}

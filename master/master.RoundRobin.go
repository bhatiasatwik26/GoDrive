package master

import (
	"godrive/config"
	"sync"
)

type RoundRobinNodeSelector struct {
	mutex     sync.Mutex
	nodeIndex int
	nodeList  []config.Node
}

func NewRoundRobinSelector(nodes []config.Node) *RoundRobinNodeSelector {
	return &RoundRobinNodeSelector{
		nodeIndex: 0,
		nodeList:  nodes,
	}
}

func (R *RoundRobinNodeSelector) GiveNode() config.Node {
	R.mutex.Lock()
	defer R.mutex.Unlock()
	node := R.nodeList[R.nodeIndex]
	R.nodeIndex = (R.nodeIndex + 1) % len(R.nodeList)
	return node
}

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
func (R RoundRobinNodeSelector) GiveNode() config.Node {
	var t config.Node
	t.Host = "127.0.0.1"
	t.Port = "6001"
	return t
}

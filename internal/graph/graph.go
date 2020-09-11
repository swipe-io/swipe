package graph

import (
	"go/types"

	"github.com/swipe-io/swipe/v2/internal/queue"
	"golang.org/x/tools/go/types/typeutil"
)

type ID struct {
	Name     string
	RecvHash uint32
	TypeHash uint32
}

type Node struct {
	Object types.Object
	values []types.TypeAndValue
}

func (n *Node) Values() []types.TypeAndValue {
	return n.values
}

func (n *Node) AddValue(values ...types.TypeAndValue) {
	n.values = append(n.values, values...)
}

type Graph struct {
	hasher typeutil.Hasher
	nodes  map[ID]*Node
	edges  map[ID][]ID
}

func NewGraph() *Graph {
	return &Graph{
		hasher: typeutil.MakeHasher(),
		nodes:  map[ID]*Node{},
		edges:  map[ID][]ID{},
	}
}

func (g *Graph) objID(obj types.Object) ID {
	var recvTypeHash uint32 = 0
	if sig, ok := obj.Type().(*types.Signature); ok {
		if sig.Recv() != nil {
			recvTypeHash = g.hasher.Hash(sig.Recv().Type())
		}
	}
	return ID{
		Name:     obj.Name(),
		RecvHash: recvTypeHash,
		TypeHash: g.hasher.Hash(obj.Type()),
	}
}

func (g *Graph) Add(n *Node) {
	id := g.objID(n.Object)
	if _, ok := g.nodes[id]; ok {
		return
	}
	g.nodes[id] = n
}

func (g *Graph) Node(obj types.Object) (nodes *Node) {
	id := g.objID(obj)
	return g.nodes[id]
}

func (g *Graph) AddEdge(n1, n2 *Node) {
	id1 := g.objID(n1.Object)
	id2 := g.objID(n2.Object)
	g.edges[id1] = append(g.edges[id1], id2)
}

func (g *Graph) Iterate(f func(n *Node)) {
	for _, n := range g.nodes {
		if f != nil {
			f(n)
		}
	}
}

func (g *Graph) Traverse(node *Node, f func(n *Node) bool) {
	q := queue.New()
	q.Enqueue(g.objID(node.Object))
	visited := make(map[ID]bool)
	for {
		if q.IsEmpty() {
			break
		}
		id := q.Dequeue().(ID)
		visited[id] = true
		near := g.edges[id]
		for i := 0; i < len(near); i++ {
			j := near[i]
			if !visited[j] {
				q.Enqueue(j)
				visited[j] = true
			}
		}
		if f != nil {
			if !f(g.nodes[id]) {
				break
			}
		}
	}
}

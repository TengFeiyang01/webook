package grpc

import (
	"fmt"
	"github.com/ecodeclub/ekit/slice"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
)

type Node struct {
	name          string
	weight        int
	currentWeight int
}

// 平滑加权轮询
// 1. 初始化每个节点的 currentWeight 为 0
// 2. 每次选择节点时，将所有节点的 currentWeight 加上自身的 weight
// 3. 选择 currentWeight 最大的节点
// 4. 将选中节点的 currentWeight 减去总权重
// 5. 重复第 2 步
func TestSmoothWRR(t *testing.T) {
	nodes := []*Node{
		{name: "a", weight: 10, currentWeight: 10},
		{name: "b", weight: 20, currentWeight: 20},
		{name: "c", weight: 30, currentWeight: 30},
	}
	b := &Balancer{
		nodes: nodes,
		t:     t,
	}

	for i := 0; i < 6; i++ {
		noPtrNodes := slice.Map[*Node, Node](nodes, func(idx int, src *Node) Node {
			return *src
		})
		t.Log("")
		target := b.wrr()
		t.Log(fmt.Sprintf("第 %d 次挑选前, nodes: %+v", i+1, noPtrNodes))
		target.Invoke()
		noPtrNodes = slice.Map[*Node, Node](nodes, func(idx int, src *Node) Node {
			return *src
		})

		t.Log(fmt.Sprintf("第 %d 次挑选后, nodes: %+v", i+1, noPtrNodes))
	}
}

type Balancer struct {
	nodes []*Node
	lock  sync.Mutex
	t     *testing.T

	idx atomic.Int32
}

func (n *Node) Invoke() {
}

// 加权轮询
func (b *Balancer) wrr() *Node {
	b.lock.Lock()
	defer b.lock.Unlock()
	// 总权重
	total := 0
	for _, n := range b.nodes {
		total += n.weight
	}
	for _, n := range b.nodes {
		n.currentWeight += n.weight
	}
	var target *Node
	for _, n := range b.nodes {
		if target == nil {
			target = n
		} else {
			if n.currentWeight > target.currentWeight {
				target = n
			}
		}
	}
	b.t.Log("选中了", target)
	target.currentWeight -= total
	b.t.Log("选中的结点减去总的权重后", target)
	return target
}

// 加权随机
func (b *Balancer) randomPick() *Node {
	total := 60
	r := rand.Int31n(int32(total))
	for _, n := range b.nodes {
		r = r - int32(n.weight)
		if r <= 0 {
			return n
		}
	}
	panic("")
}

// 随机
func (b *Balancer) random() *Node {
	r := rand.Int31()
	return b.nodes[int(r)%len(b.nodes)]
}

// 轮询 rr
func (b *Balancer) roundRobin() *Node {
	idx := int(b.idx.Add(1))
	return b.nodes[idx%len(b.nodes)]
}

func (b *Balancer) hashPick(req any) *Node {
	// 这里可以根据 req 来计算出一个 hash 值
	// 然后根据 hash 值来选择节点
	//r := hash(req)
	//return b.nodes[int(r)%len(b.nodes)]
	return nil
}

package grpc

import (
	"fmt"
	"github.com/ecodeclub/ekit/slice"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
)

// 中场休息三分钟
type Node struct {
	name          string
	weight        int
	currentWeight int
}

func (n *Node) Invoke() {
}

func TestSmoothWRR(t *testing.T) {
	nodes := []*Node{
		{
			name:          "A",
			weight:        10,
			currentWeight: 10,
		},
		{
			name:          "B",
			weight:        20,
			currentWeight: 20,
		},
		{
			name:          "C",
			weight:        30,
			currentWeight: 30,
		},
	}

	b := &Balancer{
		nodes: nodes,
		t:     t,
	}
	for i := 1; i <= 6; i++ {
		t.Log(fmt.Sprintf("第 %d 个请求挑选前，nodes: %v", i, slice.Map(nodes, func(idx int, src *Node) Node {
			return *src
		})))
		target := b.wrr()
		// 模拟发起了 RPC 调用
		target.Invoke()
		t.Log(fmt.Sprintf("第 %d 个请求挑选后，nodes: %v", i, slice.Map(nodes, func(idx int, src *Node) Node {
			return *src
		})))
	}
}

type Balancer struct {
	nodes []*Node
	lock  sync.Mutex
	t     *testing.T

	// 0
	idx *atomic.Int32
}

func (b *Balancer) wrr() *Node {
	b.lock.Lock()
	defer b.lock.Unlock()
	// 总权重
	total := 0
	for _, n := range b.nodes {
		total += n.weight
	}
	// 更新当前权重
	for _, n := range b.nodes {
		n.currentWeight = n.currentWeight + n.weight
	}
	var target *Node
	for _, n := range b.nodes {
		if target == nil {
			target = n
		} else {
			// < 或者 <= 都可以
			if target.currentWeight < n.currentWeight {
				target = n
			}
		}
	}
	b.t.Log("更新了当前权重后", slice.Map(b.nodes, func(idx int, src *Node) Node {
		return *src
	}))
	b.t.Log("选中了", target)
	target.currentWeight = target.currentWeight - total
	b.t.Log("选中的节点的当前权重，减去总权重后", target)
	return target
}

func (b *Balancer) weightedRandomPick() *Node {
	total := 60
	r := rand.Int31n(int32(total))
	for _, n := range b.nodes {
		r = r - int32(n.weight)
		if r < 0 {
			return n
		}
	}
	panic("abc")
}

//func (b *Balancer) hashPick(req any) *Node {
// 在这里选一个哈希算法
//hash.Hash()
//r := hash(req)
//return b.nodes[int(r)%len(b.nodes)]
//}

func (b *Balancer) random() *Node {
	r := rand.Int31()
	return b.nodes[int(r)%len(b.nodes)]
}

// 轮询 RR
func (b *Balancer) roundRobin() *Node {
	idx := int(b.idx.Add(1))
	return b.nodes[idx%len(b.nodes)]
}

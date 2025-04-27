package wrr

import (
	"context"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"sync"
)

const Name = "custom_weighted_round_robin"

// balance.Balancer
// balancer.Builder
// balancer.Picker
// base.PickerBuilder
// balance 是 Picker 的装饰器
func newBuilder() balancer.Builder {
	return base.NewBalancerBuilder(Name, &PickerBuilder{}, base.Config{HealthCheck: true})
}

func init() {
	balancer.Register(newBuilder())
}

type PickerBuilder struct {
}

func (p *PickerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	conns := make([]*weightConn, 0, len(info.ReadySCs))
	for sc, sci := range info.ReadySCs {
		md, _ := sci.Address.Metadata.(map[string]any)
		weightVal, _ := md["weight"]
		weight, _ := weightVal.(float64)
		//if weight == 0 {
		//
		//}
		conns = append(conns, &weightConn{
			SubConn:       sc,
			weight:        int(weight),
			currentWeight: int(weight),
		})
	}

	return &Picker{
		conns: conns,
	}
}

type Picker struct {
	conns []*weightConn
	lock  sync.Mutex
}

func (p *Picker) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	p.lock.Lock()
	defer p.lock.Unlock()
	if len(p.conns) == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}
	// 总权重
	var total int
	var maxCC *weightConn
	for _, c := range p.conns {
		total += c.weight
		c.currentWeight = c.currentWeight + c.weight
		if maxCC == nil || maxCC.currentWeight < c.currentWeight {
			maxCC = c
		}
	}

	maxCC.currentWeight = maxCC.currentWeight - total

	return balancer.PickResult{
		SubConn: maxCC.SubConn,
		Done: func(info balancer.DoneInfo) {
			// 要在这里进一步调整weight/currentWeight
			// failover 要在这里做文章
			// 根据调用结果的具体错误信息进行容错
			// 1. 如果要是触发了限流了，
			// 1.1 你可以考虑直接挪走这个节点，后面再挪回来
			// 1.2 你可以考虑直接将 weight/currentWeight 调整到极低
			// 2. 触发了熔断呢？
			// 3. 降级呢？
			err := info.Err
			if err != nil {
				return
			}
			switch err {
			case context.Canceled:
				return
			case context.DeadlineExceeded:
			// 考虑降低权重
			case io.EOF, io.ErrUnexpectedEOF:
				// 基本可以认为这个节点已经崩了
			default:
				st, ok := status.FromError(err)
				if ok {
					code := st.Code()
					switch code {
					case codes.Unavailable:
						// 可能表达的是熔断
						// 考虑挪走该节点
						maxCC.available = false

						go func() {
							// 你要开一个 goroutine 去探活
							// health check
							if p.healthCheck(maxCC) {
								// 探活成功了
								// 你要将该节点挪回来
								// 最好加点流量控制的措施
								//maxCC.currentWeight = 0
								maxCC.available = true
								//maxCC.currentWeight = maxCC.weight
							}
						}()
					case codes.DeadlineExceeded:
						// 这里可能表达的是限流
						// 你可以挪走，也可以留着
						// 留着的话，可以考虑降低权重
						// 最好是 currentWeight 和 weight 都降低
					}
				}
			}
		},
	}, nil
}

func (p *Picker) healthCheck(cc *weightConn) bool {
	return true
}

type weightConn struct {
	balancer.SubConn
	weight        int
	currentWeight int

	// 可以用来标记不可用
	available bool
}

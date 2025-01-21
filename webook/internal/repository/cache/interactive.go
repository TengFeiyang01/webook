package cache

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"strconv"
	"time"
	"webook/webook/internal/domain"
)

var (
	//go:embed lua/incr_cnt.lua
	luaIncrCnt string
	//go:embed lua/interactive_ranking_incr.lua
	luaRankingIncr string
	//go:embed lua/interactive_ranking_set.lua
	luaRankingSet string
)

var RankingUpdateErr = errors.New("指定的元素不存在")

const fieldReadCnt = "read_cnt"
const fieldLikeCnt = "like_cnt"
const fieldCollectCnt = "collect_cnt"

type InteractiveCache interface {
	IncrReadCntIfPresent(ctx context.Context, biz string, bizId int64) error
	IncrLikeCntIfPresent(ctx context.Context, biz string, id int64) error
	DecrLikeCntIfPresent(ctx context.Context, biz string, id int64) error
	IncrCollectCntIfPresent(ctx context.Context, biz string, id int64) error
	Get(ctx context.Context, biz string, id int64) (domain.Interactive, error)
	Set(ctx context.Context, biz string, bizId int64, res domain.Interactive) error
	IncrRankingIfPresent(ctx context.Context, biz string, bizId int64) error
	SetRankingScore(ctx context.Context, biz string, bizId int64, count int64) error
	BatchSetRankingScore(ctx context.Context, biz string, interactives []domain.Interactive) error
	// LikeTop 基本实现，是借助ZSET
	// 1. 前100是一个高频数据，你可以结合本地缓存
	//    你可以定时刷新缓存，比如说每 5s 调用 LikeTop， 放进去本地缓存
	// 2. 如果你有一亿的数据，你怎么实时维护？ zset 放 一亿 个元素，你的 Redis 撑不住
	//    2.1 不是真的维护，而是维护近期的数据的点赞数，比如说三天内的
	//    2.2 你要分 key，这是 Redis 解决大数据常见的方案
	// 3. 借助定时任务，我每分钟计算一次
	// 4. 定时计算，算 1000 名; 而后我借助 zset 来维护 1000 名的分数
	LikeTop(ctx context.Context, biz string) ([]domain.Interactive, error)
}

type InteractiveRedisCache struct {
	client redis.Cmdable
}

func (i *InteractiveRedisCache) LikeTop(ctx context.Context, biz string) ([]domain.Interactive, error) {
	var start int64 = 0
	var end int64 = 0
	key := fmt.Sprintf("top_100_%s", biz)
	res, err := i.client.ZRevRangeWithScores(ctx, key, start, end).Result()
	if err != nil {
		return nil, err
	}
	interactives := make([]domain.Interactive, 0, 100)
	for i := 0; i < len(res); i++ {
		id, _ := strconv.ParseInt(res[i].Member.(string), 10, 64)
		interactives = append(interactives, domain.Interactive{
			Biz:     biz,
			BizId:   id,
			LikeCnt: int64(res[i].Score),
		})
	}
	return interactives, nil
}

// BatchSetRankingScore 设置整个数据
func (i *InteractiveRedisCache) BatchSetRankingScore(ctx context.Context, biz string, interactives []domain.Interactive) error {
	members := make([]redis.Z, 0, len(interactives))
	for _, interactive := range interactives {
		members = append(members, redis.Z{
			Score:  float64(interactive.LikeCnt),
			Member: interactive.BizId,
		})
	}
	return i.client.ZAdd(ctx, i.rankingKey(biz), members...).Err()
}

func (i *InteractiveRedisCache) SetRankingScore(ctx context.Context, biz string, bizId int64, count int64) error {
	return i.client.Eval(ctx, luaRankingSet, []string{i.rankingKey(biz)}, bizId, count).Err()
}

func (i *InteractiveRedisCache) IncrRankingIfPresent(ctx context.Context, biz string, bizId int64) error {
	res, err := i.client.Eval(ctx, luaRankingIncr, []string{i.rankingKey(biz)}, bizId).Result()
	if err != nil {
		return err
	}
	if res.(int64) == 0 {
		return RankingUpdateErr
	}
	return nil
}

func NewInteractiveRedisCache(client redis.Cmdable) InteractiveCache {
	return &InteractiveRedisCache{
		client: client,
	}
}

func (i *InteractiveRedisCache) Set(ctx context.Context,
	biz string, bizId int64,
	res domain.Interactive) error {
	key := i.key(biz, bizId)
	err := i.client.HSet(ctx, key, fieldCollectCnt, res.CollectCnt,
		fieldReadCnt, res.ReadCnt,
		fieldLikeCnt, res.LikeCnt,
	).Err()
	if err != nil {
		return err
	}
	return i.client.Expire(ctx, key, time.Minute*15).Err()
}

func (i *InteractiveRedisCache) Get(ctx context.Context, biz string, id int64) (domain.Interactive, error) {
	key := i.key(biz, id)
	// 之间使用 HMGet, 即使缓存里面没有对应的 key, 也不会返回 error
	res, err := i.client.HGetAll(ctx, key).Result()
	if err != nil {
		return domain.Interactive{}, err
	}
	if len(res) == 0 {
		// 缓存不存在，系统错误
		return domain.Interactive{}, ErrKeyNotExist
	}
	var intr domain.Interactive
	// 这边是可以忽略错误的
	intr.CollectCnt, _ = strconv.ParseInt(res[fieldCollectCnt], 10, 64)
	intr.LikeCnt, _ = strconv.ParseInt(res[fieldLikeCnt], 10, 64)
	intr.ReadCnt, _ = strconv.ParseInt(res[fieldReadCnt], 10, 64)
	intr.BizId = id
	return intr, nil
}

func (i *InteractiveRedisCache) IncrCollectCntIfPresent(ctx context.Context,
	biz string, id int64) error {
	key := i.key(biz, id)
	return i.client.Eval(ctx, luaIncrCnt, []string{key}, fieldCollectCnt, 1).Err()
}

func (i *InteractiveRedisCache) IncrLikeCntIfPresent(ctx context.Context,
	biz string, bizId int64) error {
	key := i.key(biz, bizId)
	return i.client.Eval(ctx, luaIncrCnt, []string{key}, fieldLikeCnt, 1).Err()
}

func (i *InteractiveRedisCache) DecrLikeCntIfPresent(ctx context.Context,
	biz string, bizId int64) error {
	key := i.key(biz, bizId)
	return i.client.Eval(ctx, luaIncrCnt, []string{key}, fieldLikeCnt, -1).Err()
}

func (i *InteractiveRedisCache) IncrReadCntIfPresent(ctx context.Context,
	biz string, bizId int64) error {
	key := i.key(biz, bizId)
	// 不是特别需要处理 res
	//res, err := i.client.Eval(ctx, luaIncrCnt, []string{key}, fieldReadCnt, 1).Int()
	return i.client.Eval(ctx, luaIncrCnt, []string{key}, fieldReadCnt, 1).Err()
}

func (i *InteractiveRedisCache) key(biz string, bizId int64) string {
	return fmt.Sprintf("interactive:%s:%d", biz, bizId)
}

func (i *InteractiveRedisCache) rankingKey(biz string) string {
	return fmt.Sprintf("top_100_%s", biz)
}

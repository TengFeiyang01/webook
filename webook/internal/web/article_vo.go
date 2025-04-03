package web

import (
	"github.com/TengFeiyang01/webook/webook/article/domain"
)

// VO view object, 对标前段的

type CollectReq struct {
	Id  int64 `json:"id"`
	Cid int64 `json:"cid"`
}

type LikeReq struct {
	// 点赞和取消点赞都复用这个
	Id   int64 `json:"id"`
	Like bool  `json:"like"`
}

type ArticleVO struct {
	Id    int64  `json:"id"`
	Title string `json:"title"`
	// 摘要
	Abstract string `json:"abstract"`
	// 内容
	Content string `json:"content"`
	// 状态这个东西，可以是前端来处理，也可以是后端来处理
	// 0 -> unknown -> 未知状态
	// 1 -> 未发表
	Status uint8  `json:"status"`
	Author string `json:"author"`
	// 计数
	ReadCnt    int64 `json:"read_cnt"`
	LikeCnt    int64 `json:"like_cnt"`
	CollectCnt int64 `json:"collect_cnt"`

	// 我个人有没有收藏，有没有点赞
	Liked     bool `json:"liked"`
	Collected bool `json:"collected"`

	Ctime string `json:"ctime"`
	Utime string `json:"utime"`
}

type ListReq struct {
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

type ArticleReq struct {
	Id      int64  `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

func (req ArticleReq) toDomain(uid int64) domain.Article {
	return domain.Article{
		Id:      req.Id,
		Title:   req.Title,
		Content: req.Content,
		Author: domain.Author{
			Id: uid,
		},
	}
}

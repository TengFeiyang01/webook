package domain

import "time"

type ArticleStatus uint8

const (
	// ArticleStatusUnknown 为了避免零值之类的问题
	ArticleStatusUnknown ArticleStatus = iota
	ArticleStatusUnPublished
	ArticleStatusPublished
	ArticleStatusPrivate
)

// Article 可以同时表达线上库和制作库的概念吗？
// 可以同时表达，作者眼中的 Article 和 读者眼中的 Article 吗？
type Article struct {
	Id      int64         `json:"id,omitempty"`
	Title   string        `json:"title" json:"title,omitempty"`
	Content string        `json:"content" json:"content,omitempty"`
	Author  Author        `json:"author" json:"author"`
	Status  ArticleStatus `json:"status" json:"status"`
	Ctime   time.Time     `json:"ctime,omitempty"`
	Utime   time.Time     `json:"utime,omitempty"`
}

func (a Article) Abstract() string {
	// 摘要我们取前几句。
	cs := []rune(a.Content)
	if len(cs) < 100 {
		return a.Content
	}
	return a.Content[:100]
}

func (s ArticleStatus) ToUint8() uint8 {
	return uint8(s)
}

func (s ArticleStatus) NonPublished() bool {
	return s != ArticleStatusPublished
}

func (s ArticleStatus) String() string {
	switch s {
	case ArticleStatusUnPublished:
		return "UnPublished"
	case ArticleStatusPublished:
		return "Published"
	case ArticleStatusPrivate:
		return "Private"
	default:
		return "Unknown"
	}
}

// Author 在帖子这个领域内, 本质是一个值对象
type Author struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}

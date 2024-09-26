package domain

// Article 可以同时表达线上库和制作库的概念吗？
// 可以同时表达，作者眼中的 Article 和 读者眼中的 Article 吗？
type Article struct {
	Id      int64  `json:"id,omitempty"`
	Title   string `json:"title" json:"title,omitempty"`
	Content string `json:"content" json:"content,omitempty"`
	Author  Author `json:"author" json:"author"`
}

// 收藏点赞

// Author 在帖子这个领域内, 本质是一个值对象
type Author struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}

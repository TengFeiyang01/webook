package domain

type Article struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Author  Author `json:"author"`
}

type Author struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}

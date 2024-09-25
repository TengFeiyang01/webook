package domain

type Article struct {
	Id      int64  `json:"id,omitempty"`
	Title   string `json:"title" json:"title,omitempty"`
	Content string `json:"content" json:"content,omitempty"`
	Author  Author `json:"author" json:"author"`
}

type Author struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}

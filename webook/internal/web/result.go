package web

// Result 返回值规范化
// 501001 => 这里代表验证码
type Result struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}

package web

type Result struct {
	// 业务错误吗
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}

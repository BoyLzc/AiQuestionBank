package _struct

// Http 请求响应结构体
type HttpResponse struct {
	Code       int         `json:"code"`
	Msg        string      `json:"msg"`
	AiResponse ModelReturn `json:"aiRes"` // 返回题目信息
}

// 存储用户请求模型的参数
type UserRequest struct {
	Model    string `json:"model" default:"tongyi"`
	Language string `json:"language" default:"go"`
	Type     string `json:"type" default:"1"`
	Keyword  string `json:"keyword" default:"任意关键字"`
}

// 存储模型返回参数
type ModelReturn struct {
	Question    string   `json:"question"`
	Options     []string `json:"options"`
	Answer      string   `json:"answer"`
	Explanation string   `json:"explanation"`
}

// 存储json文件
type JsonResult struct {
	AiRequest        UserRequest `json:"aiReq"` // 存入用户请求
	AiResponse       ModelReturn `json:"aiRes"`
	StartRequestTime string      `json:"aiStartTime"`
	EndRequestTime   string      `json:"aiEndTime"`
	CostTime         float64     `json:"aiCostTime"`
}

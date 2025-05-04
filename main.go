package main

import (
	"AiQuestionBank/model"
	_struct "AiQuestionBank/struct"
	"AiQuestionBank/tool"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
)

// 处理用户请求，返回模型调用参数
func getUserInput(userRequest _struct.UserRequest) (string, string) {
	// 对用户上传的数据进行处理
	if userRequest.Model == "tongyi" {
		userRequest.Model = "qwen-plus" // 通义千问
	} else {
		userRequest.Model = "deepseek-r1-distill-llama-8b" // deepSeek-R1
	}
	if userRequest.Type == "1" {
		userRequest.Type = "单选题"
	} else if userRequest.Type == "2" {
		userRequest.Type = "多选题"
	} else if userRequest.Type == "3" {
		userRequest.Type = "简答题"
	}

	// 模型调用
	userMessage := fmt.Sprintf("请你生成一道%s，要求与%s语言相关，题目关键字为%s。", userRequest.Type, userRequest.Language, userRequest.Keyword)
	//fmt.Println(userMessage)
	return userMessage, userRequest.Model
}

// 获取模型返回数据的关键信息（题目信息）content
func getKeyValue(res []byte, c *gin.Context, err error) _struct.ModelReturn {
	// 定义一个 map 用于存储解析后的 Json 数据，该 map 的键为字符串类型，值为任意类型
	var result map[string]interface{}

	//将字节数据res解析到result变量中，自动匹配字段类型。若解析失败返回错误信息，成功则result存储结构化数据。
	err = json.Unmarshal(res, &result)

	// 检查解析过程中是否出现错误
	if err != nil {
		fmt.Printf("解析 JSON 出错: %v\n", err)
		// 向客户端返回 HTTP 500 状态码和错误信息，告知客户端解析模型响应的 JSON 数据时出错
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "解析模型响应的 JSON 数据出错",
		})
		// 提前结束当前处理函数，避免后续代码继续执行
		return _struct.ModelReturn{}
	}

	// fmt.Println("我是解析model返回的res的结果result", result)

	// 获取 content 字段的字符串 其包含题目关键信息
	content, ok := result["choices"].([]interface{})[0].(map[string]interface{})["message"].(map[string]interface{})["content"].(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "无法获取 content 字段",
		})
		return _struct.ModelReturn{}
	}

	// 对content字符串进行处理，使其满足json格式，方便后续使用
	contentJson := tool.CleanContentData(content)

	// 将json字符串转化为map 以有助于后续和结构体关联
	contentMap := tool.GetMap(contentJson)

	// 以结构体存储
	modelReturn := getModelReturn(contentMap)

	return modelReturn
}

// 如果请求参数包含空，则设定结构体默认参数
func getDefaultData(userRequest _struct.UserRequest) _struct.UserRequest {
	if userRequest.Model == "" {
		userRequest.Model = "tongyi"
	}
	if userRequest.Language == "" {
		userRequest.Language = "go"
	}
	if userRequest.Type == "" {
		userRequest.Type = "1"
	}
	if userRequest.Keyword == "" {
		userRequest.Keyword = "任意关键字"
	}
	return userRequest
}

// 封装JsonResult结构体
func getJsonResult(modelReturn _struct.ModelReturn, userRequest _struct.UserRequest, costTime float64, startTime string, endTime string) _struct.JsonResult {
	response := _struct.JsonResult{}
	response.AiResponse = modelReturn
	response.AiRequest = userRequest
	response.CostTime = costTime
	response.StartRequestTime = startTime
	response.EndRequestTime = endTime
	return response
}

// 封装ModelReturn结构体
func getModelReturn(result map[string]interface{}) _struct.ModelReturn {
	modelReturn := _struct.ModelReturn{}
	if question, ok := result["question"].(string); ok {
		modelReturn.Question = question
	}

	if answer, ok := result["answer"].(string); ok {
		modelReturn.Answer = answer
	}

	if explanation, ok := result["explanation"].(string); ok {
		modelReturn.Explanation = explanation
	}

	if options, ok := result["options"].(map[string]interface{}); ok {
		// 做一个处理 使用切片来存储每个选项 因为map是无序的，前端拿到选项以后，需要按照A、B、C、D顺序展示
		var strOptions []string
		for key, value := range options {
			if optStr, ok := value.(string); ok {
				strOptions = append(strOptions, key+": "+optStr)
			}
		}
		// 对切片内容进行排序 关键！
		sort.Strings(strOptions)
		modelReturn.Options = strOptions
	}
	return modelReturn
}

// 核心函数，包含处理用户请求、调用模型、处理模型返回的响应、提取关键信息并和结构体关联 最终返回结构体
func coreFunc(c *gin.Context, userRequest _struct.UserRequest, path string, httpResponse _struct.HttpResponse, jsonResults []_struct.JsonResult) (_struct.ModelReturn, time.Time, time.Time, float64, error) {
	// 如果用户输入的参数有空，则赋默认值
	userRequest = getDefaultData(userRequest)

	// 根据用户输入模型的参数，获取请求信息与模型名称
	userMessage, modelName := getUserInput(userRequest)

	// 计时
	startTime := time.Now()
	fmt.Println(startTime.String())
	// 开始调用模型 通过包名调用
	res, err := model.RunModel(userMessage, modelName)
	endTime := time.Now()
	fmt.Println(endTime.String())
	// 获取调用模型所耗费时间
	cost := endTime.Sub(startTime).Seconds()

	if err != nil {
		//fmt.Printf("调用模型出错: %v\n", err)
		return _struct.ModelReturn{}, startTime, endTime, cost, err
	}

	// 对模型返回信息进行处理，使用ModelReturn结构体存储
	modelReturn := getKeyValue(res, c, err)

	return modelReturn, startTime, endTime, cost, err
}

func main() {
	// 生成路由
	router := gin.Default()
	// 加载模板文件
	router.LoadHTMLGlob("templates/*")
	// 指定接口
	// path := "/api/questions/create"
	// 根据日期获取文件名
	filePath := tool.GetFileNameByTime()
	// 初始化json文件
	tool.InitialFile(filePath)
	// 创建一个空的 JsonResult 结构体切片
	var jsonResults []_struct.JsonResult

	// router.POST(path, func(c *gin.Context) {
	// 	// 初始化httpResponse
	// 	var httpResponse _struct.HttpResponse
	// 	httpResponse.Code = 0
	// 	httpResponse.Msg = ""

	// 	// 确保请求头的 Content-Type 为 application/json
	// 	if c.ContentType() != "application/json" {
	// 		tool.HandleAPIError(c, http.StatusBadRequest, "请设置 post 请求的请求头 Headers 中的 Content-Type 为 application/json", httpResponse)
	// 		return
	// 	}

	// 	// 获取请求体内容
	// 	body, err := c.GetRawData()
	// 	if err != nil {
	// 		tool.HandleAPIError(c, http.StatusBadRequest, "读取请求头失败", httpResponse)
	// 		return
	// 	}

	// 	var userRequest _struct.UserRequest

	// 	// 手动解析 请求体中的 Json 数据 将其映射到结构体变量userRequest 解析Json数据后的值，会覆盖 userRequest默认值
	// 	err = json.Unmarshal(body, &userRequest)
	// 	if err != nil {
	// 		tool.HandleAPIError(c, http.StatusBadRequest, "请求体解析错误，请输入正确的json格式的数据(model/language/type/keyword 均是string类型)", httpResponse)
	// 		return
	// 	}

	// 	// 判断用户输入有无错误
	// 	if userRequest.Language != "" && userRequest.Language != "java" && userRequest.Language != "go" && userRequest.Language != "python" && userRequest.Language != "c++" && userRequest.Language != "javascript" {
	// 		tool.HandleAPIError(c, http.StatusBadRequest, "请输入正确的编程语言(java/go/python/c++/javascript)", httpResponse)
	// 		return
	// 	}

	// 	if userRequest.Type != "" && userRequest.Type != "1" && userRequest.Type != "2" {
	// 		tool.HandleAPIError(c, http.StatusBadRequest, "请输入正确的题目类型(1：单选题/2：多选题)", httpResponse)
	// 		return
	// 	}

	// 	if userRequest.Model != "" && userRequest.Model != "tongyi" && userRequest.Model != "deepseek" {
	// 		tool.HandleAPIError(c, http.StatusBadRequest, "请输入正确的模型(tongyi/deepseek)", httpResponse)
	// 		return
	// 	}

	// 	// json请求传入 coreFunc中，该函数中会调用模型并处理返回的信息
	// 	modelReturn, startTime, endTime, cost, err := coreFunc(c, userRequest, filePath, httpResponse, jsonResults)
	// 	if err != nil {
	// 		tool.HandleAPIError(c, http.StatusInternalServerError, "模型调用出错"+err.Error(), httpResponse)
	// 		return
	// 	}

	// 	// 将本次请求的处理信息 绑定到 jsonResult 结构体中
	// 	var jsonResult _struct.JsonResult
	// 	// 封装此次请求的结构体
	// 	jsonResult = getJsonResult(modelReturn, userRequest, cost, startTime.String(), endTime.String())
	// 	// 将此次结果写入json切片
	// 	jsonResults = append(jsonResults, jsonResult)
	// 	// 将结构体转化成json格式写入文件
	// 	err = tool.ProcessAndWriteToFile(c, jsonResult, filePath, httpResponse, jsonResults)
	// 	if err != nil {
	// 		tool.HandleAPIError(c, http.StatusInternalServerError, "读取/写入json文件出错"+err.Error(), httpResponse)
	// 		return
	// 	}

	// 	// 赋值
	// 	httpResponse.AiResponse = jsonResult.AiResponse
	// 	c.JSON(http.StatusOK, httpResponse)
	// })

	router.GET("/index", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	router.POST("/index", func(c *gin.Context) {
		// 初始化httpResponse
		var httpResponse _struct.HttpResponse
		httpResponse.Code = 0
		httpResponse.Msg = ""

		// 确保请求头的 Content-Type 为 application/json
		if c.ContentType() != "application/json" {
			tool.HandleAPIError(c, http.StatusBadRequest, "请设置 post 请求的请求头 Headers 中的 Content-Type 为 application/json", httpResponse)
			return
		}

		// 获取请求体内容
		body, err := c.GetRawData()
		if err != nil {
			tool.HandleAPIError(c, http.StatusBadRequest, "读取请求头失败", httpResponse)
			return
		}
		fmt.Println("获取到的请求体内容:", string(body))

		var userRequest _struct.UserRequest

		// 手动解析 请求体中的 Json 数据 将其映射到结构体变量userRequest
		err = json.Unmarshal(body, &userRequest)
		if err != nil {
			tool.HandleAPIError(c, http.StatusBadRequest, "请求体解析错误，请输入正确的json格式的数据(model/language/type/keyword 均是string类型)", httpResponse)
			return
		}

		// 判断用户输入有无错误
		if userRequest.Language != "" && userRequest.Language != "java" && userRequest.Language != "go" && userRequest.Language != "python" && userRequest.Language != "c++" && userRequest.Language != "javascript" {
			tool.HandleAPIError(c, http.StatusBadRequest, "请输入正确的编程语言(java/go/python/c++/javascript)", httpResponse)
			return
		}

		if userRequest.Type != "" && userRequest.Type != "1" && userRequest.Type != "2" && userRequest.Type != "3" {
			tool.HandleAPIError(c, http.StatusBadRequest, "请输入正确的题目类型(1：单选题/2：多选题)", httpResponse)
			return
		}

		if userRequest.Model != "" && userRequest.Model != "tongyi" && userRequest.Model != "deepseek" {
			tool.HandleAPIError(c, http.StatusBadRequest, "请输入正确的模型(tongyi/deepseek)", httpResponse)
			return
		}

		// 打印获取到的表单数据，方便调试
		fmt.Printf("获取到的表单数据: 模型=%s, 语言=%s, 题型=%s, 关键词=%s\n",
			userRequest.Model, userRequest.Language, userRequest.Type, userRequest.Keyword)

		// json请求传入 coreFunc中，该函数中会调用模型并处理返回的信息
		modelReturn, startTime, endTime, cost, err := coreFunc(c, userRequest, filePath, httpResponse, jsonResults)
		if err != nil {
			tool.HandleAPIError(c, http.StatusInternalServerError, "模型调用出错"+err.Error(), httpResponse)
			return
		}

		// 将本次请求的处理信息 绑定到 jsonResult 结构体中
		var jsonResult _struct.JsonResult
		// 封装此次请求的结构体
		jsonResult = getJsonResult(modelReturn, userRequest, cost, startTime.String(), endTime.String())
		// 将此次结果写入json切片
		jsonResults = append(jsonResults, jsonResult)
		// 将结构体转化成json格式写入文件
		err = tool.ProcessAndWriteToFile(c, jsonResult, filePath, httpResponse, jsonResults)
		if err != nil {
			tool.HandleAPIError(c, http.StatusInternalServerError, "读取/写入json文件出错"+err.Error(), httpResponse)
			return
		}

		// 赋值
		httpResponse.AiResponse = jsonResult.AiResponse

		fmt.Println(httpResponse)
		c.JSON(http.StatusOK, httpResponse)
	})

	// 启动路由 由于要求端口号为8080，所以需要在启动路由时指定端口号
	router.Run(":8080")
}

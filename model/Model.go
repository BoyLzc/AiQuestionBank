package model

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func init() {
	// 加载 .env 文件
	if err := godotenv.Load(); err != nil {
		fmt.Printf("Error loading .env file: %v\n", err)
	}
}

// 定义调用模型的函数  大写以使包外文件可以调用
func RunModel(userMessage string, model string) ([]byte, error) {
	// 构建请求体
	requestBody := map[string]interface{}{
		"messages": []map[string]string{
			{
				"role": "system",
				"content": "你是清华大学计算机科学与技术系的教授。" +
					"你的任务是根据用户的需求，生成一个特定主题、多样化的与计算机专业相关的考试题库。" +
					"生成题目所涉及的编程语言，例如go、JavaScript、java、python、c++等；题目类型，例如单选题、多选题、简答题；题目关键词，例如Gin框架、Spring框架等。" +
					"你所生成的所有题目，需要给出对应的正确答案和答案解析。" +
					"生成的题目、答案以及答案解析，以json格式返回。" +
					"具体返回字段包括题目描述字段：question、题目选项字段：options、题目答案字段：answer、答案解析字段：explanation。" +
					"具体要求如下：" +
					"要求题目描述字段，遵循string类型格式；" +
					"要求题目选项字段，包含四个选项，按A、B、C、D顺序排列；" +
					"要求题目选项字段，遵循map[string]interface {}类型格式；" +
					"针对题目答案字段，如果是单选题，遵循string类型格式。" +
					"如果是多选题，要求遵循string的类型格式，同时，多选题答案要求按A、B、C、D顺序排列，且每个答案之间使用顿号隔开。" +
					"如果是简答题，options字段设置为空，其他字段全都遵循string类型格式。" +
					"要求答案解析字段，遵循string类型给出。" +
					"要求返回的json格式合理定义、确保无冗余和遗漏！",
			},
			{
				"role":    "user",
				"content": userMessage,
			},
		},
		"model": model,
	}

	// 将请求体转换为 JSON 格式
	jsonBody, _ := json.Marshal(requestBody)
	//if err != nil {
	//	return nil, err
	//}
	// fmt.Println("我是jsonBody", jsonBody)

	// 创建 HTTP 请求
	req, _ := http.NewRequestWithContext(context.TODO(), "POST", "https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions", bytes.NewBuffer(jsonBody))
	//if err != nil {
	//	return nil, err
	//}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	apiKey := os.Getenv("DASHSCOPE_API_KEY")
	// fmt.Println(apiKey)
	if apiKey == "" {
		return nil, fmt.Errorf("API Key not found in environment variables")
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)

	// fmt.Println("我是req请求头", req)

	// 发送请求
	client := &http.Client{}
	resp, _ := client.Do(req)
	//if err != nil {
	//	return nil, err
	//}
	defer resp.Body.Close()

	// 读取响应体
	respBody, _ := io.ReadAll(resp.Body)
	//if err != nil {
	//	return nil, err
	//}

	// fmt.Println("我是respBody请求头", respBody)

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("请求失败，状态码: %d，响应内容: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

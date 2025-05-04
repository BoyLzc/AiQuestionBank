package tool

import (
	_struct "AiQuestionBank/struct"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// 根据日期获取文件名
func GetFileNameByTime() string {
	now := time.Now()
	year, month, day := now.Date()

	// 检查 data 目录是否存在
	dataDir := "./data"
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		// 目录不存在，创建目录
		err := os.Mkdir(dataDir, 0755)
		if err != nil {
			fmt.Printf("创建目录失败: %v\n", err)
		}
	}

	// 定义存储最终json文件的路径
	filePath := fmt.Sprintf("./data/%d_%02d_%02d.json", year, month, day)
	return filePath
}

// 每次执行该文件 则调用一次文件初始化函数
func InitialFile(filePath string) {
	// 使用APPEND模式，防止覆盖现有文件
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		// 文件创建失败时终止程序，避免后续写入空指针异常
		fmt.Println("初始化文件失败:", err)
		return
	}
	defer file.Close()

	// 检查文件是否为空
	stat, err := file.Stat()
	if err != nil {
		fmt.Println("获取文件信息失败:", err)
		return
	}
	if stat.Size() == 0 {
		// 文件为空时初始化空数组
		_, err = file.WriteString("[]") // 必须加入"[]"否则后续插入json格式内容会爆红
		if err != nil {
			fmt.Println("写入初始化数据失败:", err)
		}
	}
}

// 将json格式的字符串转换为map
func GetMap(content string) map[string]interface{} {
	// 将字符串转换为 map
	var result map[string]interface{}
	err := json.Unmarshal([]byte(content), &result)
	if err != nil {
		fmt.Println("解析 JSON 失败:", err)
	}
	return result
}

// 清理content数据 将其转化为json格式
func CleanContentData(content string) string {
	// 对 content 做预处理，去除反引号和多余字符
	cleanContent := strings.ReplaceAll(content, "`", "")
	cleanContent = strings.TrimSpace(cleanContent)
	cleanContent = strings.ReplaceAll(cleanContent, `+`, ``)
	if strings.HasPrefix(cleanContent, "json") {
		cleanContent = strings.TrimPrefix(cleanContent, "json")
		cleanContent = strings.TrimSpace(cleanContent)
	}
	if strings.HasPrefix(cleanContent, "```") {
		cleanContent = strings.TrimPrefix(cleanContent, "```")
		cleanContent = strings.TrimSpace(cleanContent)
	}
	if strings.HasSuffix(cleanContent, "```") {
		cleanContent = strings.TrimSuffix(cleanContent, "```")
		cleanContent = strings.TrimSpace(cleanContent)
	}
	return cleanContent
}

// 打印错误信息
func HandleAPIError(c *gin.Context, statusCode int, errorMsg string, httpResponse _struct.HttpResponse) {
	httpResponse.Msg = errorMsg
	httpResponse.Code = -1
	//c.JSON(statusCode, gin.H{"error": errorMsg})
	c.JSON(statusCode, httpResponse)
}

// 将结构体 data转化为 json格式 然后将其写入指定 path文件中
func ProcessAndWriteToFile(c *gin.Context, data interface{}, path string, httpResponse _struct.HttpResponse, jsonResults []_struct.JsonResult) error {
	// 打开文件时指定编码为 UTF-8
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// 使用带编码设置的解码器
	decoder := json.NewDecoder(file)
	decoder.UseNumber() // 使用精确的数字表示

	// 将新数据追加到现有数据中
	jsonData, err := json.MarshalIndent(jsonResults, "", "  ")
	if err != nil {
		return err
	}

	// 确保写入的数据是 UTF-8 编码
	err = file.Truncate(0)
	if err != nil {
		return err
	}
	_, err = file.Seek(0, 0)
	if err != nil {
		return err
	}

	_, err = file.Write(jsonData)
	return err
}

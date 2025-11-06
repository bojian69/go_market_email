package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	"go_market_email/internal/utils"
)

type AIService struct {
	config utils.AIConfig
}

func NewAIService(config utils.AIConfig) *AIService {
	return &AIService{config: config}
}

// OpenAI API 请求结构
type OpenAIRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAIResponse struct {
	Choices []Choice `json:"choices"`
	Error   *APIError `json:"error,omitempty"`
}

type Choice struct {
	Message Message `json:"message"`
}

type APIError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
}

// ProcessWithOpenAI 使用OpenAI处理提示词
func (s *AIService) ProcessWithOpenAI(prompt string, variables map[string]interface{}) (string, error) {
	// 替换提示词中的变量
	processedPrompt := s.replaceVariables(prompt, variables)
	
	// 构建请求
	request := OpenAIRequest{
		Model: s.config.OpenAI.Model,
		Messages: []Message{
			{
				Role:    "user",
				Content: processedPrompt,
			},
		},
	}
	
	requestBody, err := json.Marshal(request)
	if err != nil {
		return "", err
	}
	
	// 发送HTTP请求
	client := &http.Client{Timeout: 60 * time.Second}
	req, err := http.NewRequest("POST", s.config.OpenAI.BaseURL+"/chat/completions", bytes.NewBuffer(requestBody))
	if err != nil {
		return "", err
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.config.OpenAI.APIKey)
	
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	
	var response OpenAIResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", err
	}
	
	if response.Error != nil {
		return "", fmt.Errorf("OpenAI API错误: %s", response.Error.Message)
	}
	
	if len(response.Choices) == 0 {
		return "", fmt.Errorf("OpenAI API返回空结果")
	}
	
	return response.Choices[0].Message.Content, nil
}

// ProcessWithCustomAPI 使用自定义API处理提示词
func (s *AIService) ProcessWithCustomAPI(prompt string, variables map[string]interface{}) (string, error) {
	if s.config.CustomAPI.URL == "" {
		return "", fmt.Errorf("自定义API URL未配置")
	}
	
	// 替换提示词中的变量
	processedPrompt := s.replaceVariables(prompt, variables)
	
	// 构建请求数据
	requestData := map[string]interface{}{
		"prompt": processedPrompt,
		"data":   variables,
	}
	
	requestBody, err := json.Marshal(requestData)
	if err != nil {
		return "", err
	}
	
	// 发送HTTP请求
	client := &http.Client{Timeout: 60 * time.Second}
	req, err := http.NewRequest("POST", s.config.CustomAPI.URL, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", err
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	// 添加自定义头部
	for key, value := range s.config.CustomAPI.Headers {
		req.Header.Set(key, value)
	}
	
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	
	// 尝试解析JSON响应
	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		// 如果不是JSON，直接返回文本
		return string(body), nil
	}
	
	// 查找结果字段
	if result, ok := response["result"].(string); ok {
		return result, nil
	}
	if content, ok := response["content"].(string); ok {
		return content, nil
	}
	if message, ok := response["message"].(string); ok {
		return message, nil
	}
	
	// 如果没有找到标准字段，返回整个响应的JSON字符串
	resultJSON, _ := json.Marshal(response)
	return string(resultJSON), nil
}

// ProcessPrompt 处理提示词（自动选择AI服务）
func (s *AIService) ProcessPrompt(prompt string, variables map[string]interface{}, useCustomAPI bool) (string, error) {
	if useCustomAPI {
		return s.ProcessWithCustomAPI(prompt, variables)
	}
	return s.ProcessWithOpenAI(prompt, variables)
}

// replaceVariables 替换提示词中的变量
func (s *AIService) replaceVariables(prompt string, variables map[string]interface{}) string {
	result := prompt
	
	for key, value := range variables {
		placeholder := "{{" + key + "}}"
		var valueStr string
		
		switch v := value.(type) {
		case string:
			valueStr = v
		case int, int64, float64:
			valueStr = fmt.Sprintf("%v", v)
		case bool:
			valueStr = fmt.Sprintf("%t", v)
		default:
			// 对于复杂类型，转换为JSON字符串
			if jsonBytes, err := json.Marshal(v); err == nil {
				valueStr = string(jsonBytes)
			} else {
				valueStr = fmt.Sprintf("%v", v)
			}
		}
		
		result = strings.ReplaceAll(result, placeholder, valueStr)
	}
	
	return result
}

// ExtractVariablesFromPrompt 从提示词中提取变量
func (s *AIService) ExtractVariablesFromPrompt(prompt string) []string {
	var variables []string
	seen := make(map[string]bool)
	
	// 使用正则表达式查找 {{variable}} 格式的变量
	start := 0
	for {
		startIdx := strings.Index(prompt[start:], "{{")
		if startIdx == -1 {
			break
		}
		startIdx += start
		
		endIdx := strings.Index(prompt[startIdx:], "}}")
		if endIdx == -1 {
			break
		}
		endIdx += startIdx
		
		variable := strings.TrimSpace(prompt[startIdx+2 : endIdx])
		if variable != "" && !seen[variable] {
			variables = append(variables, variable)
			seen[variable] = true
		}
		
		start = endIdx + 2
	}
	
	return variables
}

// ValidatePrompt 验证提示词
func (s *AIService) ValidatePrompt(prompt string) error {
	if strings.TrimSpace(prompt) == "" {
		return fmt.Errorf("提示词不能为空")
	}
	
	// 检查提示词长度
	if len(prompt) > 10000 {
		return fmt.Errorf("提示词长度不能超过10000字符")
	}
	
	// 检查变量格式
	variables := s.ExtractVariablesFromPrompt(prompt)
	for _, variable := range variables {
		if strings.Contains(variable, " ") {
			return fmt.Errorf("变量名不能包含空格: %s", variable)
		}
	}
	
	return nil
}
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

// 1. 严格对齐大模型输出的结构体
type LLMResponse struct {
	Reply   string    `json:"reply"`
	Control []Control `json:"control"`
}

type Control struct {
	Action string `json:"action"` // 必须是大写枚举
	Value  any    `json:"value"`  // 扁平化取值
}

// 2. OpenAI API 响应结构体包装
type OpenAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// AskCockpitBrain 向 vLLM 引擎发起请求并强制校验 JSON Schema
func AskCockpitBrain(vehicleState string, userInstruction string) (*LLMResponse, error) {
	// 从环境变量获取远端大模型地址 (比如 http://远端IP:8000/v1/chat/completions)
	llmURL := os.Getenv("LLM_BASE_URL")
	if llmURL == "" {
		llmURL = "http://localhost:8000/v1/chat/completions" // 默认降级地址
	}

	// 极度严苛的 System Prompt
	systemPrompt := "你是一个严谨的智能汽车座舱大脑。你需要根据车辆状态推断驾驶员的所有意图。你必须严格返回一个 JSON 对象，包含 reply 和 control 数组字段。"

	// 构建带有 JSON Schema 锁的请求体
	reqBody := map[string]any{
		"model": "driveos-cockpit-llm", // 必须和 vLLM 启动时的 --served-model-name 一致
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": fmt.Sprintf("【当前车辆状态】\n%s\n\n【驾驶员语音指令】\n\"%s\"", vehicleState, userInstruction)},
		},
		"temperature": 0.01, // 抹杀随机性
		"response_format": map[string]any{
			"type": "json_schema",
			"json_schema": map[string]any{
				"name":   "DriveOS_Control_Protocol",
				"strict": true,
				"schema": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"reply": map[string]any{
							"type": "string",
						},
						"control": map[string]any{
							"type": "array",
							"items": map[string]any{
								"type": "object",
								"properties": map[string]any{
									"action": map[string]any{
										"type": "string",
										"enum": []string{"SET_AC", "OPEN_WINDOW", "CLOSE_WINDOW", "PLAY_MUSIC", "NONE"},
									},
									"value": map[string]any{
										"type": []string{"string", "integer", "boolean", "null"},
									},
								},
								"required":             []string{"action", "value"},
								"additionalProperties": false,
							},
						},
					},
					"required":             []string{"reply", "control"},
					"additionalProperties": false,
				},
			},
		},
	}

	// 序列化请求体
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	// 发起 HTTP POST 请求
	req, err := http.NewRequest("POST", llmURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	// 如果配置了虚假的 API Key，这里带上
	if apiKey := os.Getenv("LLM_API_KEY"); apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	// 执行网络请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("vLLM request failed (check if server is running or SSH tunnel is active): %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("vLLM returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// 解析最外层响应
	var openAIResp OpenAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&openAIResp); err != nil {
		return nil, fmt.Errorf("failed to decode vLLM response: %v", err)
	}

	if len(openAIResp.Choices) == 0 {
		return nil, fmt.Errorf("vLLM returned empty choices")
	}

	// 核心：反序列化受控吐出的完美 JSON 字符串
	var result LLMResponse
	rawJSONString := openAIResp.Choices[0].Message.Content
	if err := json.Unmarshal([]byte(rawJSONString), &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal structured output (the schema lock failed!): %v", err)
	}

	return &result, nil
}
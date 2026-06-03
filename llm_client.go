package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// --- 基础通信结构体 ---

type VLLMResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// =========================================================
// Phase 2: 硬件控制链路 (直接通过 JSON Schema 呼叫 8000 端口)
// =========================================================

type ControlResult struct {
	Reply   string `json:"reply"`
	Control []struct {
		Device string `json:"device"`
		Action string `json:"action"`
		Value  any    `json:"value"` // 使用 any 兼容数字和布尔值
	} `json:"control"`
}

func AskCockpitBrain(vehicleState, query string) (*ControlResult, error) {
	// 定义强制 JSON Schema
	schemaStr := `{
		"type": "object",
		"properties": {
			"reply": {"type": "string"},
			"control": {
				"type": "array",
				"items": {
					"type": "object",
					"properties": {
						"device": {"type": "string"},
						"action": {"type": "string"},
						"value": {}
					},
					"required": ["device", "action", "value"]
				}
			}
		},
		"required": ["reply", "control"]
	}`

	var schemaObj map[string]any
	json.Unmarshal([]byte(schemaStr), &schemaObj)

	reqBody := map[string]any{
		"model": "driveos-cockpit-llm",
		"messages": []map[string]string{
			{
				"role": "system",
				"content": "你是一个智能座舱控制大脑。当前车辆状态：\n" + vehicleState,
			},
			{
				"role": "user",
				"content": query,
			},
		},
		"temperature": 0.1,
		"response_format": map[string]any{
			"type": "json_schema",
			"json_schema": map[string]any{
				"name":   "cockpit_control",
				"schema": schemaObj,
			},
		},
	}

	jsonData, _ := json.Marshal(reqBody)
	resp, err := http.Post("http://127.0.0.1:8000/v1/chat/completions", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var vllmResp VLLMResponse
	json.Unmarshal(body, &vllmResp)

	if len(vllmResp.Choices) == 0 {
		return nil, fmt.Errorf("AI 底座返回为空")
	}

	var result ControlResult
	json.Unmarshal([]byte(vllmResp.Choices[0].Message.Content), &result)
	return &result, nil
}

// =========================================================
// Phase 4: 意图路由引擎与 RAG 通信组件
// =========================================================

type Task struct {
	Intent   string `json:"intent"`    // "control" 或 "qa"
	SubQuery string `json:"sub_query"` // 剥离出的纯净子问题
}

type OrchestratorResult struct {
	Tasks []Task `json:"tasks"`
}

// OrchestratorAgent: 大模型意图解剖器
func OrchestratorAgent(query string) ([]Task, error) {
	// 核心架构跃迁：Schema 从“字符串”变成了“对象数组”
	schemaStr := `{
		"type": "object",
		"properties": {
			"tasks": {
				"type": "array",
				"items": {
					"type": "object",
					"properties": {
						"intent": {
							"type": "string",
							"enum": ["control", "qa"]
						},
						"sub_query": {
							"type": "string",
							"description": "提取出对应意图的具体指令或问题文本"
						}
					},
					"required": ["intent", "sub_query"]
				}
			}
		},
		"required": ["tasks"]
	}`

	var schemaObj map[string]any
	json.Unmarshal([]byte(schemaStr), &schemaObj)

	reqBody := map[string]any{
		"model": "driveos-cockpit-llm",
		"messages": []map[string]string{
			{
				"role": "system",
				"content": "你是一个自动驾驶座舱的意图解剖中枢。驾驶员的话语可能包含多个诉求。请将其拆解为独立的任务列表。如果是调节硬件(空调、车窗等)，归为'control'；如果是询问车辆故障、仪表盘或手册，归为'qa'。必须完整提取所有的诉求。",
			},
			{
				"role": "user",
				"content": query,
			},
		},
		"temperature": 0.0,
		"response_format": map[string]any{
			"type": "json_schema",
			"json_schema": map[string]any{
				"name":   "multi_intent_routing",
				"schema": schemaObj,
			},
		},
	}

	jsonData, _ := json.Marshal(reqBody)
	resp, err := http.Post("http://127.0.0.1:8000/v1/chat/completions", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var vllmResp VLLMResponse
	json.Unmarshal(body, &vllmResp)

	if len(vllmResp.Choices) == 0 {
		return nil, fmt.Errorf("解剖引擎返回为空")
	}

	var result OrchestratorResult
	json.Unmarshal([]byte(vllmResp.Choices[0].Message.Content), &result)
	return result.Tasks, nil
}
// AskKnowledgeBase: 穿透隧道，呼叫云端 RAG 微服务 (8001 端口)
func AskKnowledgeBase(query string) (string, error) {
	reqBody := RAGRequest{Query: query}
	jsonData, _ := json.Marshal(reqBody)

	resp, err := http.Post("http://127.0.0.1:8001/api/ask_manual", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("RAG 微服务连接失败: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var ragResp RAGResponse
	if err := json.Unmarshal(body, &ragResp); err != nil {
		return "", fmt.Errorf("RAG 响应解析失败: %v", err)
	}

	return ragResp.Answer, nil
}
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
)

type HMIRequest struct {
	Query string `json:"query"`
}

// 聚合后的统一响应结构
type AggregatedResponse struct {
	Replies  []string `json:"replies"`  // 收集所有的文本回复
	Controls []any    `json:"controls"` // 收集所有的硬件控制指令
}

func handleChat(w http.ResponseWriter, r *http.Request) {
	var req HMIRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	log.Printf("==============================================")
	log.Printf("[网关接收] 原始语音: %s", req.Query)

	// 1. 意图解剖 (Fan-out 准备)
	tasks, err := OrchestratorAgent(req.Query)
	if err != nil || len(tasks) == 0 {
		log.Printf("[网关降级] 解剖失败，执行兜底控制逻辑")
		tasks = []Task{{Intent: "control", SubQuery: req.Query}}
	}

	log.Printf("[解剖中枢] 成功拆解出 %d 个并发任务:", len(tasks))
	for i, t := range tasks {
		log.Printf("  ├── 任务 %d: [%s] -> %s", i+1, t.Intent, t.SubQuery)
	}

	// 2. 多路并发执行 (Concurrent Routing)
	var wg sync.WaitGroup
	var mu sync.Mutex
	finalRes := AggregatedResponse{
		Replies:  make([]string, 0),
		Controls: make([]any, 0),
	}

	vehicleState := "空调温度: 24, 车窗: 关闭"

	for _, task := range tasks {
		wg.Add(1)

		// 启动 Goroutine 进行物理并发
		go func(t Task) {
			defer wg.Done()

			if t.Intent == "qa" {
				log.Printf("[并发链路 QA] 正在检索知识库: %s", t.SubQuery)
				answer, err := AskKnowledgeBase(t.SubQuery)

				mu.Lock()
				if err == nil {
					finalRes.Replies = append(finalRes.Replies, answer)
					log.Printf("[并发链路 QA] 检索完成 ✅")
				}
				mu.Unlock()

			} else if t.Intent == "control" {
				log.Printf("[并发链路 Control] 正在呼叫控制大脑: %s", t.SubQuery)
				res, err := AskCockpitBrain(vehicleState, t.SubQuery)

				mu.Lock()
				if err == nil {
					finalRes.Replies = append(finalRes.Replies, res.Reply)
					for _, ctrl := range res.Control {
						finalRes.Controls = append(finalRes.Controls, ctrl)
					}
					log.Printf("[并发链路 Control] 控制指令下发完成 ✅")
				}
				mu.Unlock()
			}
		}(task)
	}

	// 3. 屏障等待 (Fan-in)
	wg.Wait()
	log.Printf("[聚合枢纽] 所有并发子任务执行完毕，正在缝合数据...")

	// 4. 返回完整结果给前端
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(finalRes)
	log.Printf("[网关响应] 最终混合数据包已发送")
	log.Printf("==============================================\n")
}

func main() {
	http.HandleFunc("/api/chat", handleChat)

	log.Println(" [DriveOS 异构聚合网关 v2.0] 启动完毕")
	log.Println(" [架构升级] 并发意图解剖引擎 (Concurrent Fan-out) 已挂载")
	log.Println(" [监听端口] :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

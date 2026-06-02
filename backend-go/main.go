package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// 定义前端发来的数据结构
type HMIRequest struct {
	Query string `json:"query"`
}

// 核心路由处理函数
func handleChat(w http.ResponseWriter, r *http.Request) {
	// 1. 解析前端发来的语音文本
	var req HMIRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// 2. 构造车辆实时状态 (在真实工程中，这里会通过 CAN 总线 SDK 实时读取)
	// 将状态格式化为大模型 System Prompt 容易理解的字符串形态
	vehicleState := fmt.Sprintf("- 空调温度: %d\n- 车窗是否打开: %v\n- 播放中的音乐: %s", 24, false, "无")

	log.Printf("[网关接收] 指令: %s", req.Query)

	// 3. 呼叫我们刚才手写的神经突触 (HTTP 直连 vLLM 引擎，自带 JSON Schema 强校验)
	res, err := AskCockpitBrain(vehicleState, req.Query)
	if err != nil {
		log.Printf("[网关异常] 大模型调用失败: %v", err)
		http.Error(w, "AI Gateway Error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 4. 将 AI 受控吐出的完美结构体，直接序列化返回给前端大屏
	w.Header().Set("Content-Type", "application/json")
	// 注意这里：res.Control 已经是强类型数组，直接返回，前端解析绝不会报错
	json.NewEncoder(w).Encode(map[string]any{
		"reply":   res.Reply,
		"control": res.Control,
	})

	log.Printf("[网关响应] 成功下发 %d 条硬件控制指令", len(res.Control))
}

func main() {
	// 挂载 HTTP 路由
	http.HandleFunc("/api/chat", handleChat)

	log.Println("=====================================================")
	log.Println("  [DriveOS 网关] 异构微服务已启动")
	log.Println("  [监听端口] :8080 (等待 HMI 大屏接入...)")
	log.Println("  [算力路由] 指向 vLLM 引擎 (带 Structured Outputs 锁)")
	log.Println("=====================================================")

	log.Fatal(http.ListenAndServe(":8080", nil))
}

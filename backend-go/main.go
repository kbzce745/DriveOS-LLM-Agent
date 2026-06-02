package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	pb "github.com/kbzce745/DriveOS-LLM-Agent/backend-go/rpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// 定义全局的 gRPC 客户端
var grpcClient pb.CockpitServiceClient

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

	// 2. 构造超时上下文，呼叫底层 AI
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// 为了 Demo 演示，车辆状态先写死。实际工程中这里会去读取 CAN 总线的数据
	grpcReq := &pb.IntentRequest{
		Query: req.Query,
		Status: &pb.VehicleStatus{
			AcTemperature: 24,
			IsWindowOpen:  false,
			CurrentMedia:  "无",
		},
	}

	// 3. 阻塞等待 Python/Gemini 推理结果
	res, err := grpcClient.ProcessIntent(ctx, grpcReq)
	if err != nil {
		http.Error(w, "AI Gateway Error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 4. 将 AI 的结果打包成 JSON 返回给前端大屏
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"reply":   res.GetReplyText(),
		"control": res.GetControlAction(),
	})
}

func main() {
	// 启动时立刻连接 Python AI 服务端
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("无法连接到 AI 大脑: %v", err)
	}
	defer conn.Close()
	
	grpcClient = pb.NewCockpitServiceClient(conn)

	// 挂载 HTTP 路由
	http.HandleFunc("/api/chat", handleChat)

	log.Println(" [Go 车机中间件] HTTP 网关已启动在 8080 端口，监听 HMI 交互...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
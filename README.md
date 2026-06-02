[English](README_en.md) | [简体中文](README.md)
---
# DriveOS-LLM-Agent: 智能座舱大模型微服务架构验证项目
**AI 引擎组件：**该网关依赖于一个经过微调的自定义 Qwen2.5-7B 模型来实现意图提取。有关核心 AI 引擎、微调脚本以及 vLLM 部署的详细信息，请访问：[DriveOS-LLM-Engine](https://github.com/kbzce745/DriveOS-LLM-Engine)
## 项目简介
本项目是一个模拟真实智能汽车座舱（Smart Cockpit）的大模型应用落地验证项目。针对车载大模型落地中存在的“意图发散”、“缺乏执行边界”以及“响应延迟”等痛点，本项目采用了**异构微服务架构**，通过大模型精确推断用户意图，并严格输出底层硬件可执行的结构化控制流（JSON）。

## 系统架构设计
本系统摒弃了单体 Python 脚本的玩具做法，严格按照工业界车载中间件标准，采用 **Go (高并发网关) + Python (AI 推理引擎)** 的分离架构，中间通过 **gRPC (Protocol Buffers)** 进行低延迟的二进制通信。

* **HMI 前端层 (Mock UI):** 基于 Gradio 构建的智能座舱可视化交互大屏，模拟驾驶员语音/文本输入。
* **中间件网关层 (Go API Gateway):** 采用 Go 语言构建的常驻守护进程。负责处理 HMI 的 HTTP 请求，打包车辆实时状态（CAN 模拟数据），并通过 gRPC 转发给 AI 大脑。
* **AI 推理层 (Python gRPC Server):** 监听网关请求，组装严格的 System Prompt，调用云端大模型引擎进行意图解析与多任务编排。
* **LLM 引擎层 (Gemini GenAI):** 当前接入 Gemini 1.5/2.5 引擎，利用其强大的 JSON Mode 能力，确保输出 100% 契合底层 C++ 硬件驱动的控制格式。

## 核心特性
* **跨语言极速通信:** 采用 gRPC 替代传统 RESTful API，利用 Protobuf 解决微服务间的序列化性能损耗。
* **多意图状态机约束:** 引入严格的白名单校验（Action Whitelist），防止 LLM 越权生成危险指令（如操作刹车/气囊），确保智驾安全底线。
* **Context-Aware 动态决策:** 模型推理强依赖于当前车辆状态（如车窗开闭、当前室温），实现千人千面的动态拟人化回复。
* **Schema Drift 防御机制:** 强制 LLM 输出规范的 JSON 数组结构，完美兼容下游强类型语言的解析逻辑。

## 技术栈
* **后端网关:** Go 1.20+, `net/http`
* **AI 引擎:** Python 3.10+, `google-genai` SDK
* **RPC 通信:** gRPC, Protocol Buffers (`protoc-gen-go`, `grpcio-tools`)
* **前端 UI:** Gradio
* **配置管理:** `python-dotenv` (严格分离密钥与代码)

## 快速启动

### 1. 克隆项目
```bash
git clone [https://github.com/kbzce745/DriveOS-LLM-Agent.git](https://github.com/kbzce745/DriveOS-LLM-Agent.git)
cd DriveOS-LLM-Agent
```

### 2. 环境配置
请在 `llm-core` 目录下创建 `.env` 文件，并填入你的大模型 API 密钥：
```env
GEMINI_API_KEY="AIzaSy_你的真实密钥"
```

### 3. 启动集群 (需三个终端)
**终端 1: 启动 Python AI 大脑**
```bash
python -m venv .venv
.\.venv\Scripts\Activate.ps1
pip install -r llm-core/requirements.txt # 需提前将依赖写入
python llm-core/server.py
```

**终端 2: 启动 Go 并发网关**
```bash
cd backend-go
go mod tidy
go run main.go
```

**终端 3: 启动车载交互大屏**
```bash
.\.venv\Scripts\Activate.ps1
python mock-hmi/app.py
```
启动成功后，浏览器访问 `http://127.0.0.1:7860` 即可体验座舱助手。

## Roadmap (演进规划)
* [x] **Phase 1:** 完成 Go + Python 微服务基建，接入 Gemini API 实现多意图控制闭环。
* [x] **Phase 2:** 摆脱云端 API 依赖，基于 LLaMA-Factory 引入开源模型（如 Qwen），并使用 LoRA 技术进行 SFT 垂直领域微调。
* [ ] **Phase 3:** 引入 RAG（检索增强生成）技术，解析百万字级的《汽车用户手册》PDF，实现车载复杂故障的精准本地问答。
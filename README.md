[English](README_en.md) | [简体中文](README.md)
---
# DriveOS-LLM-Agent: 智能座舱大模型微服务架构验证项目

**AI 引擎组件：**该网关依赖于一个经过微调的自定义 Qwen2.5-7B 模型来实现意图提取与多任务编排。有关核心 AI 引擎、微调脚本以及 vLLM 部署的详细信息，请访问：[DriveOS-LLM-Engine](https://github.com/kbzce745/DriveOS-LLM-Engine)

## 项目简介
本项目是一个模拟真实智能汽车座舱（Smart Cockpit）的大模型应用落地验证项目。针对车载大模型落地中存在的“混合意图发散”、“缺乏执行边界”以及“单线程顺序响应延迟”等痛点，本项目采用了**分布式异构微服务架构**。通过多路并发智能体路由系统，精确推断并解剖用户的复合意图，并严格输出底层硬件可执行的结构化控制流（JSON）与高纯度知识回复。

## 系统架构设计
本系统摒弃了单体 Python 脚本的玩具做法，严格按照工业界车载中间件标准，采用 **Go (高并发协同网关) + Python (分布式 AI 算力与 RAG 微服务)** 的分离架构，全链路实现低延迟、高可靠的微服务通信。

* **HMI 前端层 (Mock UI):** 基于 Gradio 构建的智能座舱可视化交互大屏，模拟驾驶员语音/文本输入。
* **中间件网关层 (Go API Gateway):** 采用 Go 语言构建的常驻守护进程。负责高并发处理 HMI 的 HTTP 请求，打包车辆实时状态。核心引入 **Fan-out/Fan-in（分发-并行-聚合）并发矩阵**，利用 `sync.WaitGroup` 与 `sync.Mutex` 将复合指令拆解为独立线程并行分发。
* **AI 推理层 (vLLM Inference Server):** 基于常驻后台的 vLLM 引擎，挂载本地微调的 `driveos-cockpit-llm` 大模型。通过严苛的 JSON Schema 强约束，在毫秒级内返回硬件控制报文。
* **知识增强层 (Python FastAPI RAG Server):** 独立运行的 RAG 知识微服务，挂载 BGE-Large-ZH 向量模型与 FAISS 向量数据库，实现百万字级本地车辆用户手册的非结构化语义检索与知识拼接。

## 核心特性
* **多意图并发协同 (Fan-out/Fan-in):** 完美破解“既要控制硬件，又要询问手册”的混合意图死局，各子链路并行不悖，逻辑无损。
* **去云端依赖的纯本地化部署:** 彻底摆脱商业 API 限制与隐私风险，全链路算力、检索、分发均在本地/云端容器物理闭环。
* **Context-Aware 动态决策:** 模型推理强依赖于当前车辆状态（如车窗开闭、当前室温），实现千人千面的动态拟人化回复。
* **Schema Drift 防御机制:** 强类型契约绑定，强制 LLM 输出规范的 JSON 数组结构，完美兼容下游强类型语言的解析逻辑。

## 技术栈
* **后端网关:** Go 1.20+, `net/http`, `sync` 原生高并发组件
* **AI 推理引擎:** Python 3.10+, `vllm==0.4.2` (黄金稳定版), `pydantic==2.7.4`
* **向量检索服务:** FastAPI, Uvicorn, FAISS, HuggingFace Embeddings (`BAAI/bge-large-zh-v1.5`)
* **配置管理:** 严格的环境变量隔离机制

## 快速启动

### 1. 克隆项目
```bash
git clone [https://github.com/kbzce745/DriveOS-LLM-Agent.git](https://github.com/kbzce745/DriveOS-LLM-Agent.git)
cd DriveOS-LLM-Agent
```

### 2. 环境配置与远端底座验证
启动本地网关前，必须确保云端隔离舱（driveos 环境）中的物理双服务已成功点火并进入监听状态：

vLLM 算力底座: 监听 127.0.0.1:8000 (提供基础大模型推理)

RAG 知识服务: 监听 127.0.0.1:8001 (提供说明书语义检索)

### 3. 启动集群 (需三个终端)
**终端 1: 运行 Go 并发协同网关**
```bash
cd backend-go
go mod tidy
go run main.go llm_client.go
```

**终端 2: 启动车载交互大屏**
```bash
python mock-hmi/app.py
```
启动成功后，浏览器访问 `http://127.0.0.1:7860` 即可体验座舱助手。

## Roadmap (演进规划)
* [x] **Phase 1:** 完成 Go + Python 微服务基建，接入 Gemini API 实现多意图控制闭环。
* [x] **Phase 2:** 摆脱云端 API 依赖，基于 LLaMA-Factory 引入开源模型（如 Qwen），并使用 LoRA 技术进行 SFT 垂直领域微调。
* [x] **Phase 3:** 引入 RAG（检索增强生成）技术，解析百万字级的《汽车用户手册》PDF，实现车载复杂故障的精准本地问答。
* [x]  **Phase 3:** 重构 Go 网关为分布式智能体架构，引入多路并发图谱路由（Concurrent Routing），完美吞吐驾驶员复合复杂指令。
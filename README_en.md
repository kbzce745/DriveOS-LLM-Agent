[English](README_en.md) | [简体中文](README.md)
---
# DriveOS-LLM-Agent: Smart Cockpit LLM Microservices Architecture
## Project Overview
This project is an application validation of Large Language Models (LLMs) in a simulated smart vehicle cockpit. Addressing common pain points in automotive LLM deployment—such as intent divergence, lack of execution boundaries, and response latency—this project adopts a **heterogeneous microservices architecture**. It accurately infers user intents via LLMs and strictly outputs structured control streams (JSON) executable by underlying hardware.

## System Architecture Design
Deviating from the amateur approach of monolithic Python scripts, this system strictly follows industrial automotive middleware standards, employing a decoupled architecture: **Go (High-Concurrency Gateway) + Python (AI Inference Engine)**. The services communicate via **gRPC (Protocol Buffers)** for low-latency binary data transmission.

* **HMI Front-end (Mock UI):** A smart cockpit visual interaction screen built with Gradio, simulating driver voice/text input.
* **Middleware Gateway (Go API Gateway):** A daemon process written in Go. It handles HTTP requests from the HMI, encapsulates real-time vehicle status (simulated CAN data), and forwards them to the AI brain via gRPC.
* **AI Inference Layer (Python gRPC Server):** Listens to gateway requests, constructs strict System Prompts, and invokes cloud-based LLM engines for intent parsing and multi-task orchestration.
* **LLM Engine Layer (Gemini GenAI):** Currently integrated with the Gemini engine, utilizing its robust JSON Mode capabilities to ensure 100% format compliance for underlying C++ hardware drivers.

## Core Features
* **Ultra-Fast Cross-Language Communication:** Utilizes gRPC over traditional RESTful APIs, leveraging Protobuf to eliminate serialization overhead between microservices.
* **Multi-Intent State Machine Constraints:** Introduces a strict Action Whitelist to prevent the LLM from generating dangerous unauthorized commands (e.g., controlling brakes/airbags), ensuring driving safety boundaries.
* **Context-Aware Dynamic Decision Making:** Model inference heavily relies on the current vehicle state (e.g., window status, cabin temperature), enabling highly personalized and dynamic anthropomorphic responses.
* **Schema Drift Defense Mechanism:** Forces the LLM to output a standardized JSON array structure, perfectly compatible with the parsing logic of downstream strongly-typed languages.

## Tech Stack
* **Backend Gateway:** Go 1.20+, `net/http`
* **AI Engine:** Python 3.10+, `google-genai` SDK
* **RPC Communication:** gRPC, Protocol Buffers (`protoc-gen-go`, `grpcio-tools`)
* **Front-end UI:** Gradio
* **Configuration Management:** `python-dotenv` (Strict separation of credentials and code)

## Quick Start

### 1. Clone the Repository
```bash
git clone [https://github.com/kbzce745/DriveOS-LLM-Agent.git](https://github.com/kbzce745/DriveOS-LLM-Agent.git)
cd DriveOS-LLM-Agent
```

### 2. Environment Configuration
Create a `.env` file in the `llm-core` directory and configure your LLM API key:
```env
GEMINI_API_KEY="AIzaSy_YOUR_REAL_KEY"
```

### 3. Launch the Cluster (Requires three terminals)
**Terminal 1: Start the Python AI Brain**
```bash
python -m venv .venv
.\.venv\Scripts\Activate.ps1
pip install -r llm-core/requirements.txt
python llm-core/server.py
```

**Terminal 2: Start the Go Concurrency Gateway**
```bash
cd backend-go
go mod tidy
go run main.go
```

**Terminal 3: Start the HMI Screen**
```bash
.\.venv\Scripts\Activate.ps1
python mock-hmi/app.py
```
Upon successful launch, visit `http://127.0.0.1:7860` in your browser to experience the cockpit assistant.

## Roadmap
* [x] **Phase 1:** Complete Go + Python microservices infrastructure, integrating Gemini API for a multi-intent control loop.
* [ ] **Phase 2:** Eliminate cloud API dependency, introduce open-source models (e.g., Qwen) via LLaMA-Factory, and conduct domain-specific SFT fine-tuning using LoRA.
* [ ] **Phase 3:** Introduce RAG (Retrieval-Augmented Generation) technology to parse massive Car User Manual PDFs, enabling precise local Q&A for complex vehicular faults.
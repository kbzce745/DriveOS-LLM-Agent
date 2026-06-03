[English](README_en.md) | [简体中文](README.md)
---
# DriveOS-LLM-Agent: Smart Cockpit LLM Microservices Architecture

**AI Engine Component:** This gateway relies on a custom fine-tuned Qwen2.5-7B model for intent extraction and multi-task orchestration. For the core AI engine, fine-tuning scripts, and vLLM deployment details, please visit: [DriveOS-LLM-Engine](https://github.com/kbzce745/DriveOS-LLM-Engine)

## Project Overview
This project is an application validation of Large Language Models (LLMs) in a simulated smart vehicle cockpit. Addressing critical production bottlenecks in automotive LLM deployment—such as mixed-intent divergence, lack of execution boundaries, and sequential single-threaded response latency—this project adopts a **distributed heterogeneous microservices architecture**. It anatomizes composite user prompts into decoupled task streams within milliseconds via a concurrent multi-agent routing topology, outputting both high-purity knowledge responses and structured execution streams (JSON) executable by underlying hardware.

## System Architecture Design
Deviating from the amateur approach of monolithic Python scripts, this system strictly follows industrial automotive middleware standards, employing a decoupled architecture: **Go (High-Concurrency Gateway) + Python (Distributed Inference & RAG Microservices)** to achieve low-latency, resilient service communication.

* **HMI Front-end (Mock UI):** A smart cockpit visual interaction screen built with Gradio, simulating driver voice/text input.
* **Middleware Gateway (Go API Gateway):** A daemon process written in Go. It handles high-concurrency HTTP requests from the HMI and encapsulates real-time vehicle status. Core to this layer is the **Fan-out/Fan-in Concurrent Matrix**, leveraging `sync.WaitGroup` and `sync.Mutex` to dispatch composite instructions into parallel threads.
* **AI Inference Layer (vLLM Inference Server):** Powered by a background vLLM daemon serving the locally fine-tuned `driveos-cockpit-llm`. It outputs hardware-driving control packets within milliseconds via strict JSON Schema layout enforcement.
* **Knowledge Augmentation Layer (Python FastAPI RAG Server):** An isolated RAG microservice driving the BGE-Large-ZH embedding model and local FAISS vector storage, enabling semantic chunking, retrieval, and contextual stitching over massive car manuals.

## Core Features
* **Multi-Intent Concurrent Orchestration (Fan-out/Fan-in):** Eliminates the zero-sum selection failure between hardware manipulation and manual lookup under composite driver instructions, running downstream paths in parallel without data loss.
* **Edge-Ready Localized Deployment:** Eliminates public API reliance and data privacy hazards by completing the entire compute, retrieval, and orchestration lifecycle within local or cloud containers.
* **Context-Aware Dynamic Decision Making:** Model inference heavily relies on the current vehicle state (e.g., window status, cabin temperature), enabling highly personalized and dynamic anthropomorphic responses.
* **Schema Drift Defense Mechanism:** Strongly-typed data contract binding forces the LLM to conform to standardized JSON array structures, ensuring absolute compatibility with downstream hardware parsing boundaries.

## Tech Stack
* **Backend Gateway:** Go 1.20+, native `net/http` and `sync` concurrency primitives
* **AI Inference Core:** Python 3.10+, `vllm==0.4.2` (Production-stable基线), `pydantic==2.7.4`
* **Vector Retrieval Infrastructure:** FastAPI, Uvicorn, FAISS, HuggingFace Embeddings (`BAAI/bge-large-zh-v1.5`)
* **Configuration:** Strict environment variable isolation mechanics

## Quick Start

### 1. Clone the Repository
```bash
git clone [https://github.com/kbzce745/DriveOS-LLM-Agent.git](https://github.com/kbzce745/DriveOS-LLM-Agent.git)
cd DriveOS-LLM-Agent
```

### 2. Environment Configuration & Backend Verification
Before launching the local gateway, ensure that the decoupled backend compute cluster is fully operational and listening within the remote container environment:

vLLM Inference Core: Listening on 127.0.0.1:8000 (serving base LLM requests)

RAG Microservice: Listening on 127.0.0.1:8001 (serving the FAISS manual query pipeline)

### 3. Launch the Cluster
**Terminal 1: Start the Python AI Brain**
```bash
cd backend-go
go mod tidy
go run main.go llm_client.go
```

**Terminal 2: Start the HMI Screen**
```bash
python mock-hmi/app.py
```

Upon successful launch, visit http://127.0.0.1:7860 in your browser to experience the concurrent, multi-intent routed cockpit assistant.

## Roadmap
* [x] **Phase 1:** Complete Go + Python microservices infrastructure, integrating Gemini API for a multi-intent control loop.
* [x] **Phase 2:** Eliminate cloud API dependency, introduce open-source models (e.g., Qwen) via LLaMA-Factory, and conduct domain-specific SFT fine-tuning using LoRA.
* [x] **Phase 3:** Introduce RAG (Retrieval-Augmented Generation) technology to parse massive Car User Manual PDFs, enabling precise local Q&A for complex vehicular faults.
* [x] **Phase 4:** Refactor the Go gateway into a distributed agent topology driven by concurrent graph routing, scaling to handle composite driver inputs flawlessly.
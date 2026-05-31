import grpc
from concurrent import futures
import json
import sys
import os
from dotenv import load_dotenv

load_dotenv()
# 导入全新的 Google GenAI SDK
from google import genai
from google.genai import types

# 确保 Python 能找到 configs 目录
sys.path.append(os.path.dirname(os.path.abspath(__file__)))
from configs import agent_pb2
from configs import agent_pb2_grpc

#  填入API Key
GEMINI_API_KEY = os.getenv("GEMINI_API_KEY")
if not GEMINI_API_KEY:
    raise ValueError(" 致命错误：请在 llm-core/.env 文件中配置 GEMINI_API_KEY")

client = genai.Client(api_key=GEMINI_API_KEY)

# 实例化全新的客户端
client = genai.Client(api_key=GEMINI_API_KEY)

class CockpitServiceServicer(agent_pb2_grpc.CockpitServiceServicer):
    
    def ProcessIntent(self, request, context):
        print(f"\n [Python AI 端] 收到语音: '{request.query}'")
        print(f" [Python AI 端] 车机状态: 空调 {request.status.ac_temperature}度, 音乐 '{request.status.current_media}', 车窗打开: {request.status.is_window_open}")
        
        prompt = f"""
        你是一个严谨的智能汽车座舱大脑。
        【当前车辆状态】
        - 空调温度: {request.status.ac_temperature}
        - 车窗是否打开: {request.status.is_window_open}
        - 播放中的音乐: {request.status.current_media}
        
        【驾驶员语音指令】
        "{request.query}"
        
        【任务】
        推断驾驶员的所有意图。你必须严格返回一个 JSON 对象，必须精确包含以下两个字段，严禁创造新字段：
        1. "reply": 对驾驶员的拟人化语音回应（保持简短亲切，一句话包含所有操作的确认）。
        2. "control": 这是一个数组（List），包含所有需要执行的控制指令。数组中每个对象必须且只能包含 "action" (可选值: SET_AC, OPEN_WINDOW, CLOSE_WINDOW, PLAY_MUSIC, NONE)。如果需要调整具体值，可附加 "value" 字段。
        
        【期望输出示例】
        {{
            "reply": "好的，已为您关闭车窗，调高空调温度，并播放音乐。",
            "control": [
                {{"action": "CLOSE_WINDOW"}},
                {{"action": "SET_AC", "value": 24}},
                {{"action": "PLAY_MUSIC", "value": "随机播放"}}
            ]
        }}
        """
        
        print(" [Python AI 端] 正在呼叫云端 Gemini 引擎推理...")
        try:
            # 使用新版 SDK 的标准调用方式，并强制要求返回 JSON
            response = client.models.generate_content(
                model='gemini-2.5-flash',
                contents=prompt,
                config=types.GenerateContentConfig(
                    response_mime_type="application/json",
                )
            )
            
            result = json.loads(response.text)
            reply_text = result.get("reply", "抱歉，系统理解出现错误。")
            control_action = result.get("control", {"action": "NONE"})
            
            print(f" [Python AI 端] 推理完成！生成指令: {control_action}")
            
            return agent_pb2.IntentResponse(
                reply_text=reply_text,
                control_action=json.dumps(control_action, ensure_ascii=False)
            )
            
        except Exception as e:
            print(f" [Python AI 端] 访问大模型失败: {e}")
            return agent_pb2.IntentResponse(
                reply_text="云端大脑连接异常，请重试。",
                control_action=json.dumps({"action": "NONE"})
            )

def serve():
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    agent_pb2_grpc.add_CockpitServiceServicer_to_server(CockpitServiceServicer(), server)
    server.add_insecure_port('[::]:50051')
    print(" [Python AI 服务端 - 全新 GenAI 引擎版] 正在 50051 端口启动...")
    server.start()
    try:
        server.wait_for_termination()
    except KeyboardInterrupt:
        print("\n 收到终止信号，服务已关闭。")

if __name__ == '__main__':
    serve()
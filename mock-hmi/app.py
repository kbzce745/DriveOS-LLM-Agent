import gradio as gr
import requests
import json

def send_to_car(query):
    try:
        # 向我们刚刚升级好的 Go 网关发送 HTTP 请求
        response = requests.post("http://localhost:8080/api/chat", json={"query": query})
        response.raise_for_status()
        data = response.json()
        
        # 将 Go 返回的压缩 JSON 格式化成带缩进的漂亮格式，方便我们在界面上查看
        control_pretty = json.dumps(json.loads(data['control']), indent=2, ensure_ascii=False)
        return data['reply'], control_pretty
    except Exception as e:
        return "网关连接失败或 AI 无响应", str(e)

# 构建工业风极简 UI
with gr.Blocks(theme=gr.themes.Soft()) as demo:
    gr.Markdown("##  DriveOS 智能座舱模拟器")
    gr.Markdown("底层链路：`前端 UI` -> `HTTP` -> `Go 并发网关` -> `gRPC` -> `Python 推理服务` -> `Gemini 引擎`")
    
    with gr.Row():
        with gr.Column():
            user_input = gr.Textbox(label=" 驾驶员语音 (输入指令)", placeholder="例如：我有点冷，顺便放首陈奕迅的歌", lines=3)
            submit_btn = gr.Button(" 发送指令", variant="primary")
            
        with gr.Column():
            reply_output = gr.Textbox(label=" 车机语音播报", interactive=False, lines=2)
            control_output = gr.Code(label=" 硬件底层控制流 (JSON)", language="json")

    # 绑定按钮点击事件
    submit_btn.click(fn=send_to_car, inputs=user_input, outputs=[reply_output, control_output])
    # 绑定回车键提交事件
    user_input.submit(fn=send_to_car, inputs=user_input, outputs=[reply_output, control_output])

if __name__ == "__main__":
    print("  [HMI 前端] 座舱大屏正在启动...")
    # 启动在本地 7860 端口
    demo.launch(server_name="127.0.0.1", server_port=7860)
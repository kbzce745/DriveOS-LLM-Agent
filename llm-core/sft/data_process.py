import json
import os
import random

# 确保输出目录存在
DATA_DIR = os.path.join(os.path.dirname(__file__), '../../data')
os.makedirs(DATA_DIR, exist_ok=True)

# 定义我们要教大模型学会的核心动作
actions = ["SET_AC", "OPEN_WINDOW", "CLOSE_WINDOW", "PLAY_MUSIC", "NAVIGATION"]

# 构造原始语料（这里只写一小部分作为 Demo，真实工业界往往有几十万条）
raw_data = [
    # 空调场景
    {"user": "太冷了", "status": "ac:20", "reply": "好的，已为您调高温度。", "action": {"action": "SET_AC", "value": 24}},
    {"user": "快热死我了", "status": "ac:26", "reply": "马上为您降温。", "action": {"action": "SET_AC", "value": 20}},
    # 车窗场景
    {"user": "透透气", "status": "window:closed", "reply": "好的，正在打开车窗。", "action": {"action": "OPEN_WINDOW"}},
    {"user": "外面风太大", "status": "window:open", "reply": "明白，已关闭车窗。", "action": {"action": "CLOSE_WINDOW"}},
    # 音乐场景
    {"user": "来点周杰伦的歌", "status": "media:none", "reply": "好的，为您播放周杰伦的精选电台。", "action": {"action": "PLAY_MUSIC", "value": "周杰伦"}},
    # 闲聊/拒答场景（极其重要！模型不能乱动汽车）
    {"user": "帮我打开车门", "status": "speed:60km/h", "reply": "抱歉，车辆行驶中无法为您开启车门，请注意安全。", "action": {"action": "NONE"}},
]

# 转换为 LLM 训练标准格式 (Alpaca format)
sft_dataset = []
for item in raw_data:
    instruction = "你是一个智能汽车座舱助手。请根据当前车辆状态和用户输入，输出用户的真实意图，并返回严格的JSON控制指令以及给用户的语音回复。"
    input_text = f"【当前车辆状态】: {item['status']}\n【用户语音】: {item['user']}"
    
    # 构造必须严格遵守的 Output 格式
    output_dict = {
        "reply": item['reply'],
        "control": item['action']
    }
    
    sft_dataset.append({
        "instruction": instruction,
        "input": input_text,
        "output": json.dumps(output_dict, ensure_ascii=False)
    })

# 保存为 json 文件
output_path = os.path.join(DATA_DIR, 'cockpit_sft_data.json')
with open(output_path, 'w', encoding='utf-8') as f:
    json.dump(sft_dataset, f, ensure_ascii=False, indent=2)

print(f" 成功生成座舱微调数据集！共 {len(sft_dataset)} 条记录。")
print(f" 数据已保存在: {output_path}")
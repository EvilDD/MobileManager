#!/usr/bin/env python
# -*- coding: utf-8 -*-

import asyncio
import json
import websockets
import argparse
import time
import struct

class ScrcpyTestClient:
    """
    Scrcpy WebSocket测试客户端
    用于测试后端WebSocket功能
    """
    # 动作常量
    ACTION_DOWN = 0
    ACTION_UP = 1
    ACTION_MOVE = 2
    
    def __init__(self, host="localhost", port=8080, udid="emulator-5554", device_port=8886):
        """初始化客户端"""
        self.ws_url = f"ws://{host}:{port}/ws/scrcpy?udid={udid}&port={device_port}"
        self.ws = None
        self.connected = False
        self.video_frames_received = 0
        self.init_info_received = False
        
    async def connect(self):
        """连接到WebSocket服务器"""
        print(f"正在连接到 {self.ws_url}...")
        self.ws = await websockets.connect(self.ws_url)
        self.connected = True
        print(f"连接成功")
        
    async def disconnect(self):
        """断开连接"""
        if self.ws:
            await self.ws.close()
            self.connected = False
            print("连接已关闭")
    
    async def send_touch_event(self, action, x, y):
        """
        发送触摸事件
        
        参数:
            action: 动作类型 (0=按下, 1=抬起, 2=移动)
            x, y: 坐标
        """
        if not self.connected:
            print("未连接到服务器")
            return
            
        command = {
            "type": "touch",
            "data": {
                "action": action,
                "x": x,
                "y": y
            }
        }
        
        await self.ws.send(json.dumps(command))
        action_name = "按下" if action == self.ACTION_DOWN else "抬起" if action == self.ACTION_UP else "移动"
        print(f"发送触摸事件: {action_name} 在 ({x}, {y})")
    
    async def send_swipe_event(self, start_x, start_y, end_x, end_y, duration=500, steps=10):
        """
        发送滑动事件
        
        参数:
            start_x, start_y: 起始坐标
            end_x, end_y: 结束坐标
            duration: 持续时间(毫秒)
            steps: 步数
        """
        if not self.connected:
            print("未连接到服务器")
            return
            
        command = {
            "type": "swipe",
            "data": {
                "startX": start_x,
                "startY": start_y,
                "endX": end_x,
                "endY": end_y,
                "duration": duration,
                "steps": steps
            }
        }
        
        await self.ws.send(json.dumps(command))
        print(f"发送滑动事件: 从 ({start_x}, {start_y}) 到 ({end_x}, {end_y}), 持续{duration}ms")
    
    async def send_video_settings(self, bitrate=8000000, max_fps=24, iframe_interval=5, width=540, height=960):
        """
        发送视频设置
        
        参数:
            bitrate: 比特率 (默认8Mbps)
            max_fps: 最大帧率
            iframe_interval: I帧间隔
            width, height: 视频分辨率
        """
        if not self.connected:
            print("未连接到服务器")
            return
            
        command = {
            "type": "videoSettings",
            "data": {
                "bitrate": bitrate,
                "maxFps": max_fps,
                "iFrameInterval": iframe_interval,
                "width": width,
                "height": height
            }
        }
        
        await self.ws.send(json.dumps(command))
        print(f"发送视频设置: {width}x{height}, {max_fps}fps, {bitrate/1000000}Mbps")
    
    async def receive_messages(self):
        """接收并处理消息"""
        if not self.connected:
            print("未连接到服务器")
            return
            
        print("开始接收消息...")
        try:
            while True:
                message = await self.ws.recv()
                
                # 如果是二进制消息
                if isinstance(message, bytes):
                    self._handle_binary_message(message)
                # 如果是文本消息
                else:
                    print(f"收到文本消息: {message}")
        except websockets.exceptions.ConnectionClosed:
            print("连接已关闭")
            self.connected = False
        except Exception as e:
            print(f"接收消息时出错: {e}")
    
    def _handle_binary_message(self, data):
        """处理二进制消息"""
        # 检查是否是初始化消息
        if len(data) > 14 and data[:14] == b'scrcpy_initial':
            self._handle_initial_info(data)
            return
            
        # 检查是否是设备消息
        if len(data) > 14 and data[:14] == b'scrcpy_message':
            self._handle_device_message(data)
            return
            
        # 否则视为视频帧数据
        self.video_frames_received += 1
        if self.video_frames_received % 10 == 0:  # 每10帧打印一次
            print(f"收到视频帧数据, 大小: {len(data)}字节, 总帧数: {self.video_frames_received}")
    
    def _handle_initial_info(self, data):
        """处理初始化信息"""
        print("收到初始化信息, 大小:", len(data), "字节")
        self.init_info_received = True
        
        # 尝试解析屏幕尺寸
        if len(data) > 80:  # 确保数据足够长
            try:
                # 跳过magic字节和设备名称(14+64=78字节)
                offset = 78
                
                # 解析显示数量
                displays_count = int.from_bytes(data[offset:offset+4], byteorder='big')
                offset += 4
                
                if displays_count > 0 and offset + 24 <= len(data):
                    # 解析第一个显示的宽高
                    width = int.from_bytes(data[offset+4:offset+8], byteorder='big')
                    height = int.from_bytes(data[offset+8:offset+12], byteorder='big')
                    print(f"解析到屏幕尺寸: {width}x{height}")
            except Exception as e:
                print(f"解析初始化数据时出错: {e}")
    
    def _handle_device_message(self, data):
        """处理设备消息"""
        print("收到设备消息, 大小:", len(data), "字节")
        
        # 尝试解析消息类型
        if len(data) > 14:
            msg_type = data[14]
            print(f"设备消息类型: {msg_type}")
    
    async def run_test(self):
        """运行测试流程"""
        try:
            # 连接到服务器
            await self.connect()
            
            # 启动接收消息任务
            receive_task = asyncio.create_task(self.receive_messages())
            
            # 等待接收初始化信息
            retry_count = 0
            while not self.init_info_received and retry_count < 10:
                await asyncio.sleep(1)
                retry_count += 1
                print(f"等待初始化信息... {retry_count}/10")
            
            if not self.init_info_received:
                print("未能接收到初始化信息，测试可能无法正常进行")
            
            # 发送视频设置
            await self.send_video_settings()
            await asyncio.sleep(2)
            
            # 测试触摸事件
            print("\n===== 测试触摸事件 =====")
            # 点击屏幕中心
            await self.send_touch_event(self.ACTION_DOWN, 270, 480)
            await asyncio.sleep(0.1)
            await self.send_touch_event(self.ACTION_UP, 270, 480)
            await asyncio.sleep(2)
            
            # 测试滑动事件
            print("\n===== 测试滑动事件 =====")
            # 从上往下滑动
            await self.send_swipe_event(270, 200, 270, 800, 800, 15)
            await asyncio.sleep(2)
            
            # 从下往上滑动
            await self.send_swipe_event(270, 800, 270, 200, 800, 15)
            await asyncio.sleep(2)
            
            # 测试持续接收视频流
            print("\n===== 测试视频流接收 =====")
            print("持续接收视频流，按Ctrl+C退出...")
            
            # 等待接收消息任务完成（实际上会一直运行，除非出错）
            await receive_task
            
        except KeyboardInterrupt:
            print("用户中断测试")
        finally:
            # 关闭连接
            await self.disconnect()

async def main():
    parser = argparse.ArgumentParser(description='Scrcpy WebSocket测试客户端')
    parser.add_argument('--host', default='localhost', help='服务器主机名')
    parser.add_argument('--port', type=int, default=8800, help='服务器端口')
    parser.add_argument('--udid', default='172.17.1.205:5555', help='设备ID')
    parser.add_argument('--device-port', type=int, default=10001, help='设备端口')
    
    args = parser.parse_args()
    
    client = ScrcpyTestClient(args.host, args.port, args.udid, args.device_port)
    await client.run_test()

if __name__ == "__main__":
    asyncio.run(main())
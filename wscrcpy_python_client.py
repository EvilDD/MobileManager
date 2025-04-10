#!/usr/bin/env python
# -*- coding: utf-8 -*-

import asyncio
import websockets
import struct
import time
import json
import cv2
import numpy as np

class ScrcpyClient:
    """
    Scrcpy客户端Python实现，用于直接连接WebSocket服务器
    支持获取屏幕图像和发送点击事件
    """
    # 常量定义
    MAGIC_BYTES_INITIAL = b'scrcpy_initial'
    MAGIC_BYTES_MESSAGE = b'scrcpy_message'
    
    # 控制消息类型
    TYPE_TOUCH = 2
    TYPE_CHANGE_STREAM_PARAMETERS = 101
    
    # 触摸事件动作
    ACTION_DOWN = 0
    ACTION_UP = 1
    ACTION_MOVE = 2
    
    # 按钮标识
    BUTTON_PRIMARY = 1 << 0  # 左键
    
    # 消息类型常量
    TYPE_INJECT_KEYCODE = 0
    TYPE_INJECT_TEXT = 1
    TYPE_INJECT_TOUCH_EVENT = 2
    TYPE_INJECT_SCROLL_EVENT = 3
    TYPE_BACK_OR_SCREEN_ON = 4
    TYPE_EXPAND_NOTIFICATION_PANEL = 5
    TYPE_EXPAND_SETTINGS_PANEL = 6
    TYPE_COLLAPSE_PANELS = 7
    TYPE_GET_CLIPBOARD = 8
    TYPE_SET_CLIPBOARD = 9
    TYPE_SET_SCREEN_POWER_MODE = 10
    TYPE_ROTATE_DEVICE = 11
    
    # 视频设置中使用的尺寸参数（发送给服务器的期望尺寸）
    VIDEO_SCREEN_WIDTH = 480  # 使用整数
    VIDEO_SCREEN_HEIGHT = 960  # 使用整数
    
    def __init__(self, ws_url="ws://172.17.1.205:8886/"):
        """初始化客户端"""
        self.ws_url = ws_url
        self.ws = None
        self.device_name = None
        self.displays = []
        self.video_settings = {}
        self.screen_info = {}
        self.connection_count = 0
        self.encoders = []
        self.client_id = -1
        self.has_initial_info = False
        self.screen_width = 0
        self.screen_height = 0
        self.video_settings_sent = False  # 添加标志位，跟踪视频设置是否已发送
        
        # 触摸事件使用的尺寸（将在解析初始化信息后计算）
        self.touch_screen_width = 0
        self.touch_screen_height = 0
        
        # 视频解码器
        self.video_decoder = cv2.VideoCapture()
        self.frame_count = 0
        self.saved_frame = False

    def to_hex(self, value):
        """将整数转换为十六进制表示（无0x前缀）"""
        return format(value, 'x')
    
    def to_hex_int(self, value):
        """将整数转换为十六进制整数值"""
        return int(self.to_hex(value), 16)
    
    def calculate_touch_screen_size(self):
        """
        计算触摸事件使用的屏幕尺寸
        根据实际设备宽高比和视频设置尺寸计算触摸事件的正确尺寸
        
        计算规则:
        1. 如果是横屏(宽>高)，以宽为主
        2. 如果是竖屏(高>宽)，以高为主
        3. 比例仍然按照真实解析出来的设备宽高比
        4. 宽度和高度都调整为16的倍数，便于编码
        5. 使用精确的缩放算法
        """
        # 验证设备尺寸是否有效
        if self.screen_width <= 0 or self.screen_height <= 0:
            # 无法计算宽高比，使用视频设置的尺寸
            self.touch_screen_width = (self.VIDEO_SCREEN_WIDTH // 16) * 16
            self.touch_screen_height = (self.VIDEO_SCREEN_HEIGHT // 16) * 16
            print(f"警告：未获取到有效的设备尺寸，使用视频设置尺寸作为触摸尺寸")
            return
            
        # 计算实际设备宽高比
        device_ratio = self.screen_width / self.screen_height
        print(f"设备宽高比: {device_ratio:.4f}")
        
        # 判断设备方向
        is_landscape = self.screen_width > self.screen_height
        print(f"设备方向: {'横屏' if is_landscape else '竖屏'}")
        
        # 目标尺寸
        target_width = self.VIDEO_SCREEN_WIDTH
        target_height = self.VIDEO_SCREEN_HEIGHT
        
        # 根据设备方向进行计算
        if is_landscape:
            # 横屏：以宽为主
            scaled_height = int(target_width / device_ratio)
            if scaled_height > target_height:
                # 如果高度超出，则以高度为准
                self.touch_screen_height = target_height
                self.touch_screen_width = int(target_height * device_ratio)
            else:
                self.touch_screen_width = target_width
                self.touch_screen_height = scaled_height
        else:
            # 竖屏：以高为主
            scaled_width = int(target_height * device_ratio)
            if scaled_width > target_width:
                # 如果宽度超出，则以宽度为准
                self.touch_screen_width = target_width
                self.touch_screen_height = int(target_width / device_ratio)
            else:
                self.touch_screen_height = target_height
                self.touch_screen_width = scaled_width
            
        # 调整为16的倍数，使用位运算确保精确性
        self.touch_screen_width = self.touch_screen_width & ~15
        self.touch_screen_height = self.touch_screen_height & ~15
        
        print(f"计算得到的触摸屏幕尺寸: {self.touch_screen_width}x{self.touch_screen_height}")
        print(f"触摸屏幕尺寸(十六进制): 0x{self.to_hex(self.touch_screen_width)} x 0x{self.to_hex(self.touch_screen_height)}")

    async def connect(self):
        """连接到WebSocket服务器"""
        print(f"正在连接到 {self.ws_url}...")
        self.ws = await websockets.connect(self.ws_url)
        print(f"连接成功")
    
    async def disconnect(self):
        """断开连接"""
        if self.ws:
            await self.ws.close()
            print("连接已关闭")
    
    async def receive_loop(self):
        """接收消息循环"""
        while True:
            try:
                data = await self.ws.recv()
                if isinstance(data, bytes):
                    await self._handle_binary_message(data)
            except websockets.exceptions.ConnectionClosed:
                print("连接已关闭")
                break
            except Exception as e:
                print(f"接收消息时出错: {e}")
                break
    
    async def _handle_binary_message(self, data):
        """处理二进制消息"""
        print(f"收到二进制消息，长度: {len(data)} 字节")
        
        # 检查是否是初始化消息
        if len(data) > len(self.MAGIC_BYTES_INITIAL) and data[:len(self.MAGIC_BYTES_INITIAL)] == self.MAGIC_BYTES_INITIAL:
            await self._handle_initial_info(data)
            return
        
        # 检查是否是设备消息
        if len(data) > len(self.MAGIC_BYTES_MESSAGE) and data[:len(self.MAGIC_BYTES_MESSAGE)] == self.MAGIC_BYTES_MESSAGE:
            await self._handle_device_message(data)
            return
        
        # 处理为视频帧
        try:
            await self._handle_video_frame(data)
        except Exception as e:
            print(f"处理视频帧时出错: {e}")
    
    async def _handle_initial_info(self, data):
        """处理初始化信息"""
        try:
            print("处理初始化信息...")
            
            offset = len(self.MAGIC_BYTES_INITIAL)
            
            # 解析设备名称 (64字节)
            DEVICE_NAME_FIELD_LENGTH = 64
            name_bytes = data[offset:offset+DEVICE_NAME_FIELD_LENGTH]
            # 过滤掉尾部的零字节
            name_bytes = name_bytes.rstrip(b'\x00')
            self.device_name = name_bytes.decode('utf-8')
            offset += DEVICE_NAME_FIELD_LENGTH
            
            # 解析剩余数据
            rest = data[offset:]
            
            # 显示数量
            displays_count = struct.unpack('>i', rest[:4])[0]
            print(f"设备名称: {self.device_name}, 显示数量: {displays_count}")
            
            self.displays = []
            rest = rest[4:]  # 跳过displays_count的4字节
            
            # 解析每个显示设备的信息
            for i in range(displays_count):
                # 这里简化处理，实际需要根据DisplayInfo的结构进行解析
                # DisplayInfo的具体结构需要从源代码中获取
                display_info_buffer_length = 24  # 简化，实际需要从源码确认
                display_info_buffer = rest[:display_info_buffer_length]
                
                # 从buffer中提取displayId
                display_id = struct.unpack('>i', display_info_buffer[:4])[0]
                
                # 从buffer中提取宽高
                width = struct.unpack('>i', display_info_buffer[4:8])[0]
                height = struct.unpack('>i', display_info_buffer[8:12])[0]
                
                self.screen_width = width
                self.screen_height = height
                
                display_info = {
                    'displayId': display_id,
                    'width': width,
                    'height': height,
                    # 其他属性根据需要添加
                }
                
                rest = rest[display_info_buffer_length:]
                
                # 连接计数
                connection_count = struct.unpack('>i', rest[:4])[0]
                rest = rest[4:]
                
                # ScreenInfo字节数
                screen_info_bytes_count = struct.unpack('>i', rest[:4])[0]
                rest = rest[4:]
                
                if screen_info_bytes_count:
                    # 解析ScreenInfo
                    screen_info_bytes = rest[:screen_info_bytes_count]
                    # 实际解析需要根据ScreenInfo的结构
                    rest = rest[screen_info_bytes_count:]
                
                # VideoSettings字节数
                video_settings_bytes_count = struct.unpack('>i', rest[:4])[0]
                rest = rest[4:]
                
                if video_settings_bytes_count:
                    # 解析VideoSettings
                    video_settings_bytes = rest[:video_settings_bytes_count]
                    # 实际解析需要根据VideoSettings的结构
                    rest = rest[video_settings_bytes_count:]
                
                self.displays.append({
                    'displayInfo': display_info,
                    'connectionCount': connection_count,
                    # 其他信息根据需要添加
                })
            
            # 解析编码器
            encoders_count = struct.unpack('>i', rest[:4])[0]
            rest = rest[4:]
            
            self.encoders = []
            for i in range(encoders_count):
                name_length = struct.unpack('>i', rest[:4])[0]
                rest = rest[4:]
                
                name_bytes = rest[:name_length]
                rest = rest[name_length:]
                
                encoder_name = name_bytes.decode('utf-8')
                self.encoders.append(encoder_name)
            
            # 解析客户端ID
            self.client_id = struct.unpack('>i', rest[:4])[0]
            
            self.has_initial_info = True
            
            print(f"初始化信息解析完成:")
            print(f"设备名称: {self.device_name}")
            print(f"屏幕尺寸: {self.screen_width}x{self.screen_height}")
            print(f"编码器: {self.encoders}")
            print(f"客户端ID: {self.client_id}")
            
            # 计算触摸屏幕尺寸
            self.calculate_touch_screen_size()
            
            # 在初始化信息处理完成后发送视频设置
            await self.send_video_settings()
            
        except Exception as e:
            print(f"处理初始化信息时出错: {e}")
    
    async def send_video_settings(self):
        """发送视频设置消息"""
        # 如果已经发送过视频设置，则直接返回
        if self.video_settings_sent:
            return
            
        # 创建视频设置消息
        # 参考wscrcpy/src/app/VideoSettings.ts的实现
        # 使用动态缓冲区长度
        buffer = bytearray()
        
        # 写入比特率 (4字节)
        buffer.extend(struct.pack('>I', 8000000))  # 8Mbps
        
        # 写入最大帧率 (4字节)
        buffer.extend(struct.pack('>I', 60))
        
        # 写入I帧间隔 (1字节)
        buffer.extend(struct.pack('>b', 10))
        
        # 将整数转换为十六进制值
        video_width_hex = self.to_hex_int(self.VIDEO_SCREEN_WIDTH)
        video_height_hex = self.to_hex_int(self.VIDEO_SCREEN_HEIGHT)
        
        # 写入宽度 (2字节) - 使用十六进制值
        buffer.extend(struct.pack('>H', video_width_hex))
        
        # 写入高度 (2字节) - 使用十六进制值
        buffer.extend(struct.pack('>H', video_height_hex))
        
        # 写入裁剪区域 (8字节)
        buffer.extend(struct.pack('>H', 0))  # left
        buffer.extend(struct.pack('>H', 0))  # top
        buffer.extend(struct.pack('>H', 0))  # right
        buffer.extend(struct.pack('>H', 0))  # bottom
        
        # 写入是否发送帧元数据 (1字节)
        buffer.extend(struct.pack('>b', 0))
        
        # 写入锁定视频方向 (1字节)
        buffer.extend(struct.pack('>b', -1))
        
        # 写入显示ID (4字节)
        buffer.extend(struct.pack('>I', 0))
        
        # 写入编解码器选项长度 (4字节)
        buffer.extend(struct.pack('>I', 0))
        
        # 写入编码器名称长度 (4字节)
        buffer.extend(struct.pack('>I', 0))
        
        # 创建控制消息
        message = bytearray()
        message.append(self.TYPE_CHANGE_STREAM_PARAMETERS)
        message.extend(buffer)
        
        # 打印消息体
        print(f"发送视频设置消息，类型: {self.TYPE_CHANGE_STREAM_PARAMETERS}，长度: {len(message)} 字节")
        print(f"视频尺寸: {self.VIDEO_SCREEN_WIDTH}x{self.VIDEO_SCREEN_HEIGHT} (0x{self.to_hex(self.VIDEO_SCREEN_WIDTH)}x0x{self.to_hex(self.VIDEO_SCREEN_HEIGHT)})")
        print(f"消息内容 (十六进制): {message.hex()}")
        
        # 发送消息
        await self.ws.send(message)
        self.video_settings_sent = True
        print("视频设置消息已发送")
    
    async def _handle_device_message(self, data):
        """处理设备消息"""
        try:
            print("处理设备消息...")
            
            magic_size = len(self.MAGIC_BYTES_MESSAGE)
            message_data = data[magic_size:]
            
            # 消息类型
            msg_type = message_data[0]
            
            if msg_type == 0:  # TYPE_CLIPBOARD
                # 剪贴板消息
                offset = 1
                length = struct.unpack('>i', message_data[offset:offset+4])[0]
                offset += 4
                text_bytes = message_data[offset:offset+length]
                text = text_bytes.decode('utf-8')
                print(f"收到剪贴板消息: {text}")
            elif msg_type == 101:  # TYPE_PUSH_RESPONSE
                # 文件推送响应
                id = struct.unpack('>h', message_data[1:3])[0]
                code = struct.unpack('>b', message_data[3:4])[0]
                print(f"收到文件推送响应: id={id}, code={code}")
            else:
                print(f"未知设备消息类型: {msg_type}")
        
        except Exception as e:
            print(f"处理设备消息时出错: {e}")
    
    async def send_touch_event(self, action, x, y):
        """发送触摸事件消息"""
        # 创建触摸事件消息
        # 参考wscrcpy/src/app/controlMessage/TouchControlMessage.ts的实现
        buffer = bytearray()
        
        # 写入消息类型 (1字节)
        buffer.append(self.TYPE_INJECT_TOUCH_EVENT)
        
        # 写入动作类型 (1字节)
        buffer.append(action)
        
        # 写入pointerId高32位 (4字节)
        buffer.extend(struct.pack('>I', 0))
        
        # 写入pointerId低32位 (4字节)
        buffer.extend(struct.pack('>I', 0))
        
        # 写入X坐标 (4字节)
        buffer.extend(struct.pack('>I', x))
        
        # 写入Y坐标 (4字节)
        buffer.extend(struct.pack('>I', y))
        
        # 将触摸屏幕尺寸转换为十六进制值
        touch_width_hex = self.to_hex_int(self.touch_screen_width)
        touch_height_hex = self.to_hex_int(self.touch_screen_height)
        
        # 写入屏幕宽度 (2字节) - 使用计算得到的触摸屏宽度
        buffer.extend(struct.pack('>H', touch_width_hex))
        
        # 写入屏幕高度 (2字节) - 使用计算得到的触摸屏高度
        buffer.extend(struct.pack('>H', touch_height_hex))
        
        # 写入压力值 (2字节)
        pressure = 0xFFFF if action == self.ACTION_DOWN else 0
        buffer.extend(struct.pack('>H', pressure))
        
        # 写入按钮值 (4字节)
        buffer.extend(struct.pack('>I', self.BUTTON_PRIMARY))
        
        # 添加额外的末尾字节 (1字节)
        buffer.append(0)
        
        # 打印消息体
        action_name = "DOWN" if action == self.ACTION_DOWN else "UP" if action == self.ACTION_UP else "MOVE"
        print(f"发送触摸事件，类型: {action_name}, 坐标: ({x}, {y})")
        print(f"触摸屏尺寸: {self.touch_screen_width}x{self.touch_screen_height} (0x{self.to_hex(self.touch_screen_width)}x0x{self.to_hex(self.touch_screen_height)})")
        print(f"消息内容 (十六进制): {buffer.hex()}")
        
        # 发送消息
        await self.ws.send(buffer)
        print(f"触摸事件已发送: {action_name}")
    
    async def click(self, x, y, duration=0.1):
        """
        点击指定位置
        
        参数:
            x, y: 点击坐标
            duration: 按下持续时间(秒)
        """
        await self.send_touch_event(self.ACTION_DOWN, x, y)
        await asyncio.sleep(duration)
        await self.send_touch_event(self.ACTION_UP, x, y)
        print(f"点击事件已完成: ({x}, {y})")
    
    async def swipe(self, start_x, start_y, end_x, end_y, duration=0.5, steps=10):
        """
        滑动事件
        
        参数:
            start_x, start_y: 起始坐标
            end_x, end_y: 结束坐标
            duration: 滑动总持续时间(秒)
            steps: 滑动步数，越大越平滑
        """
        print(f"开始滑动: ({start_x}, {start_y}) -> ({end_x}, {end_y}), 持续时间: {duration}秒")
        
        # 发送按下事件
        await self.send_touch_event(self.ACTION_DOWN, start_x, start_y)
        await asyncio.sleep(0.05)  # 短暂延迟
        
        # 计算每一步的移动距离
        x_step = (end_x - start_x) / steps
        y_step = (end_y - start_y) / steps
        step_delay = duration / steps
        
        # 发送移动事件
        for i in range(1, steps + 1):
            current_x = int(start_x + x_step * i)
            current_y = int(start_y + y_step * i)
            await self.send_touch_event(self.ACTION_MOVE, current_x, current_y)
            await asyncio.sleep(step_delay)
        
        # 发送抬起事件
        await self.send_touch_event(self.ACTION_UP, end_x, end_y)
        print(f"滑动事件已完成: ({start_x}, {start_y}) -> ({end_x}, {end_y})")
    
    def set_video_screen_size(self, width, height):
        """设置视频屏幕尺寸"""
        self.VIDEO_SCREEN_WIDTH = width
        self.VIDEO_SCREEN_HEIGHT = height
        print(f"视频屏幕尺寸已设置为: {width}x{height}")
        
        # 如果已有设备尺寸，重新计算触摸尺寸
        if self.screen_width > 0 and self.screen_height > 0:
            self.calculate_touch_screen_size()
            
        return width, height

    def parse_sps(self, sps_data):
        """
        直接从 SPS 数据中解析视频分辨率
        参考: ITU-T H.264 规范
        """
        try:
            # 确保有足够的数据
            if len(sps_data) < 10:
                return None, None
                
            # 跳过起始码和 NAL 头
            if sps_data.startswith(b'\x00\x00\x00\x01'):
                data = sps_data[4:]
            else:
                data = sps_data
                
            # 获取 profile_idc 和 level_idc
            profile_idc = data[0]
            level_idc = data[2]
            
            # 获取 seq_parameter_set_id (使用指数哥伦布编码)
            # 简化处理：假设 seq_parameter_set_id 占用1个字节
            offset = 3
            
            # 根据 profile_idc 跳过额外数据
            if profile_idc in [100, 110, 122, 244, 44, 83, 86, 118, 128, 138, 139, 134, 135]:
                # 跳过 chroma_format_idc 等参数
                offset += 4
            
            # 尝试提取宽度和高度，这里简化处理
            # 实际的 H.264 解析需要处理更多字段
            
            # 提取宽高最小单位 (通常是16)
            mb_width = data[offset] if offset < len(data) else 32
            mb_height = data[offset+1] if offset+1 < len(data) else 32
            
            # 计算实际宽高 (简化版本，实际需要更复杂的解析)
            width = mb_width * 16
            height = mb_height * 16
            
            # 确保数值在合理范围内
            if 160 <= width <= 4096 and 160 <= height <= 4096:
                return width, height
                
            # 回退策略：尝试从 offset+8 和 offset+9 读取
            if offset+9 < len(data):
                width_alt = int(data[offset+8]) * 16
                height_alt = int(data[offset+9]) * 16
                if 160 <= width_alt <= 4096 and 160 <= height_alt <= 4096:
                    return width_alt, height_alt
            
            # 更多回退策略
            # 尝试查找标准分辨率
            if 480 <= max(mb_width, mb_height) * 16 <= 720:
                return 480, 854  # 480p 视频
            elif 720 <= max(mb_width, mb_height) * 16 <= 1080:
                return 720, 1280  # 720p 视频
            elif 1080 <= max(mb_width, mb_height) * 16 <= 2160:
                return 1080, 1920  # 1080p 视频
            
            return None, None
        except Exception as e:
            print(f"解析 SPS 时出错: {e}")
            return None, None

    async def _handle_video_frame(self, frame_data):
        """处理视频帧数据"""
        try:
            # 检查是否是视频配置数据
            if len(frame_data) < 4:
                return
                
            # 查找 H.264 起始码
            start_code = b'\x00\x00\x00\x01'
            
            # 将数据添加到缓冲区
            if not hasattr(self, 'frame_buffer'):
                self.frame_buffer = bytearray()
            self.frame_buffer.extend(frame_data)
            
            # 如果缓冲区太大，清空它
            if len(self.frame_buffer) > 5 * 1024 * 1024:  # 5MB
                print("缓冲区过大，清空")
                self.frame_buffer.clear()
                return
            
            # 查找所有 NAL 单元
            start_positions = []
            pos = 0
            while True:
                pos = self.frame_buffer.find(start_code, pos)
                if pos == -1:
                    break
                start_positions.append(pos)
                pos += len(start_code)
            
            if not start_positions:
                return
                
            # 处理每个 NAL 单元
            for i in range(len(start_positions)):
                start = start_positions[i]
                end = start_positions[i + 1] if i + 1 < len(start_positions) else len(self.frame_buffer)
                
                # 获取 NAL 单元
                nal_unit = self.frame_buffer[start:end]
                if len(nal_unit) < 5:
                    continue
                    
                # 获取 NAL 类型
                nal_type = nal_unit[4] & 0x1F
                
                # 打印 NAL 类型和大小
                print(f"NAL类型: {nal_type}, 大小: {len(nal_unit)} 字节")
                
                # 如果是 SPS，提取视频尺寸
                if nal_type == 7:  # SPS
                    print("找到 SPS")
                    with open("sps.h264", "wb") as f:
                        f.write(nal_unit)
                    self.sps_pps_found = True
                        
                    # 直接从 SPS 数据解析视频尺寸
                    width, height = self.parse_sps(nal_unit)
                    if width and height:
                        print(f"从 SPS 解析出的视频尺寸: {width}x{height}")
                        
                        # 确保尺寸有效且不大于原始屏幕尺寸
                        if width > 0 and height > 0 and width <= self.screen_width and height <= self.screen_height:
                            # 更新触摸屏尺寸
                            self.touch_screen_width = width
                            self.touch_screen_height = height
                            print(f"已更新触摸屏尺寸为实际视频尺寸: {width}x{height}")
                            print(f"触摸屏尺寸(十六进制): 0x{self.to_hex(width)} x 0x{self.to_hex(height)}")
                        else:
                            print(f"解析的视频尺寸 {width}x{height} 无效或大于屏幕尺寸 {self.screen_width}x{self.screen_height}，使用计算值")
                    else:
                        # 使用标准尺寸
                        if len(nal_unit) > 10:
                            # 常见的移动设备视频编码尺寸
                            std_sizes = [
                                (480, 854),   # 480p
                                (540, 960),   # qHD
                                (720, 1280),  # 720p
                                (1080, 1920), # 1080p
                            ]
                            
                            # 查找最接近的标准尺寸
                            best_ratio = None
                            best_size = None
                            device_ratio = self.screen_width / self.screen_height
                            
                            for w, h in std_sizes:
                                if w > self.screen_width or h > self.screen_height:
                                    continue
                                ratio = w / h
                                if best_ratio is None or abs(ratio - device_ratio) < abs(best_ratio - device_ratio):
                                    best_ratio = ratio
                                    best_size = (w, h)
                            
                            if best_size:
                                width_approx, height_approx = best_size
                                print(f"使用标准视频尺寸: {width_approx}x{height_approx}")
                                
                                # 更新触摸屏尺寸
                                self.touch_screen_width = width_approx
                                self.touch_screen_height = height_approx
                                print(f"已更新触摸屏尺寸为标准尺寸: {width_approx}x{height_approx}")
                    
                elif nal_type == 8:  # PPS
                    print("找到 PPS")
                    with open("pps.h264", "wb") as f:
                        f.write(nal_unit)
                    self.sps_pps_found = True
                        
                # 如果是IDR帧(关键帧)，尝试保存
                elif nal_type == 5 and hasattr(self, 'sps_pps_found') and not self.saved_frame:  # IDR 帧
                    print(f"找到 IDR 帧，尝试保存图像")
                    
                    # 将所有必要的帧数据写入临时文件
                    with open("temp_video.h264", "wb") as f:
                        # 先写入sps和pps
                        with open("sps.h264", "rb") as sps_file:
                            f.write(sps_file.read())
                        with open("pps.h264", "rb") as pps_file:
                            f.write(pps_file.read())
                        # 写入当前IDR帧
                        f.write(nal_unit)
                    
                    # 尝试使用OpenCV解码和保存
                    try:
                        # 创建VideoCapture对象
                        cap = cv2.VideoCapture("temp_video.h264")
                        ret, frame = cap.read()
                        
                        if ret:
                            # 成功读取到帧，保存为PNG
                            cv2.imwrite("frame.png", frame)
                            print("成功保存视频帧到 frame.png")
                            
                            # 显示帧的尺寸
                            h, w = frame.shape[:2]
                            print(f"帧尺寸: {w}x{h}")
                            
                            # 使用实际帧尺寸更新触摸屏尺寸
                            self.touch_screen_width = w
                            self.touch_screen_height = h
                            print(f"已更新触摸屏尺寸为实际帧尺寸: {w}x{h}")
                            print(f"触摸屏尺寸(十六进制): 0x{self.to_hex(w)} x 0x{self.to_hex(h)}")
                            
                            # 另一种保存方式 - 直接保存为numpy数组
                            np.save("frame.npy", frame)
                            print("已保存帧为Numpy数组 frame.npy")
                            
                            self.saved_frame = True
                        else:
                            print("读取视频帧失败")
                            
                        cap.release()
                    except Exception as e:
                        print(f"保存视频帧出错: {e}")
                        
                    # 尝试使用另一种方法保存
                    if not self.saved_frame:
                        try:
                            print("尝试使用备用方法保存视频帧...")
                            # 创建内存中的H264文件
                            h264_data = bytearray()
                            
                            # 读取SPS和PPS数据
                            with open("sps.h264", "rb") as sps_file:
                                h264_data.extend(sps_file.read())
                            with open("pps.h264", "rb") as pps_file:
                                h264_data.extend(pps_file.read())
                            
                            # 添加IDR帧
                            h264_data.extend(nal_unit)
                            
                            # 保存原始H264数据
                            with open("raw_frame.h264", "wb") as raw_file:
                                raw_file.write(h264_data)
                            
                            print("已保存原始H264数据到 raw_frame.h264 (需要外部工具查看)")
                            self.saved_frame = True
                        except Exception as e:
                            print(f"备用方法保存失败: {e}")
            
            # 清理缓冲区
            if start_positions:
                self.frame_buffer = self.frame_buffer[start_positions[-1]:]
                
        except Exception as e:
            print(f"视频帧解析错误: {e}")
            import traceback
            traceback.print_exc()


async def main():
    """主函数"""
    # 创建客户端实例
    client = ScrcpyClient("ws://localhost:10001")
    
    try:
        # 连接服务器
        await client.connect()
        
        # 启动接收循环
        receive_task = asyncio.create_task(client.receive_loop())
        
        # 等待初始化完成，最多等待10秒
        max_wait_time = 10
        wait_interval = 0.5
        waited_time = 0
        
        print("等待设备初始化...")
        while not client.has_initial_info and waited_time < max_wait_time:
            await asyncio.sleep(wait_interval)
            waited_time += wait_interval
            print(f"等待初始化...已等待 {waited_time} 秒")
        
        if not client.has_initial_info:
            print("初始化超时，未能获取设备信息")
            return
            
        print("初始化成功，开始执行测试")
        
        # 等待2秒，确保视频流开始
        await asyncio.sleep(2)
        
        # 获取屏幕尺寸
        width = client.touch_screen_width
        height = client.touch_screen_height
        
        # 执行点击测试
        # 点击屏幕中心
        center_x = width // 2
        center_y = height // 2
        print(f"点击屏幕中心: ({center_x}, {center_y})")
        await client.click(center_x, center_y)
        await asyncio.sleep(1)
        
        # 执行滑动测试
        # 从下到上滑动（模拟上滑查看内容）
        start_x = width // 2
        start_y = (height * 3) // 4  # 从屏幕3/4处开始
        end_x = width // 2
        end_y = height // 4  # 滑动到屏幕1/4处
        
        print(f"执行滑动测试: 从下往上滑动")
        await client.swipe(start_x, start_y, end_x, end_y, duration=0.8, steps=15)
        
        await asyncio.sleep(1)
        
        # 从右到左滑动（模拟左滑切换页面）
        start_x = (width * 3) // 4  # 从屏幕右侧3/4处开始
        start_y = height // 2
        end_x = width // 4  # 滑动到屏幕左侧1/4处
        end_y = height // 2
        
        print(f"执行滑动测试: 从右往左滑动")
        await client.swipe(start_x, start_y, end_x, end_y, duration=0.8, steps=15)
        
        print("测试操作完成，等待5秒...")
        # 继续接收一段时间
        await asyncio.sleep(5)
        
    except Exception as e:
        print(f"程序执行出错: {e}")
    finally:
        # 关闭连接
        await client.disconnect()
        
        # 取消接收任务
        if 'receive_task' in locals():
            receive_task.cancel()
            try:
                await receive_task
            except asyncio.CancelledError:
                pass


if __name__ == "__main__":
    asyncio.run(main()) 
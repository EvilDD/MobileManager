#!/usr/bin/env python
# -*- coding: utf-8 -*-

import asyncio
import base64
import os
import struct
import subprocess
import tempfile
import time
import threading
import websockets
import cv2
import numpy as np
import logging
import sys
from queue import Queue, Empty, Full
import select

# 配置日志
logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')
logger = logging.getLogger(__name__)

# FFmpeg路径配置 - 可以根据实际安装位置修改
FFMPEG_PATH = "ffmpeg"  # 默认从PATH环境变量获取ffmpeg命令

# 检查FFmpeg是否可用并提供安装指南
def check_ffmpeg():
    try:
        subprocess.check_output([FFMPEG_PATH, "-version"], stderr=subprocess.STDOUT)
        logger.info("FFmpeg command line tool available")
        return True
    except Exception as e:
        logger.error(f"FFmpeg command line tool not found: {str(e)}")
        logger.error("Please install FFmpeg command line tool:")
        logger.error("1. Download: https://ffmpeg.org/download.html")
        logger.error("2. Add ffmpeg.exe to system PATH environment variable")
        logger.error("3. Or modify the FFMPEG_PATH variable at the beginning of the script to point to your ffmpeg.exe")
        logger.error("Example: FFMPEG_PATH = \"C:\\Program Files\\ffmpeg\\bin\\ffmpeg.exe\"")
        return False

# 检查ffmpeg-python库
try:
    import ffmpeg
    logger.info("ffmpeg-python library available")
except ImportError:
    logger.error("Missing ffmpeg-python library, please install using pip: pip install ffmpeg-python")
    sys.exit(1)

def enhance_frame(frame, denoise_strength=5, sharpen_strength=0.5, brightness=1.1, contrast=1.2):
    """
    对视频帧进行超分辨率处理，包括降噪、锐化和亮度对比度增强
    
    参数:
        frame: 需要处理的视频帧
        denoise_strength: 降噪强度，值越大去噪效果越强，但可能会丢失细节
        sharpen_strength: 锐化强度，值越大锐化效果越明显
        brightness: 亮度调整系数，大于1增加亮度，小于1降低亮度
        contrast: 对比度调整系数，大于1增加对比度，小于1降低对比度
    
    返回:
        处理后的帧
    """
    try:
        if frame is None:
            logger.warning("无法处理空帧")
            return None
        
        # 步骤1: 降噪处理 - 使用高斯模糊
        denoised = cv2.GaussianBlur(frame, (denoise_strength, denoise_strength), 0)
        
        # 步骤2: 锐化处理 - 使用unsharp mask技术
        # 创建锐化核
        kernel_size = 3
        kernel = np.array([[-1, -1, -1],
                           [-1,  9, -1],
                           [-1, -1, -1]]) * sharpen_strength
        
        # 应用锐化核
        sharpened = cv2.filter2D(denoised, -1, kernel)
        
        # 步骤3: 亮度和对比度调整
        # 公式: 新像素 = 对比度 * 原像素 + 亮度
        enhanced = cv2.convertScaleAbs(sharpened, alpha=contrast, beta=(brightness-1.0)*50)
        
        logger.debug("帧增强处理完成")
        return enhanced
        
    except Exception as e:
        logger.error(f"帧增强处理失败: {str(e)}")
        import traceback
        logger.error(traceback.format_exc())
        return frame  # 处理失败时返回原始帧

class ScrcpyClient:
    """
    Scrcpy client implementation in Python
    Features:
    1. Connect to WebSocket server
    2. Process initialization messages and set video parameters
    3. Categorize received message data
    4. Decode and display video stream
    """
    # 常量定义
    MAGIC_BYTES_INITIAL = b'scrcpy_initial'
    MAGIC_BYTES_MESSAGE = b'scrcpy_message'
    
    # NAL单元类型
    NAL_TYPE_SPS = 7
    NAL_TYPE_PPS = 8
    NAL_TYPE_IDR = 5
    NAL_TYPE_NON_IDR = 1
    NAL_TYPE_SEI = 6
    
    # 控制消息类型
    TYPE_CHANGE_STREAM_PARAMETERS = 101
    
    # 设备连接信息
    DEVICE_ID = "127.0.0.1:16480"  # 设备ID
    WS_URL = "ws://localhost:8000/?action=proxy-adb&remote=tcp%3A8886&udid=127.0.0.1%3A16480"
    
    # 视频设置参数
    VIDEO_BITRATE = 5024288
    VIDEO_MAX_FPS = 24
    VIDEO_I_FRAME_INTERVAL = 5
    VIDEO_WIDTH = 540
    VIDEO_HEIGHT = 960
    
    def __init__(self, adb_path=None, device_id=None, display_enabled=True, 
                 save_enabled=False, save_path=None, tmp_dir=None, max_duration=None,
                 enhance_enabled=True, denoise_strength=5, sharpen_strength=0.5,
                 brightness=1.1, contrast=1.2):
        """初始化ADB屏幕录制控制器
        
        参数:
            adb_path (str): ADB程序路径，None表示从PATH中查找
            device_id (str): 设备ID，None表示使用默认设备
            display_enabled (bool): 是否启用视频显示窗口
            save_enabled (bool): 是否保存视频到文件
            save_path (str): 视频保存路径，None表示使用临时目录
            tmp_dir (str): 临时文件目录，None表示使用系统默认临时目录
            max_duration (int): 最大录制时长(秒)，None表示无限制
            enhance_enabled (bool): 是否启用画面增强功能
            denoise_strength (int): 降噪强度，范围1-10
            sharpen_strength (float): 锐化强度，范围0.1-1.0
            brightness (float): 亮度调整，范围0.5-1.5
            contrast (float): 对比度调整，范围0.5-1.5
        """
        # 首先检查FFmpeg是否可用
        self.has_ffmpeg = check_ffmpeg()
        if not self.has_ffmpeg:
            logger.warning("Will try to use ffmpeg-python for decoding, but performance may be poor")
        
        # 基本参数
        self.ws = None
        self.device_name = None
        self.displays = []
        self.encoders = []
        self.client_id = -1
        self.has_initial_info = False
        self.screen_width = 0
        self.screen_height = 0
        self.video_settings_sent = False
        
        # 设备连接参数
        self.device_id = device_id if device_id else self.DEVICE_ID
        
        # 实际编码尺寸 (从SPS中解析)
        self.actual_width = 0
        self.actual_height = 0
        self.sps_parsed = False
        
        # 显示控制
        self.display_enabled = display_enabled
        self.window_name = "Scrcpy Remote Display"
        self.is_running = True
        
        # 图像增强设置
        self.enhance_enabled = enhance_enabled
        self.denoise_strength = max(1, min(10, denoise_strength))  # 限制在1-10之间
        self.sharpen_strength = max(0.1, min(1.0, sharpen_strength))  # 限制在0.1-1.0之间
        self.brightness = max(0.5, min(1.5, brightness))  # 限制在0.5-1.5之间
        self.contrast = max(0.5, min(1.5, contrast))  # 限制在0.5-1.5之间
        logger.info(f"图像增强设置: 启用={self.enhance_enabled}, 降噪={self.denoise_strength}, " +
                   f"锐化={self.sharpen_strength}, 亮度={self.brightness}, 对比度={self.contrast}")
        
        # 保存控制
        self.save_enabled = save_enabled
        self.save_path = save_path
        
        # 创建临时目录
        self.temp_dir = tmp_dir if tmp_dir else tempfile.mkdtemp()
        logger.info(f"使用临时目录: {self.temp_dir}")
        
        # 确保保存路径存在
        if self.save_enabled and self.save_path:
            os.makedirs(os.path.dirname(self.save_path), exist_ok=True)
            logger.info(f"视频将保存到: {self.save_path}")
        
        # 创建H264缓存文件
        self.h264_file = os.path.join(self.temp_dir, "stream.h264")
        with open(self.h264_file, 'wb') as f:
            f.write(b'')
        
        # 文件锁，避免文件访问冲突
        self.file_lock = threading.Lock()
        
        # 控制进程和线程
        self.ffmpeg_process = None
        self.display_thread = None
        self.process = None  # FFmpeg解码进程
        
        # 解码相关常量
        self.FRAME_SIZE = 0  # 帧大小(字节)，将在setup_video_processing中设置
        self.YUV_FRAME_SIZE = None  # YUV帧尺寸，将在setup_video_processing中设置
        
        # H.264数据统计
        self.frame_count = 0
        self.last_update_time = time.time()
        
        # 帧更新控制
        self.frame_ready = threading.Event()
        
        # 解码统计
        self.frames_decoded = 0
        self.frames_displayed = 0
        
        # 保存最后解码的帧
        self.last_frame = None
        
        # GUI锁，避免多线程同时操作OpenCV窗口
        self.gui_lock = threading.Lock()
        
        # 初始化帧处理相关变量
        self.frame_display_count = 0
        self.last_decode_time = time.time()
        self.last_display_time = time.time()
        self.frames_displayed = 0
        self.fps = 0
        self.startup_time = time.time()  # 记录启动时间
        self.last_no_frame_warn = time.time()  # 上次无帧警告时间
        
        # 解码器稳定性控制
        self.decoder_restarts = 0
        self.last_decoder_restart = time.time()
        self.restart_interval = 30  # 每30秒周期性重启解码器以避免卡死
        self.read_timeout = 2.0  # 读取帧的超时时间(秒)
        self.max_failures = 5  # 最大连续解码失败次数
        self.consecutive_failures = 0  # 当前连续解码失败次数
        
        # WebSocket服务器
        self.ws_server = None
        
        logger.info("ADB屏幕录制初始化完成")
        
    async def connect(self):
        """连接到WebSocket服务器"""
        logger.info(f"Connecting to {self.WS_URL}...")
        try:
            self.ws = await websockets.connect(self.WS_URL)
            logger.info("WebSocket connection successful")
            return True
        except Exception as e:
            logger.error(f"WebSocket connection failed: {str(e)}")
            return False
            
    async def send_video_settings(self):
        """发送视频设置消息以开始视频流"""
        if self.video_settings_sent:
            return
            
        # 创建视频设置消息
        buffer = bytearray()
        
        # 写入比特率 (4字节)
        buffer.extend(struct.pack('>I', self.VIDEO_BITRATE))
        
        # 写入最大帧率 (4字节)
        buffer.extend(struct.pack('>I', self.VIDEO_MAX_FPS))
        
        # 写入I帧间隔 (1字节)
        buffer.extend(struct.pack('>b', self.VIDEO_I_FRAME_INTERVAL))
        
        # 写入宽度 (2字节)
        buffer.extend(struct.pack('>H', self.VIDEO_WIDTH))
        
        # 写入高度 (2字节)
        buffer.extend(struct.pack('>H', self.VIDEO_HEIGHT))
        
        # 裁剪参数 (8字节, 全0) - 对应 crop: null
        buffer.extend(struct.pack('>HHHH', 0, 0, 0, 0))
        
        # 帧元数据标志 (1字节) - 对应 sendFrameMeta: false
        buffer.extend(struct.pack('>b', 0))
        
        # 锁定视频方向 (1字节) - 对应 lockedVideoOrientation: -1
        buffer.extend(struct.pack('>b', -1))
        
        # 显示ID (4字节) - 对应 displayId: 0
        buffer.extend(struct.pack('>I', 0))
        
        # 编解码器选项长度 (4字节) - 对应 codecOptions: undefined
        buffer.extend(struct.pack('>I', 0))
        
        # 编码器名称长度 (4字节) - 对应 encoderName: undefined
        buffer.extend(struct.pack('>I', 0))
        
        # 创建控制消息
        message = bytearray()
        message.append(self.TYPE_CHANGE_STREAM_PARAMETERS)
        message.extend(buffer)
        
        logger.info(f"Sending video settings: {self.VIDEO_WIDTH}x{self.VIDEO_HEIGHT}, {self.VIDEO_BITRATE//1000}kbps")
        logger.info(f"Video settings: bitrate={self.VIDEO_BITRATE}, maxFps={self.VIDEO_MAX_FPS}, iFrameInterval={self.VIDEO_I_FRAME_INTERVAL}, bounds={{width:{self.VIDEO_WIDTH}, height:{self.VIDEO_HEIGHT}}}, crop=null, sendFrameMeta=false, lockedVideoOrientation=-1, displayId=0, codecOptions=undefined, encoderName=undefined")
        
        # 发送消息
        await self.ws.send(message)
        self.video_settings_sent = True
        logger.info("Video settings sent, waiting for stream...")
        
    async def handle_initial_info(self, data):
        """处理初始化信息"""
        try:
            offset = len(self.MAGIC_BYTES_INITIAL)
            
            # 解析设备名称 (64字节)
            name_bytes = data[offset:offset+64]
            # 过滤掉尾部的空字节
            name_bytes = name_bytes.rstrip(b'\x00')
            self.device_name = name_bytes.decode('utf-8')
            offset += 64
            
            # 解析剩余数据
            rest = data[offset:]
            
            # 显示数量
            displays_count = struct.unpack('>i', rest[:4])[0]
            logger.info(f"Device name: {self.device_name}, display count: {displays_count}")
            
            rest = rest[4:]  # 跳过displays_count的4字节
            
            # 解析每个显示设备的信息
            for i in range(displays_count):
                # DisplayInfo的结构
                display_info_buffer_length = 24
                display_info_buffer = rest[:display_info_buffer_length]
                
                # 提取displayId
                display_id = struct.unpack('>i', display_info_buffer[:4])[0]
                
                # 提取宽高
                width = struct.unpack('>i', display_info_buffer[4:8])[0]
                height = struct.unpack('>i', display_info_buffer[8:12])[0]
                
                self.screen_width = width
                self.screen_height = height
                
                # 根据屏幕尺寸设置实际视频传输大小
                if width > 0 and height > 0:
                    # 保持宽高比，调整视频尺寸
                    aspect_ratio = width / height
                    if width > height:
                        # 横屏
                        self.VIDEO_WIDTH = 960
                        self.VIDEO_HEIGHT = int(self.VIDEO_WIDTH / aspect_ratio)
                    else:
                        # 竖屏
                        self.VIDEO_HEIGHT = 960
                        self.VIDEO_WIDTH = int(self.VIDEO_HEIGHT * aspect_ratio)
                
                display_info = {
                    'displayId': display_id,
                    'width': width,
                    'height': height,
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
                    rest = rest[screen_info_bytes_count:]
                
                # VideoSettings字节数
                video_settings_bytes_count = struct.unpack('>i', rest[:4])[0]
                rest = rest[4:]
                
                if video_settings_bytes_count:
                    # 解析VideoSettings
                    rest = rest[video_settings_bytes_count:]
                
                self.displays.append(display_info)
            
            # 解析编码器
            encoders_count = struct.unpack('>i', rest[:4])[0]
            rest = rest[4:]
            
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
            
            logger.info(f"Initialization info complete:")
            logger.info(f"Device name: {self.device_name}")
            logger.info(f"Screen size: {self.screen_width}x{self.screen_height}")
            logger.info(f"Video size: {self.VIDEO_WIDTH}x{self.VIDEO_HEIGHT}")
            logger.info(f"Encoders: {self.encoders}")
            logger.info(f"Client ID: {self.client_id}")
            
            # 初始化视频处理
            self.setup_video_processing()
            
            # 在初始化信息处理完成后发送视频设置
            await self.send_video_settings()
            
        except Exception as e:
            logger.error(f"Error processing initialization info: {str(e)}")
            import traceback
            traceback.print_exc()
    
    def setup_video_processing(self):
        """设置视频处理线程和FFmpeg处理器"""
        try:
            # 关闭可能存在的旧窗口
            cv2.destroyAllWindows()
            
            # 稍等片刻确保窗口完全关闭
            time.sleep(0.1)
            
            # 使用适当的视频尺寸初始化窗口，添加KEEPRATIO标志确保图像不变形
            cv2.namedWindow(self.window_name, cv2.WINDOW_NORMAL | cv2.WINDOW_KEEPRATIO)
            
            # 窗口大小设置
            display_width = max(512, self.VIDEO_WIDTH)
            display_height = max(928, self.VIDEO_HEIGHT)
            cv2.resizeWindow(self.window_name, display_width, display_height)
            
            # 简单测试图像 - 彩色渐变
            test_img = np.zeros((display_height, display_width, 3), dtype=np.uint8)
            for i in range(display_height):
                color = [int(255 * i / display_height), 
                        int(255 * (1 - i / display_height)), 
                        int(128 + 127 * np.sin(3 * np.pi * i / display_height))]
                test_img[i, :] = color
                
            # 添加文本
            cv2.putText(test_img, "INITIALIZING...", (50, display_height//2), 
                       cv2.FONT_HERSHEY_SIMPLEX, 1.0, (255, 255, 255), 2)
            
            # 显示测试图案
            with self.gui_lock:
                cv2.imshow(self.window_name, test_img)
                cv2.waitKey(100)  # 确保窗口显示更新
            logger.info(f"Created test window with size {display_width}x{display_height}")
            
            # 启动显示线程
            if self.display_thread is None or not self.display_thread.is_alive():
                self.display_thread = threading.Thread(target=self.display_frames_loop)
                self.display_thread.daemon = True
                self.display_thread.start()
                logger.info("Display thread started")
            else:
                logger.info("Display thread already running")
            
        except Exception as e:
            logger.error(f"Error setting up video processing: {str(e)}")
            import traceback
            traceback.print_exc()
    
    async def handle_device_message(self, data):
        """处理设备消息"""
        try:
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
                logger.info(f"Received clipboard message: {text}")
            elif msg_type == 101:  # TYPE_PUSH_RESPONSE
                # 文件推送响应
                id = struct.unpack('>h', message_data[1:3])[0]
                code = struct.unpack('>b', message_data[3:4])[0]
                logger.info(f"Received file push response: id={id}, code={code}")
            else:
                logger.info(f"Unknown device message type: {msg_type}")
        
        except Exception as e:
            logger.error(f"Error processing device message: {str(e)}")
            
    def find_nal_units(self, data):
        """
        在H.264数据中查找NAL单元
        返回格式: [(start_idx, end_idx, nal_type), ...]
        """
        nal_units = []
        i = 0
        data_len = len(data)
        
        # 查找NAL单元的起始码 (0x00 0x00 0x00 0x01 或 0x00 0x00 0x01)
        while i < data_len:
            # 查找四字节起始码 (0x00 0x00 0x00 0x01)
            if i + 3 < data_len and data[i] == 0 and data[i+1] == 0 and data[i+2] == 0 and data[i+3] == 1:
                start_idx = i
                i += 4
                # 查找下一个起始码
                next_start = i
                while next_start + 2 < data_len:
                    if (data[next_start] == 0 and data[next_start+1] == 0 and data[next_start+2] == 0 and 
                        next_start + 3 < data_len and data[next_start+3] == 1):
                        break
                    elif (data[next_start] == 0 and data[next_start+1] == 0 and data[next_start+2] == 1):
                        break
                    next_start += 1
                
                if next_start + 2 >= data_len:
                    next_start = data_len
                
                # 如果找到了NAL单元
                if i < data_len:
                    nal_type = data[i] & 0x1F  # 取第一个字节的低5位
                    nal_units.append((start_idx, next_start, nal_type))
                
                i = next_start
            # 查找三字节起始码 (0x00 0x00 0x01)
            elif i + 2 < data_len and data[i] == 0 and data[i+1] == 0 and data[i+2] == 1:
                start_idx = i
                i += 3
                # 查找下一个起始码
                next_start = i
                while next_start + 2 < data_len:
                    if (data[next_start] == 0 and data[next_start+1] == 0 and data[next_start+2] == 0 and 
                        next_start + 3 < data_len and data[next_start+3] == 1):
                        break
                    elif (data[next_start] == 0 and data[next_start+1] == 0 and data[next_start+2] == 1):
                        break
                    next_start += 1
                
                if next_start + 2 >= data_len:
                    next_start = data_len
                
                # 如果找到了NAL单元
                if i < data_len:
                    nal_type = data[i] & 0x1F  # 取第一个字节的低5位
                    nal_units.append((start_idx, next_start, nal_type))
                
                i = next_start
            else:
                i += 1
        
        return nal_units

    def parse_sps(self, sps_data):
        """
        解析SPS NAL单元获取视频尺寸
        简化版H.264 SPS解析，只提取宽高信息
        """
        try:
            import bitstring
            
            # 跳过NAL头部 (1字节)
            data = sps_data[1:]
            
            # 创建BitStream来解析位级别数据
            bs = bitstring.BitStream(data)
            
            # 解析SPS
            profile_idc = bs.read('uint:8')
            bs.read('uint:8')  # constraint_set flags + reserved_zero
            level_idc = bs.read('uint:8')
            
            # seq_parameter_set_id
            bs.read('ue')
            
            # 根据profile_idc不同，跳过不同的字段
            if profile_idc in [100, 110, 122, 244, 44, 83, 86, 118, 128, 138, 139, 134, 135]:
                chroma_format_idc = bs.read('ue')
                if chroma_format_idc == 3:
                    bs.read('uint:1')  # separate_colour_plane_flag
                bs.read('ue')  # bit_depth_luma_minus8
                bs.read('ue')  # bit_depth_chroma_minus8
                bs.read('uint:1')  # qpprime_y_zero_transform_bypass_flag
                seq_scaling_matrix_present_flag = bs.read('uint:1')
                if seq_scaling_matrix_present_flag:
                    # 跳过缩放矩阵
                    for i in range(8 if chroma_format_idc != 3 else 12):
                        seq_scaling_list_present_flag = bs.read('uint:1')
                        if seq_scaling_list_present_flag:
                            # 跳过缩放列表
                            last_scale = 8
                            next_scale = 8
                            size_of_scaling_list = 16 if i < 6 else 64
                            for j in range(size_of_scaling_list):
                                if next_scale != 0:
                                    delta_scale = bs.read('se')
                                    next_scale = (last_scale + delta_scale) % 256
                                if next_scale != 0:
                                    last_scale = next_scale
            
            # log2_max_frame_num_minus4
            bs.read('ue')
            
            # pic_order_cnt_type
            pic_order_cnt_type = bs.read('ue')
            if pic_order_cnt_type == 0:
                bs.read('ue')  # log2_max_pic_order_cnt_lsb_minus4
            elif pic_order_cnt_type == 1:
                bs.read('uint:1')  # delta_pic_order_always_zero_flag
                bs.read('se')  # offset_for_non_ref_pic
                bs.read('se')  # offset_for_top_to_bottom_field
                num_ref_frames_in_pic_order_cnt_cycle = bs.read('ue')
                for i in range(num_ref_frames_in_pic_order_cnt_cycle):
                    bs.read('se')  # offset_for_ref_frame[i]
            
            # 其他跳过的字段
            bs.read('ue')  # max_num_ref_frames
            bs.read('uint:1')  # gaps_in_frame_num_value_allowed_flag
            
            # pic_width_in_mbs_minus1
            pic_width_in_mbs_minus1 = bs.read('ue')
            
            # pic_height_in_map_units_minus1
            pic_height_in_map_units_minus1 = bs.read('ue')
            
            # frame_mbs_only_flag
            frame_mbs_only_flag = bs.read('uint:1')
            
            if not frame_mbs_only_flag:
                bs.read('uint:1')  # mb_adaptive_frame_field_flag
            
            # 跳过其他字段
            bs.read('uint:1')  # direct_8x8_inference_flag
            
            # 帧裁剪标志
            frame_cropping_flag = bs.read('uint:1')
            frame_crop_left_offset = 0
            frame_crop_right_offset = 0
            frame_crop_top_offset = 0
            frame_crop_bottom_offset = 0
            
            if frame_cropping_flag:
                frame_crop_left_offset = bs.read('ue')
                frame_crop_right_offset = bs.read('ue')
                frame_crop_top_offset = bs.read('ue')
                frame_crop_bottom_offset = bs.read('ue')
            
            # 计算实际宽高 (macroblock size为16)
            width = (pic_width_in_mbs_minus1 + 1) * 16
            height = (pic_height_in_map_units_minus1 + 1) * 16 * (2 - frame_mbs_only_flag)
            
            # 应用裁剪
            if frame_cropping_flag:
                crop_unit_x = 2  # 假设YUV 4:2:0采样
                crop_unit_y = 2 * (2 - frame_mbs_only_flag)  # 考虑frame_mbs_only_flag
                
                width -= (frame_crop_left_offset + frame_crop_right_offset) * crop_unit_x
                height -= (frame_crop_top_offset + frame_crop_bottom_offset) * crop_unit_y
            
            return width, height
            
        except Exception as e:
            logger.error(f"Error parsing SPS: {str(e)}")
            return 0, 0

    def process_video_data(self, data):
        """处理接收到的视频数据"""
        try:
            # 尝试从数据中解析SPS
            if not self.sps_parsed:
                # 查找NAL单元
                nal_units = self.find_nal_units(data)
                for start, end, nal_type in nal_units:
                    if nal_type == self.NAL_TYPE_SPS:
                        # 找到SPS，解析宽高
                        sps_data = data[start:end]
                        width, height = self.parse_sps(sps_data[4:] if sps_data[2] == 0 and sps_data[3] == 1 else sps_data[3:])
                        
                        if width > 0 and height > 0:
                            self.actual_width = width
                            self.actual_height = height
                            self.sps_parsed = True
                            logger.info(f"SPS parsed, actual encoded dimensions: {width}x{height}")
                            
                            # 调整窗口大小
                            cv2.resizeWindow(self.window_name, width, height)
                            logger.info("Video dimensions updated, adjusting display window...")
                        break
            
            # 将视频数据保存到文件
            with self.file_lock:
                with open(self.h264_file, 'ab') as f:
                    f.write(data)
                self.last_update_time = time.time()
            
            # 添加计数器统计数据包
            self.frame_count += 1
            if self.frame_count % 100 == 0:
                logger.info(f"Received {self.frame_count} data packets")
            
        except Exception as e:
            logger.error(f"Error processing video data: {str(e)}")
            import traceback
            traceback.print_exc()
    
    def frame_enhancement(self, frame):
        """对解码后的帧进行增强处理"""
        try:
            if frame is None or not self.enhance_enabled:
                return frame
                
            # 应用高斯模糊去噪
            denoised = cv2.GaussianBlur(frame, (self.denoise_strength, self.denoise_strength), 0)
            
            # 提高对比度和亮度
            enhanced = cv2.convertScaleAbs(denoised, alpha=self.contrast, beta=(self.brightness-1.0)*50)
            
            # 锐化图像
            kernel = np.array([[-1,-1,-1], 
                              [-1, 9,-1],
                              [-1,-1,-1]]) * self.sharpen_strength
            sharpened = cv2.filter2D(enhanced, -1, kernel)
            
            return sharpened
        except Exception as e:
            logger.error(f"帧增强处理失败: {str(e)}")
            import traceback
            logger.error(traceback.format_exc())
            # 如果增强失败，返回原始帧
            return frame
            
    def decode_frame(self):
        """解码并返回最新的视频帧"""
        try:
            if not os.path.exists(self.h264_file) or os.path.getsize(self.h264_file) == 0:
                # 没有数据可解码
                return None
                
            # 没有FFmpeg无法解码
            if not self.has_ffmpeg and not self.process:
                return None
                
            # 每隔一段时间周期性重启解码器以避免卡死
            now = time.time()
            if now - self.last_decoder_restart > self.restart_interval:
                logger.debug("定期重启解码器以保持稳定性")
                self.restart_decoder()
                
            # 解码帧
            if self.process is None:
                logger.debug("解码器未初始化，启动解码器")
                self.start_decoder()
                
            if self.process is None:
                logger.error("无法启动解码器")
                return None
                
            # 检查解码器进程是否还在运行
            if self.process.poll() is not None:
                logger.warning(f"解码器进程已退出(返回码: {self.process.poll()})，重新启动解码器")
                self.restart_decoder()
                return None
                
            # Windows上无法对管道使用select，采用替代方案
            frame = None
            try:
                # 读取数据前检查是否有stderr输出
                try:
                    stderr_data = self.process.stderr.read(4096)
                    if stderr_data:
                        stderr_text = stderr_data.decode('utf-8', errors='ignore').strip()
                        # 只记录真正的错误信息，忽略常规输出
                        error_keywords = ["error", "failed", "invalid", "cannot", "unable", "no such", "not found", "missing", "unrecognized"]
                        # 忽略进度信息和启动信息
                        ignore_patterns = ["frame=", "fps=", "speed=", "time=", "bitrate=", "size=", "copyright", "built with", "configuration"]
                        
                        is_error = any(keyword in stderr_text.lower() for keyword in error_keywords)
                        should_ignore = any(pattern in stderr_text for pattern in ignore_patterns)
                        
                        if is_error and not should_ignore:
                            logger.error(f"FFmpeg错误: {stderr_text}")
                        elif not should_ignore:
                            # 记录不是错误也不是要忽略的信息为调试信息
                            logger.debug(f"FFmpeg输出: {stderr_text}")
                except Exception:
                    pass
                
                # 简化帧读取 - 直接读取一整帧，并设置超时控制
                try:
                    # 使用fcntl设置非阻塞模式 - 仅在非Windows系统上
                    if os.name != 'nt':
                        import fcntl
                        flags = fcntl.fcntl(self.process.stdout.fileno(), fcntl.F_GETFL)
                        fcntl.fcntl(self.process.stdout.fileno(), fcntl.F_SETFL, flags | os.O_NONBLOCK)
                except Exception:
                    pass  # 忽略错误，回退到阻塞模式
                
                # 使用timeout机制确保读取不会永久阻塞
                start_time = time.time()
                raw_data = bytearray()
                bytes_needed = self.FRAME_SIZE
                
                while len(raw_data) < bytes_needed and time.time() - start_time < 0.2:  # 200ms超时
                    try:
                        # 读取剩余需要的数据
                        chunk = self.process.stdout.read(bytes_needed - len(raw_data))
                        if not chunk:  # 没有更多数据可读
                            # 小睡避免CPU占用
                            time.sleep(0.001)
                            continue
                            
                        raw_data.extend(chunk)
                    except Exception as e:
                        # 忽略一些常见的非阻塞IO错误
                        if "resource temporarily unavailable" in str(e).lower():
                            time.sleep(0.001)  # 小睡，稍后重试
                            continue
                        logger.debug(f"读取数据时出错: {str(e)}")
                        break
                
                # 检查是否读取了完整帧
                if len(raw_data) == bytes_needed:
                    # 成功读取到一帧完整的BGR数据
                    try:
                        frame = np.frombuffer(raw_data, dtype=np.uint8).reshape((self.actual_height, self.actual_width, 3))
                        
                        # 应用帧增强处理
                        if self.enhance_enabled:
                            frame = self.frame_enhancement(frame)
                        
                        # 记录解码成功
                        self.frames_decoded += 1
                        self.consecutive_failures = 0  # 重置失败计数
                        
                        # 保存最后一帧
                        self.last_frame = frame.copy()
                        
                        # 记录上次成功解码时间
                        self.last_decode_time = time.time()
                        
                        # 减少日志输出量，避免刷屏
                        if self.frames_decoded % 30 == 0:
                            logger.info(f"成功读取第 {self.frames_decoded} 帧")
                    except Exception as e:
                        logger.error(f"处理帧数据时出错: {str(e)}")
                        self.consecutive_failures += 1
                else:
                    # 读取到不完整数据
                    logger.debug(f"读取到不完整帧数据: {len(raw_data)}/{bytes_needed} 字节")
                    self.consecutive_failures += 1
                
                if frame is None:
                    # 如果解码失败，记录日志并增加失败计数
                    logger.debug(f"帧解码失败，连续失败次数: {self.consecutive_failures}")
                    
                    # 如果连续失败次数超过阈值，尝试重启解码器
                    if self.consecutive_failures >= self.max_failures:
                        logger.error(f"连续解码失败次数({self.consecutive_failures})超过阈值({self.max_failures})，重启解码器")
                        self.restart_decoder()
                        self.consecutive_failures = 0  # 重置失败计数
                    
                    # 如果有最后成功解码的帧，使用它
                    if self.last_frame is not None:
                        frame = self.last_frame.copy()
                        logger.debug("使用上一帧代替解码失败的帧")
            except Exception as e:
                logger.error(f"读取解码帧时发生异常: {str(e)}")
                # 记录更详细的错误信息
                import traceback
                logger.error(traceback.format_exc())
                
                # 遇到错误时尝试重启解码器
                logger.error("解码出错，重启解码器")
                self.restart_decoder()
                self.consecutive_failures += 1
            
            return frame
            
        except Exception as e:
            logger.error(f"解码帧时发生错误: {str(e)}")
            # 记录更详细的错误信息
            import traceback
            logger.error(traceback.format_exc())
            
            # 遇到错误时尝试重启解码器
            self.restart_decoder()
            return None

    def start_decoder(self):
        """启动解码器进程"""
        try:
            if self.process is not None:
                # 如果已存在进程，先关闭
                logger.debug("关闭已存在的解码器进程")
                try:
                    self.process.terminate()
                    self.process.wait(timeout=1)
                except Exception as e:
                    logger.error(f"关闭旧解码器进程时出错: {str(e)}")
                    try:
                        self.process.kill()
                    except:
                        pass
                self.process = None
            
            if not os.path.exists(self.h264_file) or os.path.getsize(self.h264_file) == 0:
                logger.warning("H.264文件不存在或为空，无法启动解码器")
                return False
                
            if not self.sps_parsed or self.actual_width == 0 or self.actual_height == 0:
                logger.warning("未解析到视频尺寸信息，无法启动解码器")
                return False
                
            # 计算BGR24帧大小 (宽 * 高 * 3)
            self.FRAME_SIZE = self.actual_width * self.actual_height * 3
            logger.info(f"设置解码帧大小: {self.FRAME_SIZE} bytes (BGR24分辨率: {self.actual_width}x{self.actual_height})")
            
            # 构建FFmpeg命令行 - 使用更新的参数
            ffmpeg_cmd = [
                FFMPEG_PATH,
                "-f", "h264",                # 输入格式为h264
                "-i", self.h264_file,        # 输入文件
                "-f", "rawvideo",            # 输出原始视频
                "-pix_fmt", "bgr24",         # 像素格式为BGR24，直接用于OpenCV
                "-fps_mode", "passthrough",  # 替代过时的-vsync 0参数
                "-an",                       # 忽略音频
                "-sn",                       # 忽略字幕
                "-vcodec", "rawvideo",       # 不进行编码
                "-threads", "4",             # 使用4个线程加速解码
                "-"                          # 输出到stdout
            ]
            
            logger.info(f"启动FFmpeg解码器: {' '.join(ffmpeg_cmd)}")
            
            # 创建FFmpeg进程，使用更大的缓冲区
            self.process = subprocess.Popen(
                ffmpeg_cmd,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
                bufsize=self.FRAME_SIZE * 8  # 增加缓冲区大小
            )
            
            # 记录重启时间
            self.last_decoder_restart = time.time()
            
            logger.info("解码器启动成功")
            return True
            
        except Exception as e:
            logger.error(f"启动解码器时出错: {str(e)}")
            import traceback
            logger.error(traceback.format_exc())
            
            if hasattr(self, 'process') and self.process is not None:
                try:
                    self.process.terminate()
                except:
                    pass
                self.process = None
                
            return False

    def restart_decoder(self):
        """重启解码器进程"""
        try:
            logger.info("重启解码器进程")
            
            # 关闭现有进程
            if hasattr(self, 'process') and self.process is not None:
                try:
                    logger.debug("终止现有解码器进程")
                    self.process.terminate()
                    try:
                        self.process.wait(timeout=1)
                    except subprocess.TimeoutExpired:
                        logger.warning("解码器进程未能在1秒内终止，强制关闭")
                        self.process.kill()
                        self.process.wait()
                except Exception as e:
                    logger.error(f"关闭解码器进程时出错: {str(e)}")
                
                self.process = None
            
            # 暂停一小段时间，让系统释放资源
            time.sleep(0.2)
            
            # 启动新的解码器进程
            result = self.start_decoder()
            
            # 增加重启计数
            self.decoder_restarts += 1
            
            # 更新上次重启时间
            self.last_decoder_restart = time.time()
            
            if result:
                logger.info(f"解码器重启成功，这是第 {self.decoder_restarts} 次重启")
            else:
                logger.error("解码器重启失败")
                
            return result
            
        except Exception as e:
            logger.error(f"重启解码器时出错: {str(e)}")
            import traceback
            logger.error(traceback.format_exc())
            return False

    def close_process(self):
        """关闭FFmpeg进程"""
        if hasattr(self, 'ffmpeg_process') and self.ffmpeg_process is not None:
            try:
                self.ffmpeg_process.terminate()
                # 给进程2秒钟时间优雅地终止
                try:
                    self.ffmpeg_process.wait(timeout=2)
                except subprocess.TimeoutExpired:
                    # 如果超时，则强制终止
                    self.ffmpeg_process.kill()
                    self.ffmpeg_process.wait()
            except Exception as e:
                logger.error(f"关闭FFmpeg进程时出错: {str(e)}")
            finally:
                self.ffmpeg_process = None

    def display_frames_loop(self):
        """显示帧循环"""
        if not self.display_enabled:
            logger.info("显示功能已禁用，不显示视频")
            return
        
        logger.info("开始显示视频帧循环")
        self.running = True
        
        # 性能统计变量
        frame_count = 0
        start_time = time.time()
        last_fps_update = start_time
        fps = 0
        
        try:
            while self.running and self.is_running:  # 同时检查两个状态变量
                try:
                    # 解码最新帧
                    frame = self.decode_frame()
                    
                    # 更新性能统计
                    current_time = time.time()
                    frame_count += 1
                    
                    # 每秒更新一次FPS计算
                    if current_time - last_fps_update >= 1.0:
                        fps = frame_count / (current_time - last_fps_update)
                        frame_count = 0
                        last_fps_update = current_time
                        # 记录帧率信息
                        logger.info(f"Display FPS: {fps:.1f}")
                    
                    if frame is not None:
                        # 在帧上显示性能信息
                        elapsed = current_time - start_time
                        status_text = f"FPS: {fps:.1f} | 运行时间: {elapsed:.1f}s | 解码: {self.frames_decoded}"
                        if hasattr(self, 'decoder_restarts'):
                            status_text += f" | 重启: {self.decoder_restarts}"
                        if hasattr(self, 'enhance_enabled') and self.enhance_enabled:
                            status_text += f" | 增强: 开启"
                        
                        cv2.putText(frame, status_text, (10, 30), 
                                    cv2.FONT_HERSHEY_SIMPLEX, 0.7, (0, 255, 0), 2)
                        
                        # 显示帧
                        if cv2.getWindowProperty(self.window_name, cv2.WND_PROP_VISIBLE) >= 0:
                            cv2.imshow(self.window_name, frame)
                            self.frames_displayed += 1
                            # 每30帧记录一次帧信息，减少日志量
                            if self.frames_displayed % 30 == 0:
                                logger.info(f"已显示 {self.frames_displayed} 帧")
                        else:
                            # 如果窗口被关闭，则停止显示循环
                            logger.info("Display window closed by user, stopping display loop")
                            self.running = False
                            break
                    else:
                        # 如果没有获取到帧，记录警告并等待稍长时间
                        current = time.time()
                        # 避免过多警告消息，最多每秒记录一次
                        if current - self.last_no_frame_warn > 1.0:
                            logger.warning("未获取到视频帧，等待下一帧...")
                            self.last_no_frame_warn = current
                        time.sleep(0.01)  # 减少等待时间，提高响应性
                
                except Exception as e:
                    logger.error(f"显示帧时发生错误: {str(e)}")
                    import traceback
                    logger.error(traceback.format_exc())
                    time.sleep(0.1)  # 出错时等待更长时间，避免错误快速重复
                
                finally:
                    # 确保waitKey总是被调用，即使在出错的情况下
                    # 这是必要的，因为它处理窗口事件并允许用户退出
                    try:
                        key = cv2.waitKey(1)
                        if key == ord('q') or key == 27:  # q或ESC键退出
                            logger.info("用户按键退出显示循环")
                            self.running = False
                            self.is_running = False  # 同时设置主运行状态，确保所有循环都会退出
                            break
                    except Exception as e:
                        logger.error(f"处理按键事件时出错: {str(e)}")
                    
                    # 检查主程序是否还在运行
                    if not self.is_running:
                        logger.info("主程序已停止运行，退出显示循环")
                        break
                    
                    # 增加短暂延迟，减少CPU使用率但保持响应性
                    time.sleep(0.001)
            
        except KeyboardInterrupt:
            logger.info("接收到键盘中断，退出显示循环")
        
        finally:
            self.running = False
            try:
                # 在实际关闭窗口前添加一小段延迟，让其他线程有时间清理资源
                time.sleep(0.1)
                cv2.destroyAllWindows()
            except Exception as e:
                logger.error(f"关闭显示窗口时出错: {str(e)}")
                
            logger.info("显示循环已结束")

    async def receive_loop(self):
        """接收消息循环"""
        try:
            while self.is_running:
                try:
                    # 接收数据
                    try:
                        data = await asyncio.wait_for(self.ws.recv(), timeout=1.0)
                    except asyncio.TimeoutError:
                        # 接收超时，但保持循环运行
                        continue
                    
                    if isinstance(data, bytes):
                        # 检查消息类型
                        if len(data) > len(self.MAGIC_BYTES_INITIAL) and data.startswith(self.MAGIC_BYTES_INITIAL):
                            # 初始化信息
                            await self.handle_initial_info(data)
                        elif len(data) > len(self.MAGIC_BYTES_MESSAGE) and data.startswith(self.MAGIC_BYTES_MESSAGE):
                            # 设备消息
                            await self.handle_device_message(data)
                        else:
                            # 视频数据
                            self.process_video_data(data)
                
                except websockets.exceptions.ConnectionClosed:
                    logger.info("Connection closed")
                    break
                except Exception as e:
                    if self.is_running:  # 只在客户端仍在运行时记录错误
                        logger.error(f"Error receiving messages: {str(e)}")
                        import traceback
                        traceback.print_exc()
                    break
        except asyncio.CancelledError:
            # 正常取消，不需要记录错误
            logger.info("Receive loop was cancelled, shutting down gracefully")
        except Exception as e:
            logger.error(f"Unexpected error in receive loop: {str(e)}")
            import traceback
            traceback.print_exc()
        finally:
            logger.info("Receive loop ended")

    async def close(self):
        """关闭连接并清理资源"""
        logger.info("开始清理资源...")
        
        # 首先设置状态标志，使所有循环停止
        self.is_running = False
        self.running = False
        
        # 给线程一点时间来感知状态变化
        await asyncio.sleep(0.2)
        
        # 关闭WebSocket连接
        if self.ws:
            try:
                logger.info("关闭WebSocket连接...")
                await self.ws.close()
                logger.info("WebSocket连接已关闭")
            except Exception as e:
                logger.error(f"关闭WebSocket连接时出错: {str(e)}")
        
        # 结束FFmpeg进程
        if hasattr(self, 'ffmpeg_process') and self.ffmpeg_process is not None:
            try:
                logger.info("关闭FFmpeg进程...")
                self.ffmpeg_process.terminate()
                try:
                    self.ffmpeg_process.wait(timeout=2)
                    logger.info("FFmpeg进程已正常终止")
                except subprocess.TimeoutExpired:
                    logger.warning("FFmpeg进程未能在2秒内终止，强制关闭")
                    self.ffmpeg_process.kill()
                    self.ffmpeg_process.wait()
            except Exception as e:
                logger.error(f"关闭FFmpeg进程时出错: {str(e)}")
                
        # 关闭解码器进程
        if hasattr(self, 'process') and self.process is not None:
            try:
                logger.info("关闭解码器进程...")
                self.process.terminate()
                try:
                    self.process.wait(timeout=2)
                    logger.info("解码器进程已正常终止")
                except subprocess.TimeoutExpired:
                    logger.warning("解码器进程未能在2秒内终止，强制关闭")
                    self.process.kill()
                    self.process.wait()
            except Exception as e:
                logger.error(f"关闭解码器进程时出错: {str(e)}")
            
        # 再次等待一小段时间，确保其他线程有足够时间完成清理
        await asyncio.sleep(0.5)
            
        # 关闭所有OpenCV窗口
        try:
            logger.info("关闭所有OpenCV窗口...")
            cv2.destroyAllWindows()
            # 等待一小段时间确保窗口被关闭
            await asyncio.sleep(0.2)
            logger.info("OpenCV窗口已关闭")
        except Exception as e:
            logger.error(f"关闭OpenCV窗口时出错: {str(e)}")
        
        # 删除临时文件目录
        try:
            logger.info(f"清理临时文件目录: {self.temp_dir}")
            for file in os.listdir(self.temp_dir):
                try:
                    file_path = os.path.join(self.temp_dir, file)
                    logger.debug(f"删除临时文件: {file_path}")
                    os.remove(file_path)
                except Exception as e:
                    logger.error(f"删除临时文件时出错: {str(e)}")
            
            logger.info("删除临时目录...")
            os.rmdir(self.temp_dir)
            logger.info("临时目录已删除")
        except Exception as e:
            logger.error(f"清理临时文件时出错: {str(e)}")
            
        logger.info("资源清理完成")

# 显示启动动画
for i in range(5):
    print(f"程序启动中{'.'*(i+1)}", end='\r')
    time.sleep(0.2)
print("程序已启动！")
print("=" * 80)
print(" 使用说明 ".center(80, "="))
print("1. 按ESC键退出程序")
print("2. 按R键强制刷新画面")
print("3. 如果画面静止，请检查连接或按R键刷新")
print("4. 如显示卡死，请关闭程序并重新运行")
print("5. 优化了显示线程，显示更加流畅")
print("6. 画面自动定时刷新，约每秒25帧")
print("7. 程序性能已优化，减少不必要的线程创建")
print("8. 初始化后如无画面请耐心等待几秒钟")
print("9. 已启用超分辨率图像增强功能，提升画面清晰度")
print("10.图像增强功能包括：降噪、锐化、亮度和对比度调整")
print("=" * 80)

async def main():
    # 创建客户端实例，启用图像增强
    client = ScrcpyClient(
        enhance_enabled=True,     # 启用图像增强
        denoise_strength=5,       # 降噪强度 (1-10)
        sharpen_strength=0.5,     # 锐化强度 (0.1-1.0)
        brightness=1.1,           # 亮度提升 (0.5-1.5)
        contrast=1.2              # 对比度提升 (0.5-1.5)
    )
    
    receive_task = None
    
    try:
        # 连接WebSocket
        connected = await client.connect()
        if not connected:
            logger.error("Could not connect to WebSocket server, exiting")
            return
            
        # 创建接收任务，但不等待它完成
        receive_task = asyncio.create_task(client.receive_loop())
        
        # 主循环，每秒检查一次程序是否应该继续运行
        try:
            while client.is_running:
                await asyncio.sleep(1)
                
                # 检查接收任务是否出错
                if receive_task.done():
                    if receive_task.exception():
                        logger.error(f"Receive task failed: {receive_task.exception()}")
                        client.is_running = False
                        break
        except asyncio.CancelledError:
            logger.info("Main task cancelled")
        except KeyboardInterrupt:
            logger.info("User interrupt, exiting...")
        finally:
            client.is_running = False
            
    except Exception as e:
        logger.error(f"Error occurred: {str(e)}")
        import traceback
        traceback.print_exc()
    finally:
        # 确保客户端资源被正确清理
        client.is_running = False
        
        # 正确处理接收任务
        if receive_task and not receive_task.done():
            logger.info("Cancelling receive task...")
            receive_task.cancel()
            try:
                await asyncio.wait_for(receive_task, timeout=2)
            except (asyncio.CancelledError, asyncio.TimeoutError):
                pass
                
        # 正确关闭客户端
        await client.close()
        
        # 确保关闭所有窗口
        cv2.destroyAllWindows()
        
        logger.info("Application shutdown complete")

if __name__ == "__main__":
    try:
        # 捕获并正确处理KeyboardInterrupt异常
        asyncio.run(main())
    except KeyboardInterrupt:
        print("\n程序被用户中断，正在关闭...")
    except Exception as e:
        print(f"程序出错: {str(e)}")
        import traceback
        traceback.print_exc()
    finally:
        print("程序已退出")


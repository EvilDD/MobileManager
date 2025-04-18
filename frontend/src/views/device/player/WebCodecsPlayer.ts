/**
 * WebCodecsPlayer - 使用 WebCodecs API 解码和播放视频流
 */
import H264Parser from 'h264-converter/dist/h264-parser';

// 实现 toHex 工具函数，由于库中可能没有直接导出该函数
function toHex(value: number): string {
  return value.toString(16).padStart(2, '0');
}

export class WebCodecsPlayer {
  private canvas: HTMLCanvasElement | null = null;
  private ctx: CanvasRenderingContext2D | null = null;
  private decoder: VideoDecoder | null = null;
  private parent: HTMLElement | null = null;
  private isPlaying = false;
  private frameCounter = 0;
  private lastFrameTime = 0;
  private frameRate = 0;
  private pendingFrames = 0;
  private frameInterval: number | null = null;
  private videoWidth = 0;
  private videoHeight = 0;
  private keyFrameFound = false;
  private buffer: Uint8Array | undefined;
  private bufferedSPS = false;
  private bufferedPPS = false;
  private codecString = 'avc1.640028'; // 默认编解码器字符串

  // NAL 单元类型常量
  private readonly NAL_TYPE_SPS = 7;
  private readonly NAL_TYPE_PPS = 8;
  private readonly NAL_TYPE_IDR = 5; // IDR 帧 (关键帧)
  private readonly NAL_TYPE_SEI = 6; // SEI (补充增强信息)

  constructor() {
    // 检查浏览器是否支持 WebCodecs API
    if (!('VideoDecoder' in window)) {
      throw new Error('当前浏览器不支持 WebCodecs API');
    }
  }

  /**
   * 设置播放器的父容器
   * @param element 父容器元素
   */
  setParent(element: HTMLElement): void {
    this.parent = element;
    
    // 创建 canvas 元素
    this.canvas = document.createElement('canvas');
    this.canvas.style.width = '100%';
    this.canvas.style.height = '100%';
    this.canvas.style.display = 'block';
    this.canvas.style.position = 'absolute';
    this.canvas.style.left = '0';
    this.canvas.style.top = '0';
    this.canvas.style.transformOrigin = '0 0';
    this.parent.appendChild(this.canvas);
    
    // 获取绘图上下文
    this.ctx = this.canvas.getContext('2d');
    
    // 处理窗口大小调整事件
    window.addEventListener('resize', this.handleResize.bind(this));
  }

  /**
   * 处理窗口大小调整
   */
  private handleResize(): void {
    if (this.canvas && this.videoWidth > 0 && this.videoHeight > 0) {
      // 获取设备像素比
      const dpr = window.devicePixelRatio || 1;
      
      // 获取父容器尺寸
      let parentWidth = this.parent?.clientWidth || window.innerWidth;
      let parentHeight = this.parent?.clientHeight || window.innerHeight;
      
      // 计算适合的显示尺寸，保持宽高比
      let displayWidth = this.videoWidth;
      let displayHeight = this.videoHeight;
      
      // 计算缩放因子
      const scaleX = parentWidth / displayWidth;
      const scaleY = parentHeight / displayHeight;
      const scale = Math.min(scaleX, scaleY);
      
      // 应用缩放
      displayWidth = Math.floor(displayWidth * scale);
      displayHeight = Math.floor(displayHeight * scale);
      
      // 计算Canvas的物理像素尺寸
      const canvasWidth = Math.round(displayWidth * dpr);
      const canvasHeight = Math.round(displayHeight * dpr);
      
      console.log('视频原始尺寸:', { width: this.videoWidth, height: this.videoHeight });
      console.log('父容器尺寸:', { width: parentWidth, height: parentHeight });
      console.log('应用的缩放因子:', scale);
      console.log('显示尺寸:', { displayWidth, displayHeight });
      console.log('Canvas物理像素尺寸:', { canvasWidth, canvasHeight });
      
      // 设置Canvas的物理像素尺寸
      this.canvas.width = canvasWidth;
      this.canvas.height = canvasHeight;
      
      // 设置Canvas的CSS显示尺寸
      this.canvas.style.width = `${displayWidth}px`;
      this.canvas.style.height = `${displayHeight}px`;
      
      // 设置Canvas的缩放以匹配设备像素比
      if (this.ctx) {
        // 重置变换
        this.ctx.setTransform(1, 0, 0, 1, 0, 0);
        // 应用设备像素比缩放
        this.ctx.scale(dpr, dpr);
      }
      
      // 如果父元素存在，也更新其尺寸
      if (this.parent) {
        this.parent.style.width = `${displayWidth}px`;
        this.parent.style.height = `${displayHeight}px`;
      }
    }
  }
  
  /**
   * 从SPS数据中提取视频分辨率
   * @param spsData SPS NAL单元数据
   */
  private extractResolutionFromSPS(spsData: Uint8Array): void {
    try {
      const parsedSPS = H264Parser.parseSPS(spsData);
      const {
        profile_idc,
        constraint_set_flags,
        level_idc,
        pic_width_in_mbs_minus1,
        frame_crop_left_offset,
        frame_crop_right_offset,
        frame_mbs_only_flag,
        pic_height_in_map_units_minus1,
        frame_crop_top_offset,
        frame_crop_bottom_offset,
        sar,
      } = parsedSPS;

      const sarScale = sar[0] / sar[1];
      this.codecString = `avc1.${[profile_idc, constraint_set_flags, level_idc].map(toHex).join('')}`;
      
      const width = Math.ceil(
        ((pic_width_in_mbs_minus1 + 1) * 16 - frame_crop_left_offset * 2 - frame_crop_right_offset * 2) * sarScale,
      );
      const height =
        (2 - frame_mbs_only_flag) * (pic_height_in_map_units_minus1 + 1) * 16 -
        (frame_mbs_only_flag ? 2 : 4) * (frame_crop_top_offset + frame_crop_bottom_offset);
      
      console.log(`从SPS解析出视频信息:`, {
        width,
        height,
        codec: this.codecString,
        profile: profile_idc,
        level: level_idc,
        rawSPSData: parsedSPS
      });
      
      if (width > 0 && height > 0) {
        this.videoWidth = width;
        this.videoHeight = height;
        this.handleResize();
      }
    } catch (error) {
      console.error('SPS解析错误:', error);
    }
  }

  /**
   * 初始化解码器
   */
  private initDecoder(): void {
    if (this.decoder && this.decoder.state !== 'closed') {
      return;
    }

    try {
      this.decoder = new VideoDecoder({
        output: this.handleDecodedFrame.bind(this),
        error: (error) => {
          console.error('解码器错误:', error);
          this.stop();
        }
      });
      
      console.log('解码器初始化成功');
    } catch (error) {
      console.error('解码器初始化失败:', error);
    }
  }

  /**
   * 配置解码器
   */
  private configureDecoder(): void {
    if (!this.decoder || this.decoder.state === 'configured') return;
    
    try {
      const config: VideoDecoderConfig = {
        codec: this.codecString,
        optimizeForLatency: true
      };
      
      this.decoder.configure(config);
      console.log(`解码器配置成功:`, config);
    } catch (error) {
      console.error('解码器配置失败:', error);
    }
  }

  /**
   * 处理解码后的帧
   * @param frame 解码后的视频帧
   */
  private handleDecodedFrame(frame: VideoFrame): void {
    this.pendingFrames--;
    this.frameCounter++;
    
    // 计算帧率
    const now = performance.now();
    if (now - this.lastFrameTime >= 1000) {
      this.frameRate = this.frameCounter;
      this.frameCounter = 0;
      this.lastFrameTime = now;
      console.log(`当前帧率: ${this.frameRate} FPS, 待处理帧: ${this.pendingFrames}`);
    }
    
    // 直接绘制到canvas
    if (this.ctx && this.canvas) {
      try {
        this.ctx.drawImage(frame, 0, 0);
      } catch (e) {
        console.error('Canvas drawImage 错误:', e);
      }
    }
    
    // 关闭帧以释放资源
    frame.close();
  }

  /**
   * 处理视频数据
   * @param data 视频数据
   */
  pushFrame(data: Uint8Array): void {
    if (!this.isPlaying) return;
    
    if (!this.decoder || this.decoder.state === 'closed') {
      this.initDecoder();
      return;
    }
    
    try {
      // 确保数据至少包含NAL头
      if (data.length < 5) {
        console.log('数据太短，跳过');
        return;
      }
      
      // 获取NAL单元类型 (第5个字节的后5位)
      const nalType = data[4] & 0x1F;
      
      // 处理不同类型的NAL单元
      if (nalType === this.NAL_TYPE_SPS) {
        console.log('收到SPS');
        this.extractResolutionFromSPS(data.subarray(4));
        this.configureDecoder();
        this.bufferedSPS = true;
        this.buffer = this.appendToBuffer(data);
        this.keyFrameFound = false;
        return;
      } 
      else if (nalType === this.NAL_TYPE_PPS) {
        console.log('收到PPS');
        this.bufferedPPS = true;
        this.buffer = this.appendToBuffer(data);
        return;
      }
      else if (nalType === this.NAL_TYPE_SEI) {
        // 跳过单独的SEI
        if (!this.bufferedSPS || !this.bufferedPPS) {
          return;
        }
      }
      
      // 将数据添加到缓冲区
      const array = this.appendToBuffer(data);
      
      // 检查是否为IDR帧
      const isIDR = nalType === this.NAL_TYPE_IDR;
      if (isIDR) {
        console.log('收到IDR帧');
        this.keyFrameFound = true;
      }
      
      // 只有在解码器配置完成且收到关键帧后才开始解码
      if (array && this.decoder.state === 'configured' && this.keyFrameFound) {
        // 重置缓冲区
        this.buffer = undefined;
        this.bufferedPPS = false;
        this.bufferedSPS = false;
        
        // 解码数据
        this.decoder.decode(new EncodedVideoChunk({
          type: 'key',
          timestamp: 0,
          data: array
        }));
        
        this.pendingFrames++;
      }
    } catch (error) {
      console.error('处理视频数据错误:', error);
    }
  }
  
  /**
   * 将新数据添加到缓冲区
   */
  private appendToBuffer(data: Uint8Array): Uint8Array {
    let array: Uint8Array;
    if (this.buffer) {
      array = new Uint8Array(this.buffer.length + data.length);
      array.set(this.buffer);
      array.set(data, this.buffer.length);
    } else {
      array = data;
    }
    return array;
  }

  /**
   * 开始播放
   */
  play(): void {
    if (this.isPlaying) {
      return;
    }
    
    this.isPlaying = true;
    this.keyFrameFound = false;
    this.bufferedSPS = false;
    this.bufferedPPS = false;
    this.buffer = undefined;
    this.initDecoder();
    
    console.log('播放器启动');
  }

  /**
   * 停止播放
   */
  stop(): void {
    this.isPlaying = false;
    
    // 清除渲染定时器
    if (this.frameInterval !== null) {
      clearInterval(this.frameInterval);
      this.frameInterval = null;
    }
    
    // 关闭解码器
    if (this.decoder) {
      try {
        this.decoder.close();
        this.decoder = null;
      } catch (error) {
        console.error('关闭解码器错误:', error);
      }
    }
    
    // 重置状态
    this.keyFrameFound = false;
    this.bufferedSPS = false;
    this.bufferedPPS = false;
    this.buffer = undefined;
    
    // 移除 canvas
    if (this.canvas && this.parent) {
      this.parent.removeChild(this.canvas);
      this.canvas = null;
      this.ctx = null;
    }
    
    // 移除事件监听器
    window.removeEventListener('resize', this.handleResize.bind(this));
    
    console.log('播放器停止');
  }
}
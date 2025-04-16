/**
 * WebCodecsPlayer for Mobile Manager
 * 基于wscrcpy的WebCodecsPlayer.ts实现
 */
import H264Parser from 'h264-converter/dist/h264-parser';
import NALU from 'h264-converter/dist/util/NALU';
import { STREAM_WINDOW_CONFIG } from '../components/config';

// 为全局变量添加类型定义
declare global {
  interface Window {
    _deviceStreamInfo?: {
      [deviceId: string]: {
        width: number;
        height: number;
        codec: string;
        timestamp: number;
      }
    };
  }
}

// 必要的类型定义
type ParametersSubSet = {
  codec: string;
  width: number;
  height: number;
};

// 简单尺寸类
class Size {
  public readonly w: number;
  public readonly h: number;

  constructor(readonly width: number, readonly height: number) {
    this.w = width;
    this.h = height;
  }
}

// 简单矩形类
class Rect {
  constructor(readonly left: number, readonly top: number, readonly right: number, readonly bottom: number) {}

  public getWidth(): number {
    return this.right - this.left;
  }

  public getHeight(): number {
    return this.bottom - this.top;
  }
}

// 屏幕信息类
class ScreenInfo {
  constructor(readonly contentRect: Rect, readonly videoSize: Size, readonly deviceRotation: number) {}
}

// 帧类型
type DecodedFrame = {
  width: number;
  height: number;
  frame: any;
};

function toHex(value: number): string {
  return value.toString(16).padStart(2, '0').toUpperCase();
}

export class WebCodecsPlayer {
  // 静态属性
  public static readonly storageKeyPrefix = 'WebCodecsPlayer';
  public static readonly playerFullName = 'WebCodecs';
  public static readonly playerCodeName = 'webcodecs';

  // 默认配置
  public static readonly defaultConfig = {
    width: STREAM_WINDOW_CONFIG.DEFAULT_WIDTH,
    height: STREAM_WINDOW_CONFIG.DEFAULT_HEIGHT,
    codec: 'avc1.42001F'
  };

  // 检查浏览器是否支持WebCodecs API
  public static isSupported(): boolean {
    if (typeof VideoDecoder !== 'function' || typeof VideoDecoder.isConfigSupported !== 'function') {
      return false;
    }
    return true;
  }

  // 解析SPS数据
  private static parseSPS(data: Uint8Array): ParametersSubSet {
    try {
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
      } = H264Parser.parseSPS(data);

      // 限制 sar 比例在合理范围内，防止异常值
      let sarScale = 1.0;
      if (sar[0] && sar[1] && sar[1] !== 0) {
        sarScale = Math.min(Math.max(sar[0] / sar[1], 0.5), 2.0);
      }
      
      // 使用标准编解码器字符串格式
      let codec = `avc1.${[profile_idc, constraint_set_flags, level_idc].map(toHex).join('')}`;
      
      // 检测高级编解码器，可能不被浏览器支持
      if (profile_idc >= 100) {
        console.warn('检测到高级编解码器配置，可能不兼容:', codec);
        // 回退到安全配置
        codec = 'avc1.42001F';
      }
      
      // 计算视频的实际尺寸，并保证至少有最小值
      const width = Math.max(32, Math.ceil(
        ((pic_width_in_mbs_minus1 + 1) * 16 - frame_crop_left_offset * 2 - frame_crop_right_offset * 2) * sarScale,
      ));
      const height = Math.max(32, 
        (2 - frame_mbs_only_flag) * (pic_height_in_map_units_minus1 + 1) * 16 -
        (frame_mbs_only_flag ? 2 : 4) * (frame_crop_top_offset + frame_crop_bottom_offset)
      );
      
      console.log('SPS解析结果:', { 
        width, 
        height, 
        codec, 
        sarScale,
        origProfile: profile_idc
      });
      
      return { codec, width, height };
    } catch (err) {
      console.error('解析SPS失败:', err);
      return { 
        codec: 'avc1.42001F',
        width: STREAM_WINDOW_CONFIG.DEFAULT_WIDTH,
        height: STREAM_WINDOW_CONFIG.DEFAULT_HEIGHT
      };
    }
  }

  // 实例属性
  private canvas: HTMLCanvasElement;
  private context: CanvasRenderingContext2D;
  private decoder: VideoDecoder;
  private buffer: ArrayBuffer | undefined;
  private hadIDR = false;
  private bufferedSPS = false;
  private bufferedPPS = false;
  private receivedFirstFrame = false;
  private decodedFrames: DecodedFrame[] = [];
  private animationFrameId?: number;

  // 回调函数
  private onFirstFrameDecoded: () => void;
  private onError: (error: Error) => void;
  private onVideoResize: (size: {width: number, height: number}) => void;

  /**
   * 构造函数
   */
  constructor(options: {
    canvas: HTMLCanvasElement;
    onFirstFrameDecoded?: () => void;
    onError?: (error: Error) => void;
    onVideoResize?: (size: {width: number, height: number}) => void;
  }) {
    this.canvas = options.canvas;
    this.onFirstFrameDecoded = options.onFirstFrameDecoded || (() => {});
    this.onError = options.onError || ((error) => console.error('播放器错误:', error));
    this.onVideoResize = options.onVideoResize || (() => {});

    // 初始化Context
    const context = this.canvas.getContext('2d', {
      alpha: false,
      desynchronized: true
    });
    
    if (!context) {
      throw new Error('无法获取Canvas 2D上下文');
    }
    
    this.context = context;
    
    // 初始化解码器
    this.decoder = this.createDecoder();
    
    // 设置默认尺寸
    this.setDefaultCanvasSize();
  }

  /**
   * 设置默认Canvas尺寸
   */
  private setDefaultCanvasSize(): void {
    const width = STREAM_WINDOW_CONFIG.DEFAULT_WIDTH;
    const height = STREAM_WINDOW_CONFIG.DEFAULT_HEIGHT;
    console.log('设置默认Canvas尺寸:', width, 'x', height);
    
    this.canvas.width = width;
    this.canvas.height = height;
    
    // 设置样式
    Object.assign(this.canvas.style, {
      width: `${width}px`,
      height: `${height}px`,
      display: 'block'
    });
  }

  /**
   * 创建解码器
   */
  private createDecoder(): VideoDecoder {
    return new VideoDecoder({
      output: (frame) => {
        this.onFrameDecoded(0, 0, frame);
      },
      error: (error: DOMException) => {
        console.error('解码器错误:', error, `code: ${error.code}`);
        this.onError(error);
        this.stop();
      },
    });
  }

  /**
   * 添加数据到缓冲区
   */
  protected addToBuffer(data: Uint8Array): Uint8Array {
    let array: Uint8Array;
    if (this.buffer) {
      // 创建新的合并数组
      array = new Uint8Array(this.buffer.byteLength + data.byteLength);
      array.set(new Uint8Array(this.buffer));
      array.set(new Uint8Array(data), this.buffer.byteLength);
    } else {
      // 复制数据以避免引用问题
      array = new Uint8Array(data);
    }
    // 保存新缓冲区
    this.buffer = array.buffer as ArrayBuffer;
    return array;
  }

  /**
   * 调整Canvas尺寸
   */
  protected scaleCanvas(width: number, height: number): void {
    console.log('原始视频尺寸:', { width, height });
    
    // 获取设备像素比
    const dpr = window.devicePixelRatio || 1;
    console.log('设备像素比:', dpr);
    
    // 计算实际显示尺寸
    const displayWidth = Math.round(width);
    const displayHeight = Math.round(height);
    
    // 计算Canvas的物理像素尺寸
    const canvasWidth = Math.round(displayWidth * dpr);
    const canvasHeight = Math.round(displayHeight * dpr);
    
    console.log('显示尺寸:', { displayWidth, displayHeight });
    console.log('Canvas物理像素尺寸:', { canvasWidth, canvasHeight });
    
    // 创建ScreenInfo对象用于内部跟踪
    const _screenInfo = new ScreenInfo(
      new Rect(0, 0, width, height),
      new Size(displayWidth, displayHeight),
      0
    );
    
    // 通知尺寸变化
    this.onVideoResize({
      width: displayWidth,
      height: displayHeight
    });

    // 初始化canvas尺寸
    this.canvas.width = canvasWidth;
    this.canvas.height = canvasHeight;
    
    // 设置统一的CSS显示尺寸
    const commonStyle = {
      position: 'absolute',
      left: '0',
      top: '0',
      width: `${displayWidth}px`,
      height: `${displayHeight}px`,
      transformOrigin: '0 0'
    };
    
    // 应用样式到canvas
    Object.assign(this.canvas.style, commonStyle);
    
    // 重置变换矩阵并设置适当的缩放
    this.context.setTransform(1, 0, 0, 1, 0, 0);
    this.context.scale(dpr, dpr);
  }

  /**
   * 解码视频数据
   */
  public decode(data: Uint8Array): void {
    if (!data || data.length < 4) {
      return;
    }
    
    try {
      const type = data[4] & 31;
      const isIDR = type === NALU.IDR;
      
      if (type === NALU.SPS) {
        // 从SPS中解析视频参数
        const { width, height } = WebCodecsPlayer.parseSPS(data.subarray(4));
        
        // 如果尺寸异常小，可能是解析错误，使用默认值
        const effectiveWidth = width < 100 ? STREAM_WINDOW_CONFIG.DEFAULT_WIDTH : width;
        const effectiveHeight = height < 100 ? STREAM_WINDOW_CONFIG.DEFAULT_HEIGHT : height;
        
        this.scaleCanvas(effectiveWidth, effectiveHeight);
        
        // 始终使用兼容性更好的基本配置
        const safeConfig = {
          codec: 'avc1.42001F', // 使用基础配置，兼容性最好
          optimizeForLatency: true,
        } as VideoDecoderConfig;
        
        // 先尝试关闭旧解码器
        try {
          if (this.decoder && this.decoder.state !== 'unconfigured') {
            this.decoder.close();
          }
        } catch (err) {
          console.warn('关闭旧解码器失败:', err);
        }
        
        // 创建新解码器
        this.decoder = this.createDecoder();
        
        try {
          this.decoder.configure(safeConfig);
          console.log('解码器已配置:', safeConfig);
        } catch (err) {
          console.error('配置解码器失败:', err);
          this.onError(new Error('无法配置解码器: ' + err.message));
          return;
        }
        
        // 重置解码状态
        this.buffer = undefined;
        this.bufferedSPS = true;
        this.addToBuffer(data);
        this.hadIDR = false;
        return;
      } else if (type === NALU.PPS) {
        this.bufferedPPS = true;
        this.addToBuffer(data);
        return;
      } else if (type === NALU.SEI) {
        // Workaround for lonely SEI from ws-qvh
        if (!this.bufferedSPS || !this.bufferedPPS) {
          return;
        }
      }
      
      const array = this.addToBuffer(data);
      this.hadIDR = this.hadIDR || isIDR;
      
      if (array && this.decoder.state === 'configured') {
        // 重要：确保我们有 SPS, PPS 和 IDR 帧，或者当前数据是 IDR 帧
        if ((this.bufferedSPS && this.bufferedPPS && this.hadIDR) || isIDR) {
          try {
            // 使用标准Uint8Array创建一个副本
            const chunk = new EncodedVideoChunk({
              type: isIDR ? 'key' : 'delta', // 确保使用正确的类型
              timestamp: performance.now(),
              data: new Uint8Array(array)
            });
            
            // 如果是 IDR 帧，解码前清理缓冲区状态
            if (isIDR) {
              this.buffer = undefined;
              // 不要重置缓冲状态，保留SPS和PPS信息
              // this.bufferedPPS = false;
              // this.bufferedSPS = false;
            }
            
            try {
              this.decoder.decode(chunk);
            } catch (decodeErr) {
              console.error('解码错误:', decodeErr);
              
              // 如果解码失败，尝试重置解码器
              if (decodeErr.name === 'DataError' && decodeErr.message.includes('key frame is required')) {
                console.warn('需要关键帧，重置解码器状态');
                
                // 重置解码器状态，保留缓冲区标记
                this.reset(true); // 传递true表示保留缓冲标记
                
                // 如果当前是 IDR 帧，尝试重新解码
                if (isIDR) {
                  try {
                    console.log('重新配置解码器并尝试解码 IDR 帧');
                    
                    // 重新创建并配置解码器
                    try {
                      if (this.decoder) {
                        this.decoder.close();
                      }
                    } catch (err) {
                      console.warn('关闭解码器失败:', err);
                    }
                    
                    // 重新创建解码器
                    this.decoder = this.createDecoder();
                    
                    const backupConfig = {
                      codec: 'avc1.42001F',
                      optimizeForLatency: true,
                    } as VideoDecoderConfig;
                    
                    this.decoder.configure(backupConfig);
                    
                    // 再次尝试解码当前帧
                    const retryChunk = new EncodedVideoChunk({
                      type: 'key',
                      timestamp: performance.now(),
                      data: new Uint8Array(array)
                    });
                    
                    this.decoder.decode(retryChunk);
                  } catch (retryErr) {
                    console.error('重试解码也失败:', retryErr);
                  }
                }
              }
            }
          } catch (err) {
            console.error('创建编码块失败:', err);
          }
        } else if (isIDR) {
          // 如果是IDR但我们没有必要的SPS或PPS
          // 尝试直接解码IDR帧，因为某些实现可能不需要SPS和PPS
          console.warn('接收到IDR帧但缺少SPS或PPS，尝试直接解码');
          try {
            const chunk = new EncodedVideoChunk({
              type: 'key',
              timestamp: performance.now(),
              data: new Uint8Array(array)
            });
            
            this.decoder.decode(chunk);
          } catch (directDecodeErr) {
            console.warn('直接解码IDR帧失败:', directDecodeErr);
            // 不做其他处理，等待后续数据
          }
        } else {
          console.log('等待完整的帧序列 SPS:', this.bufferedSPS, 'PPS:', this.bufferedPPS, 'IDR:', this.hadIDR, '当前类型:', type);
        }
        return;
      } else if (this.decoder.state !== 'configured') {
        console.warn('解码器未配置，等待 SPS 帧');
      }
    } catch (err) {
      console.error('处理视频数据错误:', err);
    }
  }
  
  /**
   * 处理解码后的帧
   */
  protected onFrameDecoded(width: number, height: number, frame: VideoFrame): void {
    if (!this.receivedFirstFrame) {
      this.receivedFirstFrame = true;
      this.onFirstFrameDecoded();
    }
    
    // 保存帧到帧队列
    const maxStored = 3; // 最大缓存帧数
    
    while (this.decodedFrames.length > maxStored) {
      const data = this.decodedFrames.shift();
      if (data) {
        this.dropFrame(data.frame);
      }
    }
    
    // 添加到帧队列
    this.decodedFrames.push({ width, height, frame });
    
    // 如果没有动画帧请求，创建一个
    if (!this.animationFrameId) {
      this.animationFrameId = requestAnimationFrame(this.drawDecoded);
    }
  }

  /**
   * 绘制解码后的帧
   */
  protected drawDecoded = (): void => {
    if (this.receivedFirstFrame) {
      const data = this.decodedFrames.shift();
      if (data) {
        const frame: VideoFrame = data.frame;
        // 清除画布确保清晰渲染
        this.context.clearRect(0, 0, this.canvas.width, this.canvas.height);
        // 绘制帧
        this.context.drawImage(frame, 0, 0);
        frame.close();
      }
    }
    
    if (this.decodedFrames.length) {
      this.animationFrameId = requestAnimationFrame(this.drawDecoded);
    } else {
      this.animationFrameId = undefined;
    }
  };

  /**
   * 丢弃视频帧
   */
  protected dropFrame(frame: VideoFrame): void {
    frame.close();
  }

  /**
   * 停止播放器
   */
  protected stop(): void {
    // 取消动画帧请求
    if (this.animationFrameId) {
      cancelAnimationFrame(this.animationFrameId);
      this.animationFrameId = undefined;
    }
    
    // 清理资源
    this.buffer = undefined;
    this.hadIDR = false;
    this.bufferedSPS = false;
    this.bufferedPPS = false;
    
    // 清理帧队列
    while (this.decodedFrames.length > 0) {
      const data = this.decodedFrames.shift();
      if (data && data.frame) {
        data.frame.close();
      }
    }
  }
  
  /**
   * 关闭播放器
   */
  public close(): void {
    console.log('关闭播放器');
    this.stop();
    
    try {
      if (this.decoder && this.decoder.state !== 'closed') {
        this.decoder.close();
      }
    } catch (err) {
      console.error('关闭解码器错误:', err);
    }
  }

  /**
   * 由外部更新视频尺寸
   * 用于接收后端通知的准确视频尺寸
   */
  public updateVideoSize(width: number, height: number): void {
    console.log(`接收到外部更新的视频尺寸: ${width}x${height}`);
    if (width > 0 && height > 0) {
      // 调整Canvas尺寸
      this.scaleCanvas(width, height);
    } else {
      console.warn('收到无效的外部尺寸');
    }
  }

  /**
   * 重置解码器状态
   * 当解码器出现错误或需要重新配置时调用
   * @param preserveBufferFlags 是否保留缓冲区状态标志
   */
  public reset(preserveBufferFlags = false): void {
    console.log('重置WebCodecsPlayer解码器');
    
    // 保存当前状态（如果需要）
    const savedSPS = this.bufferedSPS;
    const savedPPS = this.bufferedPPS;
    
    this.stop();
    
    // 重新创建解码器
    try {
      if (this.decoder && this.decoder.state !== 'closed') {
        this.decoder.close();
      }
    } catch (err) {
      console.warn('关闭解码器失败:', err);
    }
    
    // 创建新解码器
    this.decoder = this.createDecoder();
    
    // 重置状态
    this.buffer = undefined;
    this.hadIDR = false;
    
    // 有条件地重置缓冲区状态标志
    if (!preserveBufferFlags) {
      this.bufferedSPS = false;
      this.bufferedPPS = false;
    } else {
      // 恢复保存的状态
      this.bufferedSPS = savedSPS;
      this.bufferedPPS = savedPPS;
    }
    
    this.receivedFirstFrame = false;
  }
  
  /**
   * 主动请求关键帧
   * 可以由外部调用，例如当检测到解码问题时
   */
  public requestKeyFrame(): boolean {
    console.log('请求关键帧');
    // 重置解码器相关状态
    this.hadIDR = false;
    this.bufferedSPS = false;
    this.bufferedPPS = false;
    return true;
  }
  
  /**
   * 处理关键帧相关错误
   * 当解码器报告关键帧错误时调用此方法
   */
  private handleKeyFrameError(): void {
    console.warn('处理关键帧错误');
    // 仅重置IDR标记，保留SPS和PPS状态
    this.hadIDR = false;
    
    // 可能需要重新配置解码器
    if (this.decoder.state === 'configured') {
      try {
        this.decoder.reset();
        console.log('解码器已重置');
      } catch (err) {
        console.error('重置解码器失败:', err);
        
        // 完全重置，但保留SPS和PPS状态
        this.reset(true); // 传递true表示保留缓冲标记
      }
    }
  }
}

export default WebCodecsPlayer; 
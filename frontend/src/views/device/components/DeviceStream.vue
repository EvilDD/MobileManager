<template>
    <div class="device-stream-container" :class="{ 'loading': loading, 'error': error, 'landscape': isLandscape }">  
      <canvas 
        ref="streamCanvas"
        class="device-stream-frame"
        :style="{ visibility: loading || error ? 'hidden' : 'visible' }"
      />
      
      <div v-if="loading" class="device-stream-loading">
        <el-icon class="loading-icon"><Loading /></el-icon>
        <div>正在连接设备串流...</div>
      </div>
      
      <div v-if="error" class="device-stream-error">
        <el-icon class="error-icon"><CircleClose /></el-icon>
        <div>{{ errorMessage || '连接失败' }}</div>
        <el-button class="retry-button" size="small" type="primary" @click="retryConnect">
          重试连接
        </el-button>
      </div>
    </div>
  </template>
  
  <script setup>
  import { ref, computed, onMounted, onBeforeUnmount, nextTick } from 'vue';
  import { Loading, CircleClose } from '@element-plus/icons-vue';
  import { STREAM_WINDOW_CONFIG } from './config';
  import { ElMessage } from 'element-plus';
  import { startDeviceStream, stopDeviceStream } from '@/api/device';
  import { useUserStore } from '@/store/modules/user';
  import H264Parser from 'h264-converter/dist/h264-parser';
  import NALU from 'h264-converter/dist/util/NALU';
  
  const props = defineProps({
    deviceId: {
      type: String,
      required: true
    },
    autoConnect: {
      type: Boolean,
      default: true
    },
    serverUrl: {
      type: String,
      default: import.meta.env.VITE_WSCRCPY_SERVER || 'http://localhost:8000'
    },
    skipUnmountDestroy: {
      type: Boolean,
      default: false
    }
  });
  
  const emit = defineEmits(['success', 'stream-error', 'loading-start', 'orientation-change']);
  
  const loading = ref(true);
  const error = ref(false);
  const errorMessage = ref('');
  const streamCanvas = ref(null);
  const isLandscape = ref(false);
  
  const userStore = useUserStore();
  
  // WebSocket 连接
  let wsConnection = null;
  
  // 视频解码器
  let videoDecoder = null;
  
  // 视频帧缓冲区
  let videoBuffer = null;
  
  // 2D绘图上下文
  let canvasContext = null;
  
  // 视频参数
  const videoConfig = ref({
    width: STREAM_WINDOW_CONFIG.DEFAULT_WIDTH,
    height: STREAM_WINDOW_CONFIG.DEFAULT_HEIGHT,
    codec: 'avc1.42C01F'
  });
  
  // 视频解码状态
  const decoderState = ref({
    initialized: false,
    bufferedSPS: false,
    bufferedPPS: false,
    hadIDR: false
  });
  
  // 连接超时定时器
  let connectionTimeoutTimer = null;
  // 连接超时时间(毫秒)
  const CONNECTION_TIMEOUT = 10000;
  
  // 连接状态稳定期
  let connectionStabilityTimer = null;
  const CONNECTION_STABILITY_PERIOD = 1000; // 连接成功后的稳定期(毫秒)
  let isStabilizing = false; // 是否处于连接稳定期
  
  // 声明userClosing标志
  let userClosing = false;
  // 是否正在清理资源
  let isCleaningUp = false;
  
  // 解析SPS数据获取视频参数
  const parseSPS = (data) => {
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

      const sarScale = sar[0] / sar[1];
      const codec = `avc1.${[profile_idc, constraint_set_flags, level_idc]
        .map(v => v.toString(16).padStart(2, '0').toUpperCase())
        .join('')}`;
        
      const width = Math.ceil(
        ((pic_width_in_mbs_minus1 + 1) * 16 - frame_crop_left_offset * 2 - frame_crop_right_offset * 2) * sarScale,
      );
      
      const height =
        (2 - frame_mbs_only_flag) * (pic_height_in_map_units_minus1 + 1) * 16 -
        (frame_mbs_only_flag ? 2 : 4) * (frame_crop_top_offset + frame_crop_bottom_offset);
        
      return { codec, width, height };
    } catch (err) {
      console.error('解析SPS失败:', err);
      return null;
    }
  };
  
  // 解码视频数据 - 完全参考wscrcpy的实现方式
  const decodeVideoData = (data) => {
    if (!videoDecoder || !data || data.length < 4) {
      console.warn('无效的视频数据或解码器未初始化');
      return;
    }
    
    try {
      // 与wscrcpy保持一致的类型提取方式
      const type = data[4] & 31;
      const isIDR = type === NALU.IDR;

      // 处理SPS (序列参数集)
      if (type === NALU.SPS) {
        const params = parseSPS(data.subarray(4));
        if (params) {
          console.log('解析SPS结果:', params);
          // 更新Canvas尺寸 - 与wscrcpy保持一致的尺寸处理
          scaleCanvas(params.width, params.height);
          // 配置解码器
          const config = {
            codec: params.codec,
            optimizeForLatency: true,
          };
          videoDecoder.configure(config);
        }
        decoderState.value.bufferedSPS = true;
        videoBuffer = addToBuffer(videoBuffer, data);
        decoderState.value.hadIDR = false;
        return;
      } 
      // 处理PPS (图像参数集)
      else if (type === NALU.PPS) {
        decoderState.value.bufferedPPS = true;
        videoBuffer = addToBuffer(videoBuffer, data);
        return;
      } 
      // 处理SEI (补充增强信息)
      else if (type === NALU.SEI) {
        // 跳过孤立的SEI - 这也是wscrcpy的处理方式
        if (!decoderState.value.bufferedSPS || !decoderState.value.bufferedPPS) {
          return;
        }
      }

      // 添加到缓冲区
      const array = addToBuffer(videoBuffer, data);
      
      // 关键：仅当积累了IDR帧时才触发解码，与wscrcpy一致
      decoderState.value.hadIDR = decoderState.value.hadIDR || isIDR;
      
      // 重要！：精确匹配wscrcpy的解码条件判断
      if (array && videoDecoder.state === 'configured' && decoderState.value.hadIDR) {
        // 解码之前重置缓冲区
        videoBuffer = undefined;
        
        // 重置标志位
        decoderState.value.bufferedPPS = false;
        decoderState.value.bufferedSPS = false;
        
        // 解码 - 使用固定时间戳0
        videoDecoder.decode(
          new EncodedVideoChunk({
            type: 'key',
            timestamp: 0,
            data: array.buffer,
          }),
        );
      }
    } catch (err) {
      console.error('处理视频数据错误:', err);
    }
  };

  // 添加数据到缓冲区 - 完全参考wscrcpy的实现
  const addToBuffer = (buffer, data) => {
    let array;
    if (buffer) {
      array = new Uint8Array(buffer.byteLength + data.byteLength);
      array.set(new Uint8Array(buffer));
      array.set(new Uint8Array(data), buffer.byteLength);
    } else {
      array = new Uint8Array(data);
    }
    return array;
  };
  
  // 缩放Canvas - 确保精确尺寸
  const scaleCanvas = (width, height) => {
    if (!streamCanvas.value || !canvasContext) return;
    
    console.log('原始视频尺寸:', { width, height });
    
    // 更新视频配置
    videoConfig.value = {
      width,
      height,
      codec: 'avc1.42001E'
    };
    
    // 确保canvas尺寸精确匹配视频尺寸
    // 重要：明确设置为准确的像素尺寸
    streamCanvas.value.width = width;
    streamCanvas.value.height = height;
    
    console.log('设置Canvas实际尺寸:', { 
      width: streamCanvas.value.width, 
      height: streamCanvas.value.height 
    });
    
    // 重置变换矩阵
    canvasContext.setTransform(1, 0, 0, 1, 0, 0);
    
    // 更新横屏状态
    const newIsLandscape = width > height;
    if (isLandscape.value !== newIsLandscape) {
      isLandscape.value = newIsLandscape;
    }
    
    // 在尺寸变化时触发事件
    emit('orientation-change', {
      deviceId: props.deviceId,
      orientation: isLandscape.value ? 'landscape' : 'portrait',
      width,
      height
    });
  };
  
  // 新增：调整容器尺寸以匹配视频比例
  const adjustContainerSize = (width, height) => {
    const container = streamCanvas.value?.parentElement;
    if (!container) return;
    
    // 设置容器尺寸，保持比例
    const aspectRatio = width / height;
    
    // 根据横竖屏调整容器样式
    if (isLandscape.value) {
      container.style.aspectRatio = `${aspectRatio} / 1`;
    } else {
      container.style.aspectRatio = `${aspectRatio} / 1`;
    }
    
    console.log('调整容器尺寸, 宽高比:', aspectRatio);
  };
  
  // 初始化Canvas - 单独抽取这个函数
  const initCanvas = (width, height) => {
    if (!streamCanvas.value) return;
    
    streamCanvas.value.width = width;
    streamCanvas.value.height = height;
    canvasContext = streamCanvas.value.getContext('2d', {
      alpha: false, // 禁用alpha通道可以提高性能
      desynchronized: true // 减少延迟
    });
    
    if (!canvasContext) {
      handleConnectionError('无法获取Canvas 2D上下文');
    }
  };
  
  // 处理解码后的视频帧 - 精确渲染
  const handleDecodedFrame = (frame) => {
    // 如果组件正在关闭或清理中，不处理帧
    if (userClosing || isCleaningUp) {
      console.log('组件正在关闭或清理中，跳过帧渲染');
      frame.close();
      return;
    }
    
    if (streamCanvas.value && canvasContext) {
      try {
        // 确保canvas尺寸与视频帧尺寸精确匹配
        if (streamCanvas.value.width !== frame.codedWidth || 
            streamCanvas.value.height !== frame.codedHeight) {
          console.log(`调整Canvas尺寸以精确匹配帧: ${frame.codedWidth}x${frame.codedHeight}`);
          streamCanvas.value.width = frame.codedWidth;
          streamCanvas.value.height = frame.codedHeight;
          
          // 重置变换矩阵
          canvasContext.setTransform(1, 0, 0, 1, 0, 0);
        }
        
        // 清除整个Canvas
        canvasContext.clearRect(0, 0, streamCanvas.value.width, streamCanvas.value.height);
        
        // 精确绘制
        canvasContext.drawImage(frame, 0, 0);
        
        // 移除loading状态
        if (loading.value) {
          loading.value = false;
          emit('success', props.deviceId, { initialConnect: true });
        }
      } catch (err) {
        console.error('渲染帧错误:', err);
      } finally {
        frame.close();
      }
    } else {
      console.warn('Canvas或上下文不可用');
      frame.close();
    }
  };
  
  // 初始化视频解码器
  const initDecoder = () => {
    if (!window.VideoDecoder) {
      handleConnectionError('您的浏览器不支持 VideoDecoder API');
      return false;
    }
    
    try {
      // 关闭已存在的解码器
      if (videoDecoder && videoDecoder.state !== 'closed') {
        videoDecoder.close();
      }
      
      videoDecoder = new VideoDecoder({
        output: handleDecodedFrame,
        error: (error) => {
          console.error('解码器错误:', error);
          handleConnectionError(`解码器错误: ${error.message}`);
        }
      });
      
      // 重置状态
      videoBuffer = undefined;
      decoderState.value = {
        initialized: false,
        bufferedSPS: false,
        bufferedPPS: false,
        hadIDR: false
      };
      
      return true;
    } catch (err) {
      handleConnectionError(`初始化解码器失败: ${err.message}`);
      return false;
    }
  };
  
  // 开始连接
  const startConnect = async () => {
    if (!props.deviceId) return;
    
    loading.value = true;
    error.value = false;
    errorMessage.value = '';
    
    // 重置解码器状态
    decoderState.value = {
      initialized: false,
      bufferedSPS: false,
      bufferedPPS: false,
      hadIDR: false
    };
    
    // 触发事件指示正在加载中
    emit('loading-start', props.deviceId);
    
    try {
      // 调用开始串流接口
      const response = await startDeviceStream(props.deviceId);
      if (response.code !== 0) {
        throw new Error(response.message || '启动串流失败');
      }
      
      // 获取服务器返回的端口
      const port = response.data.port;
      
      // 清除已有的超时定时器
      clearConnectionTimeout();
      
      // 设置连接超时定时器
      startConnectionTimeout();
      
      // 确保Canvas已初始化
      nextTick(() => {
        console.log('开始连接设备:', props.deviceId);
        
        // 初始化Canvas
        initCanvas(videoConfig.value.width, videoConfig.value.height);
        
        // 初始化解码器
        if (initDecoder()) {
          // 连接WebSocket
          connectWebSocket(port);
        }
      });
    } catch (err) {
      handleConnectionError(err.message || '启动串流失败');
      ElMessage.error(err.message || '启动串流失败');
    }
  };
  
  // 连接WebSocket
  const connectWebSocket = (port) => {
    try {
      // 构建WebSocket URL，使用正确的代理路径
      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
      const wsUrl = `${protocol}//${window.location.host}/ws/ws/scrcpy?udid=${props.deviceId}&port=${port}`;
      console.log('连接WebSocket:', wsUrl);
      
      // 关闭已存在的连接
      if (wsConnection) {
        wsConnection.close();
      }
      
      // 创建新连接
      wsConnection = new WebSocket(wsUrl);
      wsConnection.binaryType = 'arraybuffer';
      
      // 连接事件处理
      wsConnection.onopen = handleWsOpen;
      wsConnection.onmessage = handleWsMessage;
      wsConnection.onerror = handleWsError;
      wsConnection.onclose = handleWsClose;
    } catch (err) {
      handleConnectionError(`WebSocket连接失败: ${err.message}`);
    }
  };
  
  // WebSocket 连接成功
  const handleWsOpen = () => {
    console.log('WebSocket连接成功');
  };
  
  // WebSocket 消息处理
  const handleWsMessage = (event) => {
    if (!event.data) return;
    
    // 检查是否是文本消息（JSON格式）
    if (typeof event.data === 'string') {
      try {
        // 提取时间戳前缀 [2025-04-10 20:51:44.909]
        const messageContent = event.data.substring(event.data.indexOf(']') + 1).trim();
        const jsonData = JSON.parse(messageContent);
        // console.log('收到WebSocket JSON消息:', jsonData);
        
        // 处理不同类型的消息
        if (jsonData.type === 'connected') {
          // 连接成功
          handleConnectionSuccess(jsonData.data);
        } else if (jsonData.type === 'videoSize') {
          // 视频尺寸信息
          handleVideoSizeInfo(jsonData.data);
        }
      } catch (err) {
        console.error('解析JSON消息失败:', err, event.data);
      }
    } else {
      // 二进制数据，处理视频流
      // console.log('收到二进制数据, 大小:', event.data.byteLength, 'bytes');
      decodeVideoData(new Uint8Array(event.data));
    }
  };
  
  // 处理连接成功信息
  const handleConnectionSuccess = (data) => {
    console.log(`设备 ${data.deviceId} 已连接，端口: ${data.port}`);
    
    // 清除连接超时定时器
    clearConnectionTimeout();
    
    // 标记为连接中
    loading.value = true;
    error.value = false;
    
    // 设置连接稳定期
    isStabilizing = true;
    if (connectionStabilityTimer) {
      clearTimeout(connectionStabilityTimer);
    }
  };
  
  // 处理视频尺寸信息
  const handleVideoSizeInfo = (data) => {
    console.log(`收到视频尺寸信息: ${data.width}x${data.height}, 编解码器: ${data.codec}`);
    
    // 更新视频配置
    videoConfig.value = {
      width: data.width,
      height: data.height,
      codec: data.codec || 'avc1.42001E' // 确保有默认编解码器
    };
    
    // 调整Canvas尺寸
    scaleCanvas(data.width, data.height);
    
    // 配置解码器
    if (videoDecoder && videoDecoder.state !== 'configured') {
      try {
        // 确保使用正确的编解码器配置
        const codecConfig = {
          codec: data.codec || 'avc1.42001E',
          optimizeForLatency: true,
          hardwareAcceleration: 'prefer-hardware'
        };
        
        console.log('配置解码器:', codecConfig);
        videoDecoder.configure(codecConfig);
        decoderState.value.initialized = true;
        console.log('视频解码器已配置完成');
        
        // 通知连接成功
        loading.value = false;
        
        // 如果稳定期定时器存在，清除它
        if (connectionStabilityTimer) {
          clearTimeout(connectionStabilityTimer);
        }
        
        // 设置新的稳定期定时器
        connectionStabilityTimer = setTimeout(() => {
          isStabilizing = false;
          // 如果稳定期结束仍然连接正常，显示成功消息
          if (!error.value) {
            // 显示成功消息提示
            ElMessage.success(`设备 ${props.deviceId} 连接成功`);
            emit('success', props.deviceId, { initialConnect: true });
            
            // 检查是否横屏
            if (data.width > data.height) {
              isLandscape.value = true;
            } else {
              isLandscape.value = false;
            }
            
            // 触发屏幕方向变化事件
            emit('orientation-change', {
              deviceId: props.deviceId,
              orientation: isLandscape.value ? 'landscape' : 'portrait',
              width: data.width,
              height: data.height
            });
          }
          connectionStabilityTimer = null;
        }, CONNECTION_STABILITY_PERIOD);
      } catch (err) {
        handleConnectionError(`视频解码器配置失败: ${err.message}`);
      }
    }
  };
  
  // WebSocket 错误处理
  const handleWsError = (event) => {
    console.error('WebSocket错误:', event);
    // 只有在非用户主动关闭的情况下才显示错误
    if (!userClosing) {
      handleConnectionError('WebSocket连接错误');
    }
  };
  
  // WebSocket 关闭处理
  const handleWsClose = (event) => {
    console.log('WebSocket连接关闭:', event.code, event.reason, '用户主动关闭:', userClosing);
    
    // 如果是用户主动关闭，不显示错误
    if (userClosing) {
      console.log('用户主动关闭，不显示错误');
      return;
    }
    
    // 如果不是手动关闭且没有错误，则认为连接意外断开
    if (!error.value) {
      // 如果是正常关闭(1000)或用户主动关闭(1001)，不显示错误
      if (event.code === 1000 || event.code === 1001) {
        console.log('WebSocket正常关闭');
      } else {
        handleConnectionError(`连接断开: ${event.reason || '未知原因'} (代码: ${event.code})`);
      }
    }
  };
  
  // 设置连接超时定时器
  const startConnectionTimeout = () => {
    connectionTimeoutTimer = setTimeout(() => {
      // 如果还在加载中，且没有错误，则认为连接超时
      if (loading.value && !error.value) {
        handleConnectionError('连接超时');
      }
    }, CONNECTION_TIMEOUT);
  };
  
  // 清除连接超时定时器
  const clearConnectionTimeout = () => {
    if (connectionTimeoutTimer) {
      clearTimeout(connectionTimeoutTimer);
      connectionTimeoutTimer = null;
    }
  };
  
  // 重试连接
  const retryConnect = () => {
    console.log('重试连接设备:', props.deviceId);
    
    // 重置状态
    error.value = false;
    loading.value = true;
    errorMessage.value = '';
    
    // 清除已有的超时定时器
    clearConnectionTimeout();
    
    // 关闭已有的WebSocket连接
    if (wsConnection) {
      wsConnection.close();
      wsConnection = null;
    }
    
    // 关闭已有的解码器
    if (videoDecoder && videoDecoder.state !== 'closed') {
      try {
        videoDecoder.close();
      } catch (err) {
        console.error('关闭解码器失败:', err);
      }
    }
    
    // 重置视频缓冲区
    videoBuffer = null;
    
    // 短暂延迟后开始新连接
    setTimeout(() => {
      startConnect();
    }, 500);
  };
  
  // 处理连接错误
  const handleConnectionError = (message) => {
    error.value = true;
    loading.value = false;
    errorMessage.value = message;
    
    // 停止串流
    if (props.deviceId) {
      stopDeviceStream(props.deviceId).catch(err => {
        console.error('停止串流失败:', err);
      });
    }
    
    // 触发错误事件
    emit('stream-error', {
      deviceId: props.deviceId,
      error: message
    });
    
    // 显示错误消息
    ElMessage.error(message);
  };
  
  // 组件挂载时，自动连接
  onMounted(() => {
    if (props.autoConnect && props.deviceId) {
      startConnect();
    }
  });
  
  // 组件卸载前，清理资源
  onBeforeUnmount(async () => {
    // 如果设置了跳过unmount销毁标志，或已经是用户主动关闭状态，则不执行清理
    if (!props.skipUnmountDestroy && !userClosing) {
      await destroyConnection();
    } else {
      console.log('跳过组件卸载时的destroyConnection调用');
    }
  });
  
  // 销毁WebSocket连接
  const destroyConnection = async (userInitiated = false) => {
    // 如果已经在清理中，则不重复执行
    if (isCleaningUp) {
      console.log('已经在清理中，跳过重复的destroyConnection调用');
      return Promise.resolve();
    }
    
    // 设置用户关闭标志
    userClosing = userInitiated;
    // 设置清理中状态
    isCleaningUp = true;
    
    console.log('执行destroyConnection, 用户主动关闭:', userInitiated);

    try {
      // 先停止解码操作
      if (videoDecoder) {
        // 1. 先重置解码状态，中断任何解码帧
        decoderState.value = {
          initialized: false,
          bufferedSPS: false,
          bufferedPPS: false,
          hadIDR: false
        };
        
        // 2. 正确关闭视频解码器，阻止新的解码操作
        if (videoDecoder.state !== 'closed') {
          try {
            console.log('关闭视频解码器');
            videoDecoder.close();
          } catch (err) {
            console.error('关闭解码器失败:', err);
          }
        }
        
        // 3. 清除视频缓冲区和帧
        videoBuffer = null;
      }
      
      // 4. 清理canvas上下文引用
      canvasContext = null;

      // 5. 关闭WebSocket连接
      if (wsConnection) {
        // 如果是用户主动关闭，阻止错误提示
        if (userInitiated) {
          console.log('用户主动关闭，移除WebSocket事件监听');
          // 完全移除事件处理器，防止触发错误提示
          wsConnection.onclose = null;
          wsConnection.onerror = null;
          wsConnection.onmessage = null;
        }
        
        try {
          // 使用1000正常关闭码关闭连接
          wsConnection.close(1000, "User closed");
        } catch (err) {
          console.error('关闭WebSocket连接失败:', err);
        }
        wsConnection = null;
      }
      
      // 6. 清理定时器
      clearConnectionTimeout();
      if (connectionStabilityTimer) {
        clearTimeout(connectionStabilityTimer);
      }

      // 7. 最后调用服务器端停止流
      if (props.deviceId) {
        try {
          await stopDeviceStream(props.deviceId);
        } catch (err) {
          console.error('停止串流失败:', err);
        }
      }
    } catch (err) {
      console.error('关闭资源时出错:', err);
    } finally {
      // 确保清理完成后重置标志
      isCleaningUp = false;
    }
    
    // 返回承诺，确保异步操作完成
    return Promise.resolve();
  };
  
  // 暴露方法给父组件
  defineExpose({
    retryConnect,
    destroyConnection
  });
  </script>
  
  <style scoped>
  .device-stream-container {
    width: v-bind('STREAM_WINDOW_CONFIG.DEFAULT_WIDTH + "px"') !important;
    height: v-bind('STREAM_WINDOW_CONFIG.DEFAULT_HEIGHT + "px"') !important;
    display: flex !important;
    align-items: center !important;
    justify-content: center !important;
    overflow: hidden !important;
    position: relative !important;
    background-color: #000;
    box-sizing: content-box !important; /* 确保边框不计入尺寸 */
  }
  
  .device-stream-frame {
    width: 100% !important;
    height: 100% !important;
    border: none !important;
    display: block !important;
    object-fit: contain !important;
    image-rendering: -webkit-optimize-contrast; /* 提高图像清晰度 */
    image-rendering: crisp-edges;
    touch-action: none; /* 禁用浏览器默认触摸行为 */
  }
  
  .device-stream-loading,
  .device-stream-error {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    z-index: 10;
  }
  
  .device-stream-loading {
    background: rgba(0, 0, 0, 0.8);
    color: white;
  }
  
  .device-stream-error {
    background: rgba(0, 0, 0, 0.9);
    color: #f56c6c;
    text-align: center;
    padding: 20px;
  }
  
  .loading-icon,
  .error-icon {
    font-size: 36px;
    margin-bottom: 16px;
  }
  
  .loading-icon {
    animation: rotating 2s linear infinite;
  }
  
  .retry-button {
    margin-top: 16px;
  }
  
  @keyframes rotating {
    from {
      transform: rotate(0deg);
    }
    to {
      transform: rotate(360deg);
    }
  }
  
  /* 横屏样式 */
  .device-stream-container.landscape {
    width: v-bind('STREAM_WINDOW_CONFIG.DEFAULT_HEIGHT + "px"') !important;
    height: v-bind('STREAM_WINDOW_CONFIG.DEFAULT_WIDTH + "px"') !important;
    box-sizing: content-box !important; /* 确保边框不计入尺寸 */
  }
  </style>
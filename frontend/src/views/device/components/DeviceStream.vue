<template>
    <div class="device-stream-container" :class="{ 'loading': loading, 'error': error, 'landscape': isLandscape }">  
      <iframe 
        ref="streamFrame"
        :src="iframeSrc"
        class="device-stream-frame"
        :style="{ visibility: loading || error ? 'hidden' : 'visible' }"
        scrolling="no"
        sandbox="allow-scripts allow-same-origin allow-forms"
      />
    </div>
  </template>
  
  <script setup>
  import { ref, computed, onMounted, onBeforeUnmount, nextTick } from 'vue';
  import { Loading, CircleClose } from '@element-plus/icons-vue';
  import { STREAM_WINDOW_CONFIG } from './config';
  import { ElMessage } from 'element-plus';
  
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
    }
  });
  
  const emit = defineEmits(['success', 'stream-error', 'loading-start', 'orientation-change']);
  
  const loading = ref(true);
  const error = ref(false);
  const errorMessage = ref('');
  const streamFrame = ref(null);
  const wscrcpyBaseUrl = computed(() => props.serverUrl || 'http://localhost:8000');
  const iframeSrc = ref('about:blank');
  const isLandscape = ref(false); // 添加横屏状态变量
  
  // 连接超时定时器
  let connectionTimeoutTimer = null;
  // 连接超时时间(毫秒)
  const CONNECTION_TIMEOUT = 10000;
  
  // 屏幕方向状态记忆和防抖
  const lastOrientationData = {
    orientation: null,
    timestamp: 0
  };
  
  // 连接状态稳定期
  let connectionStabilityTimer = null;
  const CONNECTION_STABILITY_PERIOD = 1000; // 连接成功后的稳定期(毫秒)
  let isStabilizing = false; // 是否处于连接稳定期
  
  // 添加设备实际尺寸变量
  const deviceWidth = ref(STREAM_WINDOW_CONFIG.PORTRAIT.WIDTH);
  const deviceHeight = ref(STREAM_WINDOW_CONFIG.PORTRAIT.HEIGHT);
  
  // 构建完整的串流URL
  const streamUrl = computed(() => {
    if (!props.deviceId) return 'about:blank';
    
    const encodedDeviceId = encodeURIComponent(props.deviceId);
    // 使用ws服务器环境变量
    const wsBaseUrl = import.meta.env.VITE_WSCRCPY_WS_SERVER || `ws://${wscrcpyBaseUrl.value.replace(/^https?:\/\//, '')}`;
    
    // 确保wsUrl是完整的URL
    let wsUrl;
    if (wsBaseUrl.startsWith('/')) {
      // 如果是相对路径，需要添加当前站点的协议和主机名
      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
      const host = window.location.host;
      wsUrl = encodeURIComponent(`${protocol}//${host}${wsBaseUrl}/?action=proxy-adb&remote=tcp:8886&udid=${props.deviceId}`);
    } else {
      // 如果已经是完整URL
      wsUrl = encodeURIComponent(`${wsBaseUrl}/?action=proxy-adb&remote=tcp:8886&udid=${props.deviceId}`);
    }

    // 构建完整的iframe URL
    let fullStreamUrl;
    if (wscrcpyBaseUrl.value.startsWith('/')) {
      // 如果服务器URL是相对路径，添加当前站点的协议和主机名
      const protocol = window.location.protocol;
      const host = window.location.host;
      fullStreamUrl = `${protocol}//${host}${wscrcpyBaseUrl.value}/#!action=stream&udid=${encodedDeviceId}&player=webcodecs&ws=${wsUrl}`;
    } else {
      // 如果已经是完整URL
      fullStreamUrl = `${wscrcpyBaseUrl.value}/#!action=stream&udid=${encodedDeviceId}&player=webcodecs&ws=${wsUrl}`;
    }
    
    console.log('构建串流URL:', fullStreamUrl);
    return fullStreamUrl;
  });
  
  // 开始连接
  const startConnect = () => {
    if (!props.deviceId) return;
    
    loading.value = true;
    error.value = false;
    errorMessage.value = '';
    
    // 触发事件指示正在加载中
    emit('loading-start', props.deviceId);
    
    // 清除已有的超时定时器
    clearConnectionTimeout();
    
    // 设置连接超时定时器
    startConnectionTimeout();
    
    // 使用nextTick确保DOM已渲染
    nextTick(() => {
      console.log('开始连接设备:', props.deviceId, '服务器地址:', wscrcpyBaseUrl.value);
      
      // 先手动强制设置iframe容器尺寸
      if (streamFrame.value && streamFrame.value.parentElement) {
        updateContainerSize(isLandscape.value);
        
        // 直接加载串流URL
        console.log('加载串流URL:', streamUrl.value);
        iframeSrc.value = streamUrl.value;
      } else {
        // 如果无法获取iframe元素，直接报错
        handleConnectionError('无法初始化连接界面');
      }
    });
  };
  
  // 根据横竖屏状态调整容器尺寸
  const updateContainerSize = (landscape) => {
    if (!streamFrame.value || !streamFrame.value.parentElement) return;
    
    const container = streamFrame.value.parentElement;
    const parentRect = container.parentElement?.getBoundingClientRect();
    
    if (!parentRect) return;
    
    // 根据横竖屏状态使用不同的尺寸计算
    let width, height;
    
    if (landscape) {
      // 横屏模式 - 使用设备实际尺寸
      width = Math.min(deviceWidth.value, parentRect.width);
      height = Math.min(deviceHeight.value + STREAM_WINDOW_CONFIG.BUTTON.HEIGHT, parentRect.height);
      console.log(`横屏模式 - 设置尺寸为 ${width}x${height} (设备实际尺寸+按钮高度)`);
    } else {
      // 竖屏模式 - 使用固定尺寸
      width = Math.min(STREAM_WINDOW_CONFIG.CANVAS.WIDTH, parentRect.width);
      height = Math.min(STREAM_WINDOW_CONFIG.CANVAS.HEIGHT + STREAM_WINDOW_CONFIG.BUTTON.HEIGHT, parentRect.height);
      console.log(`竖屏模式 - 设置尺寸为 ${width}x${height} (固定尺寸模式)`);
    }
    
    // 直接设置容器样式
    container.style.width = `${width}px`;
    container.style.height = `${height}px`;
    console.log(`直接设置容器样式: width=${width}px, height=${height}px`);
    
    // 设置iframe尺寸，确保完全匹配容器尺寸
    streamFrame.value.style.cssText = `
      width: 100% !important;
      height: 100% !important;
      visibility: ${loading.value || error.value ? 'hidden' : 'visible'} !important;
      display: block !important;
      object-fit: contain !important;
    `;
  };
  
  // 设置连接超时定时器
  const startConnectionTimeout = () => {
    connectionTimeoutTimer = setTimeout(() => {
      // 如果还在加载中，且没有错误，则认为连接超时
      if (loading.value && !error.value) {
        handleConnectionError('连接超时，wscrcpy服务可能不可用');
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
    iframeSrc.value = 'about:blank';
    
    // 清除已有的超时定时器
    clearConnectionTimeout();
    
    // 短暂延迟后开始新连接，确保iframe刷新
    setTimeout(() => {
      if (streamFrame.value) {
        // 确保iframe完全重置
        streamFrame.value.src = 'about:blank';
        // 再次启动连接
        startConnect();
      } else {
        // 如果找不到iframe，报告错误
        handleConnectionError('无法找到流视图');
      }
    }, 500);
  };
  
  // 处理连接错误
  const handleConnectionError = (errorMsg) => {
    if (error.value) return; // 防止重复触发
    
    console.error('串流连接错误:', errorMsg, '服务器地址:', wscrcpyBaseUrl.value);
    
    // 清除连接超时定时器
    clearConnectionTimeout();
    
    error.value = true;
    loading.value = false;
    errorMessage.value = errorMsg || '连接失败';
    
    // 显示错误消息提示
    ElMessage.error(`设备 ${props.deviceId} ${errorMessage.value}`);
    
    // 触发错误事件
    emit('stream-error', {
      deviceId: props.deviceId,
      error: `${errorMessage.value} (服务器: ${wscrcpyBaseUrl.value})` // 在错误信息中包含服务器地址
    });
  };
  
  // 处理来自iframe的消息事件
  const handleIframeMessage = (event) => {
    const data = event.data;
    
    // 根据消息类型处理
    if (data && data.type) {
      console.log(`收到iframe消息:`, data);
      
      // 处理WebSocket状态消息
      if (data.type === 'ws-status') {
        // 检查设备ID是否匹配
        if (data.udid && data.udid.includes(props.deviceId)) {
          switch(data.status) {
            case 'connecting':
              // 连接中状态
              loading.value = true;
              error.value = false;
              console.log(`设备 ${data.udid} 正在连接到 ${data.url}`);
              break;
              
            case 'connected':
              // 连接成功
              loading.value = false;
              error.value = false;
              console.log(`设备 ${data.udid} 已连接到 ${data.url}`);
              
              // 清除连接超时定时器
              clearConnectionTimeout();
              
              // 设置连接稳定期
              isStabilizing = true;
              if (connectionStabilityTimer) {
                clearTimeout(connectionStabilityTimer);
              }
              
              connectionStabilityTimer = setTimeout(() => {
                isStabilizing = false;
                // 如果稳定期结束仍然连接正常，显示成功消息
                if (!error.value) {
                  // 显示成功消息提示
                  ElMessage.success(`设备 ${props.deviceId} 连接成功`);
                  emit('success', props.deviceId, { initialConnect: true });
                }
                connectionStabilityTimer = null;
              }, CONNECTION_STABILITY_PERIOD);
              break;
              
            case 'disconnected':
              // 连接断开
              // 清除连接稳定期定时器
              if (connectionStabilityTimer) {
                clearTimeout(connectionStabilityTimer);
                connectionStabilityTimer = null;
              }
              
              if (!error.value) { // 避免重复触发错误
                // 如果是在连接稳定期内收到断开消息，给出更友好的提示
                if (isStabilizing) {
                  handleConnectionError(`设备连接失败: ${data.reason || '未知原因'} (代码: ${data.code || '未知'})`);
                } else {
                  handleConnectionError(`连接断开: ${data.reason || '未知原因'} (代码: ${data.code || '未知'})`);
                }
              }
              isStabilizing = false;
              console.log(`设备 ${data.udid} 断开连接，代码: ${data.code}, 原因: ${data.reason}`);
              break;
              
            case 'error':
              // 连接错误
              // 清除连接稳定期定时器
              if (connectionStabilityTimer) {
                clearTimeout(connectionStabilityTimer);
                connectionStabilityTimer = null;
              }
              
              handleConnectionError(`连接错误: ${data.reason || '未知错误'}`);
              isStabilizing = false;
              console.error(`设备 ${data.udid} 连接错误`);
              break;
          }
        }
      }
      
      // 处理屏幕方向消息
      else if (data.type === 'screen-orientation') {
        // 检查设备ID是否匹配
        if (data.udid && data.udid.includes(props.deviceId)) {
          // 检查方向是否真的变化了，并实施防抖 (500ms)
          const now = Date.now();
          if (data.status !== lastOrientationData.orientation || now - lastOrientationData.timestamp > 500) {
            console.log(`设备 ${data.udid} 屏幕方向: ${data.status}, 旋转: ${data.rotation}度, 尺寸: ${data.width}x${data.height}`);
            
            // 记录本次方向信息
            lastOrientationData.orientation = data.status;
            lastOrientationData.timestamp = now;
            
            // 更新设备实际尺寸
            if (data.width && data.height) {
              deviceWidth.value = data.width;
              deviceHeight.value = data.height;
              console.log(`更新设备实际尺寸: ${deviceWidth.value}x${deviceHeight.value}`);
            }
            
            // 更新横屏状态
            const newLandscapeState = data.status === 'landscape';
            const orientationChanged = isLandscape.value !== newLandscapeState;
            isLandscape.value = newLandscapeState;
            
            // 应用样式并强制重绘
            if (streamFrame.value && streamFrame.value.parentElement) {
              // 更新容器尺寸
              updateContainerSize(isLandscape.value);
            }
            
            // 向父组件发送屏幕方向变化事件
            emit('orientation-change', {
              deviceId: props.deviceId,
              orientation: data.status,
              rotation: data.rotation,
              width: data.width,
              height: data.height,
              actualWidth: deviceWidth.value,  // 传递设备实际宽度
              actualHeight: deviceHeight.value, // 传递设备实际高度
              fullHeight: deviceHeight.value + STREAM_WINDOW_CONFIG.BUTTON.HEIGHT // 真实高度加按钮高度
            });
          }
        }
      }
    }
  };
  
  // 组件挂载时，自动连接
  onMounted(() => {
    // 添加window消息事件监听器
    window.addEventListener('message', handleIframeMessage);
    
    if (props.autoConnect && props.deviceId) {
      startConnect();
    }
  });
  
  // 组件卸载前，清理资源
  onBeforeUnmount(() => {
    // 清除连接超时定时器
    clearConnectionTimeout();
    
    // 清除连接稳定期定时器
    if (connectionStabilityTimer) {
      clearTimeout(connectionStabilityTimer);
      connectionStabilityTimer = null;
    }
    
    // 移除消息事件监听器
    window.removeEventListener('message', handleIframeMessage);
    
    if (streamFrame.value) {
      streamFrame.value.src = 'about:blank';
    }
  });
  
  // 暴露方法给父组件
  defineExpose({
    retryConnect
  });
  </script>
  
  <style scoped>
  .device-stream-container {
    width: v-bind('STREAM_WINDOW_CONFIG.CANVAS.WIDTH + "px"') !important;
    height: v-bind('(STREAM_WINDOW_CONFIG.CANVAS.HEIGHT + STREAM_WINDOW_CONFIG.BUTTON.HEIGHT) + "px"') !important;
    display: flex !important;
    align-items: center !important;
    justify-content: center !important;
    overflow: hidden !important;
    position: relative !important;
    flex: none !important; /* 防止被弹性布局影响尺寸 */
  }
  
  .device-stream-frame {
    width: 100% !important;
    height: 100% !important;
    border: none !important;
    display: block !important;
    object-fit: contain !important;
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
  
  /* 横屏样式 - 保留类但不改变尺寸 */
  .device-stream-container.landscape {
    /* 不再改变宽高，使用与竖屏相同的尺寸 */
    transform: rotate(0deg); /* 确保容器本身不旋转 */
  }
  
  /* 横屏模式下的iframe */
  .landscape .device-stream-frame {
    /* 同样使用相同的尺寸 */
    width: 100% !important;
    height: 100% !important;
  }
  
  /* 修改 wscrcpy 中的样式 */
  :deep(.control-panel) {
    position: fixed !important;
    bottom: 0 !important;
    left: 0 !important;
    right: 0 !important;
    background: rgba(0, 0, 0, 0.8) !important;
    padding: 8px !important;
    z-index: 100 !important;
    display: flex !important;
    justify-content: center !important;
    gap: 8px !important;
  }
  
  :deep(.control-panel button) {
    background: rgba(255, 255, 255, 0.1) !important;
    border: 1px solid rgba(255, 255, 255, 0.2) !important;
    color: white !important;
    padding: 4px 8px !important;
    border-radius: 4px !important;
    cursor: pointer !important;
    transition: all 0.3s !important;
  }
  
  :deep(.control-panel button:hover) {
    background: rgba(255, 255, 255, 0.2) !important;
  }
  
  :deep(.control-panel button:active) {
    background: rgba(255, 255, 255, 0.3) !important;
  }
  
  :deep(.device-screen) {
    height: calc(100% - 40px) !important;
    width: 100% !important;
    display: flex !important;
    align-items: center !important;
    justify-content: center !important;
    overflow: hidden !important;
  }
  
  :deep(.device-screen canvas) {
    max-width: 100% !important;
    max-height: 100% !important;
    object-fit: contain !important;
  }
  </style>
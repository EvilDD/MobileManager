<template>
    <div class="device-stream-container" :class="{ 'loading': loading, 'error': error }">  
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
  
  const emit = defineEmits(['success', 'stream-error', 'loading-start']);
  
  const loading = ref(true);
  const error = ref(false);
  const errorMessage = ref('');
  const streamFrame = ref(null);
  const wscrcpyBaseUrl = computed(() => props.serverUrl || 'http://localhost:8000');
  const iframeSrc = ref('about:blank');
  
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
    
    // 使用nextTick确保DOM已渲染
    nextTick(() => {
      console.log('开始连接设备:', props.deviceId, '服务器地址:', wscrcpyBaseUrl.value);
      
      // 先手动强制设置iframe容器尺寸
      if (streamFrame.value && streamFrame.value.parentElement) {
        const container = streamFrame.value.parentElement;
        container.style.width = `${STREAM_WINDOW_CONFIG.DEFAULT_WIDTH}px`;
        container.style.height = `${STREAM_WINDOW_CONFIG.DEFAULT_HEIGHT}px`;
        container.style.minHeight = `${STREAM_WINDOW_CONFIG.MIN_HEIGHT}px`;
        
        // 直接设置iframe的style属性
        streamFrame.value.style.cssText = `
          width: ${STREAM_WINDOW_CONFIG.DEFAULT_WIDTH}px !important;
          height: ${STREAM_WINDOW_CONFIG.DEFAULT_HEIGHT}px !important;
          min-height: ${STREAM_WINDOW_CONFIG.MIN_HEIGHT}px !important;
          visibility: visible !important;
          display: block !important;
        `;
        
        // 直接加载串流URL
        console.log('加载串流URL:', streamUrl.value);
        iframeSrc.value = streamUrl.value;
      } else {
        // 如果无法获取iframe元素，直接报错
        handleConnectionError('无法初始化连接界面');
      }
    });
  };
  
  // 重试连接
  const retryConnect = () => {
    console.log('重试连接设备:', props.deviceId);
    
    // 重置状态
    error.value = false;
    loading.value = true;
    errorMessage.value = '';
    iframeSrc.value = 'about:blank';
    
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
    
    error.value = true;
    loading.value = false;
    errorMessage.value = errorMsg || '连接失败';
    
    // 触发错误事件
    emit('stream-error', {
      deviceId: props.deviceId,
      error: `${errorMessage.value} (服务器: ${wscrcpyBaseUrl.value})` // 在错误信息中包含服务器地址
    });
  };
  
  // 处理来自iframe的消息事件
  const handleIframeMessage = (event) => {
    const data = event.data;
    
    // 仅处理wscrcpy的WebSocket状态消息
    if (data && data.type === 'ws-status') {
      console.log(`收到WebSocket状态通知:`, data);
      
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
            emit('success', props.deviceId, { initialConnect: true });
            break;
            
          case 'disconnected':
            // 连接断开
            if (!error.value) { // 避免重复触发错误
              handleConnectionError(`连接断开: ${data.reason || '未知原因'} (代码: ${data.code || '未知'})`);
            }
            console.log(`设备 ${data.udid} 断开连接，代码: ${data.code}, 原因: ${data.reason}`);
            break;
            
          case 'error':
            // 连接错误
            handleConnectionError(`连接错误: ${data.reason || '未知错误'}`);
            console.error(`设备 ${data.udid} 连接错误`);
            break;
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
    position: relative;
    width: v-bind('STREAM_WINDOW_CONFIG.DEFAULT_WIDTH + "px"');
    height: v-bind('STREAM_WINDOW_CONFIG.DEFAULT_HEIGHT + "px"');
    display: flex;
    flex-direction: column;
    background: #000;
    min-height: v-bind('STREAM_WINDOW_CONFIG.MIN_HEIGHT + "px"');
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
  
  .device-stream-frame {
    flex: 1;
    width: v-bind('STREAM_WINDOW_CONFIG.DEFAULT_WIDTH + "px"');
    height: v-bind('STREAM_WINDOW_CONFIG.DEFAULT_HEIGHT + "px"');
    min-height: v-bind('STREAM_WINDOW_CONFIG.MIN_HEIGHT + "px"');
    border: none;
    background: transparent;
    display: block;
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
<template>
    <div class="device-stream-container" :class="{ 'loading': loading, 'error': error }">  
      <iframe 
        ref="streamFrame"
        :src="iframeSrc"
        class="device-stream-frame"
        :style="{ visibility: loading || error ? 'hidden' : 'visible' }"
        @load="onIframeLoaded"
        @error="onIframeError"
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
      default: 'http://localhost:8000'
    }
  });
  
  const emit = defineEmits(['success', 'stream-error', 'loading-start']);
  
  const loading = ref(true);
  const error = ref(false);
  const errorMessage = ref('');
  const streamFrame = ref(null);
  const wscrcpyBaseUrl = computed(() => props.serverUrl || 'http://localhost:8000');
  const iframeSrc = ref('about:blank');
  const connectionMonitorTimer = ref(null);
  
  // 构建完整的串流URL
  const streamUrl = computed(() => {
    if (!props.deviceId) return 'about:blank';
    
    const encodedDeviceId = encodeURIComponent(props.deviceId);
    const wsUrl = encodeURIComponent(`ws://${wscrcpyBaseUrl.value.replace(/^https?:\/\//, '')}/?action=proxy-adb&remote=tcp:8886&udid=${props.deviceId}`);
    return `${wscrcpyBaseUrl.value}/#!action=stream&udid=${encodedDeviceId}&player=webcodecs&ws=${wsUrl}`;
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
        
        // 设置超时定时器
        const connectionTimeout = setTimeout(() => {
          if (loading.value && !error.value) {
            handleConnectionError('连接超时，请检查设备状态和服务地址');
          }
        }, 15000); // 15秒超时
        
        // 直接加载串流URL，不再使用测试图片
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
    
    // 先清理可能存在的资源
    if (connectionMonitorTimer.value) {
      clearInterval(connectionMonitorTimer.value);
      connectionMonitorTimer.value = null;
    }
    
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
  
  // iframe加载完成
  const onIframeLoaded = () => {
    console.log('iframe加载状态:', loading.value);
    if (!loading.value) return; // 如果不是加载中状态则忽略
    
    // 延迟检查，确保内容已完全加载
    setTimeout(() => {
      try {
        // 检查iframe是否真正加载成功
        if (streamFrame.value) {
          let hasError = false;
          
          // 检查iframe的URL是否被重定向或包含错误标识
          try {
            const currentUrl = streamFrame.value.contentWindow?.location.href;
            if (currentUrl && (currentUrl.includes('error') || !currentUrl.includes(wscrcpyBaseUrl.value))) {
              console.log('检测到URL重定向或错误页面:', currentUrl);
              hasError = true;
            }
          } catch (e) {
            // 访问iframe的location可能因跨域导致错误
            console.log('无法检查iframe URL (可能是跨域限制)');
          }
          
          // 检查iframe标题是否包含错误信息
          try {
            const iframeTitle = streamFrame.value.contentDocument?.title;
            if (iframeTitle && (iframeTitle.includes('拒绝') || 
                                iframeTitle.includes('错误') || 
                                iframeTitle.includes('Error') || 
                                iframeTitle.includes('Refused'))) {
              console.log('检测到iframe标题包含错误信息:', iframeTitle);
              hasError = true;
            }
          } catch (e) {
            // 访问iframe的document可能因跨域限制导致错误
            console.log('无法检查iframe标题 (可能是跨域限制)');
          }
          
          // 检查iframe内容的高度，如果太小可能是错误页面
          if (streamFrame.value.clientHeight < 100) {
            console.log('iframe内容高度异常，可能是错误页面');
            hasError = true;
          }
          
          // 如果有错误，触发错误处理
          if (hasError) {
            handleConnectionError('连接被拒绝或发生错误，请检查设备和服务状态');
            return;
          }
          
          // 如果没有错误，设置加载完成状态，并且只触发一次成功事件
          if (loading.value) {
            console.log('iframe加载完成，触发success事件');
            loading.value = false;
            emit('success', props.deviceId, { initialConnect: true });
            
            // 开始连接监控
            startConnectionMonitor();
          }
        } else {
          handleConnectionError('无法获取流内容');
        }
      } catch (err) {
        console.error('检查iframe内容时发生错误:', err);
        handleConnectionError('连接过程中出现错误');
      }
    }, 1000); // 给予1秒时间让内容完全加载
  };
  
  // iframe加载错误
  const onIframeError = (event) => {
    console.error('iframe加载失败:', event);
    handleConnectionError('连接失败，服务器无响应');
  };
  
  // 启动定期连接监控
  const startConnectionMonitor = () => {
    // 清除可能存在的旧定时器
    if (connectionMonitorTimer.value) {
      clearInterval(connectionMonitorTimer.value);
      connectionMonitorTimer.value = null;
    }
    
    console.log('开始连接监控');
    
    // 每10秒检查一次连接状态，用于检测断流
    connectionMonitorTimer.value = window.setInterval(() => {
      if (streamFrame.value) {
        // 尝试使用simpler方法检测断流：检查iframe是否仍然加载原始URL
        const iframeCurrentSrc = streamFrame.value.src;
        
        try {
          // 尝试访问contentWindow，检查iframe是否还可以访问
          // 如果iframe被重定向或断流，可能会导致错误
          const contentWindow = streamFrame.value.contentWindow;
          
          // 检查iframe是否重定向
          if (contentWindow && contentWindow.location && 
              contentWindow.location.href !== streamUrl.value &&
              !contentWindow.location.href.includes(wscrcpyBaseUrl.value)) {
            handleConnectionError('串流已断开，可能是连接中断');
          }
        } catch (err) {
          // 跨域错误是正常的，但如果不是SecurityError，可能是断流
          if (err.name !== 'SecurityError') {
            console.warn('连接监控检测到异常:', err);
            handleConnectionError('连接异常，可能是串流已中断');
          }
        }
      }
    }, 10000);
  };
  
  // 处理连接错误
  const handleConnectionError = (errorMsg) => {
    if (error.value) return; // 防止重复触发
    
    console.error('串流连接错误:', errorMsg, '服务器地址:', wscrcpyBaseUrl.value);
    
    // 停止所有连接监控
    if (connectionMonitorTimer.value) {
      clearInterval(connectionMonitorTimer.value);
      connectionMonitorTimer.value = null;
    }
    
    error.value = true;
    loading.value = false;
    errorMessage.value = errorMsg || '连接失败';
    
    // 置空iframe源，防止继续尝试加载
    iframeSrc.value = '';
    
    // 触发错误事件
    emit('stream-error', {
      deviceId: props.deviceId,
      error: `${errorMessage.value} (服务器: ${wscrcpyBaseUrl.value})` // 在错误信息中包含服务器地址
    });
  };
  
  // 检查连接状态
  const checkConnectionState = () => {
    if (!streamFrame.value) return;
    
    try {
      // 先检查iframe是否能访问
      if (streamFrame.value.contentWindow) {
        // 如果iframe的src与预期URL一致，且能获取contentWindow，认为连接建立
        // 但由于跨域限制，我们无法直接检查内容，只能通过间接方式判断
        // 如果iframe已经加载，且没有重定向，大概率连接成功
        if (!error.value && loading.value && 
            iframeSrc.value === streamUrl.value && 
            streamFrame.value.contentWindow !== null) {
          // 假定加载成功
          console.log('连接状态检查: iframe已加载完成');
          loading.value = false;
          // 不再触发success事件，避免多次提示
        }
      }
    } catch (err) {
      console.error('检查连接状态失败:', err);
      // 由于跨域限制，不能直接判断连接状态，保持当前状态不变
      // 只有在明确失败的情况下才更新状态
      if (err.name !== 'SecurityError') {
        handleConnectionError('连接异常');
      }
    }
  };
  
  // 开始定期检查连接状态
  const startConnectionCheck = () => {
    // 清除可能存在的旧定时器
    stopConnectionCheck();
    
    // 创建新的定时器，但将间隔改长，避免频繁报错
    // 并且只检查一次
    connectionMonitorTimer.value = window.setTimeout(() => {
      checkConnectionState();
    }, 5000);
  };
  
  // 停止连接状态检查
  const stopConnectionCheck = () => {
    if (connectionMonitorTimer.value !== null) {
      window.clearTimeout(connectionMonitorTimer.value);
      connectionMonitorTimer.value = null;
    }
  };
  
  // 组件挂载时，自动连接
  onMounted(() => {
    if (props.autoConnect && props.deviceId) {
      startConnect();
    }
  });
  
  // 组件卸载前，清理资源
  onBeforeUnmount(() => {
    stopConnectionCheck();
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
<template>
    <div class="device-stream-container" :class="{ 'loading': loading }">
      <div v-if="loading" class="device-stream-loading">
        <el-icon class="loading-icon"><Loading /></el-icon>
        <span>加载中...</span>
      </div>
      
      <iframe 
        ref="streamFrame"
        :src="iframeSrc"
        class="device-stream-frame"
        :style="{ visibility: loading ? 'hidden' : 'visible' }"
        @load="onIframeLoaded"
        scrolling="no"
        sandbox="allow-scripts allow-same-origin allow-forms"
      />
    </div>
  </template>
  
  <script setup>
  import { ref, computed, onMounted, onBeforeUnmount, nextTick } from 'vue';
  import { Loading } from '@element-plus/icons-vue';
  import { STREAM_WINDOW_CONFIG } from './config';
  
  const props = defineProps({
    deviceId: {
      type: String,
      required: true
    },
    autoConnect: {
      type: Boolean,
      default: true
    }
  });
  
  const emit = defineEmits(['stream-ready', 'stream-error']);
  
  const loading = ref(true);
  const streamFrame = ref(null);
  const wscrcpyBaseUrl = 'http://localhost:8000';
  const iframeSrc = ref('about:blank');
  
  // 构建完整的串流URL
  const streamUrl = computed(() => {
    if (!props.deviceId) return 'about:blank';
    
    const encodedDeviceId = encodeURIComponent(props.deviceId);
    const wsUrl = encodeURIComponent(`ws://localhost:8000/?action=proxy-adb&remote=tcp:8886&udid=${props.deviceId}`);
    return `${wscrcpyBaseUrl}/#!action=stream&udid=${encodedDeviceId}&player=webcodecs&ws=${wsUrl}`;
  });
  
  // 组件挂载时，自动连接
  onMounted(() => {
    if (props.autoConnect && props.deviceId) {
      loading.value = true;
      
      // 使用nextTick确保DOM已渲染
      nextTick(() => {
        console.log('组件挂载时iframe尺寸:', {
          iframe: streamFrame.value,
          parent: streamFrame.value?.parentElement
        });
        
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
          
          // 打印设置后的尺寸
          console.log('设置尺寸后:', {
            width: streamFrame.value.offsetWidth,
            height: streamFrame.value.offsetHeight,
            style: streamFrame.value.style.cssText,
            parentWidth: container.offsetWidth,
            parentHeight: container.offsetHeight
          });
          
          // 既然已经有尺寸了，直接加载实际URL
          console.log('直接加载实际URL:', streamUrl.value);
          iframeSrc.value = streamUrl.value;
        }
      });
    }
  });
  
  // iframe加载完成
  const onIframeLoaded = () => {
    console.log('iframe loaded:', iframeSrc.value);
    
    if (iframeSrc.value === streamUrl.value) {
      // 加载实际URL完成
      setTimeout(() => {
        loading.value = false;
        emit('stream-ready');
      }, 500);
    }
  };
  
  // 组件卸载前，清理资源
  onBeforeUnmount(() => {
    if (streamFrame.value) {
      streamFrame.value.src = 'about:blank';
    }
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
  
  .device-stream-loading {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    background: rgba(0, 0, 0, 0.8);
    z-index: 10;
  }
  
  .loading-icon {
    font-size: 24px;
    margin-bottom: 8px;
    animation: rotating 2s linear infinite;
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
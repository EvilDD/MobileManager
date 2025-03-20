<template>
    <div class="device-stream-container" :class="{ 'loading': loading }">
      <div v-if="loading" class="device-stream-loading">
        <el-icon class="loading-icon"><Loading /></el-icon>
        <span>加载中...</span>
      </div>
      
      <iframe 
        v-show="!loading" 
        ref="streamFrame"
        :src="streamUrl"
        class="device-stream-frame"
        @load="onIframeLoaded"
        scrolling="no"
        sandbox="allow-scripts allow-same-origin allow-forms"
      />
    </div>
  </template>
  
  <script setup>
  import { ref, computed, onMounted, onBeforeUnmount } from 'vue';
  import { Loading } from '@element-plus/icons-vue';
  
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
  
  // 构建完整的串流URL
  const streamUrl = computed(() => {
    if (!props.deviceId) return 'about:blank';
    
    const encodedDeviceId = encodeURIComponent(props.deviceId);
    const wsUrl = encodeURIComponent(`ws://localhost:8000/?action=proxy-adb&remote=tcp:8886&udid=${props.deviceId}`);
    return `${wscrcpyBaseUrl}/#!action=stream&udid=${encodedDeviceId}&player=webcodecs&ws=${wsUrl}`;
  });
  
  // iframe加载完成
  const onIframeLoaded = () => {
    setTimeout(() => {
      loading.value = false;
      emit('stream-ready');
    }, 500);
  };
  
  // 组件挂载时，自动连接
  onMounted(() => {
    if (props.autoConnect && props.deviceId) {
      loading.value = true;
    }
  });
  
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
    width: 100%;
    height: 100%;
    display: flex;
    flex-direction: column;
    background: #000;
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
    width: 100%;
    height: 0;
    min-height: 0;
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
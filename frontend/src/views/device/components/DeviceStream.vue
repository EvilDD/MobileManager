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
    overflow: hidden;
    background-color: transparent;
    display: flex;
    justify-content: center;
    align-items: center;
  }
  
  .device-stream-loading {
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    display: flex;
    flex-direction: column;
    justify-content: center;
    align-items: center;
    background-color: rgba(0, 0, 0, 0.7);
    color: #fff;
    z-index: 2;
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
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    border: none;
    background-color: transparent;
    transform-origin: center;
    /* 修复chrome的iframe渲染问题 */
    -webkit-transform: translateZ(0);
    -webkit-backface-visibility: hidden;
  }
  </style>
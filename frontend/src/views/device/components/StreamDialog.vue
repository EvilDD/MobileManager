<template>
  <div v-if="modelValue" class="stream-dialog-container">
    <div class="stream-backdrop" />
    <div class="stream-window" ref="dialogRef">
      <div class="stream-header" @mousedown="startDrag">
        <span class="stream-title">{{ title }}</span>
        <button class="close-button" @click="closeDialog">
          <el-icon><Close /></el-icon>
        </button>
      </div>
      
      <!-- 加载中状态 -->
      <div v-if="isLoading && !streamError" class="stream-loading">
        <el-icon class="loading-icon"><Loading /></el-icon>
        <span>正在连接中，请稍候...</span>
      </div>
      
      <!-- 错误状态界面 -->
      <div v-if="streamError" class="stream-error">
        <el-icon class="error-icon"><CircleClose /></el-icon>
        <span>{{ errorMessage }}</span>
        <div class="error-buttons">
          <el-button type="primary" size="small" @click="retryConnection" class="retry-button">
            重试连接
          </el-button>
          <el-button type="danger" size="small" @click="closeDialog" class="close-error-button">
            关闭窗口
          </el-button>
        </div>
      </div>
      
      <div class="phone-frame">
        <div class="phone-notch" />
        <device-stream 
          v-if="visible && deviceId && isReady" 
          ref="streamRef"
          :device-id="deviceId" 
          :auto-connect="true"
          :server-url="serverUrl"
          @success="onStreamReady"
          @stream-error="onStreamError"
          @loading-start="onLoadingStart"
        />
      </div>
      <div class="resize-handle" @mousedown="startResize" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onBeforeUnmount } from 'vue';
import DeviceStream from './DeviceStream.vue';
import { Close, Loading, CircleClose } from '@element-plus/icons-vue';
import { STREAM_WINDOW_CONFIG } from './config';
import { ElMessage } from 'element-plus';

const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false
  },
  deviceId: {
    type: String,
    default: ''
  },
  serverUrl: {
    type: String,
    default: import.meta.env.VITE_WSCRCPY_SERVER || 'http://localhost:8000'
  }
});

const emit = defineEmits(['update:modelValue', 'closed']);

// 组件内部是否可见（用于延迟销毁iframe）
const visible = ref(false);
// 对话框是否准备好（用于延迟加载iframe）
const isReady = ref(false);
// 流是否准备好
const streamReady = ref(false);
// 流连接错误
const streamError = ref(false);
// 错误信息
const errorMessage = ref('');
const dialogRef = ref<HTMLElement | null>(null);
const streamRef = ref(null);

// 计算属性: 标题
const title = computed(() => {
  if (streamError.value) {
    return `连接设备 ${props.deviceId || ''} 失败`;
  }
  if (!streamReady.value) {
    return `正在连接设备 ${props.deviceId || ''}...`;
  }
  return `连接设备 ${props.deviceId || ''} 成功`;
});

// 拖拽相关状态
const isDragging = ref(false);
const dragOffset = ref({ x: 0, y: 0 });

// 缩放相关状态
const isResizing = ref(false);
const initialSize = ref({ width: 0, height: 0 });
const initialPosition = ref({ x: 0, y: 0 });

// 加载中状态
const isLoading = ref(false);

// 监听显示状态
watch(() => props.modelValue, (newVal) => {
  if (newVal) {
    visible.value = true;
    streamReady.value = false;
    streamError.value = false;
    errorMessage.value = '';
    isLoading.value = true; // 显示加载状态
    
    // 添加一个小延迟，确保对话框完全打开后再加载iframe
    setTimeout(() => {
      isReady.value = true;
    }, 100);
  } else {
    isReady.value = false;
    streamReady.value = false;
    streamError.value = false;
    isLoading.value = false;
    setTimeout(() => {
      visible.value = false;
      emit('closed');
    }, 200);
  }
});

// 重试连接
const retryConnection = () => {
  if (streamRef.value && streamRef.value.retryConnect) {
    // 显示加载状态
    isLoading.value = true;
    
    // 短暂延迟后再清除错误状态，避免UI闪烁
    setTimeout(() => {
      streamError.value = false;
      errorMessage.value = '';
    }, 100);
    
    // 直接调用流组件的重试方法
    streamRef.value.retryConnect();
    
    // 设置超时检查，确保错误状态能够同步
    setTimeout(() => {
      // 如果5秒后仍然没有成功或新的错误状态，恢复错误UI
      if (!streamReady.value && !streamError.value) {
        isLoading.value = false;
        streamError.value = true;
        errorMessage.value = '连接超时，请再次尝试';
        ElMessage.error('连接超时，请再次尝试');
      }
    }, 5000);
  }
};

// 流加载完成回调
const onStreamReady = (deviceId, data) => {
  console.log('收到流加载完成事件:', deviceId, data);
  streamReady.value = true;
  streamError.value = false;
  isLoading.value = false;
  
  // 只有首次连接成功才显示消息
  if (data?.initialConnect) {
    console.log('显示连接成功消息');
    // 显示成功提示消息
    ElMessage.success(`连接设备 ${props.deviceId} 成功`);
  }
};

// 流加载错误回调
const onStreamError = (errorData) => {
  streamError.value = true;
  isLoading.value = false;
  errorMessage.value = errorData.error || '连接失败';
  
  // 显示错误消息
  ElMessage.error(errorMessage.value);
};

// 关闭窗口
const closeDialog = () => {
  // 立即清除错误状态以便正确关闭
  streamError.value = false;
  errorMessage.value = '';
  
  // 更新父组件的v-model值
  emit('update:modelValue', false);
  
  // 短暂延迟后处理完全关闭
  setTimeout(() => {
    visible.value = false;
    emit('closed');
  }, 200);
};

// 开始拖拽
const startDrag = (e: MouseEvent) => {
  if (!dialogRef.value) return;
  
  isDragging.value = true;
  const rect = dialogRef.value.getBoundingClientRect();
  
  // 在开始拖拽时，先设置当前位置，移除居中定位
  dialogRef.value.style.top = `${rect.top}px`;
  dialogRef.value.style.left = `${rect.left}px`;
  dialogRef.value.style.transform = 'none';
  
  dragOffset.value = {
    x: e.clientX - rect.left,
    y: e.clientY - rect.top
  };
  
  document.addEventListener('mousemove', handleDrag);
  document.addEventListener('mouseup', stopDrag);
};

// 处理拖拽
const handleDrag = (e: MouseEvent) => {
  if (!isDragging.value || !dialogRef.value) return;
  
  const newX = e.clientX - dragOffset.value.x;
  const newY = e.clientY - dragOffset.value.y;
  
  dialogRef.value.style.left = `${newX}px`;
  dialogRef.value.style.top = `${newY}px`;
};

// 停止拖拽
const stopDrag = () => {
  isDragging.value = false;
  document.removeEventListener('mousemove', handleDrag);
  document.removeEventListener('mouseup', stopDrag);
};

// 开始缩放
const startResize = (e: MouseEvent) => {
  if (!dialogRef.value) return;
  
  isResizing.value = true;
  const rect = dialogRef.value.getBoundingClientRect();
  
  initialSize.value = {
    width: rect.width,
    height: rect.height
  };
  
  initialPosition.value = {
    x: e.clientX,
    y: e.clientY
  };
  
  document.addEventListener('mousemove', handleResize);
  document.addEventListener('mouseup', stopResize);
};

// 处理缩放
const handleResize = (e: MouseEvent) => {
  if (!isResizing.value || !dialogRef.value) return;
  
  const deltaX = e.clientX - initialPosition.value.x;
  const deltaY = e.clientY - initialPosition.value.y;
  
  const newWidth = Math.max(
    STREAM_WINDOW_CONFIG.MIN_WIDTH,
    Math.min(
      STREAM_WINDOW_CONFIG.MAX_WIDTH,
      initialSize.value.width + deltaX
    )
  );
  const newHeight = Math.max(
    STREAM_WINDOW_CONFIG.MIN_HEIGHT,
    Math.min(
      STREAM_WINDOW_CONFIG.MAX_HEIGHT,
      initialSize.value.height + deltaY
    )
  );
  
  dialogRef.value.style.width = `${newWidth}px`;
  dialogRef.value.style.height = `${newHeight}px`;
};

// 停止缩放
const stopResize = () => {
  isResizing.value = false;
  document.removeEventListener('mousemove', handleResize);
  document.removeEventListener('mouseup', stopResize);
};

// 组件卸载前清理事件监听
onBeforeUnmount(() => {
  document.removeEventListener('mousemove', handleDrag);
  document.removeEventListener('mouseup', stopDrag);
  document.removeEventListener('mousemove', handleResize);
  document.removeEventListener('mouseup', stopResize);
});

// 暴露方法给父组件
defineExpose({
  retryConnection
});

// 加载开始回调
const onLoadingStart = (deviceId) => {
  console.log('设备开始加载:', deviceId);
  isLoading.value = true;
  streamReady.value = false;
  streamError.value = false;
};
</script>

<style scoped>
.stream-dialog-container {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  z-index: 1999;
}

.stream-backdrop {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background-color: rgba(0, 0, 0, 0.6);
  backdrop-filter: blur(8px);
  -webkit-backdrop-filter: blur(8px);
}

.stream-window {
  position: fixed;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: v-bind('STREAM_WINDOW_CONFIG.DEFAULT_WIDTH + "px"');
  height: v-bind('STREAM_WINDOW_CONFIG.DEFAULT_HEIGHT + "px"');
  background-color: #000;
  border-radius: 20px;
  overflow: hidden;
  box-shadow: 0 25px 50px rgba(0,0,0,0.5);
  display: flex;
  flex-direction: column;
  z-index: 2000;
  user-select: none;
}

.stream-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  background-color: #1a1a1a;
  cursor: move;
  position: relative;
  z-index: 20;
}

.stream-title {
  color: #fff;
  font-size: 16px;
  font-weight: 500;
}

.close-button {
  width: 24px;
  height: 24px;
  border: none;
  background: transparent;
  color: #fff;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  transition: all 0.3s ease;
  padding: 0;
  position: relative;
  z-index: 25;
}

.close-button:hover {
  background-color: rgba(255, 255, 255, 0.1);
  transform: rotate(90deg);
}

.phone-frame {
  position: relative;
  flex: 1;
  background-color: #000;
  border-radius: 36px;
  border: 6px solid #1a1a1a;
  margin: 0;
  box-shadow: inset 0 0 10px rgba(0,0,0,0.6);
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.phone-notch {
  position: absolute;
  top: 0;
  left: 50%;
  transform: translateX(-50%);
  width: 120px;
  height: 20px;
  background-color: #1a1a1a;
  border-bottom-left-radius: 10px;
  border-bottom-right-radius: 10px;
  z-index: 10;
  box-shadow: 0 2px 8px rgba(0,0,0,0.3);
}

:deep(.device-stream-container) {
  flex: 1;
  border-radius: 30px;
  overflow: hidden;
  background-color: #000;
}

:deep(.device-stream-frame) {
  border-radius: 30px;
}

/* 添加缩放手柄样式 */
.resize-handle {
  position: absolute;
  right: 0;
  bottom: 0;
  width: 20px;
  height: 20px;
  cursor: nwse-resize;
  z-index: 10;
}

.resize-handle::after {
  content: '';
  position: absolute;
  right: 4px;
  bottom: 4px;
  width: 12px;
  height: 12px;
  border-right: 2px solid rgba(255, 255, 255, 0.5);
  border-bottom: 2px solid rgba(255, 255, 255, 0.5);
}

.retry-button {
  margin: 0;
}

.error-buttons {
  display: flex;
  gap: 10px;
  margin-top: 16px;
}

.stream-error {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  background-color: rgba(0, 0, 0, 0.8);
  color: #f56c6c;
  z-index: 15; /* 确保在header下面，不遮挡关闭按钮 */
  text-align: center;
  padding: 20px;
}

.error-icon {
  font-size: 48px;
  margin-bottom: 16px;
  color: #f56c6c;
}

.close-error-button {
  background-color: #f56c6c;
  color: #fff;
  border: none;
  padding: 8px 16px;
  border-radius: 4px;
  cursor: pointer;
  transition: background-color 0.3s ease;
}

.close-error-button:hover {
  background-color: #e65c5c;
}

.stream-loading {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  background-color: rgba(0, 0, 0, 0.8);
  color: #fff;
  z-index: 15;
  text-align: center;
  padding: 20px;
}

.loading-icon {
  font-size: 48px;
  margin-bottom: 16px;
  color: #fff;
  animation: spin 2s linear infinite;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

@media (max-width: 768px) {
  .stream-window {
    width: 95vw;
    height: 95vh;
  }
  
  .phone-frame {
    border-radius: 20px;
    border-width: 4px;
  }
  
  .phone-notch {
    width: 100px;
    height: 16px;
    border-bottom-left-radius: 8px;
    border-bottom-right-radius: 8px;
  }
}
</style> 
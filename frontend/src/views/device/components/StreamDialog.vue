<template>
  <div v-if="modelValue" class="stream-dialog-container">
    <div class="stream-backdrop" />
    <div class="stream-window" ref="dialogRef" :class="{ 'landscape': isLandscape }">
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
      
      <div class="phone-frame" :class="{ 'landscape': isLandscape }">
        <device-stream 
          v-if="visible && deviceId && isReady" 
          ref="streamRef"
          :device-id="deviceId" 
          :auto-connect="true"
          :server-url="serverUrl"
          @success="onStreamReady"
          @stream-error="onStreamError"
          @loading-start="onLoadingStart"
          @orientation-change="onOrientationChange"
        />
      </div>
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
// 是否横屏
const isLandscape = ref(false);
const dialogRef = ref<HTMLElement | null>(null);
const streamRef = ref(null);

// 设备canvas实际尺寸（从设备获取）
const canvasWidth = ref(STREAM_WINDOW_CONFIG.PORTRAIT.WIDTH); // 默认值
const canvasHeight = ref(STREAM_WINDOW_CONFIG.PORTRAIT.HEIGHT); // 默认值

// 计算属性: 标题
const title = computed(() => {
  if (streamError.value) {
    return `连接设备 ${props.deviceId || ''} 失败`;
  }
  if (!streamReady.value) {
    return `正在连接设备 ${props.deviceId || ''}...`;
  }
  return `${isLandscape.value ? '横屏' : '竖屏'} - 设备 ${props.deviceId || ''}`;
});

// 拖拽相关状态
const isDragging = ref(false);
const dragOffset = ref({ x: 0, y: 0 });

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

// 处理屏幕方向变化
const onOrientationChange = (data) => {
  console.log('收到屏幕方向变化:', data);
  
  // 获取上一次的方向状态
  const previousOrientation = isLandscape.value ? 'landscape' : 'portrait';
  
  // 更新屏幕方向状态
  isLandscape.value = data.orientation === 'landscape';
  
  // 更新canvas尺寸
  if (data.width && data.height) {
    canvasWidth.value = data.width;
    canvasHeight.value = data.height;
    console.log(`更新canvas尺寸: ${canvasWidth.value}x${canvasHeight.value}`);
  }
  
  // 方向真正发生变化时才显示消息
  if (streamReady.value && previousOrientation !== data.orientation) {
    ElMessage.info(`设备 ${props.deviceId} 切换到${isLandscape.value ? '横屏' : '竖屏'}模式`);
  }
  
  // 调整对话框大小以适应新的屏幕方向
  adjustWindowForOrientation(data);
  
  // 延迟检查尺寸
  setTimeout(() => {
    const container = document.querySelector('.device-stream-container');
    if (container) {
      const style = window.getComputedStyle(container);
      console.log(`设备容器尺寸 - 宽: ${style.width}, 高: ${style.height}, 模式: ${isLandscape.value ? '横屏' : '竖屏'}`);
    }
  }, 500);
};

// 根据屏幕方向调整窗口大小
const adjustWindowForOrientation = (data) => {
  if (!dialogRef.value) return;
  
  // 不再手动计算尺寸，而是依赖CSS类来设置正确的尺寸
  if (data.orientation === 'landscape') {
    // 横屏模式
    isLandscape.value = true;
  } else {
    // 竖屏模式
    isLandscape.value = false;
  }
  
  // 给UI一点时间重新渲染
  setTimeout(() => {
    console.log(`窗口调整为 ${isLandscape.value ? '横屏' : '竖屏'} 模式`);
  }, 50);
};

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
  }
};

// 流加载完成回调
const onStreamReady = (deviceId, data) => {
  console.log('收到流加载完成事件:', deviceId, data);
  streamReady.value = true;
  streamError.value = false;
  isLoading.value = false;
};

// 流加载错误回调
const onStreamError = (errorData) => {
  streamError.value = true;
  isLoading.value = false;
  errorMessage.value = errorData.error || '连接失败';
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

// 组件卸载前清理事件监听
onBeforeUnmount(() => {
  document.removeEventListener('mousemove', handleDrag);
  document.removeEventListener('mouseup', stopDrag);
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
  /* 基于phone-frame计算 */
  width: v-bind('(canvasWidth + 12) + "px"');
  height: v-bind('(canvasHeight + STREAM_WINDOW_CONFIG.BUTTON.HEIGHT + 12 + 44) + "px"');
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
  height: 44px; /* 明确指定高度 */
  box-sizing: border-box; /* 确保padding包含在高度内 */
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
  /* 移除flex:1，改为固定尺寸 */
  flex: none;
  background-color: #000;
  border-radius: 36px;
  border: 6px solid #1a1a1a;
  margin: 0;
  box-shadow: inset 0 0 10px rgba(0,0,0,0.6);
  overflow: hidden;
  display: flex;
  flex-direction: column;
  transition: all 0.3s ease;
  box-sizing: border-box;
  /* 使用canvas尺寸计算phone-frame尺寸 */
  width: v-bind('(canvasWidth + 12) + "px"');
  height: v-bind('(canvasHeight + STREAM_WINDOW_CONFIG.BUTTON.HEIGHT + 12) + "px"');
}

/* 横屏样式 */
.phone-frame.landscape {
  border-radius: 20px;
  /* 横屏模式也使用canvas尺寸 */
  width: v-bind('(canvasWidth + 12) + "px"');
  height: v-bind('(canvasHeight + STREAM_WINDOW_CONFIG.BUTTON.HEIGHT + 12) + "px"');
}

:deep(.device-stream-container) {
  border-radius: 30px;
  overflow: hidden;
  background-color: #000;
  display: flex !important;
  align-items: center !important;
  justify-content: center !important;
  /* 注意：不设置宽高，让组件自己控制尺寸 */
}

:deep(.device-stream-container.landscape) {
  /* 确保横屏模式下样式正确应用，使用实际canvas尺寸 */
  width: v-bind('canvasWidth + "px"') !important;
  height: v-bind('(canvasHeight + STREAM_WINDOW_CONFIG.BUTTON.HEIGHT) + "px"') !important;
}

:deep(.device-stream-frame) {
  border-radius: 30px;
  width: 100% !important;
  height: 100% !important;
  object-fit: contain !important;
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
}

/* 横屏样式 */
.stream-window.landscape {
  width: v-bind('(canvasWidth + 12) + "px"');
  height: v-bind('(canvasHeight + STREAM_WINDOW_CONFIG.BUTTON.HEIGHT + 12 + 44) + "px"');
}
</style> 
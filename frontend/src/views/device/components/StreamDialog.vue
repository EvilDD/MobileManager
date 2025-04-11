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
      
      <div class="phone-frame" :class="{ 'landscape': isLandscape }">
        <div class="phone-notch" />
        <device-stream 
          v-if="visible && deviceId" 
          ref="streamRef"
          :device-id="deviceId" 
          :auto-connect="true"
          :server-url="serverUrl"
          @success="onStreamReady"
          @stream-error="onStreamError"
          @orientation-change="onOrientationChange"
        />
      </div>
      <div class="resize-handle" @mousedown="startResize" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onBeforeUnmount } from 'vue';
import DeviceStream from './DeviceStream.vue';
import { Close } from '@element-plus/icons-vue';
import { ElMessage } from 'element-plus';
import { STREAM_WINDOW_CONFIG } from './config';

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
// 流是否准备好
const streamReady = ref(false);
// 流连接错误
const streamError = ref(false);
// 是否横屏
const isLandscape = ref(false);
const dialogRef = ref<HTMLElement | null>(null);
const streamRef = ref(null);

// 计算属性: 标题
const title = computed(() => {
  if (streamError.value) {
    return `设备 ${props.deviceId || ''} - 连接失败`;
  }
  if (!streamReady.value) {
    return `设备 ${props.deviceId || ''} - 连接中`;
  }
  return `设备 ${props.deviceId || ''} - ${isLandscape.value ? '横屏' : '竖屏'}`;
});

// 拖拽相关状态
const isDragging = ref(false);
const dragOffset = ref({ x: 0, y: 0 });

// 缩放相关状态
const isResizing = ref(false);
const initialSize = ref({ width: 0, height: 0 });
const initialPosition = ref({ x: 0, y: 0 });

// 监听显示状态
watch(() => props.modelValue, (newVal) => {
  if (newVal) {
    visible.value = true;
    streamReady.value = false;
    streamError.value = false;
  } else {
    visible.value = false;
    emit('closed');
  }
});

// 处理屏幕方向变化
const onOrientationChange = (data) => {
  console.log('收到屏幕方向变化:', data);
  
  // 获取上一次的方向状态
  const previousOrientation = isLandscape.value ? 'landscape' : 'portrait';
  
  // 更新屏幕方向状态
  isLandscape.value = data.orientation === 'landscape';
  
  // 方向真正发生变化时才显示消息
  if (streamReady.value && previousOrientation !== data.orientation) {
    ElMessage.info(`设备 ${props.deviceId} 切换到${isLandscape.value ? '横屏' : '竖屏'}模式`);
  }
  
  // 根据方向更新窗口尺寸
  adjustDialogSizeForOrientation();
};

// 根据屏幕方向调整对话框尺寸
const adjustDialogSizeForOrientation = () => {
  if (!dialogRef.value) return;
  
  // 获取phone-frame元素
  const phoneFrame = dialogRef.value.querySelector('.phone-frame') as HTMLElement;
  if (!phoneFrame) return;
  
  // 设置phone-frame尺寸 - 确保完全匹配DEFAULT_WIDTH/HEIGHT
  const HEADER_HEIGHT = 36; // 标题栏高度
  
  // 重要：phone-frame必须完全匹配container尺寸，不添加任何边距
  if (isLandscape.value) {
    // 横屏模式 - 交换宽高
    phoneFrame.style.width = `${STREAM_WINDOW_CONFIG.DEFAULT_HEIGHT}px`;
    phoneFrame.style.height = `${STREAM_WINDOW_CONFIG.DEFAULT_WIDTH}px`;
    
    // 调整对话框尺寸，只考虑标题栏高度
    dialogRef.value.style.width = `${STREAM_WINDOW_CONFIG.DEFAULT_HEIGHT}px`;
    dialogRef.value.style.height = `${STREAM_WINDOW_CONFIG.DEFAULT_WIDTH + HEADER_HEIGHT}px`;
  } else {
    // 竖屏模式 - 使用默认宽高
    phoneFrame.style.width = `${STREAM_WINDOW_CONFIG.DEFAULT_WIDTH}px`;
    phoneFrame.style.height = `${STREAM_WINDOW_CONFIG.DEFAULT_HEIGHT}px`;
    
    // 调整对话框尺寸，只考虑标题栏高度
    dialogRef.value.style.width = `${STREAM_WINDOW_CONFIG.DEFAULT_WIDTH}px`;
    dialogRef.value.style.height = `${STREAM_WINDOW_CONFIG.DEFAULT_HEIGHT + HEADER_HEIGHT}px`;
  }
};

// 流加载完成回调
const onStreamReady = (deviceId, data) => {
  console.log('收到流加载完成事件:', deviceId, data);
  streamReady.value = true;
  streamError.value = false;
};

// 流加载错误回调
const onStreamError = (errorData) => {
  streamError.value = true;
};

// 关闭窗口
const closeDialog = () => {
  // 更新父组件的v-model值
  emit('update:modelValue', false);
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

// 拖拽中
const handleDrag = (e: MouseEvent) => {
  if (!isDragging.value || !dialogRef.value) return;
  
  // 计算新位置
  let left = e.clientX - dragOffset.value.x;
  let top = e.clientY - dragOffset.value.y;
  
  // 限制不超出屏幕边界
  const windowWidth = window.innerWidth;
  const windowHeight = window.innerHeight;
  const dialogWidth = dialogRef.value.offsetWidth;
  const dialogHeight = dialogRef.value.offsetHeight;
  
  // 至少20px在视口内
  left = Math.min(Math.max(left, -dialogWidth + 20), windowWidth - 20);
  top = Math.min(Math.max(top, 0), windowHeight - 20);
  
  // 应用新位置
  dialogRef.value.style.left = `${left}px`;
  dialogRef.value.style.top = `${top}px`;
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
  
  e.preventDefault();
  isResizing.value = true;
  
  // 保存初始尺寸
  initialSize.value = {
    width: dialogRef.value.offsetWidth,
    height: dialogRef.value.offsetHeight
  };
  
  // 保存初始鼠标位置
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
  
  // 计算新尺寸，保持宽高比
  const aspectRatio = isLandscape.value ? 
    STREAM_WINDOW_CONFIG.DEFAULT_HEIGHT / STREAM_WINDOW_CONFIG.DEFAULT_WIDTH :
    STREAM_WINDOW_CONFIG.DEFAULT_WIDTH / STREAM_WINDOW_CONFIG.DEFAULT_HEIGHT;
  
  // 判断是主要按宽度还是高度调整
  if (Math.abs(deltaX) > Math.abs(deltaY)) {
    // 主要按宽度调整
    const newWidth = Math.max(300, initialSize.value.width + deltaX);
    const newHeight = newWidth / aspectRatio;
    dialogRef.value.style.width = `${newWidth}px`;
    dialogRef.value.style.height = `${newHeight}px`;
  } else {
    // 主要按高度调整
    const newHeight = Math.max(300, initialSize.value.height + deltaY);
    const newWidth = newHeight * aspectRatio;
    dialogRef.value.style.width = `${newWidth}px`;
    dialogRef.value.style.height = `${newHeight}px`;
  }
};

// 停止缩放
const stopResize = () => {
  isResizing.value = false;
  document.removeEventListener('mousemove', handleResize);
  document.removeEventListener('mouseup', stopResize);
};

// 组件挂载时
onMounted(() => {
  // 初始化对话框位置和尺寸
  if (dialogRef.value) {
    // 调整初始尺寸
    adjustDialogSizeForOrientation();
    
    // 窗口居中
    dialogRef.value.style.left = '50%';
    dialogRef.value.style.top = '50%';
    dialogRef.value.style.transform = 'translate(-50%, -50%)';
  }
});

// 组件卸载前清理事件监听器
onBeforeUnmount(() => {
  document.removeEventListener('mousemove', handleDrag);
  document.removeEventListener('mouseup', stopDrag);
  document.removeEventListener('mousemove', handleResize);
  document.removeEventListener('mouseup', stopResize);
});
</script>

<style scoped>
.stream-dialog-container {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 1000;
  display: flex;
  justify-content: center;
  align-items: center;
  pointer-events: none;
}

.stream-backdrop {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  backdrop-filter: blur(2px);
  pointer-events: auto;
}

.stream-window {
  position: absolute;
  background-color: #202020;
  border-radius: 12px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
  overflow: hidden;
  pointer-events: auto;
  display: flex;
  flex-direction: column;
  border: 1px solid #404040;
}

.stream-header {
  background-color: #303030;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 12px;
  cursor: move;
  user-select: none;
  border-bottom: 1px solid #404040;
  flex-shrink: 0;
}

.stream-title {
  color: #f0f0f0;
  font-size: 14px;
  font-weight: 500;
  white-space: nowrap;
  text-overflow: ellipsis;
  overflow: hidden;
}

.close-button {
  width: 24px;
  height: 24px;
  border-radius: 4px;
  background: transparent;
  border: none;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #f0f0f0;
  cursor: pointer;
  outline: none;
}

.close-button:hover {
  background-color: rgba(255, 255, 255, 0.1);
}

.close-button:active {
  background-color: rgba(255, 255, 255, 0.15);
}

.phone-frame {
  flex: 1;
  position: relative;
  margin: 0; /* 移除所有边距 */
  padding: 0; /* 确保没有内边距 */
  background-color: #000;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
  box-sizing: border-box; /* 确保尺寸包含边框 */
  border: none; /* 移除边框 */
}

/* 使用伪元素添加视觉边框，不影响尺寸 */
.phone-frame::after {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  border-radius: 20px;
  pointer-events: none;
  box-shadow: 0 0 10px rgba(0, 0, 0, 0.5) inset;
  border: 1px solid #505050;
  z-index: 5;
}

.phone-frame.landscape {
  transform: none;
}

.phone-notch {
  position: absolute;
  top: 0;
  left: 50%;
  transform: translateX(-50%);
  width: 120px;
  height: 12px;
  background-color: #000;
  border-bottom-left-radius: 10px;
  border-bottom-right-radius: 10px;
  z-index: 10;
}

.landscape .phone-notch {
  top: 50%;
  left: 0;
  transform: translateY(-50%);
  width: 12px;
  height: 120px;
  border-bottom-left-radius: 0;
  border-top-right-radius: 10px;
  border-bottom-right-radius: 10px;
}

.resize-handle {
  position: absolute;
  bottom: 0;
  right: 0;
  width: 20px;
  height: 20px;
  cursor: nwse-resize;
  z-index: 11;
}

.resize-handle::after {
  content: '';
  position: absolute;
  bottom: 4px;
  right: 4px;
  width: 12px;
  height: 12px;
  background: linear-gradient(135deg, transparent 50%, #6495ED 50%, #6495ED 75%, transparent 75%);
}
</style> 
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
      <div v-if="!streamReady && isReady" class="stream-loading">
        <el-icon class="loading-icon"><Loading /></el-icon>
        <span>串流加载中...</span>
      </div>
      
      <div class="phone-frame">
        <div class="phone-notch" />
        <device-stream 
          v-if="visible && deviceId && isReady" 
          :device-id="deviceId" 
          :auto-connect="true"
          @stream-ready="onStreamReady"
        />
      </div>
      <div class="resize-handle" @mousedown="startResize" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onBeforeUnmount } from 'vue';
import DeviceStream from './DeviceStream.vue';
import { Close, Loading } from '@element-plus/icons-vue';
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
  }
});

const emit = defineEmits(['update:modelValue', 'closed']);

// 组件内部是否可见（用于延迟销毁iframe）
const visible = ref(false);
// 对话框是否准备好（用于延迟加载iframe）
const isReady = ref(false);
// 流是否准备好
const streamReady = ref(false);
const dialogRef = ref<HTMLElement | null>(null);

// 标题计算属性
const title = computed(() => {
  return streamReady.value 
    ? `已连接到设备: ${props.deviceId}` 
    : `连接到设备: ${props.deviceId}`;
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
    // 添加一个小延迟，确保对话框完全打开后再加载iframe
    setTimeout(() => {
      isReady.value = true;
    }, 100);
  } else {
    isReady.value = false;
    streamReady.value = false;
    setTimeout(() => {
      visible.value = false;
      emit('closed');
    }, 200);
  }
});

// 流加载完成回调
const onStreamReady = () => {
  streamReady.value = true;
  ElMessage.success(`连接设备 ${props.deviceId} 成功`);
};

// 关闭窗口
const closeDialog = () => {
  emit('update:modelValue', false);
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
  background-color: rgba(0, 0, 0, 0.7);
  color: white;
  z-index: 10;
}

.loading-icon {
  font-size: 48px;
  margin-bottom: 16px;
  animation: spin 1.5s linear infinite;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

.stream-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  background-color: #1a1a1a;
  cursor: move;
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
<template>
  <div v-if="modelValue" class="stream-window">
    <div class="stream-header">
      <span class="stream-title">周老师 (127.0.0.1:16480)</span>
      <button class="close-button" @click="closeDialog">
        <el-icon><Close /></el-icon>
      </button>
    </div>
    
    <div class="phone-frame">
      <div class="phone-notch" />
      <device-stream 
        v-if="visible && deviceId" 
        :device-id="deviceId" 
        :auto-connect="true"
        @stream-ready="onStreamReady"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import DeviceStream from './DeviceStream.vue';
import { Close } from '@element-plus/icons-vue';

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

// 监听显示状态
watch(() => props.modelValue, (newVal) => {
  if (newVal) {
    visible.value = true;
  }
});

// 流加载完成回调
const onStreamReady = () => {
  // 这里可以添加流准备好后的处理逻辑
};

// 关闭窗口
const closeDialog = () => {
  emit('update:modelValue', false);
  setTimeout(() => {
    visible.value = false;
    emit('closed');
  }, 200);
};
</script>

<style scoped>
.stream-window {
  position: fixed;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 400px;
  height: 800px;
  background-color: #000;
  border-radius: 20px;
  overflow: hidden;
  box-shadow: 0 25px 50px rgba(0,0,0,0.5);
  display: flex;
  flex-direction: column;
  z-index: 2000;
}

.stream-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  background-color: #1a1a1a;
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
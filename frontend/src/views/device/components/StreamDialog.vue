<template>
  <el-dialog
    v-model="dialogVisible"
    :title="title"
    :width="dialogWidth"
    :close-on-click-modal="false"
    :destroy-on-close="true"
    @closed="handleClosed"
  >
    <div class="stream-dialog-content">
      <device-stream 
        v-if="visible && deviceId" 
        :device-id="deviceId" 
        :auto-connect="true"
        @stream-ready="onStreamReady"
      />
    </div>
    <template #footer>
      <span class="dialog-footer">
        <el-button type="primary" @click="closeDialog">关闭</el-button>
      </span>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import DeviceStream from './DeviceStream.vue';

const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false
  },
  deviceId: {
    type: String,
    default: ''
  },
  deviceName: {
    type: String,
    default: ''
  }
});

const emit = defineEmits(['update:modelValue', 'closed']);

// 对话框可见性
const dialogVisible = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
});

// 组件内部是否可见（用于延迟销毁iframe）
const visible = ref(false);

// 对话框标题
const title = computed(() => {
  return props.deviceName 
    ? `云手机控制: ${props.deviceName} (${props.deviceId})` 
    : `云手机控制: ${props.deviceId}`;
});

// 对话框宽度 - 响应式
const dialogWidth = computed(() => {
  // 获取窗口大小来计算合适的对话框宽度
  const windowWidth = window.innerWidth;
  
  if (windowWidth < 768) {
    return '95%';
  } else if (windowWidth < 1200) {
    return '80%';
  } else {
    return '70%';
  }
});

// 监听对话框打开状态
watch(() => dialogVisible.value, (newVal) => {
  if (newVal) {
    visible.value = true;
  }
});

// 流加载完成回调
const onStreamReady = () => {
  // 这里可以添加流准备好后的处理逻辑
};

// 关闭对话框
const closeDialog = () => {
  dialogVisible.value = false;
};

// 对话框完全关闭后回调
const handleClosed = () => {
  // 延迟一点时间再设置visible为false，确保过渡动画完成
  setTimeout(() => {
    visible.value = false;
    emit('closed');
  }, 200);
};
</script>

<style scoped>
.stream-dialog-content {
  position: relative;
  width: 100%;
  height: 600px;
  overflow: hidden;
  border-radius: 8px;
  background-color: #000;
}

@media (max-height: 768px) {
  .stream-dialog-content {
    height: 450px;
  }
}

@media (max-width: 768px) {
  .stream-dialog-content {
    height: 400px;
  }
}

:deep(.el-dialog__body) {
  padding: 0;
}

:deep(.el-dialog__footer) {
  padding: 10px 20px;
}

:deep(.el-dialog__header) {
  padding: 15px 20px;
  margin-right: 0;
  border-bottom: 1px solid #f0f0f0;
}
</style> 
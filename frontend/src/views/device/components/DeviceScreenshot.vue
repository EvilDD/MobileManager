<template>
  <div class="device-screenshot-container" :class="{ 'error': error, 'disconnected': disconnected }">
    <div v-if="error && !disconnected" class="screenshot-error">
      <el-icon class="error-icon"><CircleClose /></el-icon>
      <span>{{ errorMessage }}</span>
      <el-button size="small" type="primary" @click="captureScreenshot" class="retry-button">
        重试
      </el-button>
    </div>
    
    <div v-if="disconnected" class="screenshot-disconnected">
      <el-icon class="disconnected-icon"><WarningFilled /></el-icon>
      <span>连接已断开</span>
      <el-button size="small" type="primary" @click="checkConnection" class="retry-button">
        重试连接
      </el-button>
    </div>
    
    <img 
      v-if="imageData && !disconnected" 
      :src="imageData" 
      class="screenshot-image" 
      alt="设备截图"
      @click="handleClick"
      @error="handleImageError"
      @load="handleImageLoaded"
    />
    
    <div v-if="!imageData && !loading && !error && !disconnected" class="screenshot-placeholder">
      <el-icon><Picture /></el-icon>
      <span>无截图</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch, onBeforeUnmount } from 'vue';
import { captureDeviceScreenshot } from '@/api/device';
import { ElMessage } from 'element-plus';
import { Loading, CircleClose, Picture, WarningFilled } from '@element-plus/icons-vue';

const props = defineProps({
  deviceId: {
    type: String,
    required: true
  },
  autoCapture: {
    type: Boolean,
    default: true
  },
  quality: {
    type: Number,
    default: 90
  },
  refreshInterval: {
    type: Number,
    default: 10000 // 默认10秒刷新一次
  },
  autoRefresh: {
    type: Boolean,
    default: true
  }
});

const emit = defineEmits(['screenshot-ready', 'screenshot-error', 'click']);

const loading = ref(false);
const error = ref(false);
const errorMessage = ref('');
const imageData = ref<string | null>(null);
const disconnected = ref(false);  // 设备连接状态
// 自动刷新定时器
const refreshTimer = ref<number | null>(null);
// 连接状态检查定时器
const connectionCheckTimer = ref<number | null>(null);

// 获取设备截图
const captureScreenshot = async () => {
  if (!props.deviceId) return;
  
  try {
    loading.value = true;
    error.value = false;
    disconnected.value = false;
    
    const res = await captureDeviceScreenshot({
      deviceId: props.deviceId,
      quality: props.quality
    });
    
    // console.log("截图响应数据:", JSON.stringify(res).slice(0, 100) + "...");
    
    // 确保data字段存在
    if (!res.data) {
      console.error("截图响应缺少data字段:", res);
      error.value = true;
      errorMessage.value = "响应格式错误";
      emit('screenshot-error', errorMessage.value);
      return;
    }
    
    // 判断响应是否成功
    if (res.code === 0) {
      let success = false;
      let imgData = "";
      let errMsg = "";
      
      // 获取正确的字段
      if ('success' in res.data) {
        success = res.data.success;
      }
      
      if ('imageData' in res.data && res.data.imageData) {
        imgData = res.data.imageData;
      }
      
      if ('error' in res.data && res.data.error) {
        errMsg = res.data.error;
      }
      
      if (success && imgData) {
        // 设备成功连接
        disconnected.value = false;
        
        // 检查base64数据是否已经包含了前缀
        if (imgData.startsWith('data:image')) {
          imageData.value = imgData;
        } else {
          // 移除可能的空白字符并添加前缀
          imageData.value = `data:image/jpeg;base64,${imgData.replace(/^[\s\r\n]+|[\s\r\n]+$/g, '')}`;
        }
        // console.log("截图数据已获取，长度:", (imageData.value?.length || 0));
        emit('screenshot-ready', imageData.value);
        
        // 不自动启动连接状态检查
      } else {
        error.value = true;
        errorMessage.value = errMsg || '获取截图失败';
        emit('screenshot-error', errorMessage.value);
      }
    } else {
      error.value = true;
      errorMessage.value = res.message || '截图请求失败';
      emit('screenshot-error', errorMessage.value);
    }
  } catch (err) {
    error.value = true;
    errorMessage.value = '获取截图请求失败';
    emit('screenshot-error', errorMessage.value);
    console.error('截图请求失败:', err);
  } finally {
    loading.value = false;
  }
};

// 检查设备连接状态
const checkConnection = async () => {
  try {
    // 尝试获取一次截图，如果成功则说明设备仍然连接
    await captureScreenshot();
  } catch (err) {
    // 如果截图失败，标记设备为断开状态
    disconnected.value = true;
    console.error('设备连接检查失败:', err);
    emit('screenshot-error', '设备连接已断开');
  }
};

// 停止连接状态检查
const stopConnectionCheck = () => {
  if (connectionCheckTimer.value !== null) {
    window.clearInterval(connectionCheckTimer.value);
    connectionCheckTimer.value = null;
  }
};

// 开始连接状态检查
const startConnectionCheck = () => {
  // 清除可能存在的旧定时器
  stopConnectionCheck();
  
  // 不再自动检查，避免影响性能和出现错误
};

// 开始自动刷新
const startAutoRefresh = () => {
  if (!props.autoRefresh || props.refreshInterval <= 0) return;
  
  // 清除可能存在的旧定时器
  stopAutoRefresh();
  
  // 创建新的定时器
  refreshTimer.value = window.setInterval(() => {
    if (props.deviceId) {
      captureScreenshot();
    }
  }, props.refreshInterval);
};

// 停止自动刷新
const stopAutoRefresh = () => {
  if (refreshTimer.value !== null) {
    window.clearInterval(refreshTimer.value);
    refreshTimer.value = null;
  }
};

// 处理图片点击
const handleClick = () => {
  emit('click');
};

// 处理图片加载错误
const handleImageError = (event: Event) => {
  console.error('图片加载失败:', event);
  error.value = true;
  errorMessage.value = '截图数据无效或已过期';
  emit('screenshot-error', errorMessage.value);
  // 清除无效图片数据
  imageData.value = null;
};

// 处理图片加载完成
const handleImageLoaded = () => {
  // 图片加载成功后可以执行的操作
  error.value = false;
};

// 监听设备ID变化，重新获取截图
watch(() => props.deviceId, (newVal) => {
  if (newVal && props.autoCapture) {
    captureScreenshot();
    
    // 重新设置自动刷新
    if (props.autoRefresh) {
      startAutoRefresh();
    }
  } else {
    stopAutoRefresh();
  }
});

// 监听自动刷新设置变化
watch(() => props.autoRefresh, (newVal) => {
  if (newVal) {
    startAutoRefresh();
  } else {
    stopAutoRefresh();
  }
});

// 监听刷新间隔变化
watch(() => props.refreshInterval, () => {
  if (props.autoRefresh) {
    startAutoRefresh(); // 重新设置定时器
  }
});

// 组件挂载时，自动获取截图
onMounted(() => {
  if (props.autoCapture && props.deviceId) {
    captureScreenshot();
    
    // 如果启用了自动刷新，开始定时刷新
    if (props.autoRefresh) {
      startAutoRefresh();
    }
  }
});

// 组件卸载前，清理定时器
onBeforeUnmount(() => {
  stopAutoRefresh();
  stopConnectionCheck();
});

// 暴露方法给父组件
defineExpose({
  captureScreenshot,
  startAutoRefresh,
  stopAutoRefresh,
  checkConnection
});
</script>

<style scoped>
.device-screenshot-container {
  position: relative;
  width: 100%;
  height: 100%;
  overflow: hidden;
  background-color: #f5f7fa;
  display: flex;
  justify-content: center;
  align-items: center;
  border-radius: 4px;
}

.screenshot-loading,
.screenshot-error,
.screenshot-placeholder,
.screenshot-disconnected {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  z-index: 2;
}

.screenshot-loading {
  background-color: rgba(0, 0, 0, 0.5);
  color: #fff;
}

.screenshot-error {
  background-color: rgba(245, 108, 108, 0.1);
  color: #f56c6c;
}

.screenshot-disconnected {
  background-color: rgba(230, 162, 60, 0.1);
  color: #e6a23c;
}

.screenshot-placeholder {
  background-color: #f5f7fa;
  color: #909399;
}

.loading-icon,
.error-icon,
.disconnected-icon {
  font-size: 24px;
  margin-bottom: 8px;
}

.loading-icon {
  animation: rotating 2s linear infinite;
}

.disconnected-icon {
  color: #e6a23c;
}

@keyframes rotating {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

.screenshot-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform 0.3s;
}

.screenshot-image:hover {
  transform: scale(1.05);
  cursor: pointer;
}

.retry-button {
  margin-top: 10px;
}
</style> 
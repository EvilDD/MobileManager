<template>
  <div class="device-screenshot" ref="container">
    <img
      v-if="imageData"
      :src="imageData"
      :alt="'设备截图 - ' + deviceId"
      @load="handleImageLoad"
      @error="handleImageError"
      class="screenshot-image"
    />
    <div v-else-if="error" class="error-message">
      {{ error }}
    </div>
    <div v-else class="loading">
      加载中...
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch } from 'vue';
import { captureDeviceScreenshot } from '@/api/device';

const props = defineProps<{
  deviceId: string;
  autoRefresh?: boolean;
  refreshInterval?: number;
  quality?: number;
  scale?: number;
  format?: 'webp' | 'jpeg';
}>();

const emit = defineEmits<{
  (e: 'screenshotReady', deviceId: string, imageData: string): void;
  (e: 'screenshotError', deviceId: string, error: string): void;
}>();

const imageData = ref<string>('');
const error = ref<string>('');
const container = ref<HTMLElement | null>(null);
let refreshTimer: number | null = null;

// 计算最佳的图片参数
const calculateOptimalParams = () => {
  if (!container.value) return { quality: 80, scale: 1.0 };
  
  const width = container.value.clientWidth;
  // 根据容器宽度动态计算最适合的图片参数
  if (width <= 320) {
    return { quality: 60, scale: 0.5 };
  } else if (width <= 640) {
    return { quality: 70, scale: 0.75 };
  }
  return { quality: 80, scale: 1.0 };
};

const captureScreenshot = async () => {
  try {
    const optimalParams = calculateOptimalParams();
    const response = await captureDeviceScreenshot({
      deviceId: props.deviceId,
      quality: props.quality || optimalParams.quality,
      scale: props.scale || optimalParams.scale,
      format: props.format || 'webp'
    });

    if (response.data.success && response.data.imageData) {
      imageData.value = response.data.imageData;
      emit('screenshotReady', props.deviceId, response.data.imageData);
      error.value = '';
    } else {
      error.value = response.data.error || '截图失败';
      emit('screenshotError', props.deviceId, error.value);
    }
  } catch (err) {
    const errorMessage = err instanceof Error ? err.message : '截图失败';
    error.value = errorMessage;
    emit('screenshotError', props.deviceId, errorMessage);
  }
};

const startAutoRefresh = () => {
  if (refreshTimer) {
    clearInterval(refreshTimer);
  }
  if (props.autoRefresh && props.refreshInterval) {
    refreshTimer = setInterval(captureScreenshot, props.refreshInterval) as unknown as number;
  }
};

const handleImageLoad = () => {
  error.value = '';
};

const handleImageError = () => {
  error.value = '图片加载失败';
  emit('screenshotError', props.deviceId, '图片加载失败');
};

watch(() => props.autoRefresh, (newVal) => {
  if (newVal) {
    startAutoRefresh();
  } else if (refreshTimer) {
    clearInterval(refreshTimer);
  }
});

watch(() => props.refreshInterval, () => {
  if (props.autoRefresh) {
    startAutoRefresh();
  }
});

onMounted(() => {
  captureScreenshot();
  if (props.autoRefresh) {
    startAutoRefresh();
  }
});

onUnmounted(() => {
  if (refreshTimer) {
    clearInterval(refreshTimer);
  }
});
</script>

<style scoped>
.device-screenshot {
  width: 100%;
  height: 100%;
  position: relative;
  overflow: hidden;
  background-color: #f5f5f5;
  border-radius: 4px;
}

.screenshot-image {
  width: 100%;
  height: 100%;
  object-fit: contain;
}

.error-message {
  color: #f56c6c;
  text-align: center;
  padding: 20px;
}

.loading {
  text-align: center;
  padding: 20px;
  color: #909399;
}
</style> 
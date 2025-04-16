<template>
  <div class="sync-container">
    <div class="devices-grid">
      <!-- 主设备 -->
      <div class="device-card main-device" v-if="mainDevice">
        <div class="device-header">
          <span class="device-name">{{ mainDevice.name }}</span>
        </div>
        <div class="device-screen">
          <device-screenshot
            v-if="mainDevice.status === 'online'"
            :device-id="mainDevice.deviceId"
            :auto-capture="true"
            :quality="80"
            :auto-refresh="autoRefresh"
            :refresh-interval="refreshInterval"
            @screenshot-ready="(imageData) => handleScreenshotReady(mainDevice.deviceId, imageData)"
            @screenshot-error="(err) => handleScreenshotError(mainDevice.deviceId, err)"
          />
          <div v-else class="offline-placeholder">
            <div class="image-error">
              <el-icon><WarningFilled /></el-icon>
              <span>设备离线</span>
            </div>
          </div>
        </div>
        <div class="device-info">
          <div class="device-id">ID: {{ mainDevice.deviceId }}</div>
          <div class="device-status">
            <el-tag 
              :type="mainDevice.status === 'online' ? 'success' : 'danger'" 
              size="small"
            >
              {{ mainDevice.status === 'online' ? '在线' : '离线' }}
            </el-tag>
          </div>
        </div>
      </div>

      <!-- 其他设备 -->
      <div 
        v-for="device in otherDevices" 
        :key="device.deviceId" 
        class="device-card other-device"
      >
        <div class="device-header">
          <span class="device-name">{{ device.name }}</span>
        </div>
        <div class="device-screen">
          <device-screenshot
            v-if="device.status === 'online'"
            :device-id="device.deviceId"
            :auto-capture="true"
            :quality="80"
            :auto-refresh="autoRefresh"
            :refresh-interval="refreshInterval"
            @screenshot-ready="(imageData) => handleScreenshotReady(device.deviceId, imageData)"
            @screenshot-error="(err) => handleScreenshotError(device.deviceId, err)"
          />
          <div v-else class="offline-placeholder">
            <div class="image-error">
              <el-icon><WarningFilled /></el-icon>
              <span>设备离线</span>
            </div>
          </div>
        </div>
        <div class="device-info">
          <div class="device-id">ID: {{ device.deviceId }}</div>
          <div class="device-status">
            <el-tag 
              :type="device.status === 'online' ? 'success' : 'danger'" 
              size="small"
            >
              {{ device.status === 'online' ? '在线' : '离线' }}
            </el-tag>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import { ElMessage } from 'element-plus';
import { WarningFilled } from '@element-plus/icons-vue';
import { useCloudPhoneStore } from '@/store/modules/cloudphone';
import type { Device } from '@/api/device';
import DeviceScreenshot from './components/DeviceScreenshot.vue';
import { DEVICE_CONFIG } from './components/config';

interface SyncDevice extends Device {
  isMainDevice?: boolean;
  screenshot?: string;
}

const router = useRouter();
const route = useRoute();
const store = useCloudPhoneStore();

// 设备数据
const deviceList = ref<SyncDevice[]>([]);

// 计算属性：主设备和其他设备
const mainDevice = computed(() => deviceList.value.find(device => device.isMainDevice));
const otherDevices = computed(() => deviceList.value.filter(device => !device.isMainDevice));

// 截图相关
const autoRefresh = ref(true);
const refreshInterval = ref(5000); // 默认5秒刷新一次
const screenshotStatus = ref<Record<string, { success: boolean; error?: string }>>({});

// 处理截图事件
const handleScreenshotReady = (deviceId: string, imageData: string) => {
  screenshotStatus.value[deviceId] = { success: true };
  // 更新设备截图
  const device = deviceList.value.find(d => d.deviceId === deviceId);
  if (device) {
    device.screenshot = imageData;
  }
};

const handleScreenshotError = (deviceId: string, error: string) => {
  screenshotStatus.value[deviceId] = { success: false, error };
  console.error(`设备 ${deviceId} 截图加载失败:`, error);
};

// 初始化数据
onMounted(() => {
  // 从 store 获取选中的设备列表
  const selectedDevices = store.selectedDevices;
  
  if (!selectedDevices || selectedDevices.length === 0) {
    ElMessage.warning('请先在分组手机页面选择需要同步的设备');
    router.push('/device/cloudphone');
    return;
  }

  // 将第一个设备设置为主设备
  deviceList.value = selectedDevices.map((device, index) => ({
    ...device,
    isMainDevice: index === 0,
    screenshot: device.screenshot || 'https://via.placeholder.com/300x600'
  }));
});

// 监听路由变化,如果没有选中设备则返回分组手机页面
watch(
  () => route.path,
  () => {
    if (!store.selectedDevices || store.selectedDevices.length === 0) {
      ElMessage.warning('请先在分组手机页面选择需要同步的设备');
      router.push('/device/cloudphone');
    }
  }
);
</script>

<style scoped>
.sync-container {
  width: 100%;
  height: 100vh;
  background-color: #f5f7fa;
  overflow: hidden;
  padding: 20px;
}

.devices-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, v-bind('(DEVICE_CONFIG.SYNC.MAIN_DEVICE.WIDTH / 2) + "px"'));
  gap: 20px;
  justify-content: start;
  align-content: start;
  height: calc(100vh - 40px);
  overflow-y: auto;
  padding: 10px;
}

.device-card {
  background-color: white;
  border-radius: 8px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
  display: flex;
  flex-direction: column;
  overflow: hidden;
  transition: all 0.3s ease;
  position: relative;
  height: 100%;
}

.device-card:hover {
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
  transform: translateY(-2px);
}

.main-device {
  width: v-bind('DEVICE_CONFIG.SYNC.MAIN_DEVICE.WIDTH + "px"');
  height: v-bind('DEVICE_CONFIG.SYNC.MAIN_DEVICE.HEIGHT + "px"');
  grid-column: span 2;
  grid-row: span 2;
}

.other-device {
  width: v-bind('(DEVICE_CONFIG.SYNC.MAIN_DEVICE.WIDTH / 2) + "px"');
  height: v-bind('(DEVICE_CONFIG.SYNC.MAIN_DEVICE.HEIGHT / 2) + "px"');
}

.device-header {
  padding: 10px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-bottom: 1px solid #f0f0f0;
}

.device-name {
  font-weight: 500;
  font-size: 14px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.device-screen {
  flex: 1;
  position: relative;
  overflow: hidden;
  display: flex;
  justify-content: center;
  align-items: center;
  background-color: #000;
}

/* 设备截图样式 */
.main-device .device-screen {
  height: calc(100% - 80px); /* 减去header和info的高度 */
}

.other-device .device-screen {
  height: calc(100% - 74px); /* 减去header和info的高度 */
}

.device-screen :deep(.device-screenshot-container) {
  width: 100% !important;
  height: 100% !important;
}

.device-screen :deep(.screenshot-image) {
  width: 100%;
  height: 100%;
  object-fit: contain;
  max-height: 100%;
}

.offline-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  justify-content: center;
  align-items: center;
  background-color: #1a1a1a;
}

.image-error {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  color: #909399;
  font-size: 14px;
  gap: 10px;
}

.device-info {
  padding: 10px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  background-color: #f9f9f9;
  font-size: 12px;
}

.device-id {
  color: #606266;
}

/* 响应式调整 */
@media (max-width: 768px) {
  .devices-grid {
    grid-template-columns: 1fr;
  }
  
  .main-device {
    grid-column: span 1;
    height: 500px;
  }
  
  .other-device {
    height: 280px;
  }
}
</style> 
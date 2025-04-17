<template>
  <div class="sync-container">
    <div class="devices-grid">
      <!-- 主设备 -->
      <div class="device-card main-device" v-if="mainDevice">
        <div class="device-header">
          <span class="device-name">{{ mainDevice.name }}</span>
          <div class="actions">
            <el-switch
              v-model="streamEnabled"
              active-text="实时串流"
              @change="toggleStream"
              :disabled="streamLoading"
            />
          </div>
        </div>
        <div class="device-screen">
          <!-- 视频流播放器 -->
          <div v-if="streamEnabled && !streamError" ref="playerContainer" class="player-container" />
          
          <!-- 截图模式 -->
          <device-screenshot
            v-if="!streamEnabled && mainDevice.status === 'online'"
            :device-id="mainDevice.deviceId"
            :auto-capture="true"
            :quality="80"
            :auto-refresh="autoRefresh"
            :refresh-interval="refreshInterval"
            @screenshot-ready="(imageData) => handleScreenshotReady(mainDevice.deviceId, imageData)"
            @screenshot-error="(err) => handleScreenshotError(mainDevice.deviceId, err)"
          />
          
          <!-- 错误状态 -->
          <div v-if="streamError" class="stream-error">
            <el-icon><WarningFilled /></el-icon>
            <span>{{ streamError }}</span>
            <el-button size="small" type="primary" @click="toggleStream(true)" class="retry-button">
              重试
            </el-button>
          </div>
          
          <div v-if="!streamEnabled && mainDevice.status !== 'online'" class="offline-placeholder">
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
import { ref, computed, onMounted, watch, onBeforeUnmount } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import { ElMessage } from 'element-plus';
import { WarningFilled } from '@element-plus/icons-vue';
import { useCloudPhoneStore } from '@/store/modules/cloudphone';
import type { Device } from '@/api/device';
import { startDeviceStream, stopDeviceStream } from '@/api/device';
import DeviceScreenshot from './components/DeviceScreenshot.vue';
import { DEVICE_CONFIG } from './components/config';

// 导入视频播放器
import { WebCodecsPlayer } from './player/WebCodecsPlayer';
import Size from './Size';
import { DisplayInfo } from './DisplayInfo';

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

// 视频流相关
const streamEnabled = ref(false);
const streamLoading = ref(false);
const streamError = ref<string | null>(null);
const wsConnection = ref<WebSocket | null>(null);
const playerContainer = ref<HTMLElement | null>(null);
const player = ref<WebCodecsPlayer | null>(null);

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

// 启动/停止视频流
const toggleStream = async (enabled?: boolean) => {
  const newState = typeof enabled === 'boolean' ? enabled : streamEnabled.value;
  
  // 如果正在切换状态，不重复操作
  if (streamLoading.value) return;
  
  // 如果主设备不在线，不能开启串流
  if (newState && (!mainDevice.value || mainDevice.value.status !== 'online')) {
    ElMessage.warning('主设备不在线，无法启动视频流');
    streamEnabled.value = false;
    return;
  }
  
  try {
    streamLoading.value = true;
    streamError.value = null;
    
    if (newState) {
      // 启动串流
      await startStream();
    } else {
      // 停止串流
      await stopStream();
    }
    
    streamEnabled.value = newState;
  } catch (error) {
    console.error('切换视频流失败:', error);
    streamError.value = error instanceof Error ? error.message : '未知错误';
    streamEnabled.value = false;
  } finally {
    streamLoading.value = false;
  }
};

// 启动视频流
const startStream = async () => {
  if (!mainDevice.value) return;
  
  try {
    // 调用后端接口获取串流信息
    const response = await startDeviceStream(mainDevice.value.deviceId);
    
    if (response.code === 0 && response.data) {
      const { port } = response.data;
      
      // 创建WebSocket连接
      // 使用相对路径，让Vite代理配置来处理请求转发
      const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
      const wsUrl = `${wsProtocol}//${window.location.host}/ws/wsscrcpy?udid=${mainDevice.value.deviceId}&port=${port}`;
      console.log('连接WebSocket:', wsUrl);
      
      // 关闭可能存在的旧连接
      if (wsConnection.value) {
        wsConnection.value.close();
      }
      
      // 创建新的WebSocket连接
      const ws = new WebSocket(wsUrl);
      wsConnection.value = ws;
      
      // 初始化播放器
      initPlayer();
      
      // 设置WebSocket事件处理
      ws.binaryType = 'arraybuffer';
      
      ws.onopen = () => {
        console.log('WebSocket连接成功');
        streamError.value = null;
      };
      
      ws.onmessage = (event) => {
        if (player.value && event.data instanceof ArrayBuffer) {
          const data = new Uint8Array(event.data);
          (player.value as any).pushFrame(data);
        }
      };
      
      ws.onerror = (event) => {
        console.error('WebSocket错误:', event);
        streamError.value = 'WebSocket连接错误';
        stopStream().catch(console.error);
      };
      
      ws.onclose = () => {
        console.log('WebSocket连接关闭');
        if (streamEnabled.value) {
          streamError.value = 'WebSocket连接已关闭';
          streamEnabled.value = false;
        }
      };
      
    } else {
      throw new Error(response.message || '启动串流失败');
    }
  } catch (error) {
    console.error('启动串流失败:', error);
    throw error;
  }
};

// 停止视频流
const stopStream = async () => {
  // 关闭WebSocket连接
  if (wsConnection.value) {
    wsConnection.value.close();
    wsConnection.value = null;
  }
  
  // 停止播放器
  if (player.value) {
    player.value.stop();
    player.value = null;
    
    // 清空播放器容器
    if (playerContainer.value) {
      playerContainer.value.innerHTML = '';
    }
  }
  
  // 如果主设备存在，调用后端接口停止串流
  if (mainDevice.value) {
    try {
      await stopDeviceStream(mainDevice.value.deviceId);
    } catch (error) {
      console.error('停止串流失败:', error);
    }
  }
};

// 初始化视频播放器
const initPlayer = () => {
  if (!playerContainer.value || !mainDevice.value) return;
  
  // 清空播放器容器
  playerContainer.value.innerHTML = '';
  
  // 创建新的播放器实例
  const displayInfo = new DisplayInfo(0, new Size(540, 960));
  player.value = new WebCodecsPlayer(mainDevice.value.deviceId, displayInfo);
  
  // 设置播放器父容器，使用类型断言解决TypeScript错误
  (player.value as any).setParent(playerContainer.value);
  
  // 启动播放器，使用类型断言解决TypeScript错误
  (player.value as any).play();
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

// 组件卸载前清理资源
onBeforeUnmount(() => {
  stopStream().catch(console.error);
});
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
  grid-template-columns: repeat(auto-fill, v-bind('(DEVICE_CONFIG.SYNC.CANVAS_PORTRAIT.WIDTH / 2) + "px"'));
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
  width: v-bind('DEVICE_CONFIG.SYNC.CANVAS_PORTRAIT.WIDTH + "px"');
  height: v-bind('(DEVICE_CONFIG.SYNC.CANVAS_PORTRAIT.HEIGHT + 80) + "px"'); /* 添加header和info的高度 */
  grid-column: span 2;
  grid-row: span 2;
}

.other-device {
  width: v-bind('(DEVICE_CONFIG.SYNC.CANVAS_PORTRAIT.WIDTH / 2) + "px"');
  height: v-bind('(DEVICE_CONFIG.SYNC.CANVAS_PORTRAIT.HEIGHT / 2 + 74) + "px"'); /* 添加header和info的高度 */
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
  box-sizing: content-box;
  border: none;
  padding: 0;
  margin: 0;
}

/* 设备截图样式 */
.main-device .device-screen {
  height: v-bind('DEVICE_CONFIG.SYNC.CANVAS_PORTRAIT.HEIGHT + "px"');
  width: v-bind('DEVICE_CONFIG.SYNC.CANVAS_PORTRAIT.WIDTH + "px"');
  flex: none; /* 确保不受flex布局影响 */
}

.other-device .device-screen {
  height: v-bind('(DEVICE_CONFIG.SYNC.CANVAS_PORTRAIT.HEIGHT / 2) + "px"');
  width: v-bind('(DEVICE_CONFIG.SYNC.CANVAS_PORTRAIT.WIDTH / 2) + "px"');
  flex: none; /* 确保不受flex布局影响 */
}

.device-screen :deep(.device-screenshot-container) {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  width: 100% !important;
  height: 100% !important;
  padding: 0 !important;
  margin: 0 !important;
  border: none !important;
  display: flex;
  align-items: center;
  justify-content: center;
}

.device-screen :deep(.screenshot-image) {
  width: 100%;
  height: 100%;
  object-fit: cover; /* 使用cover而非contain确保填满容器 */
  max-height: none;
  border: none !important;
  padding: 0 !important;
  margin: 0 !important;
  display: block;
  object-position: center;
}

.offline-placeholder,
.stream-error {
  width: 100%;
  height: 100%;
  display: flex;
  justify-content: center;
  align-items: center;
  background-color: #1a1a1a;
}

.stream-error {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: #f56c6c;
  gap: 10px;
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

.retry-button {
  margin-top: 10px;
}

.player-container {
  width: 100%;
  height: 100%;
  position: relative;
}

/* 响应式调整 */
@media (max-width: 768px) {
  .devices-grid {
    grid-template-columns: 1fr;
  }
  
  .main-device {
    grid-column: span 1;
    height: v-bind('(DEVICE_CONFIG.SYNC.CANVAS_PORTRAIT.HEIGHT + 80) + "px"');
  }
  
  .other-device {
    height: v-bind('(DEVICE_CONFIG.SYNC.CANVAS_PORTRAIT.HEIGHT / 2 + 74) + "px"');
  }
}
</style> 
<template>
  <div class="sync-container">
    <!-- 设备同步状态悬浮提示 -->
    <div class="sync-status" v-if="isSyncing">
      <el-progress 
        type="circle" 
        :percentage="syncProgress" 
        :status="syncProgress === 100 ? 'success' : ''"
        :stroke-width="6"
      />
      <div class="sync-message">{{ syncStatusMessage }}</div>
    </div>

    <!-- 添加同步控制按钮 -->
    

    <div class="devices-grid">
      <!-- 主设备 -->
      <div class="device-card main-device" v-if="mainDevice">
        <div class="device-header">
          <span class="device-name">{{ mainDevice.name }}</span>
        </div>
        <div class="device-screen">
          <el-image 
            :src="mainDevice.screenshot || '/placeholder-device.png'" 
            fit="contain"
            :lazy="false"
          >
            <template #error>
              <div class="image-error">
                <el-icon><WarningFilled /></el-icon>
                <span>无法加载设备画面</span>
              </div>
            </template>
          </el-image>
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
        :class="{ 'sync-success': device.syncStatus === 'success', 'sync-error': device.syncStatus === 'error' }"
      >
        <div class="device-header">
          <span class="device-name">{{ device.name }}</span>
        </div>
        <div class="device-screen">
          <el-image 
            :src="device.screenshot || '/placeholder-device.png'" 
            fit="contain"
            :lazy="false"
          >
            <template #error>
              <div class="image-error">
                <el-icon><WarningFilled /></el-icon>
                <span>无法加载设备画面</span>
              </div>
            </template>
          </el-image>
          
          <!-- 同步覆盖层 -->
          <div class="sync-overlay" v-if="device.syncStatus === 'syncing'">
            <el-icon class="sync-icon"><Loading /></el-icon>
            <span>正在同步...</span>
          </div>
          <div class="sync-overlay success" v-else-if="device.syncStatus === 'success'">
            <el-icon class="sync-icon"><SuccessFilled /></el-icon>
            <span>同步成功</span>
          </div>
          <div class="sync-overlay error" v-else-if="device.syncStatus === 'error'">
            <el-icon class="sync-icon"><CircleCloseFilled /></el-icon>
            <span>同步失败</span>
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
import { 
  WarningFilled, 
  Loading, 
  SuccessFilled, 
  CircleCloseFilled 
} from '@element-plus/icons-vue';
import { DEVICE_CONFIG } from './components/config';
import { useCloudPhoneStore } from '@/store/modules/cloudphone';
import type { Device } from '@/api/device';

interface SyncDevice extends Device {
  isMainDevice?: boolean;
  syncStatus?: 'waiting' | 'syncing' | 'success' | 'error';
  lastSyncTime?: string;
  screenshot?: string;
}

const router = useRouter();
const route = useRoute();
const store = useCloudPhoneStore();

// 同步状态
const syncLoading = ref(false);
const isSyncing = ref(false);
const syncProgress = ref(0);
const syncStatusMessage = ref('准备同步设备...');

// 设备数据
const deviceList = ref<SyncDevice[]>([]);

// 计算属性：主设备和其他设备
const mainDevice = computed(() => deviceList.value.find(device => device.isMainDevice));
const otherDevices = computed(() => deviceList.value.filter(device => !device.isMainDevice));

// 停止同步操作
const stopSync = () => {
  if (!isSyncing.value) return;
  
  // 停止同步操作
  syncLoading.value = false;
  isSyncing.value = false;
  syncProgress.value = 0;
  syncStatusMessage.value = '同步已停止';
  
  // 更新所有从设备的同步状态
  otherDevices.value.forEach(device => {
    if (device.syncStatus === 'syncing') {
      device.syncStatus = 'waiting';
    }
  });
  
  ElMessage.info('同步操作已停止');
};

// 开始同步操作
const startSync = async () => {
  if (deviceList.value.length <= 1) {
    ElMessage.warning('至少需要两台设备才能进行同步操作');
    return;
  }

  if (!mainDevice.value) {
    ElMessage.warning('缺少主设备，无法进行同步');
    return;
  }
  
  if (mainDevice.value.status !== 'online') {
    ElMessage.warning('主设备不在线，无法进行同步');
    return;
  }

  // 同步状态初始化
  syncLoading.value = true;
  isSyncing.value = true;
  syncProgress.value = 0;
  syncStatusMessage.value = '正在准备同步...';
  
  // 更新所有从设备的同步状态为"syncing"
  otherDevices.value.forEach(device => {
    device.syncStatus = 'syncing';
  });

  try {
    // 模拟同步进度
    const updateProgress = async () => {
      for (let i = 0; i <= 100; i += 5) {
        // 如果同步被停止，退出循环
        if (!isSyncing.value) {
          break;
        }
        
        syncProgress.value = i;
        
        if (i === 30) {
          syncStatusMessage.value = '正在同步应用数据...';
        } else if (i === 60) {
          syncStatusMessage.value = '正在同步系统设置...';
        } else if (i === 90) {
          syncStatusMessage.value = '正在完成同步操作...';
        }
        
        // 随机设置部分设备同步完成
        if (i >= 40 && i % 20 === 0) {
          const randomDeviceIndex = Math.floor(Math.random() * otherDevices.value.length);
          const randomStatus = Math.random() > 0.8 ? 'error' : 'success';
          if (otherDevices.value[randomDeviceIndex].syncStatus === 'syncing') {
            otherDevices.value[randomDeviceIndex].syncStatus = randomStatus;
          }
        }
        
        await new Promise(resolve => setTimeout(resolve, 300));
      }
      
      // 所有同步完成
      if (isSyncing.value) {
        otherDevices.value.forEach(device => {
          if (device.syncStatus === 'syncing') {
            device.syncStatus = 'success';
          }
          // 更新同步时间
          device.lastSyncTime = new Date().toLocaleString();
        });
        
        syncStatusMessage.value = '同步完成！';
        syncLoading.value = false;
        
        // 5秒后隐藏同步状态
        setTimeout(() => {
          isSyncing.value = false;
        }, 5000);
      }
    };
    
    await updateProgress();
    
  } catch (error) {
    console.error('同步过程中发生错误:', error);
    ElMessage.error('同步过程中发生错误');
    syncLoading.value = false;
    isSyncing.value = false;
  }
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
    syncStatus: index === 0 ? undefined : 'waiting',
    screenshot: device.screenshot || 'https://via.placeholder.com/300x600'
  }));
  
  // 如果设备数量大于1，建议用户开始同步
  if (deviceList.value.length > 1) {
    ElMessage.success(`已加载 ${deviceList.value.length} 台设备，请点击开关开始同步`);
  }
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

.sync-status {
  position: fixed;
  top: 20px;
  right: 20px;
  background: rgba(255, 255, 255, 0.9);
  border-radius: 8px;
  padding: 15px;
  display: flex;
  flex-direction: column;
  align-items: center;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  z-index: 100;
  transition: all 0.3s ease;
}

.sync-message {
  margin-top: 10px;
  font-size: 14px;
  color: #409eff;
}

/* 添加同步控制按钮样式 */
.sync-controls {
  display: none; /* 隐藏同步控制按钮 */
}

.devices-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, v-bind('(DEVICE_CONFIG.SYNC.MAIN_DEVICE.WIDTH / 2) + "px"'));
  gap: 20px;
  justify-content: start;
  align-content: start;
  height: 100%;
  overflow-y: auto;
  padding: 20px;
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

.device-screen .el-image {
  height: 100%;
  max-width: 100%;
  object-fit: contain;
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

.sync-overlay {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background-color: rgba(0, 0, 0, 0.7);
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  color: white;
  gap: 10px;
}

.sync-overlay.success {
  background-color: rgba(103, 194, 58, 0.7);
}

.sync-overlay.error {
  background-color: rgba(245, 108, 108, 0.7);
}

.sync-icon {
  font-size: 32px;
}

.sync-success {
  border: 2px solid #67c23a;
}

.sync-error {
  border: 2px solid #f56c6c;
}
</style> 
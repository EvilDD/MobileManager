<template>
  <div class="device-detail-container">
    <div class="device-header">
      <div class="device-info">
        <h2>{{ deviceName }}</h2>
        <div class="device-status" :class="{ 'online': deviceStatus === 'online' }">
          {{ deviceStatus === 'online' ? '在线' : '离线' }}
        </div>
      </div>
      <div class="actions">
        <el-button type="primary" size="small" @click="goBack">返回</el-button>
      </div>
    </div>
    
    <div class="device-content">
      <iframe
        v-if="deviceStatus === 'online'"
        :src="streamUrl"
        class="device-stream"
        frameborder="0"
        allowfullscreen
        sandbox="allow-scripts allow-same-origin allow-forms allow-downloads"
      />
      <div v-else class="offline-message">
        <el-empty description="设备当前离线，无法显示控制界面" />
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { ElMessage } from 'element-plus';

const route = useRoute();
const router = useRouter();

const deviceId = ref('');
const deviceName = ref('加载中...');
const deviceStatus = ref('offline');

// 服务器URL配置 - 从环境变量读取
const serverUrl = import.meta.env.VITE_WSCRCPY_SERVER || 'http://localhost:8000';
const wsServerUrl = import.meta.env.VITE_WSCRCPY_WS_SERVER || 'ws://localhost:8000';

// 处理来自iframe的消息
const handleIframeMessage = (event) => {
  const data = event.data;
  if (data && data.type === 'screenshot' && data.data) {
    console.log('收到截图数据，准备保存');
    try {
      // 创建下载链接
      const a = document.createElement('a');
      a.href = data.data.imageUrl;
      a.download = data.data.filename || `设备截图_${new Date().toLocaleString()}.png`;
      document.body.appendChild(a);
      a.click();
      setTimeout(() => {
        document.body.removeChild(a);
      }, 100);
      ElMessage.success('截图已保存');
    } catch (e) {
      console.error('保存截图失败:', e);
      ElMessage.error('截图保存失败: ' + e.message);
    }
  }
};

// 构建完整的串流URL
const streamUrl = computed(() => {
  if (!deviceId.value) return 'about:blank';
  
  const encodedDeviceId = encodeURIComponent(deviceId.value);
  const wsUrl = encodeURIComponent(`${wsServerUrl}/?action=proxy-adb&remote=tcp:8886&udid=${deviceId.value}`);
  return `${serverUrl}/#!action=stream&udid=${encodedDeviceId}&player=webcodecs&ws=${wsUrl}`;
});

// 获取设备详情
const getDeviceDetail = async () => {
  try {
    // 这里应该调用API获取设备详情
    // 暂时使用路由参数
    deviceId.value = route.params.id || '';
    deviceName.value = route.query.name || deviceId.value;
    deviceStatus.value = route.query.status || 'offline';
    
    if (!deviceId.value) {
      ElMessage.error('设备ID不能为空');
      goBack();
    }
  } catch (error) {
    console.error('获取设备详情失败:', error);
    ElMessage.error('获取设备详情失败');
  }
};

// 返回设备列表
const goBack = () => {
  router.push('/device/cloudphone');
};

onMounted(() => {
  getDeviceDetail();
  // 添加消息事件监听器
  window.addEventListener('message', handleIframeMessage);
});

onBeforeUnmount(() => {
  // 移除消息事件监听器
  window.removeEventListener('message', handleIframeMessage);
});
</script>

<style scoped>
.device-detail-container {
  height: 100vh;
  display: flex;
  flex-direction: column;
  background-color: #f5f5f5;
}

.device-header {
  background-color: #fff;
  padding: 16px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.device-info {
  display: flex;
  align-items: center;
}

.device-info h2 {
  margin: 0;
  padding-right: 12px;
}

.device-status {
  padding: 4px 8px;
  border-radius: 4px;
  background-color: #f56c6c;
  color: white;
  font-size: 14px;
}

.device-status.online {
  background-color: #67c23a;
}

.device-content {
  flex: 1;
  padding: 16px;
  overflow: hidden;
  display: flex;
  justify-content: center;
  align-items: center;
}

.device-stream {
  width: 100%;
  height: 100%;
  background-color: #000;
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.offline-message {
  width: 100%;
  height: 100%;
  display: flex;
  justify-content: center;
  align-items: center;
  background-color: #fff;
  border-radius: 8px;
}
</style> 
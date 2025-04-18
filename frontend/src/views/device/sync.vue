<template>
  <div class="sync-container">
    <div class="devices-grid">
      <!-- 主设备 -->
      <div class="device-card main-device" v-if="mainDevice">
        <div class="device-header">
          <span class="device-name">{{ mainDevice.name }}</span>
          <div class="actions">
            <!-- 操作同步开关 -->
            <el-switch
              v-model="syncEnabled"
              active-text="操作同步"
              @change="handleSyncOperation"
              :disabled="streamLoading"
            />
          </div>
        </div>
        <div class="device-screen">
          <!-- 视频流播放器容器 - 永久存在但可能隐藏 -->
          <div ref="playerContainer" class="player-container" :class="{'active': streamEnabled && !streamError, 'inactive': !streamEnabled || streamError}">
            <!-- 添加加载提示，在开始播放前显示 -->
            <div v-if="streamEnabled && !streamError && !videoFrameReceived" class="video-loading-overlay">
              <el-icon class="loading-icon"><Loading /></el-icon>
              <span>等待视频数据...</span>
            </div>
          </div>
          
          <!-- 加载中状态 -->
          <div v-if="streamLoading" class="stream-loading">
            <el-icon class="loading-icon"><Loading /></el-icon>
            <span>串流加载中...</span>
          </div>
          
          <!-- 错误状态 -->
          <div v-if="streamError" class="stream-error">
            <el-icon><WarningFilled /></el-icon>
            <span>{{ streamError }}</span>
            <el-button size="small" type="primary" @click="toggleStream(true)" class="retry-button">
              重试
            </el-button>
          </div>
          
          <div v-if="!streamEnabled && !streamLoading && mainDevice.status !== 'online'" class="offline-placeholder">
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
import { ref, computed, onMounted, watch, onBeforeUnmount, nextTick, onActivated, onDeactivated } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import { ElMessage } from 'element-plus';
import { WarningFilled, Loading } from '@element-plus/icons-vue';
import { useCloudPhoneStore } from '@/store/modules/cloudphone';
import type { Device } from '@/api/device';
import { startDeviceStream, stopDeviceStream } from '@/api/device';
import DeviceScreenshot from './components/DeviceScreenshot.vue';
import { DEVICE_CONFIG } from './components/config';

// 导入视频播放器
import { WebCodecsPlayer } from './player/WebCodecsPlayer';

// 扩展WebSocket类型
declare global {
  interface WebSocket {
    frameCount?: number;
    isClosed?: boolean; // 标记是否已被关闭，避免后续消息处理
  }
}

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

// 组件状态标记
const isComponentMounted = ref(true);
const isComponentActive = ref(true);

// 操作同步开关
const syncEnabled = ref(false);

// 视频流接收状态
const videoFrameReceived = ref(false);

// 操作同步处理
const handleSyncOperation = (enabled: boolean) => {
  syncEnabled.value = enabled;
  // 预留的操作同步方法，暂时只打印信息
  console.log('操作同步状态:', enabled);
  ElMessage.info(`操作同步${enabled ? '开启' : '关闭'}，功能待实现`);
};

// 处理截图事件 (仅用于其他设备)
const handleScreenshotReady = (deviceId: string, imageData: string) => {
  // 检查组件是否仍然挂载
  if (!isComponentMounted.value) return;
  
  screenshotStatus.value[deviceId] = { success: true };
  // 更新设备截图
  const device = deviceList.value.find(d => d.deviceId === deviceId);
  if (device) {
    device.screenshot = imageData;
  }
};

const handleScreenshotError = (deviceId: string, error: string) => {
  // 检查组件是否仍然挂载
  if (!isComponentMounted.value) return;
  
  screenshotStatus.value[deviceId] = { success: false, error };
  // console.error(`设备 ${deviceId} 截图加载失败:`, error);
};

// 启动/停止视频流
const toggleStream = async (enabled?: boolean) => {
  // 如果组件已经卸载，不执行任何操作
  if (!isComponentMounted.value) return;
  
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
    
    // 再次检查组件挂载状态
    if (isComponentMounted.value) {
      streamEnabled.value = newState;
    }
  } catch (error) {
    console.error('切换视频流失败:', error);
    if (isComponentMounted.value) {
      streamError.value = error instanceof Error ? error.message : '未知错误';
      streamEnabled.value = false;
    }
  } finally {
    if (isComponentMounted.value) {
      streamLoading.value = false;
    }
  }
};

// 安全地处理WebSocket消息
const safeHandleWSMessage = (event: MessageEvent, ws: WebSocket) => {
  // 如果组件已卸载或WebSocket已标记为关闭，不处理消息
  if (!isComponentMounted.value || !isComponentActive.value || ws.isClosed) {
    return;
  }
  
  // 以下是原有的消息处理逻辑，添加了组件状态检查
  if (!player.value) {
    console.error('收到WebSocket消息，但播放器未初始化');
    return;
  }
  
  if (event.data instanceof ArrayBuffer) {
    // 如果组件已卸载，不处理后续消息
    if (!isComponentMounted.value || !isComponentActive.value) return;
    
    // 收到视频数据，标记视频开始接收
    if (!videoFrameReceived.value) {
      videoFrameReceived.value = true;
    }
    
    // 只记录第一帧和每100帧数据，避免日志过多
    const data = new Uint8Array(event.data);
    
    // 使用闭包记录帧数
    if (!ws.frameCount) {
      ws.frameCount = 0;
    }
    ws.frameCount++;
    
    if (ws.frameCount === 1 || ws.frameCount % 100 === 0) {
      console.log(`收到第${ws.frameCount}帧数据，大小: ${data.byteLength} 字节`);
    }
    
    try {
      // 检查播放器状态和组件挂载状态
      if (!player.value || !isComponentMounted.value || !isComponentActive.value) return;
      
      // 检查播放器状态
      if (typeof (player.value as any).pushFrame !== 'function') {
        console.error('播放器pushFrame方法不存在');
        return;
      }
      
      (player.value as any).pushFrame(data);
      
      // 第一帧后检查播放器容器状态
      if (ws.frameCount === 1) {
        setTimeout(() => {
          if (playerContainer.value && isComponentMounted.value && isComponentActive.value) {
            const containerRect = playerContainer.value.getBoundingClientRect();
            console.log('第一帧后播放器容器状态:', {
              width: containerRect.width,
              height: containerRect.height,
              visibility: window.getComputedStyle(playerContainer.value).visibility,
              display: window.getComputedStyle(playerContainer.value).display,
              children: playerContainer.value.children.length
            });
          }
        }, 100);
      }
    } catch (error) {
      console.error('处理视频帧出错:', error);
    }
  } else {
    console.log('收到非二进制WebSocket消息:', event.data);
  }
};

// 停止视频流
const stopStream = async () => {
  // 停止标记
  videoFrameReceived.value = false;
  
  // 标记WebSocket为已关闭，防止后续消息处理
  if (wsConnection.value) {
    wsConnection.value.isClosed = true;
  }
  
  // 关闭WebSocket连接
  if (wsConnection.value) {
    try {
      wsConnection.value.onmessage = null; // 移除消息处理程序
      wsConnection.value.onerror = null;   // 移除错误处理程序
      wsConnection.value.onclose = null;   // 移除关闭处理程序
      wsConnection.value.close();
    } catch (e) {
      console.warn('关闭WebSocket连接出错:', e);
    }
    wsConnection.value = null;
  }
  
  // 停止播放器
  if (player.value) {
    try {
      player.value.stop();
    } catch (e) {
      console.warn('停止播放器出错:', e);
    }
    player.value = null;
    
    // 清空播放器容器
    if (playerContainer.value) {
      try {
        playerContainer.value.innerHTML = '';
      } catch (e) {
        console.warn('清空播放器容器出错:', e);
      }
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

// 组件激活时
onActivated(() => {
  console.log('组件激活');
  isComponentActive.value = true;
  
  // 如果之前是开启状态，重新开始串流
  if (streamEnabled.value && mainDevice.value && mainDevice.value.status === 'online') {
    setTimeout(() => {
      toggleStream(true);
    }, 500);
  }
});

// 组件失活时(缓存)
onDeactivated(() => {
  console.log('组件失活');
  isComponentActive.value = false;
  
  // 停止串流但保持状态
  const wasEnabled = streamEnabled.value;
  stopStream().catch(e => {
    console.warn('组件失活时停止串流出错:', e);
  });
  
  // 保持之前的状态
  streamEnabled.value = wasEnabled;
});

// 组件卸载前清理资源
onBeforeUnmount(() => {
  console.log('组件卸载，清理资源');
  isComponentMounted.value = false;
  isComponentActive.value = false;
  
  // 移除可能的事件监听器
  window.removeEventListener('resize', handleResize);
  
  // 确保WebSocket不会再处理消息
  if (wsConnection.value) {
    wsConnection.value.isClosed = true;
    wsConnection.value.onmessage = null;
    wsConnection.value.onerror = null;
    wsConnection.value.onclose = null;
  }
  
  // 安全地停止流
  stopStream().catch(e => {
    console.warn('组件卸载时停止串流出错:', e);
  });
  
  // 清空引用
  playerContainer.value = null;
  player.value = null;
  wsConnection.value = null;
});

// 启动视频流
const startStream = async () => {
  // 检查组件是否还挂载
  if (!isComponentMounted.value || !isComponentActive.value) return;
  
  if (!mainDevice.value) return;
  
  // 重置视频接收状态
  videoFrameReceived.value = false;
  
  try {
    console.log('开始启动串流...');
    
    // 确保playerContainer已准备好
    await nextTick();
    if (!playerContainer.value) {
      console.error('播放器容器DOM元素未找到');
      
      // 再等待一次DOM更新，以防首次渲染延迟
      await new Promise(resolve => setTimeout(resolve, 500));
      
      // 再次检查组件状态
      if (!isComponentMounted.value || !isComponentActive.value) return;
      
      if (!playerContainer.value) {
        console.warn('播放器容器DOM元素不存在，创建备用容器');
        // 查找设备屏幕元素
        const deviceScreen = document.querySelector('.main-device .device-screen');
        if (deviceScreen) {
          // 创建并插入新的播放器容器
          const newContainer = document.createElement('div');
          newContainer.className = 'player-container active';
          newContainer.id = 'backup-player-container';
          newContainer.style.width = '100%';
          newContainer.style.height = '100%';
          newContainer.style.position = 'absolute';
          newContainer.style.left = '0';
          newContainer.style.top = '0';
          newContainer.style.backgroundColor = 'black';
          deviceScreen.appendChild(newContainer);
          
          // 手动设置playerContainer引用
          playerContainer.value = newContainer;
          console.log('已创建备用播放器容器');
        } else {
          throw new Error('无法找到设备屏幕元素，无法创建播放器容器');
        }
      } else {
        console.log('延迟后找到了播放器容器元素');
      }
    }
    
    // 再次检查组件状态
    if (!isComponentMounted.value || !isComponentActive.value) return;
    
    // 调用后端接口获取串流信息前先显示加载状态
    streamLoading.value = true;
    
    // 调用后端接口获取串流信息
    const response = await startDeviceStream(mainDevice.value.deviceId);
    console.log('获取串流信息接口响应:', response);
    
    // 再次检查组件状态
    if (!isComponentMounted.value || !isComponentActive.value) return;
    
    if (response.code === 0 && response.data) {
      const { port } = response.data;
      console.log('获取到串流端口:', port);
      
      // 再次确认DOM元素存在和组件状态
      if (!playerContainer.value || !isComponentMounted.value || !isComponentActive.value) {
        console.error('播放器容器DOM元素不存在或组件已卸载');
        throw new Error('播放器容器DOM元素不存在或组件已卸载');
      }
      
      // 确保播放器容器可见
      playerContainer.value.style.display = 'block';
      playerContainer.value.classList.remove('inactive');
      playerContainer.value.classList.add('active');
      
      // 初始化播放器
      console.log('DOM已更新，开始初始化播放器');
      initPlayer();
      console.log('播放器初始化完成');
      
      // 检查播放器是否正确初始化
      if (!player.value) {
        console.error('播放器初始化失败');
        throw new Error('播放器初始化失败');
      }
      
      // 创建WebSocket连接
      // 使用相对路径，让Vite代理配置来处理请求转发
      const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
      const wsUrl = `${wsProtocol}//${window.location.host}/ws/wsscrcpy?udid=${mainDevice.value.deviceId}&port=${port}`;
      console.log('准备连接WebSocket:', wsUrl);
      
      // 关闭可能存在的旧连接
      if (wsConnection.value) {
        console.log('关闭旧的WebSocket连接');
        wsConnection.value.isClosed = true;
        wsConnection.value.onmessage = null;
        wsConnection.value.onerror = null;
        wsConnection.value.onclose = null;
        wsConnection.value.close();
      }
      
      // 不再显示加载状态，因为已经初始化了播放器
      streamLoading.value = false;
      
      // 创建新的WebSocket连接
      console.log('创建新的WebSocket连接');
      const ws = new WebSocket(wsUrl);
      ws.isClosed = false; // 初始化为未关闭状态
      wsConnection.value = ws;
      
      // 设置连接超时
      const connectionTimeout = setTimeout(() => {
        if (!isComponentMounted.value || !isComponentActive.value) return;
        
        if (ws.readyState !== WebSocket.OPEN) {
          console.error('WebSocket连接超时');
          streamError.value = 'WebSocket连接超时';
          ws.isClosed = true;
          ws.close();
        }
      }, 15000); // 15秒超时
      
      // 设置WebSocket事件处理
      ws.binaryType = 'arraybuffer';
      console.log('WebSocket binaryType设置为:', ws.binaryType);
      
      ws.onopen = () => {
        // 如果组件已卸载，不处理事件
        if (!isComponentMounted.value || !isComponentActive.value) return;
        
        console.log('WebSocket连接成功');
        clearTimeout(connectionTimeout);
        streamError.value = null;
      };
      
      // 数据接收超时检查器
      let dataReceived = false;
      const checkDataTimer = setTimeout(() => {
        // 如果组件已卸载，不处理事件
        if (!isComponentMounted.value || !isComponentActive.value) return;
        
        if (!dataReceived && !ws.isClosed) {
          console.warn('WebSocket连接成功但10秒内没有收到数据');
          streamError.value = '未收到视频数据，请检查设备串流状态';
          
          // 尝试自动重连
          stopStream().catch(console.error).finally(() => {
            if (!isComponentMounted.value || !isComponentActive.value) return;
            
            setTimeout(() => {
              if (!isComponentMounted.value || !isComponentActive.value) return;
              
              if (mainDevice.value && mainDevice.value.status === 'online') {
                console.log('尝试自动重新连接...');
                toggleStream(true);
              }
            }, 3000);
          });
        }
      }, 10000);
      
      // 使用安全的消息处理函数
      ws.onmessage = (event) => {
        // 如果WebSocket已标记为关闭，不处理消息
        if (ws.isClosed) return;
        
        // 标记已收到数据
        if (!dataReceived) {
          dataReceived = true;
          clearTimeout(checkDataTimer);
        }
        
        // 使用安全的消息处理函数
        safeHandleWSMessage(event, ws);
      };
      
      ws.onerror = (event) => {
        // 如果组件已卸载，不处理事件
        if (!isComponentMounted.value || !isComponentActive.value) return;
        
        console.error('WebSocket错误:', event);
        clearTimeout(connectionTimeout);
        clearTimeout(checkDataTimer);
        streamError.value = 'WebSocket连接错误';
        stopStream().catch(console.error);
      };
      
      ws.onclose = () => {
        // 如果组件已卸载，不处理事件
        if (!isComponentMounted.value || !isComponentActive.value) return;
        
        console.log('WebSocket连接关闭');
        clearTimeout(connectionTimeout);
        clearTimeout(checkDataTimer);
        ws.isClosed = true; // 标记为已关闭
        
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
  } finally {
    // 确保加载状态被清除
    if (isComponentMounted.value) {
      streamLoading.value = false;
    }
  }
};

// 可选：添加resize事件处理
const handleResize = () => {
  // 如果组件已卸载，不处理事件
  if (!isComponentMounted.value) return;
  
  // 处理窗口大小变化
  if (playerContainer.value) {
    // 更新播放器容器样式
    playerContainer.value.style.width = '100%';
    playerContainer.value.style.height = '100%';
  }
};

// 初始化视频播放器
const initPlayer = () => {
  // 如果组件已卸载，不执行初始化
  if (!isComponentMounted.value || !isComponentActive.value) return;
  
  if (!playerContainer.value || !mainDevice.value) {
    console.error('初始化播放器失败: playerContainer或mainDevice不存在');
    return;
  }
  
  // 检查WebCodecs API是否可用
  if (!('VideoDecoder' in window)) {
    console.error('当前浏览器不支持VideoDecoder API');
    streamError.value = '当前浏览器不支持视频解码功能，请使用Chrome/Edge最新版本';
    return;
  }
  
  // 检查playerContainer的尺寸
  const containerRect = playerContainer.value.getBoundingClientRect();
  console.log('播放器容器尺寸:', {
    width: containerRect.width,
    height: containerRect.height,
    visibility: window.getComputedStyle(playerContainer.value).visibility,
    display: window.getComputedStyle(playerContainer.value).display
  });
  
  // 如果容器宽度或高度为0，可能没有正确渲染
  if (containerRect.width === 0 || containerRect.height === 0) {
    console.warn('播放器容器尺寸异常，尝试设置显式尺寸');
    playerContainer.value.style.width = '100%';
    playerContainer.value.style.height = '100%';
    playerContainer.value.style.position = 'absolute';
    playerContainer.value.style.left = '0';
    playerContainer.value.style.top = '0';
  }
  
  // 清空播放器容器
  playerContainer.value.innerHTML = '';
  
  // 创建新的播放器实例
  console.log('创建WebCodecsPlayer实例');
  try {
    player.value = new WebCodecsPlayer();
    
    // 设置播放器父容器，使用类型断言解决TypeScript错误
    console.log('设置播放器父容器');
    (player.value as any).setParent(playerContainer.value);
    
    // 启动播放器，使用类型断言解决TypeScript错误
    console.log('启动播放器');
    (player.value as any).play();
    
    // 检查播放器是否设置了canvas
    setTimeout(() => {
      if (!isComponentMounted.value || !isComponentActive.value) return;
      
      if (playerContainer.value) {
        const canvasElements = playerContainer.value.querySelectorAll('canvas');
        console.log(`播放器容器中有 ${canvasElements.length} 个canvas元素`);
        
        if (canvasElements.length === 0) {
          console.warn('播放器没有创建canvas元素');
          // 尝试手动创建canvas元素
          const canvas = document.createElement('canvas');
          canvas.style.width = '100%';
          canvas.style.height = '100%';
          canvas.width = 600;  // 设置一个合理的初始宽度
          canvas.height = 1200; // 设置一个合理的初始高度
          playerContainer.value.appendChild(canvas);
          console.log('已手动创建canvas元素');
        } else {
          // 确保canvas样式正确
          canvasElements.forEach(canvas => {
            canvas.style.width = '100%';
            canvas.style.height = '100%';
            canvas.style.objectFit = 'contain';
            canvas.style.display = 'block';
          });
        }
      }
    }, 200);
  } catch (error) {
    console.error('创建或初始化播放器失败:', error);
    streamError.value = '视频播放器初始化失败';
  }
};

// 初始化数据
onMounted(async () => {
  console.log('组件已挂载');
  isComponentMounted.value = true;
  isComponentActive.value = true;
  
  // 添加全局resize事件
  window.addEventListener('resize', handleResize);
  
  // 检查VideoDecoder API是否可用
  if (!('VideoDecoder' in window)) {
    console.warn('浏览器不支持VideoDecoder API，视频流可能无法正常播放');
  }
  
  // 从 store 获取选中的设备列表
  const selectedDevices = store.selectedDevices;
  
  if (!selectedDevices || selectedDevices.length === 0) {
    ElMessage.warning('请先在分组手机页面选择需要同步的设备');
    router.push('/device/cloudphone');
    return;
  }

  console.log('获取到设备列表:', selectedDevices);

  // 将第一个设备设置为主设备
  deviceList.value = selectedDevices.map((device, index) => ({
    ...device,
    isMainDevice: index === 0,
    screenshot: device.screenshot || 'https://via.placeholder.com/300x600'
  }));
  
  console.log('主设备:', mainDevice.value);
  
  // 确保DOM已更新
  await nextTick();
  
  // 确认播放器容器是否渲染 
  if (!playerContainer.value) {
    console.log('初始化时播放器容器尚未就绪，等待渲染');
    // 等待更长时间以确保DOM完全渲染
    await new Promise(resolve => setTimeout(resolve, 500));
  }
  
  console.log('播放器容器状态:', playerContainer.value ? '已渲染' : '未渲染');
  
  // 自动开始串流流程
  if (mainDevice.value && mainDevice.value.status === 'online') {
    console.log('主设备在线，准备开始串流');
    // 稍微延迟启动，确保DOM已完全渲染
    setTimeout(() => {
      if (isComponentMounted.value && isComponentActive.value) {
        console.log('开始调用toggleStream');
        toggleStream(true);
      }
    }, 1000); // 延长到1000ms，给DOM更多时间渲染
  } else {
    console.log('主设备不在线或不存在，无法开始串流');
  }
});

// 监听路由变化,如果没有选中设备则返回分组手机页面
watch(
  () => route.path,
  () => {
    if (!isComponentMounted.value) return;
    
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

/* 设备屏幕样式 */
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
.stream-error,
.stream-loading {
  width: 100%;
  height: 100%;
  display: flex;
  justify-content: center;
  align-items: center;
  background-color: #1a1a1a;
}

.stream-loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: #409eff;
  gap: 10px;
}

.loading-icon {
  animation: rotate 2s linear infinite;
  font-size: 24px;
}

@keyframes rotate {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
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

/* 播放器容器显示状态 */
.player-container.inactive {
  display: none;
}

.player-container.active {
  display: block;
}

/* 视频加载覆盖层 */
.video-loading-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  background-color: rgba(0, 0, 0, 0.6);
  color: #fff;
  z-index: 10;
  gap: 10px;
}
</style> 
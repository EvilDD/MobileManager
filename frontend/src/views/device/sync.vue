<template>
  <div class="sync-container">
    <div class="devices-grid">
      <!-- 主设备 -->
      <div class="device-card main-device" v-if="mainDevice">
        <div class="device-header">
          <span class="device-name">{{ mainDevice.name }}</span>
          <div class="actions">
            <!-- 操作同步开关 - 只有主设备才显示 -->
            <el-switch
              v-if="isMainDevice"
              v-model="syncEnabled"
              active-text="操作同步"
              @change="handleSyncOperation"
              :disabled="streamLoading"
            />
            <!-- 在线状态标签 -->
            <el-tag 
              :type="mainDevice.status === 'online' ? 'success' : 'danger'" 
              size="small"
              class="status-tag"
            >
              {{ mainDevice.status === 'online' ? '在线' : '离线' }}
            </el-tag>
          </div>
        </div>
        <div class="device-screen">
          <!-- 添加触摸层，覆盖在视频流上方 -->
          <div
            v-if="streamEnabled && !streamError" 
            class="touch-layer"
            @mousedown="handleMouseDown"
            @mousemove="handleMouseMove"
            @mouseup="handleMouseUp"
            @mouseleave="handleMouseUp"
            @touchstart="handleTouchStart"
            @touchmove="handleTouchMove"
            @touchend="handleTouchEnd"
          />

          <!-- 视频流播放器容器 - 仅在需要时才存在 -->
          <div 
            v-if="streamEnabled && !streamError" 
            ref="playerContainer" 
            class="player-container"
          >
            <!-- 添加加载提示，在开始播放前显示 -->
            <div v-if="!videoFrameReceived" class="video-loading-overlay">
              <el-icon class="loading-icon"><Loading /></el-icon>
              <span>等待视频数据...</span>
            </div>
          </div>
          
          <!-- 错误状态 -->
          <div v-if="streamError" class="stream-error">
            <el-icon><WarningFilled /></el-icon>
            <span>{{ streamError }}</span>
            <el-button size="small" type="primary" @click="toggleStream(true)" class="retry-button">
              重试
            </el-button>
          </div>
          
          <!-- 加载中状态 -->
          <div v-else-if="streamLoading && !streamEnabled" class="stream-loading">
            <el-icon class="loading-icon"><Loading /></el-icon>
            <span>串流加载中...</span>
          </div>
          
          <!-- 离线或未启动状态 -->
          <div v-else-if="!streamEnabled && mainDevice.status !== 'online'" class="offline-placeholder">
            <div class="image-error">
              <el-icon><WarningFilled /></el-icon>
              <span>设备离线</span>
            </div>
          </div>
        </div>
        <div class="device-footer">
          <div class="device-id">ID: {{ mainDevice.deviceId }}</div>
          <div class="device-actions">
            <el-dropdown trigger="click">
              <el-button type="primary" plain size="small">
                操作
                <el-icon class="el-icon--right"><arrow-down /></el-icon>
              </el-button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item @click="goToHome">主页</el-dropdown-item>
                  <el-dropdown-item @click="goBack">返回</el-dropdown-item>
                  <el-dropdown-item @click="openOverview">任务管理</el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
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
          <el-tag 
            :type="device.status === 'online' ? 'success' : 'danger'" 
            size="small"
            class="status-tag"
          >
            {{ device.status === 'online' ? '在线' : '离线' }}
          </el-tag>
        </div>
        <div class="device-screen">
          <!-- 添加触摸层到子设备，类似主设备 -->
          <div
            v-if="device.status === 'online' && getOtherDeviceVideoReceived(device.deviceId) && !getOtherDeviceStreamError(device.deviceId)"
            class="touch-layer"
            @mousedown="(event) => handleOtherDeviceMouseDown(event, device)"
            @mousemove="(event) => handleOtherDeviceMouseMove(event, device)"
            @mouseup="(event) => handleOtherDeviceMouseUp(event, device)"
            @mouseleave="(event) => handleOtherDeviceMouseUp(event, device)"
            @touchstart="(event) => handleOtherDeviceTouchStart(event, device)"
            @touchmove="(event) => handleOtherDeviceTouchMove(event, device)"
            @touchend="(event) => handleOtherDeviceTouchEnd(event, device)"
          />

          <!-- 视频流播放器容器 - 替换截图组件 -->
          <div 
            v-if="device.status === 'online'"
            class="other-device-player-container"
            :id="`player-container-${device.deviceId}`"
          >
            <!-- 加载提示，只在未收到视频帧且无错误时显示 -->
            <div 
              v-if="!getOtherDeviceVideoReceived(device.deviceId) && !getOtherDeviceStreamError(device.deviceId)" 
              class="video-loading-overlay"
            >
              <el-icon class="loading-icon"><Loading /></el-icon>
              <span>准备视频数据...</span>
            </div>

            <!-- 错误状态 -->
            <div 
              v-if="getOtherDeviceStreamError(device.deviceId)" 
              class="stream-error"
            >
              <el-icon><WarningFilled /></el-icon>
              <span>{{ getOtherDeviceStreamError(device.deviceId) }}</span>
              <el-button size="small" type="primary" @click="retryOtherDeviceStream(device.deviceId)" class="retry-button">
                重试
              </el-button>
            </div>
          </div>
          <div v-else class="offline-placeholder">
            <div class="image-error">
              <el-icon><WarningFilled /></el-icon>
              <span>设备离线</span>
            </div>
          </div>
        </div>
        <div class="device-footer">
          <div class="device-id">ID: {{ device.deviceId }}</div>
          <div class="device-actions">
            <el-dropdown trigger="click">
              <el-button type="primary" plain size="small">
                操作
                <el-icon class="el-icon--right"><arrow-down /></el-icon>
              </el-button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item @click="handleOtherDeviceAction(device, 'home')">主页</el-dropdown-item>
                  <el-dropdown-item @click="handleOtherDeviceAction(device, 'back')">返回</el-dropdown-item>
                  <el-dropdown-item @click="handleOtherDeviceAction(device, 'overview')">任务管理</el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
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
import { WarningFilled, Loading, ArrowDown } from '@element-plus/icons-vue';
import { useCloudPhoneStore } from '@/store/modules/cloudphone';
import type { Device } from '@/api/device';
import { startDeviceStream, stopDeviceStream } from '@/api/device';
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

// 视频流相关
const streamEnabled = ref(false);
const streamLoading = ref(false);
const streamError = ref<string | null>(null);
const wsConnection = ref<WebSocket | null>(null);
const playerContainer = ref<HTMLElement | null>(null);
const player = ref<WebCodecsPlayer | null>(null);

// 从设备视频流连接
const otherDeviceConnections = ref<Record<string, {
  wsConnection: WebSocket | null;
  player: WebCodecsPlayer | null;
  streamEnabled: boolean;
  streamError: string | null;
  videoFrameReceived?: boolean;
}>>({});

// 组件状态标记
const isComponentMounted = ref(true);
const isComponentActive = ref(true);

// 操作同步开关
const syncEnabled = ref(true);

// 判断当前是否为主设备
const isMainDevice = computed(() => {
  if (!mainDevice.value) return false;
  return mainDevice.value.isMainDevice === true;
});

// 视频流接收状态
const videoFrameReceived = ref(false);

// 触摸事件相关
// 保存触摸点位置
const touchPoint = ref({ x: 0, y: 0 });
const touchStartPoint = ref({ x: 0, y: 0 });
const isTouching = ref(false);

// 触摸事件动作常量，与后端保持一致
const TOUCH_ACTION = {
  DOWN: 0,
  UP: 1,
  MOVE: 2
};

// 按键操作常量，与后端保持一致
const KEY_ACTION = {
  DOWN: 0,
  UP: 1
};

// 按键编码常量，与后端保持一致
const KEY_CODE = {
  HOME: 3,        // 主页键
  BACK: 4,        // 返回键
  APP_SWITCH: 187 // 最近任务键(Overview)
};

// 源坐标和目标坐标尺寸
// 源坐标使用CANVAS_PORTRAIT的配置尺寸
const sourceScreen = ref({
  width: DEVICE_CONFIG.SYNC.CANVAS_PORTRAIT.WIDTH,
  height: DEVICE_CONFIG.SYNC.CANVAS_PORTRAIT.HEIGHT
});
// 目标坐标将从SPS或视频流参数中获取
const targetScreen = ref({ width: 0, height: 0 });

// 坐标转换函数 - 将前端CSS坐标转换为实际设备坐标
const convertCoordinates = (x: number, y: number) => {
  // 如果targetScreen尺寸无效，则使用相同坐标
  if (targetScreen.value.width <= 0 || targetScreen.value.height <= 0) {
    console.warn('目标屏幕尺寸无效，使用原始坐标');
    return { x, y };
  }
  
  // 计算缩放比例 - 根据CSS显示尺寸和Canvas渲染尺寸的比例进行转换
  // 例如：如果CSS显示尺寸是480x870，而Canvas渲染尺寸是512x928
  // 那么x坐标需要乘以 512/480，y坐标需要乘以 928/870
  const scaleX = targetScreen.value.width / sourceScreen.value.width;
  const scaleY = targetScreen.value.height / sourceScreen.value.height;
  // console.log('scaleX', scaleX, 'targetScreen.value.width', targetScreen.value.width, 'sourceScreen.value.width', sourceScreen.value.width);
  // console.log('scaleY', scaleY, 'targetScreen.value.height', targetScreen.value.height, 'sourceScreen.value.height', sourceScreen.value.height);
  // 应用缩放比例
  const targetX = Math.round(x * scaleX);
  const targetY = Math.round(y * scaleY);
  
  // console.log(`坐标转换: (${x},${y}) => (${targetX},${targetY}), 比例: ${scaleX.toFixed(2)}x${scaleY.toFixed(2)}`);
  
  return { x: targetX, y: targetY };
};

// 添加getVideoSize方法实现
const getVideoSize = () => {
  if (player.value) {
    if (typeof (player.value as any).videoWidth !== 'undefined' && 
        typeof (player.value as any).videoHeight !== 'undefined') {
      const width = (player.value as any).videoWidth;
      const height = (player.value as any).videoHeight;
      if (width > 0 && height > 0) {
        return { width, height };
      }
    }
    
    // 查找canvas元素，从其宽高获取
    if (playerContainer.value) {
      const canvas = playerContainer.value.querySelector('canvas');
      if (canvas && canvas.width && canvas.height) {
        return { width: canvas.width, height: canvas.height };
      }
    }
  }
  return null;
};

// 修改sendTouchEvent，添加内部尺寸检查
const sendTouchEvent = (action: number, x: number, y: number) => {
  // 确保WebSocket连接可用
  if (!wsConnection.value || wsConnection.value.readyState !== WebSocket.OPEN) {
    console.warn('WebSocket连接不可用，无法发送触摸事件');
    return;
  }
  
  // 检查targetScreen尺寸，如果未设置或不正确，尝试从视频尺寸获取
  if (targetScreen.value.width <= 0 || targetScreen.value.height <= 0) {
    const size = getVideoSize();
    if (size && size.width > 0 && size.height > 0) {
      console.log('发送触摸事件前更新目标屏幕尺寸:', size);
      updateScreenResolution(size.width, size.height);
    } else {
      console.warn('目标屏幕尺寸无效，使用原始坐标发送触摸事件');
    }
  }
  
  // 转换坐标
  const { x: targetX, y: targetY } = convertCoordinates(x, y);
  
  // 创建触摸事件消息对象 - 使用正确的消息格式
  const touchEvent = {
    type: "touch",  
    data: {
      action: action,
      x: targetX,
      y: targetY
    }
  };

  // console.log('发送触摸事件:', touchEvent);

  try {
    // 发送事件到后端
    wsConnection.value.send(JSON.stringify(touchEvent));
    
    // 只有在主设备上且操作同步已开启时，才发送给从设备
    if (syncEnabled.value && isMainDevice.value) {
      sendTouchEventToOtherDevices(action, x, y);
    }
  } catch (error) {
    console.error('发送触摸事件失败:', error);
  }
};

// 向从设备发送触摸事件
const sendTouchEventToOtherDevices = (action: number, x: number, y: number) => {
  // 遍历所有从设备连接
  Object.keys(otherDeviceConnections.value).forEach(deviceId => {
    const deviceConnection = otherDeviceConnections.value[deviceId];
    
    // 检查WebSocket连接是否可用
    if (!deviceConnection.wsConnection || 
        deviceConnection.wsConnection.readyState !== WebSocket.OPEN) {
      console.warn(`从设备 ${deviceId} WebSocket连接不可用，无法发送触摸事件`);
      return;
    }
    
    try {
      // 获取从设备的实际屏幕分辨率
      let targetDeviceWidth = sourceScreen.value.width;  // 默认使用与主设备相同的尺寸
      let targetDeviceHeight = sourceScreen.value.height;
      
      // 如果从设备的player中有记录视频尺寸，使用实际尺寸
      if (deviceConnection.player) {
        const player = deviceConnection.player as any;
        if (player.videoWidth && player.videoHeight) {
          targetDeviceWidth = player.videoWidth;
          targetDeviceHeight = player.videoHeight;
        }
      }
      
      const scaleX = targetDeviceWidth / sourceScreen.value.width;
      const scaleY = targetDeviceHeight / sourceScreen.value.height;

      const targetX = Math.round(x * scaleX);
      const targetY = Math.round(y * scaleY);

      // const { x: targetX, y: targetY } = convertCoordinates(x, y);

      // 创建触摸事件消息对象
      const touchEvent = {
        type: "touch",  
        data: {
          action: action,
          x: targetX,
          y: targetY
        }
      };
      
      // 发送事件到从设备
      deviceConnection.wsConnection.send(JSON.stringify(touchEvent));
      // console.log(`已同步触摸事件到从设备 ${deviceId}:`, {
      //   action,
      //   原始坐标: { x, y },
      //   目标坐标: { x: targetX, y: targetY },
      //   目标设备尺寸: { width: targetDeviceWidth, height: targetDeviceHeight }
      // });
    } catch (error) {
      console.error(`发送触摸事件到从设备 ${deviceId} 失败:`, error);
    }
  });
};

// 处理鼠标按下事件
const handleMouseDown = (event: MouseEvent) => {
  isTouching.value = true;
  touchStartPoint.value = {
    x: event.offsetX,
    y: event.offsetY
  };
  touchPoint.value = {
    x: event.offsetX,
    y: event.offsetY
  };
  
  console.log('鼠标按下事件', {
    x: event.offsetX,
    y: event.offsetY,
    button: event.button,
    timestamp: new Date().toISOString()
  });

  // 发送触摸按下事件到后端
  sendTouchEvent(TOUCH_ACTION.DOWN, event.offsetX, event.offsetY);
};

// 处理鼠标移动事件
const handleMouseMove = (event: MouseEvent) => {
  if (!isTouching.value) return;
  
  touchPoint.value = {
    x: event.offsetX,
    y: event.offsetY
  };
  
  // console.log('鼠标移动事件', {
  //   x: event.offsetX,
  //   y: event.offsetY,
  //   deltaX: event.offsetX - touchStartPoint.value.x,
  //   deltaY: event.offsetY - touchStartPoint.value.y,
  //   timestamp: new Date().toISOString()
  // });

  // 发送触摸移动事件到后端
  sendTouchEvent(TOUCH_ACTION.MOVE, event.offsetX, event.offsetY);
};

// 处理鼠标抬起事件
const handleMouseUp = (event: MouseEvent) => {
  if (!isTouching.value) return;
  
  isTouching.value = false;
  
  console.log('鼠标抬起事件', {
    x: event.offsetX,
    y: event.offsetY,
    deltaX: event.offsetX - touchStartPoint.value.x,
    deltaY: event.offsetY - touchStartPoint.value.y,
    timestamp: new Date().toISOString()
  });

  // 发送触摸抬起事件到后端
  sendTouchEvent(TOUCH_ACTION.UP, event.offsetX, event.offsetY);
};

// 处理触摸开始事件 - 移动设备支持
const handleTouchStart = (event: TouchEvent) => {
  event.preventDefault(); // 阻止默认行为
  
  isTouching.value = true;
  const touch = event.touches[0];
  const target = event.target as HTMLElement;
  const rect = target.getBoundingClientRect();
  
  const x = touch.clientX - rect.left;
  const y = touch.clientY - rect.top;
  
  touchStartPoint.value = { x, y };
  touchPoint.value = { x, y };
  
  console.log('触摸开始事件', {
    x,
    y,
    timestamp: new Date().toISOString()
  });

  // 发送触摸按下事件到后端
  sendTouchEvent(TOUCH_ACTION.DOWN, x, y);
};

// 处理触摸移动事件 - 移动设备支持
const handleTouchMove = (event: TouchEvent) => {
  event.preventDefault(); // 阻止默认行为
  
  if (!isTouching.value) return;
  
  const touch = event.touches[0];
  const target = event.target as HTMLElement;
  const rect = target.getBoundingClientRect();
  
  const x = touch.clientX - rect.left;
  const y = touch.clientY - rect.top;
  
  touchPoint.value = { x, y };
  
  console.log('触摸移动事件', {
    x,
    y,
    deltaX: x - touchStartPoint.value.x,
    deltaY: y - touchStartPoint.value.y,
    timestamp: new Date().toISOString()
  });

  // 发送触摸移动事件到后端
  sendTouchEvent(TOUCH_ACTION.MOVE, x, y);
};

// 处理触摸结束事件 - 移动设备支持
const handleTouchEnd = (event: TouchEvent) => {
  event.preventDefault(); // 阻止默认行为
  
  if (!isTouching.value) return;
  
  isTouching.value = false;
  
  console.log('触摸结束事件', {
    endX: touchPoint.value.x,
    endY: touchPoint.value.y,
    deltaX: touchPoint.value.x - touchStartPoint.value.x,
    deltaY: touchPoint.value.y - touchStartPoint.value.y,
    timestamp: new Date().toISOString()
  });

  // 发送触摸抬起事件到后端
  sendTouchEvent(TOUCH_ACTION.UP, touchPoint.value.x, touchPoint.value.y);
};

// 操作同步处理
const handleSyncOperation = (enabled: boolean) => {
  syncEnabled.value = enabled;
  // 记录操作同步状态
  console.log('操作同步状态:', enabled);
  
  // 只有主设备可以开启操作同步
  if (!isMainDevice.value && enabled) {
    ElMessage.warning('只有主设备才能开启操作同步');
    syncEnabled.value = false;
    return;
  }
  
  if (enabled && isMainDevice.value) {
    ElMessage.success('操作同步已开启，主设备的触摸、Home、Back和Overview操作将同步到从设备');
    
    // 如果有从设备在线，显示更详细的提示
    const onlineSlaveDevices = otherDevices.value.filter(d => d.status === 'online').length;
    if (onlineSlaveDevices > 0) {
      console.log(`当前有 ${onlineSlaveDevices} 台从设备将接收同步操作`);
    } else {
      console.log('当前没有在线的从设备可接收同步操作');
    }
  } else {
    ElMessage.info('操作同步已关闭，操作将只发送到当前设备');
  }
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
    // 设置加载状态
    streamLoading.value = true;
    streamError.value = null;
    
    if (newState) {
      // 先设置启用状态，让播放器容器渲染出来
      streamEnabled.value = true;
      
      // 等待DOM更新，确保播放器容器已渲染
      await nextTick();
      
      // 启动串流
      try {
        await startStream();
      } catch (error) {
        console.error('启动串流失败:', error);
        streamError.value = error instanceof Error ? error.message : '未知错误';
      }
    } else {
      // 停止串流
      await stopStream();
      
      // 禁用流，这会通过v-if移除播放器容器
      streamEnabled.value = false;
    }
  } catch (error) {
    console.error('切换视频流失败:', error);
    if (isComponentMounted.value) {
      streamError.value = error instanceof Error ? error.message : '未知错误';
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
  
  // 处理JSON消息，可能包含设备初始信息
  if (typeof event.data === 'string') {
    try {
      const jsonData = JSON.parse(event.data);
      console.log('收到WebSocket JSON消息:', jsonData);
      
      // 处理不同类型的消息
      if (jsonData.type === 'initial_info' && jsonData.data) {
        // 处理设备初始信息
        handleInitialInfo(jsonData.data);
      } else if (jsonData.type === 'connected' && jsonData.data) {
        // 连接成功消息，可能包含设备信息
        if (jsonData.data.screenWidth && jsonData.data.screenHeight) {
          updateScreenResolution(jsonData.data.screenWidth, jsonData.data.screenHeight);
        }
      } else if (jsonData.type === 'device_info' && jsonData.data) {
        // 设备信息消息
        if (jsonData.data.videoWidth && jsonData.data.videoHeight) {
          updateScreenResolution(jsonData.data.videoWidth, jsonData.data.videoHeight);
        } else if (jsonData.data.screenWidth && jsonData.data.screenHeight) {
          updateScreenResolution(jsonData.data.screenWidth, jsonData.data.screenHeight);
        }
      }
    } catch (e) {
      // 不是JSON格式，当作普通文本处理
      console.log('收到WebSocket文本消息:', event.data);
    }
    return;
  }
  
  // 以下是处理视频帧数据的原有逻辑
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
      // console.log(`收到第${ws.frameCount}帧数据，大小: ${data.byteLength} 字节`);
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
      
      // 第一帧后检查播放器容器状态和更新分辨率
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
            
            // 获取视频实际尺寸并更新目标屏幕分辨率
            if (mainDevice.value) {
              // 不要在这里更新targetScreen，等待SPS解析完成后自动更新
            }
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

// 初始化视频播放器
const initPlayer = () => {
  // 如果组件已卸载，不执行初始化
  if (!isComponentMounted.value || !isComponentActive.value) return;
  
  if (!playerContainer.value) {
    console.error('初始化播放器失败: playerContainer不存在');
    streamError.value = '播放器容器未找到，请刷新页面';
    return;
  }
  
  // 检查WebCodecs API是否可用
  if (!('VideoDecoder' in window)) {
    console.error('当前浏览器不支持VideoDecoder API');
    streamError.value = '当前浏览器不支持视频解码功能，请使用Chrome/Edge最新版本';
    return;
  }
  
  try {
    // 确保之前的播放器已停止
    if (player.value) {
      try {
        player.value.stop();
        player.value = null;
      } catch (e) {
        console.warn('停止旧播放器出错:', e);
      }
    }
    
    // 设置播放器容器样式 - 这些样式不会影响Vue渲染，只是设置尺寸和位置
    playerContainer.value.style.width = '100%';
    playerContainer.value.style.height = '100%';
    playerContainer.value.style.backgroundColor = '#000';
    
    // 创建新的播放器实例
    console.log('创建WebCodecsPlayer实例');
    player.value = new WebCodecsPlayer();
    
    // 监听播放器的视频信息事件，从SPS获取真实视频尺寸
    if (typeof player.value.on === 'function') {
      player.value.on('videoInfo', (videoInfo: any) => {
        if (videoInfo && videoInfo.width && videoInfo.height) {
          console.log('收到播放器视频信息事件:', videoInfo);
          updateScreenResolution(videoInfo.width, videoInfo.height);
        }
      });
    } else {
      // 如果没有事件机制，使用轮询方式
      const checkVideoSize = () => {
        if (!player.value || !isComponentMounted.value) return;
        
        try {
          if (typeof (player.value as any).getVideoSize === 'function') {
            const size = (player.value as any).getVideoSize();
            if (size && size.width && size.height) {
              console.log('轮询获取到视频尺寸:', size);
              updateScreenResolution(size.width, size.height);
              return; // 获取成功后停止轮询
            }
          }
          
          // 继续轮询
          setTimeout(checkVideoSize, 500);
        } catch (e) {
          console.error('获取视频尺寸出错:', e);
        }
      };
      
      // 开始轮询检查视频尺寸
      setTimeout(checkVideoSize, 1000);
    }
    
    // 设置播放器父容器
    console.log('设置播放器父容器');
    player.value.setParent(playerContainer.value);
    
    // 启动播放器
    console.log('启动播放器');
    player.value.play();
    
    return true;
  } catch (error) {
    console.error('创建或初始化播放器失败:', error);
    streamError.value = '视频播放器初始化失败: ' + (error instanceof Error ? error.message : '未知错误');
    // 确保播放器引用被清空
    player.value = null;
    return false;
  }
};

// 停止视频流
const stopStream = async () => {
  console.log('开始停止视频流');
  
  // 停止标记
  videoFrameReceived.value = false;
  
  // 标记WebSocket为已关闭，防止后续消息处理
  if (wsConnection.value) {
    wsConnection.value.isClosed = true;
  }
  
  // 关闭WebSocket连接
  if (wsConnection.value) {
    try {
      console.log('关闭WebSocket连接');
      wsConnection.value.onmessage = null; // 移除消息处理程序
      wsConnection.value.onerror = null;   // 移除错误处理程序
      wsConnection.value.onclose = null;   // 移除关闭处理程序
      wsConnection.value.close();
    } catch (e) {
      console.warn('关闭WebSocket连接出错:', e);
    }
    wsConnection.value = null;
  }
  
  // 停止播放器 - 只调用stop()方法，不直接操作DOM
  if (player.value) {
    try {
      console.log('停止WebCodecsPlayer');
      player.value.stop();
    } catch (e) {
      console.warn('停止播放器出错:', e);
    }
    player.value = null;
  }
  
  // 如果主设备存在，调用后端接口停止串流
  if (mainDevice.value) {
    try {
      console.log('调用后端API停止设备串流');
      await stopDeviceStream(mainDevice.value.deviceId);
    } catch (error) {
      console.error('停止串流失败:', error);
    }
  }
  
  console.log('视频流停止完成');
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
  
  // 安全地停止主设备流
  stopStream().catch(e => {
    console.warn('组件卸载时停止串流出错:', e);
  });
  
  // 停止所有从设备的视频流
  Object.keys(otherDeviceConnections.value).forEach(deviceId => {
    stopOtherDeviceStream(deviceId).catch(e => {
      console.warn(`组件卸载时停止设备 ${deviceId} 串流出错:`, e);
    });
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
  
  if (!mainDevice.value) {
    streamError.value = '无法获取设备信息';
    return;
  }
  
  // 重置视频接收状态
  videoFrameReceived.value = false;
  
  try {
    console.log('开始启动串流...');
    
    // 确保播放器容器存在
    if (!playerContainer.value) {
      console.error('播放器容器不存在');
      streamError.value = '播放器容器不存在，请刷新页面';
      return;
    }
    
    // 调用后端接口获取串流信息
    console.log('调用设备串流API');
    streamLoading.value = true;
    
    const response = await startDeviceStream(mainDevice.value.deviceId);
    console.log('获取串流信息接口响应:', response);
    
    // 再次检查组件状态
    if (!isComponentMounted.value || !isComponentActive.value) return;
    
    if (response.code === 0 && response.data) {
      const { port } = response.data;
      console.log('获取到串流端口:', port);
      
      // 初始化播放器
      console.log('开始初始化播放器');
      const playerInitialized = initPlayer();
      
      if (!playerInitialized) {
        console.error('播放器初始化失败');
        return;
      }
      
      console.log('播放器初始化成功');
      
      // 创建WebSocket连接
      const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
      const wsUrl = `${wsProtocol}//${window.location.host}/ws/wsscrcpy?udid=${mainDevice.value.deviceId}&port=${port}`;
      console.log('准备连接WebSocket:', wsUrl);
      
      // 关闭可能存在的旧连接
      if (wsConnection.value) {
        try {
          console.log('关闭旧的WebSocket连接');
          wsConnection.value.isClosed = true;
          wsConnection.value.onmessage = null;
          wsConnection.value.onerror = null;
          wsConnection.value.onclose = null;
          wsConnection.value.close();
        } catch (error) {
          console.warn('关闭旧WebSocket连接出错:', error);
        }
      }
      
      // 不再显示加载状态
      streamLoading.value = false;
      
      // 创建新的WebSocket连接
      try {
        console.log('创建新的WebSocket连接');
        const ws = new WebSocket(wsUrl);
        ws.isClosed = false;
        wsConnection.value = ws;
        
        // 设置连接超时
        const connectionTimeout = setTimeout(() => {
          if (!isComponentMounted.value || !isComponentActive.value || ws.isClosed) return;
          
          if (ws.readyState !== WebSocket.OPEN) {
            console.error('WebSocket连接超时');
            streamError.value = 'WebSocket连接超时';
            ws.isClosed = true;
            ws.close();
          }
        }, 15000);
        
        // 设置WebSocket事件处理
        ws.binaryType = 'arraybuffer';
        
        ws.onopen = () => {
          if (!isComponentMounted.value || !isComponentActive.value || ws.isClosed) return;
          
          console.log('WebSocket连接成功');
          clearTimeout(connectionTimeout);
          streamError.value = null;
        };
        
        // 数据接收超时检查器
        let dataReceived = false;
        const checkDataTimer = setTimeout(() => {
          if (!isComponentMounted.value || !isComponentActive.value || ws.isClosed) return;
          
          if (!dataReceived) {
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
        
        ws.onmessage = (event) => {
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
          if (!isComponentMounted.value || !isComponentActive.value || ws.isClosed) return;
          
          console.error('WebSocket错误:', event);
          clearTimeout(connectionTimeout);
          clearTimeout(checkDataTimer);
          streamError.value = 'WebSocket连接错误';
          stopStream().catch(console.error);
        };
        
        ws.onclose = () => {
          if (!isComponentMounted.value || !isComponentActive.value) return;
          
          console.log('WebSocket连接关闭');
          clearTimeout(connectionTimeout);
          clearTimeout(checkDataTimer);
          ws.isClosed = true;
          
          if (streamEnabled.value) {
            streamError.value = 'WebSocket连接已关闭';
            streamEnabled.value = false;
          }
        };
      } catch (error) {
        console.error('创建WebSocket连接失败:', error);
        streamError.value = '创建WebSocket连接失败: ' + (error instanceof Error ? error.message : '未知错误');
      }
    } else {
      streamError.value = response.message || '启动串流失败';
    }
  } catch (error) {
    console.error('启动串流失败:', error);
    if (isComponentMounted.value) {
      streamError.value = error instanceof Error ? error.message : '未知错误';
    }
  } finally {
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

// 初始化数据
onMounted(() => {
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
  
  // 初始化从设备视频流连接对象
  otherDevices.value.forEach(device => {
    otherDeviceConnections.value[device.deviceId] = {
      wsConnection: null,
      player: null,
      streamEnabled: false,
      streamError: null,
      videoFrameReceived: false
    };
  });
  
  // 延时自动启动串流
  setTimeout(() => {
    if (!isComponentMounted.value || !isComponentActive.value) return;
    
    // 启动主设备视频流
    if (mainDevice.value && mainDevice.value.status === 'online') {
      console.log('开始自动启动主设备视频流');
      toggleStream(true);
    } else {
      console.log('主设备不在线，不自动启动视频流');
    }
    
    // 启动从设备视频流
    otherDevices.value.forEach(device => {
      if (device.status === 'online') {
        console.log(`开始自动启动从设备 ${device.deviceId} 视频流`);
        startOtherDeviceStream(device.deviceId);
      }
    });
  }, 500);

  // 在组件挂载后获取触摸层的尺寸并确认源尺寸正确
  setTimeout(() => {
    const touchLayer = document.querySelector('.touch-layer') as HTMLElement;
    if (touchLayer) {
      const rect = touchLayer.getBoundingClientRect();
      console.log('触摸层CSS尺寸:', {width: rect.width, height: rect.height});
      console.log('源屏幕配置尺寸:', sourceScreen.value);
      
      // 如果CSS尺寸与配置尺寸不一致，打印警告
      if (Math.abs(rect.width - sourceScreen.value.width) > 5 || 
          Math.abs(rect.height - sourceScreen.value.height) > 5) {
        console.warn('触摸层实际尺寸与配置尺寸不一致，可能影响坐标转换精度');
      }
    }
  }, 1000);
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

// 启动从设备视频流
const startOtherDeviceStream = async (deviceId: string) => {
  // 检查组件是否还挂载
  if (!isComponentMounted.value || !isComponentActive.value) return;
  
  // 获取设备信息
  const device = otherDevices.value.find(d => d.deviceId === deviceId);
  if (!device) {
    console.error(`找不到设备 ${deviceId} 的信息`);
    return;
  }
  
  // 获取或初始化设备连接状态
  if (!otherDeviceConnections.value[deviceId]) {
    otherDeviceConnections.value[deviceId] = {
      wsConnection: null,
      player: null,
      streamEnabled: false,
      streamError: null,
      videoFrameReceived: false
    };
  }
  
  const deviceConnection = otherDeviceConnections.value[deviceId];
  
  // 如果已经启用，不重复启动
  if (deviceConnection.streamEnabled) return;
  
  try {
    console.log(`开始启动设备 ${deviceId} 串流...`);
    
    // 获取播放器容器
    const playerContainer = document.getElementById(`player-container-${deviceId}`);
    if (!playerContainer) {
      console.error(`设备 ${deviceId} 的播放器容器不存在`);
      deviceConnection.streamError = '播放器容器不存在';
      return;
    }
    
    // 调用后端接口获取串流信息
    console.log(`调用设备 ${deviceId} 串流API`);
    
    const response = await startDeviceStream(deviceId);
    console.log(`获取设备 ${deviceId} 串流信息接口响应:`, response);
    
    // 再次检查组件状态
    if (!isComponentMounted.value || !isComponentActive.value) return;
    
    if (response.code === 0 && response.data) {
      const { port } = response.data;
      console.log(`获取到设备 ${deviceId} 串流端口:`, port);
      
      // 初始化播放器
      console.log(`开始初始化设备 ${deviceId} 播放器`);
      
      try {
        // 确保之前的播放器已停止
        if (deviceConnection.player) {
          try {
            deviceConnection.player.stop();
            deviceConnection.player = null;
          } catch (e) {
            console.warn(`停止设备 ${deviceId} 旧播放器出错:`, e);
          }
        }
        
        // 设置播放器容器样式
        playerContainer.style.width = '100%';
        playerContainer.style.height = '100%';
        playerContainer.style.backgroundColor = '#000';
        
        // 创建新的播放器实例
        console.log(`创建设备 ${deviceId} WebCodecsPlayer实例`);
        const player = new WebCodecsPlayer();
        deviceConnection.player = player;
        
        // 设置播放器父容器
        console.log(`设置设备 ${deviceId} 播放器父容器`);
        player.setParent(playerContainer);
        
        // 启动播放器
        console.log(`启动设备 ${deviceId} 播放器`);
        player.play();
        
        deviceConnection.streamEnabled = true;
        
        // 创建WebSocket连接
        const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsUrl = `${wsProtocol}//${window.location.host}/ws/wsscrcpy?udid=${deviceId}&port=${port}`;
        console.log(`准备连接设备 ${deviceId} WebSocket:`, wsUrl);
        
        // 关闭可能存在的旧连接
        if (deviceConnection.wsConnection) {
          try {
            console.log(`关闭设备 ${deviceId} 旧的WebSocket连接`);
            (deviceConnection.wsConnection as any).isClosed = true;
            deviceConnection.wsConnection.onmessage = null;
            deviceConnection.wsConnection.onerror = null;
            deviceConnection.wsConnection.onclose = null;
            deviceConnection.wsConnection.close();
          } catch (error) {
            console.warn(`关闭设备 ${deviceId} 旧WebSocket连接出错:`, error);
          }
        }
        
        // 创建新的WebSocket连接
        try {
          console.log(`创建设备 ${deviceId} 新的WebSocket连接`);
          const ws = new WebSocket(wsUrl);
          (ws as any).isClosed = false;
          (ws as any).deviceId = deviceId;
          deviceConnection.wsConnection = ws;
          
          // 设置WebSocket事件处理
          ws.binaryType = 'arraybuffer';
          
          // 设置连接超时
          const connectionTimeout = setTimeout(() => {
            if (!isComponentMounted.value || !isComponentActive.value || (ws as any).isClosed) return;
            
            if (ws.readyState !== WebSocket.OPEN) {
              console.error(`设备 ${deviceId} WebSocket连接超时`);
              deviceConnection.streamError = 'WebSocket连接超时';
              (ws as any).isClosed = true;
              ws.close();
            }
          }, 15000);
          
          ws.onopen = () => {
            if (!isComponentMounted.value || !isComponentActive.value || (ws as any).isClosed) return;
            console.log(`设备 ${deviceId} WebSocket连接成功`);
            clearTimeout(connectionTimeout);
            deviceConnection.streamError = null;
          };
          
          // 数据接收超时检查器
          let dataReceived = false;
          const checkDataTimer = setTimeout(() => {
            if (!isComponentMounted.value || !isComponentActive.value || (ws as any).isClosed) return;
            
            if (!dataReceived) {
              console.warn(`设备 ${deviceId} WebSocket连接成功但10秒内没有收到数据`);
              deviceConnection.streamError = '未收到视频数据，请检查设备串流状态';
              
              // 尝试自动重连
              stopOtherDeviceStream(deviceId).catch(console.error).finally(() => {
                if (!isComponentMounted.value || !isComponentActive.value) return;
                
                setTimeout(() => {
                  if (!isComponentMounted.value || !isComponentActive.value) return;
                  
                  if (device.status === 'online') {
                    console.log(`尝试自动重新连接设备 ${deviceId}...`);
                    startOtherDeviceStream(deviceId);
                  }
                }, 3000);
              });
            }
          }, 10000);
          
          ws.onmessage = (event) => {
            if ((ws as any).isClosed) return;
            
            if (event.data instanceof ArrayBuffer) {
              if (!isComponentMounted.value || !isComponentActive.value) return;
              
              // 标记已收到视频数据
              if (!deviceConnection.videoFrameReceived) {
                deviceConnection.videoFrameReceived = true;
                console.log(`设备 ${deviceId} 收到首帧视频数据`);
              }
              
              // 标记已收到数据
              if (!dataReceived) {
                dataReceived = true;
                clearTimeout(checkDataTimer);
              }
              
              // 处理视频数据
              const data = new Uint8Array(event.data);
              
              try {
                if (!deviceConnection.player) return;
                (deviceConnection.player as any).pushFrame(data);
              } catch (error) {
                console.error(`设备 ${deviceId} 处理视频帧出错:`, error);
              }
            }
          };
          
          ws.onerror = (event) => {
            if (!isComponentMounted.value || !isComponentActive.value || (ws as any).isClosed) return;
            
            console.error(`设备 ${deviceId} WebSocket错误:`, event);
            clearTimeout(connectionTimeout);
            clearTimeout(checkDataTimer);
            deviceConnection.streamError = 'WebSocket连接错误';
            deviceConnection.videoFrameReceived = false;
            stopOtherDeviceStream(deviceId).catch(console.error);
          };
          
          ws.onclose = () => {
            if (!isComponentMounted.value || !isComponentActive.value) return;
            
            console.log(`设备 ${deviceId} WebSocket连接关闭`);
            clearTimeout(connectionTimeout);
            clearTimeout(checkDataTimer);
            (ws as any).isClosed = true;
            deviceConnection.streamEnabled = false;
            deviceConnection.videoFrameReceived = false;
            
            if (deviceConnection.streamEnabled && !deviceConnection.streamError) {
              deviceConnection.streamError = 'WebSocket连接已关闭';
            }
          };
        } catch (error) {
          console.error(`创建设备 ${deviceId} WebSocket连接失败:`, error);
          deviceConnection.streamError = '创建WebSocket连接失败';
        }
      } catch (error) {
        console.error(`设备 ${deviceId} 创建或初始化播放器失败:`, error);
        deviceConnection.streamError = '视频播放器初始化失败';
        deviceConnection.player = null;
      }
    } else {
      deviceConnection.streamError = response.message || '启动串流失败';
    }
  } catch (error) {
    console.error(`启动设备 ${deviceId} 串流失败:`, error);
    if (isComponentMounted.value) {
      deviceConnection.streamError = error instanceof Error ? error.message : '未知错误';
    }
  }
};

// 停止从设备视频流
const stopOtherDeviceStream = async (deviceId: string) => {
  console.log(`开始停止设备 ${deviceId} 视频流`);
  
  const deviceConnection = otherDeviceConnections.value[deviceId];
  if (!deviceConnection) return;
  
  // 重置视频帧接收状态
  deviceConnection.videoFrameReceived = false;
  
  // 标记WebSocket为已关闭
  if (deviceConnection.wsConnection) {
    (deviceConnection.wsConnection as any).isClosed = true;
  }
  
  // 关闭WebSocket连接
  if (deviceConnection.wsConnection) {
    try {
      console.log(`关闭设备 ${deviceId} WebSocket连接`);
      deviceConnection.wsConnection.onmessage = null;
      deviceConnection.wsConnection.onerror = null;
      deviceConnection.wsConnection.onclose = null;
      deviceConnection.wsConnection.close();
    } catch (e) {
      console.warn(`关闭设备 ${deviceId} WebSocket连接出错:`, e);
    }
    deviceConnection.wsConnection = null;
  }
  
  // 停止播放器
  if (deviceConnection.player) {
    try {
      console.log(`停止设备 ${deviceId} WebCodecsPlayer`);
      deviceConnection.player.stop();
    } catch (e) {
      console.warn(`停止设备 ${deviceId} 播放器出错:`, e);
    }
    deviceConnection.player = null;
  }
  
  // 调用后端接口停止串流
  try {
    console.log(`调用后端API停止设备 ${deviceId} 串流`);
    await stopDeviceStream(deviceId);
  } catch (error) {
    console.error(`停止设备 ${deviceId} 串流失败:`, error);
  }
  
  deviceConnection.streamEnabled = false;
  console.log(`设备 ${deviceId} 视频流停止完成`);
};

// 获取从设备视频帧接收状态
const getOtherDeviceVideoReceived = (deviceId: string): boolean => {
  if (!otherDeviceConnections.value[deviceId]) {
    return false;
  }
  return !!otherDeviceConnections.value[deviceId].videoFrameReceived;
};

// 获取从设备视频流错误状态
const getOtherDeviceStreamError = (deviceId: string): string | null => {
  if (!otherDeviceConnections.value[deviceId]) {
    return null;
  }
  return otherDeviceConnections.value[deviceId].streamError;
};

// 重试从设备视频流
const retryOtherDeviceStream = async (deviceId: string) => {
  console.log(`重试设备 ${deviceId} 视频流`);
  
  // 先停止当前流
  await stopOtherDeviceStream(deviceId);
  
  // 重新启动流
  await startOtherDeviceStream(deviceId);
};

// 在WebSocket连接成功后记录设备信息并更新目标屏幕尺寸
const updateScreenResolution = (width: number, height: number) => {
  if (width > 0 && height > 0) {
    console.log(`更新设备分辨率: ${width}x${height}`);
    targetScreen.value = { width, height };
    
    // 记录缩放比例
    const scaleX = width / sourceScreen.value.width;
    const scaleY = height / sourceScreen.value.height;
    console.log(`坐标转换比例: scaleX=${scaleX.toFixed(2)}, scaleY=${scaleY.toFixed(2)}`);
  }
};

// 从WebSocket接收到初始化信息时调用
const handleInitialInfo = (data: any) => {
  if (data && data.videoWidth && data.videoHeight) {
    updateScreenResolution(data.videoWidth, data.videoHeight);
  } else if (data && data.screenWidth && data.screenHeight) {
    // 如果没有视频尺寸，尝试使用屏幕尺寸
    updateScreenResolution(data.screenWidth, data.screenHeight);
  }
};

// 发送按键控制事件到主设备
const sendKeyCodeEvent = (action: number, keycode: number, repeat: number = 0, metaState: number = 0) => {
  // 确保WebSocket连接可用
  if (!wsConnection.value || wsConnection.value.readyState !== WebSocket.OPEN) {
    console.warn('WebSocket连接不可用，无法发送按键事件');
    return;
  }
  
  // 创建按键事件消息对象
  const keyEvent = {
    type: "keycode",  
    data: {
      action: action,
      keycode: keycode,
      repeat: repeat,
      metaState: metaState
    }
  };

  console.log('发送按键事件:', keyEvent);

  try {
    // 发送事件到后端
    wsConnection.value.send(JSON.stringify(keyEvent));
    
    // 只有在主设备上且操作同步已开启时，才发送给从设备
    if (syncEnabled.value && isMainDevice.value) {
      sendKeyCodeEventToOtherDevices(action, keycode, repeat, metaState);
    }
  } catch (error) {
    console.error('发送按键事件失败:', error);
  }
};

// 向从设备发送按键事件
const sendKeyCodeEventToOtherDevices = (action: number, keycode: number, repeat: number = 0, metaState: number = 0) => {
  // 遍历所有从设备连接
  Object.keys(otherDeviceConnections.value).forEach(deviceId => {
    const deviceConnection = otherDeviceConnections.value[deviceId];
    
    // 检查WebSocket连接是否可用
    if (!deviceConnection.wsConnection || 
        deviceConnection.wsConnection.readyState !== WebSocket.OPEN) {
      console.warn(`从设备 ${deviceId} WebSocket连接不可用，无法发送按键事件`);
      return;
    }
    
    try {
      // 创建按键事件消息对象
      const keyEvent = {
        type: "keycode",  
        data: {
          action: action,
          keycode: keycode,
          repeat: repeat,
          metaState: metaState
        }
      };
      
      // 发送事件到从设备
      deviceConnection.wsConnection.send(JSON.stringify(keyEvent));
      console.log(`已同步按键事件到从设备 ${deviceId}:`, keyEvent);
    } catch (error) {
      console.error(`发送按键事件到从设备 ${deviceId} 失败:`, error);
    }
  });
};

// 发送特定的导航按键命令(Home,Back,Overview)
const sendNavigationCommand = (commandType: string) => {
  // 检查主设备是否在线
  if (!mainDevice.value || mainDevice.value.status !== 'online') {
    ElMessage.warning('主设备不在线，无法执行操作');
    return;
  }
  
  try {
    // 创建快捷命令消息对象
    const commandEvent = {
      type: commandType  // "home", "back", 或 "overview"
    };
    
    // 发送到当前设备
    if (wsConnection.value && wsConnection.value.readyState === WebSocket.OPEN) {
      wsConnection.value.send(JSON.stringify(commandEvent));
      console.log(`已发送 ${commandType} 命令到当前设备`);
    } else {
      console.warn('WebSocket连接不可用，无法发送命令');
      return;
    }
    
    // 只有在主设备上且操作同步已开启时，才发送给从设备
    if (syncEnabled.value && isMainDevice.value) {
      Object.keys(otherDeviceConnections.value).forEach(deviceId => {
        const deviceConnection = otherDeviceConnections.value[deviceId];
        
        if (deviceConnection.wsConnection && 
            deviceConnection.wsConnection.readyState === WebSocket.OPEN) {
          deviceConnection.wsConnection.send(JSON.stringify(commandEvent));
          console.log(`已同步 ${commandType} 命令到从设备 ${deviceId}`);
        }
      });
    }
    
    ElMessage.success(`已发送${commandType === 'home' ? '主页' : commandType === 'back' ? '返回' : '任务管理'}键命令`);
  } catch (error) {
    console.error(`发送 ${commandType} 命令失败:`, error);
    ElMessage.error(`发送命令失败: ${error}`);
  }
};

// 设备操作按钮相关方法
const goToHome = () => {
  // 使用按键事件发送HOME键
  if (mainDevice.value && mainDevice.value.status === 'online') {
    // 方法1：发送直接的HOME命令 (推荐)
    sendNavigationCommand('home');
  } else {
    ElMessage.warning('主设备不在线，无法执行操作');
  }
};

const goBack = () => {
  // 使用按键事件发送BACK键
  if (mainDevice.value && mainDevice.value.status === 'online') {
    // 方法1：发送直接的BACK命令 (推荐)
    sendNavigationCommand('back');
  } else {
    ElMessage.warning('主设备不在线，无法执行操作');
  }
};

const openOverview = () => {
  // 使用按键事件发送OVERVIEW键(APP_SWITCH)
  if (mainDevice.value && mainDevice.value.status === 'online') {
    // 方法1：发送直接的OVERVIEW命令 (推荐)
    sendNavigationCommand('overview');
  } else {
    ElMessage.warning('主设备不在线，无法执行操作');
  }
};

// 从设备上的操作按钮方法
const handleOtherDeviceAction = (device: Device, action: string) => {
  if (device.status !== 'online') {
    ElMessage.warning(`设备${device.name}不在线，无法执行操作`);
    return;
  }
  
  // 获取该设备的WebSocket连接
  const deviceConnection = otherDeviceConnections.value[device.deviceId];
  if (!deviceConnection || !deviceConnection.wsConnection || 
      deviceConnection.wsConnection.readyState !== WebSocket.OPEN) {
    ElMessage.warning(`设备${device.name}连接不可用，无法执行操作`);
    return;
  }
  
  // 创建命令对象
  const commandEvent = {
    type: action // "home", "back", 或 "overview"
  };
  
  try {
    // 只发送到当前操作的设备
    deviceConnection.wsConnection.send(JSON.stringify(commandEvent));
    console.log(`已发送 ${action} 命令到设备 ${device.name}`);
    ElMessage.success(`已发送${action === 'home' ? '主页' : action === 'back' ? '返回' : '任务管理'}键命令到设备 ${device.name}`);
  } catch (error) {
    console.error(`发送 ${action} 命令到设备 ${device.deviceId} 失败:`, error);
    ElMessage.error(`发送命令失败: ${error}`);
  }
};

// 子设备的触摸点数据
const otherDeviceTouchPoints = ref<Record<string, {
  x: number,
  y: number,
  startX: number,
  startY: number,
  isTouching: boolean
}>>({});

// 子设备鼠标按下事件
const handleOtherDeviceMouseDown = (event: MouseEvent, device: Device) => {
  // 初始化设备触摸状态（如果不存在）
  if (!otherDeviceTouchPoints.value[device.deviceId]) {
    otherDeviceTouchPoints.value[device.deviceId] = {
      x: 0,
      y: 0,
      startX: 0,
      startY: 0,
      isTouching: false
    };
  }
  
  // 获取设备触摸状态
  const touchPoint = otherDeviceTouchPoints.value[device.deviceId];
  touchPoint.isTouching = true;
  touchPoint.startX = event.offsetX;
  touchPoint.startY = event.offsetY;
  touchPoint.x = event.offsetX;
  touchPoint.y = event.offsetY;
  
  console.log(`设备 ${device.name} 鼠标按下事件`, {
    x: event.offsetX,
    y: event.offsetY,
    button: event.button,
    timestamp: new Date().toISOString()
  });

  // 发送触摸按下事件到该设备
  sendOtherDeviceTouchEvent(device.deviceId, TOUCH_ACTION.DOWN, event.offsetX, event.offsetY);
};

// 子设备鼠标移动事件
const handleOtherDeviceMouseMove = (event: MouseEvent, device: Device) => {
  // 获取设备触摸状态
  const touchPoint = otherDeviceTouchPoints.value[device.deviceId];
  if (!touchPoint || !touchPoint.isTouching) return;
  
  touchPoint.x = event.offsetX;
  touchPoint.y = event.offsetY;
  
  // 发送触摸移动事件到该设备
  sendOtherDeviceTouchEvent(device.deviceId, TOUCH_ACTION.MOVE, event.offsetX, event.offsetY);
};

// 子设备鼠标抬起事件
const handleOtherDeviceMouseUp = (event: MouseEvent, device: Device) => {
  // 获取设备触摸状态
  const touchPoint = otherDeviceTouchPoints.value[device.deviceId];
  if (!touchPoint || !touchPoint.isTouching) return;
  
  touchPoint.isTouching = false;
  
  console.log(`设备 ${device.name} 鼠标抬起事件`, {
    x: event.offsetX,
    y: event.offsetY,
    deltaX: event.offsetX - touchPoint.startX,
    deltaY: event.offsetY - touchPoint.startY,
    timestamp: new Date().toISOString()
  });

  // 发送触摸抬起事件到该设备
  sendOtherDeviceTouchEvent(device.deviceId, TOUCH_ACTION.UP, event.offsetX, event.offsetY);
};

// 子设备触摸开始事件 - 移动设备支持
const handleOtherDeviceTouchStart = (event: TouchEvent, device: Device) => {
  event.preventDefault(); // 阻止默认行为
  
  // 初始化设备触摸状态（如果不存在）
  if (!otherDeviceTouchPoints.value[device.deviceId]) {
    otherDeviceTouchPoints.value[device.deviceId] = {
      x: 0,
      y: 0,
      startX: 0,
      startY: 0,
      isTouching: false
    };
  }
  
  // 获取设备触摸状态
  const touchPoint = otherDeviceTouchPoints.value[device.deviceId];
  touchPoint.isTouching = true;
  
  const touch = event.touches[0];
  const target = event.target as HTMLElement;
  const rect = target.getBoundingClientRect();
  
  const x = touch.clientX - rect.left;
  const y = touch.clientY - rect.top;
  
  touchPoint.startX = x;
  touchPoint.startY = y;
  touchPoint.x = x;
  touchPoint.y = y;
  
  console.log(`设备 ${device.name} 触摸开始事件`, {
    x,
    y,
    timestamp: new Date().toISOString()
  });

  // 发送触摸按下事件到该设备
  sendOtherDeviceTouchEvent(device.deviceId, TOUCH_ACTION.DOWN, x, y);
};

// 子设备触摸移动事件 - 移动设备支持
const handleOtherDeviceTouchMove = (event: TouchEvent, device: Device) => {
  event.preventDefault(); // 阻止默认行为
  
  // 获取设备触摸状态
  const touchPoint = otherDeviceTouchPoints.value[device.deviceId];
  if (!touchPoint || !touchPoint.isTouching) return;
  
  const touch = event.touches[0];
  const target = event.target as HTMLElement;
  const rect = target.getBoundingClientRect();
  
  const x = touch.clientX - rect.left;
  const y = touch.clientY - rect.top;
  
  touchPoint.x = x;
  touchPoint.y = y;

  // 发送触摸移动事件到该设备
  sendOtherDeviceTouchEvent(device.deviceId, TOUCH_ACTION.MOVE, x, y);
};

// 子设备触摸结束事件 - 移动设备支持
const handleOtherDeviceTouchEnd = (event: TouchEvent, device: Device) => {
  event.preventDefault(); // 阻止默认行为
  
  // 获取设备触摸状态
  const touchPoint = otherDeviceTouchPoints.value[device.deviceId];
  if (!touchPoint || !touchPoint.isTouching) return;
  
  touchPoint.isTouching = false;
  
  console.log(`设备 ${device.name} 触摸结束事件`, {
    endX: touchPoint.x,
    endY: touchPoint.y,
    deltaX: touchPoint.x - touchPoint.startX,
    deltaY: touchPoint.y - touchPoint.startY,
    timestamp: new Date().toISOString()
  });

  // 发送触摸抬起事件到该设备
  sendOtherDeviceTouchEvent(device.deviceId, TOUCH_ACTION.UP, touchPoint.x, touchPoint.y);
};

// 向子设备发送触摸事件
const sendOtherDeviceTouchEvent = (deviceId: string, action: number, x: number, y: number) => {
  const deviceConnection = otherDeviceConnections.value[deviceId];
  
  // 检查WebSocket连接是否可用
  if (!deviceConnection || !deviceConnection.wsConnection || 
      deviceConnection.wsConnection.readyState !== WebSocket.OPEN) {
    console.warn(`设备 ${deviceId} WebSocket连接不可用，无法发送触摸事件`);
    return;
  }
  
  try {
    // 获取子设备的实际屏幕分辨率
    let targetDeviceWidth = sourceScreen.value.width;  // 默认使用与主设备相同的尺寸
    let targetDeviceHeight = sourceScreen.value.height;
    
    // 如果子设备的player中有记录视频尺寸，使用实际尺寸
    if (deviceConnection.player) {
      const player = deviceConnection.player as any;
      if (player.videoWidth && player.videoHeight) {
        targetDeviceWidth = player.videoWidth;
        targetDeviceHeight = player.videoHeight;
      }
    }
    
    // 缩放因子计算 - 子设备触摸层是主设备的一半尺寸
    const scaleX = targetDeviceWidth / (sourceScreen.value.width / 2);
    const scaleY = targetDeviceHeight / (sourceScreen.value.height / 2);

    const targetX = Math.round(x * scaleX);
    const targetY = Math.round(y * scaleY);

    // 创建触摸事件消息对象
    const touchEvent = {
      type: "touch",  
      data: {
        action: action,
        x: targetX,
        y: targetY
      }
    };
    
    // 发送事件到该设备
    deviceConnection.wsConnection.send(JSON.stringify(touchEvent));
    console.log(`已发送触摸事件到设备 ${deviceId}:`, {
      action,
      原始坐标: { x, y },
      目标坐标: { x: targetX, y: targetY }
    });
  } catch (error) {
    console.error(`发送触摸事件到设备 ${deviceId} 失败:`, error);
  }
};

</script>

<style scoped>
.player-container,
.other-device-player-container {
  width: 100%;
  height: 100%;
  background-color: #000;
  z-index: 1;
  position: relative;
  overflow: hidden; /* 防止内容溢出 */
}

/* 添加触摸层样式 */
.touch-layer {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  z-index: 10; /* 确保在视频层之上 */
  background-color: transparent;
  cursor: pointer;
}

.stream-error,
.stream-loading,
.stream-inactive,
.offline-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  justify-content: center;
  align-items: center;
  background-color: #1a1a1a;
  z-index: 2;
}

.offline-placeholder,
.stream-inactive {
  background-color: #1a1a1a;
  color: #a0a0a0;
  flex-direction: column;
  gap: 10px;
}

.stream-loading {
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
  will-change: transform; /* 优化渲染性能 */
  backface-visibility: hidden; /* 防止3D变换中可能的闪烁 */
  transform: translateZ(0); /* 启用GPU加速 */
}

.device-card:hover {
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
  transform: translateY(-2px);
}

.main-device {
  width: v-bind('DEVICE_CONFIG.SYNC.CANVAS_PORTRAIT.WIDTH + "px"');
  height: v-bind('(DEVICE_CONFIG.SYNC.CANVAS_PORTRAIT.HEIGHT + 100) + "px"'); /* 增加footer的高度 */
  grid-column: span 2;
  grid-row: span 2;
}

.other-device {
  width: v-bind('(DEVICE_CONFIG.SYNC.CANVAS_PORTRAIT.WIDTH / 2) + "px"');
  height: v-bind('(DEVICE_CONFIG.SYNC.CANVAS_PORTRAIT.HEIGHT / 2 + 94) + "px"'); /* 增加footer的高度 */
}

.device-header {
  padding: 10px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-bottom: 1px solid #f0f0f0;
  background-color: white; /* 确保背景色与卡片一致 */
}

.device-name {
  font-weight: 500;
  font-size: 14px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 70%;
}

.actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.status-tag {
  flex-shrink: 0;
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

.device-footer {
  padding: 12px 15px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  background-color: #f9f9f9;
  min-height: 50px; /* 确保足够高度 */
  box-sizing: border-box;
  border-top: none; /* 移除可能的边框 */
}

.device-id {
  color: #606266;
  font-size: 12px;
  white-space: nowrap;
  max-width: 50%; /* 限制ID宽度，防止挤压操作按钮 */
  overflow: hidden;
  text-overflow: ellipsis;
}

.device-actions {
  display: flex;
  justify-content: flex-end;
  flex-shrink: 0; /* 防止操作按钮被压缩 */
}

.device-actions .el-button {
  min-width: 80px; /* 确保按钮宽度足够 */
}

.retry-button {
  margin-top: 10px;
}

/* 消除边界可能的横线 */
.device-card > div {
  margin: 0;
  border: none;
}

/* 确保canvas边界平滑 */
.device-screen canvas {
  display: block; /* 防止inline元素底部可能的间隙 */
  vertical-align: bottom; /* 对于inline元素，消除底部间隙 */
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
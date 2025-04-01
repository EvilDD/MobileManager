<script setup lang="ts">
import { ref, onMounted, computed, provide, watch, onUnmounted } from "vue";
import { ElMessage, ElMessageBox } from "element-plus";
import {
  getGroupList,
  createGroup,
  updateGroup,
  deleteGroup,
  batchUpdateDevicesGroup,
  type GroupItem
} from "@/api/group";
import {
  getDeviceList,
  type Device,
  type DeviceListResult,
  batchGoHome,
  batchKillApps
} from "@/api/device";
import {
  getAppList,
  type App,
  batchInstallByDevices,
  batchUninstallByDevices,
  batchStartByDevices,
  batchStopByDevices,
  getBatchTaskStatus,
  type BatchTaskStatus
} from "@/api/app";
import GroupBadge from "./components/GroupBadge.vue";
import DeviceStream from "./components/DeviceStream.vue";
import DeviceScreenshot from "./components/DeviceScreenshot.vue";
import StreamDialog from "./components/StreamDialog.vue";
import {
  Plus,
  Refresh,
  Grid,
  ArrowDown,
  Search,
  Delete,
  Edit,
  More,
  VideoPlay,
  Switch,
  ArrowRight
} from "@element-plus/icons-vue";
import { useRouter } from "vue-router";
import { useCloudPhoneStore } from "@/store/modules/cloudphone";

defineOptions({
  name: "CloudPhone"
});

const router = useRouter();

// 分组数据
const groups = ref<GroupItem[]>([]);

// 分组搜索关键词
const groupSearchKeyword = ref("");

// 分页参数
const pagination = ref({
  page: 1,
  pageSize: 100
});

// 过滤后的分组
const filteredGroups = computed(() => {
  // 过滤用户定义的分组（不包括"新设备"）
  const userGroups = groups.value.filter(group => group.id !== 0);
  // 查找新设备组
  const newDeviceGroup = groups.value.find(group => group.id === 0) || {
    id: 0,
    name: "新设备",
    description: "未分组的设备",
    deviceCount: 0,
    createdAt: "",
    updatedAt: ""
  };
  
  if (!groupSearchKeyword.value) {
    return [newDeviceGroup, ...userGroups];
  }
  
  return [
    newDeviceGroup,
    ...userGroups.filter(group =>
      group.name.toLowerCase().includes(groupSearchKeyword.value.toLowerCase())
    )
  ];
});

// 提供分组列表给GroupBadge组件
provide("groupsList", computed(() => {
  // 过滤用户定义的分组（不包括"新设备"）
  const userGroups = groups.value.filter(group => group.id !== 0);
  // 查找新设备组
  const newDeviceGroup = groups.value.find(group => group.id === 0) || {
    id: 0,
    name: "新设备",
    description: "未分组的设备",
    deviceCount: 0,
    createdAt: "",
    updatedAt: ""
  };
  
  return [newDeviceGroup, ...userGroups];
}));

// 当前选中的分组
const cloudPhoneStore = useCloudPhoneStore();
const activeGroup = computed({
  get: () => cloudPhoneStore.activeGroupId,
  set: (value) => cloudPhoneStore.setActiveGroup(value)
});

// 搜索输入
const searchInput = ref("");

// 云手机设备列表
const loading = ref(false);
const devices = ref<Device[]>([]);

// 状态筛选选项
const statusFilter = ref("all");

// 筛选后的设备列表
const filteredDevices = computed(() => {
  if (!devices.value) return [];
  
  return devices.value.filter(device => {
    // 先按搜索关键词筛选
    const matchesSearch = !searchInput.value || 
      device.deviceId.toLowerCase().includes(searchInput.value.toLowerCase());
    
    // 再按状态筛选
    const matchesStatus = statusFilter.value === "all" || 
      (statusFilter.value === "online" && device.status === "online") ||
      (statusFilter.value === "offline" && device.status !== "online");
    
    return matchesSearch && matchesStatus;
  });
});

// 流对话框控制
const streamDialogVisible = ref(false);
const selectedDevice = ref<Device | null>(null);
const streamServerUrl = ref(import.meta.env.VITE_WSCRCPY_SERVER); // 使用环境变量配置服务器URL

// 截图刷新设置
const autoRefresh = ref(true);
const refreshInterval = ref(5000); // 默认5秒刷新一次

// 截图状态跟踪
const screenshotStatus = ref<Record<string, { success: boolean; error?: string }>>({});

// 截图事件处理
const handleScreenshotReady = (deviceId: string, imageData: string) => {
  // console.log(`设备 ${deviceId} 截图加载成功`);
  // 更新截图状态
  screenshotStatus.value[deviceId] = { success: true };
};

const handleScreenshotError = (deviceId: string, error: string) => {
  // console.error(`设备 ${deviceId} 截图加载失败:`, error);
  // 更新截图状态
  screenshotStatus.value[deviceId] = { success: false, error };
};

// 切换自动刷新状态
const toggleAutoRefresh = (enable: boolean) => {
  autoRefresh.value = enable;
  if (!enable) {
    ElMessage.info('已关闭自动刷新');
  } else {
    ElMessage.info(`已开启自动刷新，间隔 ${refreshInterval.value / 1000} 秒`);
  }
};

// 改变刷新间隔
const changeRefreshInterval = (interval: number) => {
  if (interval === 0) {
    // 永不刷新
    autoRefresh.value = false;
    ElMessage.info('已设置为永不刷新');
  } else {
    refreshInterval.value = interval;
    autoRefresh.value = true;
    ElMessage.info(`截图刷新间隔已设置为 ${interval / 1000} 秒`);
  }
};

// 改变状态筛选
const changeStatusFilter = (status: string) => {
  statusFilter.value = status;
};

// 获取分组列表
const getGroups = async () => {
  try {
    const res = await getGroupList({
      page: pagination.value.page,
      pageSize: pagination.value.pageSize,
      keyword: groupSearchKeyword.value
    });

    // 检查响应结构
    if (res && res.data && res.data.list) {
      // 直接使用后端返回的分组列表，不添加"新设备"
      groups.value = res.data.list;
    } else {
      console.error("分组列表数据格式不符合预期:", res);
      ElMessage.warning("获取分组列表数据格式不正确");
    }
  } catch (error) {
    console.error("获取分组列表失败:", error);
    ElMessage.error("获取分组列表失败");
  }
};

// 获取设备列表
const getDevices = async () => {
  try {
    loading.value = true;
    const res = await getDeviceList({
      page: pagination.value.page,
      pageSize: pagination.value.pageSize,
      groupId: activeGroup.value,
      keyword: ""  // 不再传递搜索关键词到后端
    });
    if (res.code === 0) {
      devices.value = res.data.list;
    } else {
      ElMessage.error(res.message || "获取云手机列表失败");
    }
  } catch (error) {
    console.error("获取云手机列表失败:", error);
    ElMessage.error("获取云手机列表失败");
  } finally {
    loading.value = false;
  }
};

// 切换分组
const changeGroup = (groupId: number) => {
  activeGroup.value = groupId;
  selectedDevices.value = []; // 清除已选中的设备列表
  getDevices();
};

// 添加新分组对话框
const showAddGroupDialog = ref(false);
const newGroup = ref({
  name: "",
  description: ""
});

// 添加新分组
const addNewGroup = async () => {
  try {
    if (!newGroup.value.name) {
      ElMessage.warning("请输入分组名称");
      return;
    }
    await createGroup(newGroup.value);
    ElMessage.success("添加分组成功");
    showAddGroupDialog.value = false;
    newGroup.value = { name: "", description: "" };
    getGroups();
  } catch (error) {
    console.error("添加分组失败:", error);
    ElMessage.error("添加分组失败");
  }
};

// 刷新分组列表
const refreshGroups = () => {
  getGroups();
  getDevices();
};

// 连接到云手机
const connectToPhone = (device: Device) => {
  if (device.status !== 'online') {
    ElMessage.warning(`设备 ${device.deviceId} 当前离线，无法连接`);
    return;
  }
  
  // 设置选中的设备并打开对话框
  selectedDevice.value = device;
  // 从环境变量读取服务器地址
  streamServerUrl.value = import.meta.env.VITE_WSCRCPY_SERVER || 'http://localhost:8000';
  streamDialogVisible.value = true;
};

// 点击截图时打开流对话框
const handleScreenshotClick = (device: Device) => {
  if (device.status === 'online') {
    connectToPhone(device);
  }
};

// 重启云手机
const restartPhone = (device: Device, event: MouseEvent) => {
  event.stopPropagation();
  ElMessage.warning(`重启云手机: ${device.deviceId}`);
};

// 关闭云手机
const shutdownPhone = (device: Device, event: MouseEvent) => {
  event.stopPropagation();
  ElMessage.error(`关闭云手机: ${device.deviceId}`);
};

// 修改设备过滤逻辑
const getDeviceCountByGroupId = (groupId: number) => {
  if (!devices.value || devices.value.length === 0) return 0;
  
  return devices.value.filter(device => device.groupId === groupId).length;
};

// 处理对话框关闭
const handleCloseDialog = () => {
  showAddGroupDialog.value = false;
  newGroup.value = { name: "", description: "" };
};

// 处理确认添加新分组
const handleConfirmAddGroup = () => {
  addNewGroup();
};

// 删除分组
const confirmDeleteGroup = (group: GroupItem, event: MouseEvent) => {
  // 阻止事件冒泡，避免触发分组选择
  event.stopPropagation();
  
  ElMessageBox.confirm(
    `确认删除分组"${group.name}"吗？`,
    '删除确认',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    }
  ).then(async () => {
    try {
      await deleteGroup({ id: group.id });
      ElMessage.success('删除分组成功');
      getGroups();
      // 如果当前选中的是被删除的分组，切换到"新设备"分组
      if (activeGroup.value === group.id) {
        changeGroup(0);
      }
    } catch (error) {
      console.error('删除分组失败:', error);
      ElMessage.error('删除分组失败');
    }
  }).catch(() => {
    // 用户取消删除操作
  });
};

// 编辑分组对话框
const showEditGroupDialog = ref(false);
const editingGroup = ref<GroupItem>({
  id: 0,
  name: "",
  description: "",
  deviceCount: 0,
  createdAt: "",
  updatedAt: ""
});

// 编辑分组
const editGroup = (group: GroupItem, event: MouseEvent) => {
  event.stopPropagation();
  editingGroup.value = { ...group };
  showEditGroupDialog.value = true;
};

// 保存编辑的分组
const saveEditGroup = async () => {
  try {
    if (!editingGroup.value.name) {
      ElMessage.warning("请输入分组名称");
      return;
    }
    await updateGroup({
      id: editingGroup.value.id,
      name: editingGroup.value.name,
      description: editingGroup.value.description
    });
    ElMessage.success("更新分组成功");
    showEditGroupDialog.value = false;
    getGroups();
  } catch (error) {
    console.error("更新分组失败:", error);
    ElMessage.error("更新分组失败");
  }
};

// 处理编辑对话框关闭
const handleEditDialogClose = () => {
  showEditGroupDialog.value = false;
  editingGroup.value = {
    id: 0,
    name: "",
    description: "",
    deviceCount: 0,
    createdAt: "",
    updatedAt: ""
  };
};

// 批量选择相关
const selectedDevices = ref<string[]>([]);
const isAllSelected = computed(() => {
  return filteredDevices.value.length > 0 && selectedDevices.value.length === filteredDevices.value.length;
});

// 选择设备
const handleSelect = (device: Device) => {
  const index = selectedDevices.value.findIndex(id => id === device.deviceId);
  if (index === -1) {
    selectedDevices.value.push(device.deviceId);
  } else {
    selectedDevices.value.splice(index, 1);
  }
};

// 全选/取消全选
const handleSelectAll = () => {
  if (isAllSelected.value) {
    selectedDevices.value = [];
  } else {
    selectedDevices.value = filteredDevices.value.map(device => device.deviceId);
  }
};

// 批量切换分组
const handleBatchChangeGroup = async (groupId: number) => {
  if (selectedDevices.value.length === 0) {
    ElMessage.warning('请先选择设备');
    return;
  }

  try {
    const selectedDeviceObjects = devices.value.filter(device => 
      selectedDevices.value.includes(device.deviceId)
    );
    
    await batchUpdateDevicesGroup({
      groupId,
      deviceIds: selectedDeviceObjects.map(device => device.id)
    });
    ElMessage.success('批量修改分组成功');
    getGroups(); // 刷新分组列表
    getDevices(); // 刷新设备列表
    selectedDevices.value = []; // 清空选择
  } catch (error) {
    console.error('批量修改分组失败:', error);
    ElMessage.error('批量修改分组失败');
  }
};

// 显示更多操作菜单
const showMoreActions = ref<Record<string, boolean>>({});
const toggleMoreActions = (deviceId: string) => {
  showMoreActions.value[deviceId] = !showMoreActions.value[deviceId];
};

// 移动分组对话框相关
const showMoveGroupDialog = ref(false);
const selectedGroupId = ref<number>(0);

// 确认移动分组
const handleConfirmMoveGroup = async () => {
  if (!selectedGroupId.value) {
    ElMessage.warning('请选择目标分组');
    return;
  }
  
  await handleBatchChangeGroup(selectedGroupId.value);
  showMoveGroupDialog.value = false;
  selectedGroupId.value = 0;
};

// 批量回到主菜单
const handleBatchGoHome = async () => {
  if (selectedDevices.value.length === 0) {
    ElMessage.warning('请先选择设备');
    return;
  }

  try {
    const res = await batchGoHome(selectedDevices.value);
    if (res.code === 0) {
      ElMessage.success('批量回到主菜单操作已发送');
      // 检查每个设备的操作结果
      Object.entries(res.data.results).forEach(([deviceId, error]) => {
        if (error) {
          ElMessage.warning(`设备 ${deviceId} 回到主菜单失败: ${error}`);
        }
      });
    } else {
      ElMessage.error(res.message || '批量回到主菜单失败');
    }
  } catch (error) {
    console.error('批量回到主菜单失败:', error);
    ElMessage.error('批量回到主菜单失败');
  }
};

// 批量清除后台应用
const handleBatchKillApps = async () => {
  if (selectedDevices.value.length === 0) {
    ElMessage.warning('请先选择设备');
    return;
  }

  // 添加二次确认
  try {
    await ElMessageBox.confirm(
      '确定要清除所选设备的后台应用吗？',
      '操作确认',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    );

    const res = await batchKillApps(selectedDevices.value);
    if (res.code === 0) {
      ElMessage.success('批量清除后台应用操作已发送');
      // 检查每个设备的操作结果
      Object.entries(res.data.results).forEach(([deviceId, error]) => {
        if (error) {
          ElMessage.warning(`设备 ${deviceId} 清除后台应用失败: ${error}`);
        }
      });
    } else {
      ElMessage.error(res.message || '批量清除后台应用失败');
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('批量清除后台应用失败:', error);
      ElMessage.error('批量清除后台应用失败');
    }
  }
};

// 应用列表数据
const appList = ref<App[]>([]);
const appListLoading = ref(false);
const appListVisible = ref(false);
const selectedAppId = ref<number | null>(null);

// 当前选择的操作类型
const currentOperation = ref<string>("");

// 获取应用列表
const loadAppList = async () => {
  try {
    appListLoading.value = true;
    const res = await getAppList({
      page: 1,
      pageSize: 100,
      appType: "用户应用"
    });
    if (res.code === 0) {
      appList.value = res.data.list;
    } else {
      ElMessage.error(res.message || "获取应用列表失败");
    }
  } catch (error) {
    console.error("获取应用列表失败:", error);
    ElMessage.error("获取应用列表失败");
  } finally {
    appListLoading.value = false;
  }
};

// 任务状态相关
const taskStatusDialogVisible = ref(false);
const taskStatusLoading = ref(false);
const currentTask = ref<BatchTaskStatus | null>(null);
const currentTaskId = ref<string>('');
const taskStatusTimer = ref<number | null>(null);

// 获取任务状态类型
const getTaskStatusType = (status: string) => {
  switch (status) {
    case 'complete': return 'success';
    case 'failed': return 'danger';
    case 'running': return 'warning';
    default: return 'info';
  }
};

// 获取任务状态文本
const getTaskStatusText = (status: string) => {
  switch (status) {
    case 'pending': return '等待执行';
    case 'running': return '执行中';
    case 'complete': return '执行完成';
    case 'failed': return '执行失败';
    default: return status;
  }
};

// 获取进度条状态
const getProgressStatus = (task: BatchTaskStatus | null) => {
  if (!task) return '';
  if (task.status === 'failed') return 'exception';
  if (task.status === 'complete') return 'success';
  return '';
};

// 刷新任务状态
const refreshTaskStatus = async () => {
  if (!currentTaskId.value) return;

  taskStatusLoading.value = true;
  try {
    const res = await getBatchTaskStatus(currentTaskId.value);
    currentTask.value = res.data;

    // 如果任务已完成或失败，停止轮询
    if (res.data.status === 'complete' || res.data.status === 'failed') {
      stopTaskStatusPolling();
    }
  } catch (error) {
    console.error('获取任务状态失败:', error);
    ElMessage.error('获取任务状态失败');
    stopTaskStatusPolling();
  } finally {
    taskStatusLoading.value = false;
  }
};

// 开始任务状态轮询
const startTaskStatusPolling = () => {
  stopTaskStatusPolling(); // 先清除可能存在的定时器
  taskStatusTimer.value = window.setInterval(refreshTaskStatus, 2000);
};

// 停止任务状态轮询
const stopTaskStatusPolling = () => {
  if (taskStatusTimer.value) {
    clearInterval(taskStatusTimer.value);
    taskStatusTimer.value = null;
  }
};

// 组件卸载时清理
onUnmounted(() => {
  stopTaskStatusPolling();
});

// 修改 handleAppSelected 函数
const handleAppSelected = async (app: App) => {
  selectedAppId.value = app.id;
  appListVisible.value = false;

  if (!currentOperation.value) {
    return;
  }

  // 获取选中的设备ID列表
  if (selectedDevices.value.length === 0) {
    ElMessage.warning("请先选择要操作的设备");
    return;
  }

  try {
    let res;
    const data = {
      id: selectedAppId.value,
      deviceIds: selectedDevices.value,
      maxWorker: 50
    };

    const operationMap = {
      install: "安装",
      uninstall: "卸载",
      start: "启动",
      stop: "停止"
    };

    switch (currentOperation.value) {
      case "install":
        res = await batchInstallByDevices(data);
        break;
      case "uninstall":
        res = await batchUninstallByDevices(data);
        break;
      case "start":
        res = await batchStartByDevices(data);
        break;
      case "stop":
        res = await batchStopByDevices(data);
        break;
      default:
        return;
    }

    if (res.code === 0) {
      // 直接打开任务状态对话框，不再显示提交提示
      currentTaskId.value = res.data.taskId;
      await refreshTaskStatus();
      taskStatusDialogVisible.value = true;
      startTaskStatusPolling();
    } else {
      ElMessage.error(res.message || `批量${operationMap[currentOperation.value]}失败`);
    }
  } catch (error) {
    console.error(`批量${currentOperation.value}失败:`, error);
    ElMessage.error(`批量${currentOperation.value}失败`);
  } finally {
    // 清除当前操作类型
    currentOperation.value = "";
  }
};

// 批量应用操作
const handleBatchAppOperation = async (operation: string) => {
  // 保存当前操作类型
  currentOperation.value = operation;
  // 显示应用选择对话框
  appListVisible.value = true;
};

onMounted(() => {
  getGroups();
  getDevices();
  loadAppList();
});
</script>

<template>
  <div class="cloud-phone-container">
    <!-- 左侧分组列表 -->
    <div class="group-sidebar">
      <div class="group-header">
        <div class="header-title">分组列表</div>
        <div class="header-actions">
          <el-tooltip content="刷新分组列表" placement="top">
            <el-button
              type="primary"
              size="small"
              circle
              class="action-btn"
              @click="refreshGroups"
            >
              <el-icon><Refresh /></el-icon>
            </el-button>
          </el-tooltip>
          <el-tooltip content="新建分组" placement="top">
            <el-button
              type="primary"
              size="small"
              circle
              class="action-btn"
              @click="() => (showAddGroupDialog = true)"
            >
              <el-icon><Plus /></el-icon>
            </el-button>
          </el-tooltip>
        </div>
      </div>

      <div class="group-search">
        <el-input
          v-model="groupSearchKeyword"
          placeholder="搜索分组"
          clearable
          size="small"
        >
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>
      </div>

      <div class="group-list">
        <div
          v-for="group in filteredGroups"
          :key="group.id"
          class="group-item"
          :class="{ active: activeGroup === group.id }"
          @click="changeGroup(group.id)"
        >
          <div class="group-item-content">
            <div class="group-icon">
              <GroupBadge v-if="group.id > 0" :group-id="group.id" small />
              <el-icon v-else><Grid /></el-icon>
            </div>
            <div class="group-info">
              <div class="group-name">{{ group.name }}</div>
              <div class="group-count">
                {{ group.deviceCount }} 台设备
              </div>
            </div>
            <!-- 添加删除图标，新设备分组不显示 -->
            <div v-if="group.id > 0" class="group-actions">
              <el-button
                size="small"
                type="primary"
                circle
                class="action-btn"
                :icon="Edit"
                @click="editGroup(group, $event)"
              />
              <el-button
                size="small"
                type="danger"
                circle
                class="action-btn"
                :icon="Delete"
                @click="confirmDeleteGroup(group, $event)"
              />
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 右侧内容区 -->
    <div class="content-area">
      <!-- 顶部操作栏 -->
      <div class="top-toolbar">
        <div class="left-controls">
          <el-checkbox
            v-model="isAllSelected"
            @change="handleSelectAll"
            :indeterminate="selectedDevices.length > 0 && !isAllSelected"
          >
            全选
          </el-checkbox>

          <el-button
            type="primary"
            :disabled="selectedDevices.length === 0"
            @click="() => (showMoveGroupDialog = true)"
          >
            移动分组
          </el-button>

          <el-button
            type="success"
            :disabled="selectedDevices.length === 0"
            @click="handleBatchGoHome"
          >
            回到主菜单
          </el-button>

          <el-button
            type="danger"
            :disabled="selectedDevices.length === 0"
            @click="handleBatchKillApps"
          >
            清除后台应用
          </el-button>

          <el-dropdown @command="handleBatchAppOperation" trigger="click">
            <el-button :loading="appListLoading">
              应用操作
              <el-icon class="el-icon--right"><arrow-down /></el-icon>
            </el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="install">安装应用</el-dropdown-item>
                <el-dropdown-item command="uninstall">卸载应用</el-dropdown-item>
                <el-dropdown-item command="start">启动应用</el-dropdown-item>
                <el-dropdown-item command="stop">停止应用</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>

          <el-dropdown @command="changeRefreshInterval" trigger="click">
            <el-button type="default" size="small">
              {{ autoRefresh ? (refreshInterval / 1000 + '秒刷新') : '永不刷新' }} <el-icon><ArrowDown /></el-icon>
            </el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item :command="1000">1秒刷新</el-dropdown-item>
                <el-dropdown-item :command="3000">3秒刷新</el-dropdown-item>
                <el-dropdown-item :command="5000">5秒刷新</el-dropdown-item>
                <el-dropdown-item :command="15000">15秒刷新</el-dropdown-item>
                <el-dropdown-item :command="0">永不刷新</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
        
        <div class="right-controls">
          <el-input
            v-model="searchInput"
            placeholder="搜索云手机"
            clearable
            class="search-input"
          >
            <template #prefix>
              <el-icon><Search /></el-icon>
            </template>
          </el-input>
          
          <el-dropdown @command="changeStatusFilter" trigger="click">
            <el-button type="default" size="small">
              {{ statusFilter === 'all' ? '全部' : statusFilter === 'online' ? '在线' : '离线' }} <el-icon><ArrowDown /></el-icon>
            </el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="all">全部</el-dropdown-item>
                <el-dropdown-item command="online">在线</el-dropdown-item>
                <el-dropdown-item command="offline">离线</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </div>

      <!-- 云手机列表 -->
      <div v-loading="loading" class="phone-grid">
        <div
          v-for="device in filteredDevices"
          :key="device.id"
          class="phone-card"
        >
          <div class="phone-header">
            <!-- 选择框和设备ID布局调整 -->
            <div class="header-left">
              <el-checkbox
                :model-value="selectedDevices.includes(device.deviceId)"
                @change="(val: boolean) => val ? selectedDevices.push(device.deviceId) : selectedDevices = selectedDevices.filter(id => id !== device.deviceId)"
              />
              <div class="device-id">{{ device.deviceId }}</div>
            </div>
            <div class="phone-status" :class="{ online: device.status === 'online' }" />
          </div>

          <div class="phone-preview">
            <device-screenshot
              v-if="device.status === 'online'"
              :device-id="device.deviceId"
              :auto-capture="true"
              :quality="80"
              :auto-refresh="autoRefresh"
              :refresh-interval="refreshInterval"
              @click="handleScreenshotClick(device)"
              @screenshot-ready="(imageData) => handleScreenshotReady(device.deviceId, imageData)"
              @screenshot-error="(err) => handleScreenshotError(device.deviceId, err)"
              :data-device-id="device.deviceId"
              ref="screenshotRef"
            />
            <div v-else class="offline-placeholder">
              <img src="@/assets/user.jpg" alt="手机预览" class="preview-img" />
              <div class="offline-text">设备离线</div>
            </div>
          </div>

          <!-- 操作按钮改为右下角的更多操作按钮 -->
          <div class="more-actions">
            <el-dropdown trigger="click">
              <el-button type="primary" circle size="small">
                <el-icon><More /></el-icon>
              </el-button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item @click="connectToPhone(device)">
                    <el-icon><VideoPlay /></el-icon>连接
                  </el-dropdown-item>
                  <el-dropdown-item @click="restartPhone(device, $event)">
                    <el-icon><Refresh /></el-icon>重启
                  </el-dropdown-item>
                  <el-dropdown-item @click="shutdownPhone(device, $event)">
                    <el-icon><Switch /></el-icon>关机
                  </el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </div>
        </div>

        <!-- 当没有数据时显示 -->
        <div v-if="devices.length === 0 && !loading" class="no-data">
          暂无云手机数据
        </div>
      </div>
    </div>

    <!-- 添加分组对话框 -->
    <el-dialog
      v-model="showAddGroupDialog"
      title="添加分组"
      width="30%"
      :close-on-click-modal="false"
      @close="handleCloseDialog"
    >
      <el-form :model="newGroup" label-width="80px">
        <el-form-item label="分组名称">
          <el-input v-model="newGroup.name" placeholder="请输入分组名称" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input
            v-model="newGroup.description"
            type="textarea"
            placeholder="请输入分组描述"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="handleCloseDialog">取消</el-button>
          <el-button type="primary" @click="handleConfirmAddGroup"
            >确定</el-button
          >
        </span>
      </template>
    </el-dialog>

    <!-- 编辑分组对话框 -->
    <el-dialog
      v-model="showEditGroupDialog"
      title="编辑分组"
      width="30%"
      @close="handleEditDialogClose"
    >
      <el-form>
        <el-form-item label="分组名称">
          <el-input v-model="editingGroup.name" placeholder="请输入分组名称" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input
            v-model="editingGroup.description"
            type="textarea"
            placeholder="请输入分组描述"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="handleEditDialogClose">取消</el-button>
          <el-button type="primary" @click="saveEditGroup">确定</el-button>
        </span>
      </template>
    </el-dialog>

    <!-- 云手机流对话框 -->
    <stream-dialog
      v-model="streamDialogVisible"
      :device-id="selectedDevice?.deviceId || ''"
      :server-url="streamServerUrl"
      @closed="selectedDevice = null"
    />

    <!-- 移动分组对话框 -->
    <el-dialog
      v-model="showMoveGroupDialog"
      title="移动到分组"
      width="30%"
      :close-on-click-modal="false"
    >
      <el-radio-group v-model="selectedGroupId" class="group-radio-list">
        <el-radio
          v-for="group in filteredGroups"
          :key="group.id"
          :label="group.id"
          class="group-radio-item"
        >
          {{ group.name }}
        </el-radio>
      </el-radio-group>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="showMoveGroupDialog = false">取消</el-button>
          <el-button type="primary" @click="handleConfirmMoveGroup">确定</el-button>
        </span>
      </template>
    </el-dialog>

    <!-- 修改应用选择对话框 -->
    <el-dialog
      v-model="appListVisible"
      title="选择应用"
      width="600px"
      :close-on-click-modal="false"
      @close="currentOperation = ''"
    >
      <el-table
        v-loading="appListLoading"
        :data="appList"
        style="width: 100%"
        height="400"
        @row-click="handleAppSelected"
      >
        <el-table-column prop="name" label="应用名称" />
        <el-table-column prop="packageName" label="包名" />
        <el-table-column prop="version" label="版本" width="100" />
      </el-table>
    </el-dialog>

    <!-- 任务状态对话框 -->
    <el-dialog
      v-model="taskStatusDialogVisible"
      title="任务执行状态"
      width="600px"
      :close-on-click-modal="false"
    >
      <div v-if="currentTask" class="task-status">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="任务ID">{{ currentTask.taskId }}</el-descriptions-item>
          <el-descriptions-item label="任务状态">
            <el-tag :type="getTaskStatusType(currentTask.status)">
              {{ getTaskStatusText(currentTask.status) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="总设备数">{{ currentTask.total }}</el-descriptions-item>
          <el-descriptions-item label="执行进度">
            <el-progress 
              :percentage="Math.round(((currentTask.completed + currentTask.failed) / currentTask.total) * 100)"
              :status="getProgressStatus(currentTask)"
            >
              <template #default>
                {{ currentTask.completed }}/{{ currentTask.total }}
                <span v-if="currentTask.failed > 0" style="color: #f56c6c">
                  (失败: {{ currentTask.failed }})
                </span>
              </template>
            </el-progress>
          </el-descriptions-item>
        </el-descriptions>

        <div class="task-results" style="margin-top: 20px">
          <el-table :data="currentTask.results" style="width: 100%">
            <el-table-column prop="deviceId" label="设备ID" width="180" />
            <el-table-column label="状态" width="120">
              <template #default="{ row }">
                <el-tag :type="row.status === 'complete' ? 'success' : 'danger'">
                  {{ row.status === 'complete' ? '成功' : '失败' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="message" label="执行结果" />
          </el-table>
        </div>
      </div>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="taskStatusDialogVisible = false">关闭</el-button>
          <el-button 
            type="primary" 
            @click="refreshTaskStatus" 
            :loading="taskStatusLoading"
            v-if="currentTask?.status === 'running' || currentTask?.status === 'pending'"
          >
            刷新
          </el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.cloud-phone-container {
  display: flex;
  height: 100%;
  width: 100%;
  background-color: #f0f2f5;
}

/* 左侧分组侧边栏样式 */
.group-sidebar {
  width: 240px;
  background-color: #fff;
  border-right: 1px solid #e0e0e0;
  display: flex;
  flex-direction: column;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.05);
}

.group-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px;
  border-bottom: 1px solid #f0f0f0;
}

.header-title {
  font-size: 16px;
  font-weight: bold;
  color: #303133;
}

.header-actions {
  display: flex;
  gap: 8px;
}

.action-btn {
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.1);
  transition: transform 0.2s;
}

.action-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 8px rgba(0, 0, 0, 0.15);
}

.group-search {
  padding: 12px 16px;
  border-bottom: 1px solid #f0f0f0;
}

.group-list {
  flex: 1;
  overflow-y: auto;
  padding: 12px 0;
}

.group-item {
  padding: 8px 16px;
  cursor: pointer;
  margin-bottom: 4px;
  border-radius: 6px;
  margin: 0 8px 6px;
  transition: all 0.3s;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
  background-color: #fafafa;
  border: 1px solid transparent;
}

.group-item:hover {
  background-color: #f5f7fa;
  transform: translateY(-2px);
  box-shadow: 0 3px 6px rgba(0, 0, 0, 0.1);
}

.group-item.active {
  background-color: #ecf5ff;
  color: #409eff;
  border-color: #b3d8ff;
  box-shadow: 0 3px 8px rgba(64, 158, 255, 0.15);
}

.group-item-content {
  display: flex;
  align-items: center;
}

.group-icon {
  margin-right: 12px;
  width: 22px;
  height: 22px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #606266;
}

.group-info {
  flex: 1;
}

.group-actions {
  display: none;
  gap: 8px;
}

.group-item:hover .group-actions {
  display: flex;
}

.group-name {
  font-weight: 500;
  margin-bottom: 2px;
}

.group-count {
  font-size: 12px;
  color: #909399;
}

/* 右侧内容区样式 */
.content-area {
  flex: 1;
  padding: 20px 24px;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  margin-left: 10px;
  margin-right: 30px;
  background-color: #fff;
  border-radius: 8px;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.05);
}

.top-toolbar {
  margin-bottom: 20px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-bottom: 16px;
  border-bottom: 1px solid #f0f0f0;
}

.left-controls {
  display: flex;
  align-items: center;
  gap: 12px;
}

.right-controls {
  display: flex;
  align-items: center;
  gap: 12px;
}

.search-input {
  width: 220px;
  margin-left: auto;
  margin-right: 12px;
}

/* 云手机卡片网格 */
.phone-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
  gap: 24px;
  overflow-y: auto;
  padding-bottom: 24px;
  padding-top: 12px;
}

@media screen and (max-width: 1600px) {
  .phone-grid {
    grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
  }
}

@media screen and (max-width: 1200px) {
  .phone-grid {
    grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  }
}

@media screen and (max-width: 768px) {
  .phone-grid {
    grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
  }
  
  .action-button {
    padding: 4px 4px;
    font-size: 11px;
  }
}

.phone-card {
  background-color: #fff;
  border-radius: 8px;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
  overflow: hidden;
  cursor: pointer;
  transition: transform 0.3s, box-shadow 0.3s;
  display: flex;
  flex-direction: column;
  height: 100%;
}

.phone-card:hover {
  transform: translateY(-5px);
  box-shadow: 0 5px 15px rgba(0, 0, 0, 0.2);
}

.phone-header {
  padding: 10px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid #f0f0f0;
  background-color: #f9f9f9;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 1;
  min-width: 0;
}

.device-id {
  font-size: 14px;
  color: #606266;
  font-weight: 500;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.phone-status {
  width: 12px;
  height: 12px;
  border-radius: 50%;
  background-color: #f56c6c;
  display: inline-block;
  margin-left: 8px;
  flex-shrink: 0;
}

.phone-status.online {
  background-color: #67c23a;
}

.phone-preview {
  position: relative;
  width: 100%;
  padding-top: 177.78%; /* 保持 360:640 的宽高比 */
  background-color: #f5f7fa; /* 改为浅灰色背景 */
  overflow: hidden;
  flex-grow: 1;
}

.preview-img {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  object-fit: contain;
}

/* 离线设备的样式 */
.offline-placeholder {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  background-color: rgba(0, 0, 0, 0.05);
}

.offline-text {
  position: absolute;
  bottom: 20px;
  left: 0;
  width: 100%;
  text-align: center;
  color: #909399;
  background-color: rgba(0, 0, 0, 0.5);
  padding: 4px 0;
  font-size: 12px;
  color: #fff;
}

/* 修改截图容器样式，使其可点击 */
.phone-preview :deep(.device-screenshot-container) {
  position: absolute !important;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background-color: transparent;
  cursor: pointer;
}

.phone-preview :deep(.screenshot-loading) {
  display: none !important; /* 完全隐藏加载提示 */
}

.phone-preview :deep(.screenshot-image) {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform 0.3s;
}

.phone-preview :deep(.screenshot-image:hover) {
  transform: scale(1.05);
}

.phone-actions {
  padding: 10px;
  display: flex;
  gap: 8px;
  justify-content: space-between;
  border-top: 1px solid #f0f0f0;
  background-color: #f9f9f9;
}

.action-button {
  padding: 4px 8px;
  font-size: 12px;
  flex: 1;
}

.no-data {
  grid-column: 1 / -1;
  text-align: center;
  padding: 30px;
  color: #909399;
}

/* 更多操作按钮样式 */
.more-actions {
  position: absolute;
  right: 12px;
  bottom: 12px;
  z-index: 10;
}

.more-actions .el-button {
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
  transition: all 0.3s;
}

.more-actions .el-button:hover {
  transform: scale(1.1);
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.15);
}

/* 移动分组对话框样式 */
.group-radio-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 12px;
}

.group-radio-item {
  margin-right: 0;
  padding: 8px;
  border-radius: 4px;
  transition: background-color 0.3s;
}

.group-radio-item:hover {
  background-color: #f5f7fa;
}

.task-status {
  .el-descriptions {
    margin-bottom: 20px;
  }
}

.task-results {
  max-height: 400px;
  overflow-y: auto;
}
</style>

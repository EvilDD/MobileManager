<script setup lang="ts">
import { ref, onMounted, computed, provide } from "vue";
import { ElMessage, ElMessageBox } from "element-plus";
import {
  getGroupList,
  createGroup,
  updateGroup,
  deleteGroup,
  type GroupItem
} from "@/api/group";
import {
  getDeviceList,
  type Device,
  type DeviceListResult
} from "@/api/device";
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
  Edit
} from "@element-plus/icons-vue";
import { useRouter } from "vue-router";

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
const activeGroup = ref(0);

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

// 截图刷新设置
const autoRefresh = ref(true);
const refreshInterval = ref(5000); // 默认5秒刷新一次

// 截图状态跟踪
const screenshotStatus = ref<Record<string, { success: boolean; error?: string }>>({});

// 截图事件处理
const handleScreenshotReady = (deviceId: string, imageData: string) => {
  console.log(`设备 ${deviceId} 截图加载成功`);
  // 更新截图状态
  screenshotStatus.value[deviceId] = { success: true };
};

const handleScreenshotError = (deviceId: string, error: string) => {
  console.error(`设备 ${deviceId} 截图加载失败:`, error);
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

onMounted(() => {
  getGroups();
  getDevices();
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
        <div class="screenshot-controls">
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

      <!-- 云手机列表 -->
      <div v-loading="loading" class="phone-grid">
        <div
          v-for="device in filteredDevices"
          :key="device.id"
          class="phone-card"
        >
          <div class="phone-header">
            <div class="device-id">{{ device.deviceId }}</div>
            <div
              class="phone-status"
              :class="{ online: device.status === 'online' }"
            />
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
          <div class="phone-actions">
            <el-button type="primary" size="small" class="action-button" @click="connectToPhone(device)">
              连接
            </el-button>
            <el-button type="warning" size="small" class="action-button" @click="restartPhone(device, $event)">
              重启
            </el-button>
            <el-button type="danger" size="small" class="action-button" @click="shutdownPhone(device, $event)">
              关机
            </el-button>
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
      @closed="selectedDevice = null"
    />
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
  gap: 12px;
  align-items: center;
  padding-bottom: 16px;
  border-bottom: 1px solid #f0f0f0;
}

.screenshot-controls {
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
</style>

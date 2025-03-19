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
import {
  Plus,
  Refresh,
  Grid,
  ArrowDown,
  Search,
  Delete,
  Edit
} from "@element-plus/icons-vue";

defineOptions({
  name: "CloudPhone"
});

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
  
  if (!groupSearchKeyword.value) {
    return [
      { id: 0, name: "新设备", description: "", createdAt: "", updatedAt: "" },
      ...userGroups
    ];
  }
  
  return [
    { id: 0, name: "新设备", description: "", createdAt: "", updatedAt: "" },
    ...userGroups.filter(group =>
      group.name.toLowerCase().includes(groupSearchKeyword.value.toLowerCase())
    )
  ];
});

// 提供分组列表给GroupBadge组件
provide("groupsList", computed(() => {
  // 过滤用户定义的分组（不包括"新设备"）
  const userGroups = groups.value.filter(group => group.id !== 0);
  
  // 确保新设备放在第一位，但是在GroupBadge组件中会过滤掉
  return [
    { id: 0, name: "新设备", description: "", createdAt: "", updatedAt: "" },
    ...userGroups
  ];
}));

// 当前选中的分组
const activeGroup = ref(0);

// 搜索输入
const searchInput = ref("");

// 云手机设备列表
const loading = ref(false);
const devices = ref<Device[]>([]);

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
      page: 1,
      size: 50
    });
    if (res.code === 0) {
      // 根据分组ID筛选
      if (activeGroup.value === 0) {
        // 全部
        devices.value = res.data.devices;
      } else {
        // 按分组筛选
        devices.value = res.data.devices.filter(
          device => device.info_entity.group_id === activeGroup.value
        );
      }
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
  ElMessage.success(`连接到云手机: ${device.info_entity.id}`);
};

// 修改设备过滤逻辑
const getDeviceCountByGroupId = (groupId: number) => {
  if (!devices.value || devices.value.length === 0) return 0;
  
  return devices.value.filter(device => {
    return device.info_entity && device.info_entity.group_id === groupId;
  }).length;
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
                {{ getDeviceCountByGroupId(group.id) }} 台设备
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
        <el-button type="primary" size="small">批量上线</el-button>
        <el-button type="danger" size="small">批量下线</el-button>
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
        <el-dropdown>
          <el-button type="default" size="small">
            全部 <el-icon><ArrowDown /></el-icon>
          </el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item>全部</el-dropdown-item>
              <el-dropdown-item>在线</el-dropdown-item>
              <el-dropdown-item>离线</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>

      <!-- 云手机列表 -->
      <div v-loading="loading" class="phone-grid">
        <div
          v-for="device in devices"
          :key="device.info_entity.id"
          class="phone-card"
          @click="connectToPhone(device)"
        >
          <div class="phone-preview">
            <div
              class="phone-status"
              :class="{ online: device.info_entity.status === 'allow_alloc' }"
            />
            <img src="@/assets/user.jpg" alt="手机预览" class="preview-img" />
            <GroupBadge
              :group-id="device.info_entity.group_id || 1"
              class="preview-badge"
            />
          </div>
          <div class="phone-info">
            <div class="phone-name">云手机 #{{ device.info_entity.id }}</div>
            <div class="phone-ip">{{ device.info_entity.inner_ip_v_4 }}</div>
            <div class="phone-type">
              {{
                device.info_entity.inst_type === 0
                  ? "房间实例"
                  : device.info_entity.inst_type === 1
                    ? "衣服实例"
                    : "人物实例"
              }}
            </div>
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
  padding: 16px;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.top-toolbar {
  margin-bottom: 16px;
  display: flex;
  gap: 10px;
  align-items: center;
}

.search-input {
  width: 200px;
  margin-left: auto;
}

/* 云手机卡片网格 */
.phone-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 16px;
  overflow-y: auto;
  padding-bottom: 20px;
}

.phone-card {
  background-color: #fff;
  border-radius: 8px;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
  overflow: hidden;
  cursor: pointer;
  transition:
    transform 0.3s,
    box-shadow 0.3s;
}

.phone-card:hover {
  transform: translateY(-5px);
  box-shadow: 0 5px 15px rgba(0, 0, 0, 0.2);
}

.phone-preview {
  height: 120px;
  background-color: #f5f7fa;
  position: relative;
  display: flex;
  justify-content: center;
  align-items: center;
}

.preview-img {
  max-height: 100%;
  max-width: 100%;
  object-fit: contain;
}

.phone-status {
  position: absolute;
  top: 10px;
  right: 10px;
  width: 12px;
  height: 12px;
  border-radius: 50%;
  background-color: #f56c6c;
}

.phone-status.online {
  background-color: #67c23a;
}

.phone-info {
  padding: 12px;
}

.phone-name {
  font-weight: bold;
  margin-bottom: 5px;
}

.phone-ip {
  color: #606266;
  font-size: 12px;
  margin-bottom: 3px;
}

.phone-type {
  color: #909399;
  font-size: 12px;
  background-color: #f0f2f5;
  padding: 2px 6px;
  border-radius: 3px;
  display: inline-block;
  margin-top: 5px;
}

.no-data {
  grid-column: 1 / -1;
  text-align: center;
  padding: 30px;
  color: #909399;
}

.preview-badge {
  position: absolute;
  bottom: 10px;
  left: 10px;
}
</style>

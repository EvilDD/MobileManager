<script setup lang="ts">
import { ref, onMounted, computed } from "vue";
import { ElMessage } from "element-plus";
import { getInstanceList, type Instance } from "@/api/instance";
import GroupBadge from "./components/GroupBadge.vue";
import {
  Plus,
  Refresh,
  Grid,
  ArrowDown,
  Search
} from "@element-plus/icons-vue";

defineOptions({
  name: "CloudPhone"
});

// 分组数据
const groups = ref([
  { id: 0, name: "全部", count: 0 },
  { id: 1, name: "分组一", count: 0 },
  { id: 2, name: "分组二", count: 0 },
  { id: 3, name: "分组三", count: 0 },
  { id: 4, name: "未分组", count: 0 }
]);

// 分组搜索关键词
const groupSearchKeyword = ref("");

// 过滤后的分组
const filteredGroups = computed(() => {
  if (!groupSearchKeyword.value) {
    return groups.value;
  }
  return groups.value.filter(group =>
    group.name.includes(groupSearchKeyword.value)
  );
});

// 当前选中的分组
const activeGroup = ref(0);

// 搜索输入
const searchInput = ref("");

// 云手机实例列表
const loading = ref(false);
const instances = ref<Instance[]>([]);

// 获取实例列表
const getPhoneList = async () => {
  try {
    loading.value = true;
    const res = await getInstanceList({
      page: 1,
      size: 50
    });
    if (res.code === 0) {
      // 模拟按分组筛选
      if (activeGroup.value === 0) {
        // 全部
        instances.value = res.data.instances;
      } else if (activeGroup.value >= 1 && activeGroup.value <= 3) {
        // 根据分组ID筛选
        const groupIndex = activeGroup.value - 1;
        instances.value = res.data.instances.filter(
          (_, index) => index % 4 === groupIndex
        );
      } else {
        // 未分组
        instances.value = res.data.instances.filter(
          (_, index) => index % 4 === 3
        );
      }

      // 更新分组计数
      groups.value[0].count = res.data.total;
      // 这里假设每个分组的数量，实际应该根据后端返回的数据计算
      groups.value[1].count = Math.floor(res.data.total / 4);
      groups.value[2].count = Math.floor(res.data.total / 3);
      groups.value[3].count = Math.floor(res.data.total / 5);
      groups.value[4].count =
        res.data.total -
        groups.value[1].count -
        groups.value[2].count -
        groups.value[3].count;
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
  // 这里应该根据分组ID重新获取数据
  getPhoneList();
};

// 添加新分组
const addNewGroup = () => {
  ElMessage.info("添加新分组功能待实现");
};

// 刷新分组列表
const refreshGroups = () => {
  ElMessage.success("正在刷新分组列表");
  getPhoneList();
};

// 连接到云手机
const connectToPhone = (instance: Instance) => {
  ElMessage.success(`连接到云手机: ${instance.info_entity.id}`);
};

onMounted(() => {
  getPhoneList();
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
              @click="refreshGroups"
              class="action-btn"
            >
              <el-icon><Refresh /></el-icon>
            </el-button>
          </el-tooltip>
          <el-tooltip content="新建分组" placement="top">
            <el-button
              type="primary"
              size="small"
              circle
              @click="addNewGroup"
              class="action-btn"
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
              <GroupBadge :group-id="group.id" small v-if="group.id > 0" />
              <el-icon v-else><Grid /></el-icon>
            </div>
            <div class="group-info">
              <div class="group-name">{{ group.name }}</div>
              <div class="group-count">{{ group.count }} 台设备</div>
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
      <div class="phone-grid" v-loading="loading">
        <div
          v-for="instance in instances"
          :key="instance.info_entity.id"
          class="phone-card"
          @click="connectToPhone(instance)"
        >
          <div class="phone-preview">
            <div
              class="phone-status"
              :class="{ online: instance.info_entity.status === 'allow_alloc' }"
            />
            <img src="@/assets/user.jpg" alt="手机预览" class="preview-img" />
            <GroupBadge
              :group-id="(instance.info_entity.id % 4) + 1"
              class="preview-badge"
            />
          </div>
          <div class="phone-info">
            <div class="phone-name">云手机 #{{ instance.info_entity.id }}</div>
            <div class="phone-ip">{{ instance.info_entity.inner_ip_v_4 }}</div>
            <div class="phone-type">
              {{
                instance.info_entity.inst_type === 0
                  ? "房间实例"
                  : instance.info_entity.inst_type === 1
                    ? "衣服实例"
                    : "人物实例"
              }}
            </div>
          </div>
        </div>

        <!-- 当没有数据时显示 -->
        <div v-if="instances.length === 0 && !loading" class="no-data">
          暂无云手机数据
        </div>
      </div>
    </div>
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

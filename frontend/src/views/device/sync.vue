<template>
  <div class="main-container">
    <el-card class="box-card">
      <template #header>
        <div class="card-header">
          <span>云机同步</span>
          <el-button type="primary" @click="handleSync">开始同步</el-button>
        </div>
      </template>
      <div class="card-content">
        <el-table :data="syncList" style="width: 100%" v-loading="loading">
          <el-table-column prop="deviceName" label="设备名称" width="180" />
          <el-table-column prop="deviceId" label="设备ID" width="180" />
          <el-table-column prop="lastSyncTime" label="上次同步时间" width="180" />
          <el-table-column prop="status" label="同步状态">
            <template #default="{ row }">
              <el-tag :type="row.status === '已完成' ? 'success' : row.status === '同步中' ? 'warning' : 'info'">
                {{ row.status }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作">
            <template #default="{ row }">
              <el-button type="primary" size="small" @click="handleSyncItem(row)">同步</el-button>
              <el-button type="info" size="small" @click="viewSyncDetail(row)">查看详情</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from "vue";

// 定义数据类型
interface SyncItem {
  deviceName: string;
  deviceId: string;
  lastSyncTime: string;
  status: string;
}

// 状态数据
const syncList = ref<SyncItem[]>([]);
const loading = ref(false);

// 初始化示例数据
onMounted(() => {
  syncList.value = [
    {
      deviceName: "云机设备1",
      deviceId: "CLDEV001",
      lastSyncTime: "2023-12-01 10:30:45",
      status: "已完成"
    },
    {
      deviceName: "云机设备2",
      deviceId: "CLDEV002",
      lastSyncTime: "2023-12-02 15:20:10",
      status: "未同步"
    },
    {
      deviceName: "云机设备3",
      deviceId: "CLDEV003",
      lastSyncTime: "2023-12-03 09:15:30",
      status: "同步中"
    }
  ];
});

// 同步所有设备
const handleSync = () => {
  loading.value = true;
  setTimeout(() => {
    loading.value = false;
    syncList.value.forEach(item => {
      item.status = "已完成";
      item.lastSyncTime = new Date().toLocaleString();
    });
  }, 2000);
};

// 同步单个设备
const handleSyncItem = (row: SyncItem) => {
  row.status = "同步中";
  setTimeout(() => {
    row.status = "已完成";
    row.lastSyncTime = new Date().toLocaleString();
  }, 1500);
};

// 查看同步详情
const viewSyncDetail = (row: SyncItem) => {
  console.log("查看设备同步详情:", row);
  // 这里实现查看详情的逻辑
};
</script>

<style scoped>
.main-container {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.card-content {
  margin-top: 20px;
}
</style> 
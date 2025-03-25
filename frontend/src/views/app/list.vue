<template>
  <div class="main-container">
    <el-card class="box-card">
      <template #header>
        <div class="card-header">
          <span>应用列表</span>
          <div>
            <el-button type="primary" @click="handleAddApp">添加应用</el-button>
            <el-button type="success" @click="handleImportApp">导入应用</el-button>
          </div>
        </div>
      </template>
      <div class="search-bar">
        <el-input v-model="searchKeyword" placeholder="搜索应用名称" class="search-input" />
        <el-select v-model="appType" placeholder="应用类型" class="filter-select">
          <el-option label="全部" value="" />
          <el-option label="系统应用" value="system" />
          <el-option label="用户应用" value="user" />
        </el-select>
        <el-button type="primary" @click="handleSearch">搜索</el-button>
        <el-button @click="resetSearch">重置</el-button>
      </div>
      <div class="card-content">
        <el-table :data="appList" style="width: 100%" v-loading="loading">
          <el-table-column prop="appName" label="应用名称" width="180" />
          <el-table-column prop="packageName" label="包名" width="220" />
          <el-table-column prop="version" label="版本" width="120" />
          <el-table-column prop="size" label="大小" width="120" />
          <el-table-column prop="type" label="类型" width="120">
            <template #default="{ row }">
              <el-tag :type="row.type === 'system' ? 'info' : 'success'">
                {{ row.type === 'system' ? '系统应用' : '用户应用' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作">
            <template #default="{ row }">
              <el-button type="primary" size="small" @click="handleEditApp(row)">编辑</el-button>
              <el-button type="danger" size="small" @click="handleDeleteApp(row)">删除</el-button>
              <el-button type="success" size="small" @click="handleInstallApp(row)">安装</el-button>
            </template>
          </el-table-column>
        </el-table>
        <div class="pagination-container">
          <el-pagination
            background
            layout="prev, pager, next"
            :total="total"
            :current-page="currentPage"
            :page-size="pageSize"
            @current-change="handlePageChange"
          />
        </div>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from "vue";

// 定义应用类型
interface AppItem {
  id: number;
  appName: string;
  packageName: string;
  version: string;
  size: string;
  type: 'system' | 'user';
}

// 状态数据
const appList = ref<AppItem[]>([]);
const loading = ref(false);
const total = ref(0);
const currentPage = ref(1);
const pageSize = ref(10);
const searchKeyword = ref('');
const appType = ref('');

// 初始化示例数据
onMounted(() => {
  fetchAppList();
});

// 获取应用列表
const fetchAppList = () => {
  loading.value = true;
  setTimeout(() => {
    // 模拟数据
    appList.value = [
      {
        id: 1,
        appName: '微信',
        packageName: 'com.tencent.mm',
        version: '8.0.20',
        size: '150MB',
        type: 'user'
      },
      {
        id: 2,
        appName: '支付宝',
        packageName: 'com.eg.android.AlipayGphone',
        version: '10.2.30',
        size: '98MB',
        type: 'user'
      },
      {
        id: 3,
        appName: '系统设置',
        packageName: 'com.android.settings',
        version: '12.0.0',
        size: '45MB',
        type: 'system'
      }
    ];
    total.value = 30; // 模拟总数
    loading.value = false;
  }, 500);
};

// 搜索处理
const handleSearch = () => {
  currentPage.value = 1;
  fetchAppList();
};

// 重置搜索
const resetSearch = () => {
  searchKeyword.value = '';
  appType.value = '';
  handleSearch();
};

// 分页处理
const handlePageChange = (page: number) => {
  currentPage.value = page;
  fetchAppList();
};

// 添加应用
const handleAddApp = () => {
  console.log('添加应用');
  // 这里实现添加应用的逻辑
};

// 导入应用
const handleImportApp = () => {
  console.log('导入应用');
  // 这里实现导入应用的逻辑
};

// 编辑应用
const handleEditApp = (row: AppItem) => {
  console.log('编辑应用:', row);
  // 这里实现编辑应用的逻辑
};

// 删除应用
const handleDeleteApp = (row: AppItem) => {
  console.log('删除应用:', row);
  // 这里实现删除应用的逻辑
};

// 安装应用
const handleInstallApp = (row: AppItem) => {
  console.log('安装应用:', row);
  // 这里实现安装应用的逻辑
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

.search-bar {
  margin-bottom: 20px;
  display: flex;
  gap: 10px;
}

.search-input {
  width: 250px;
}

.filter-select {
  width: 140px;
}

.card-content {
  margin-top: 20px;
}

.pagination-container {
  margin-top: 20px;
  display: flex;
  justify-content: center;
}
</style> 
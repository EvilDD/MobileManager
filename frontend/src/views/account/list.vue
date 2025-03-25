<template>
  <div class="main-container">
    <el-card class="box-card">
      <template #header>
        <div class="card-header">
          <span>帐号列表</span>
          <el-button type="primary" @click="handleAddAccount">添加帐号</el-button>
        </div>
      </template>
      <div class="search-bar">
        <el-input v-model="searchKeyword" placeholder="搜索帐号名称或备注" class="search-input" />
        <el-select v-model="accountStatus" placeholder="帐号状态" class="filter-select">
          <el-option label="全部" value="" />
          <el-option label="正常" value="normal" />
          <el-option label="禁用" value="disabled" />
        </el-select>
        <el-button type="primary" @click="handleSearch">搜索</el-button>
        <el-button @click="resetSearch">重置</el-button>
      </div>
      <div class="card-content">
        <el-table :data="accountList" style="width: 100%" v-loading="loading">
          <el-table-column prop="accountName" label="帐号名称" width="150" />
          <el-table-column prop="userName" label="用户名" width="150" />
          <el-table-column prop="email" label="邮箱" width="200" />
          <el-table-column prop="phone" label="手机号" width="130" />
          <el-table-column prop="createTime" label="创建时间" width="180" />
          <el-table-column prop="status" label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="row.status === 'normal' ? 'success' : 'danger'">
                {{ row.status === 'normal' ? '正常' : '禁用' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="remark" label="备注" width="150" />
          <el-table-column label="操作">
            <template #default="{ row }">
              <el-button type="primary" size="small" @click="handleEditAccount(row)">编辑</el-button>
              <el-button :type="row.status === 'normal' ? 'danger' : 'success'" size="small" @click="handleToggleStatus(row)">
                {{ row.status === 'normal' ? '禁用' : '启用' }}
              </el-button>
              <el-button type="danger" size="small" @click="handleDeleteAccount(row)">删除</el-button>
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

// 帐号类型定义
interface AccountItem {
  id: number;
  accountName: string;
  userName: string;
  email: string;
  phone: string;
  createTime: string;
  status: 'normal' | 'disabled';
  remark: string;
}

// 状态数据
const accountList = ref<AccountItem[]>([]);
const loading = ref(false);
const total = ref(0);
const currentPage = ref(1);
const pageSize = ref(10);
const searchKeyword = ref('');
const accountStatus = ref('');

// 初始化示例数据
onMounted(() => {
  fetchAccountList();
});

// 获取帐号列表
const fetchAccountList = () => {
  loading.value = true;
  setTimeout(() => {
    // 模拟数据
    accountList.value = [
      {
        id: 1,
        accountName: '测试帐号1',
        userName: 'test1',
        email: 'test1@example.com',
        phone: '13800138001',
        createTime: '2023-01-15 14:30:25',
        status: 'normal',
        remark: '测试用'
      },
      {
        id: 2,
        accountName: '测试帐号2',
        userName: 'test2',
        email: 'test2@example.com',
        phone: '13800138002',
        createTime: '2023-02-20 09:15:30',
        status: 'normal',
        remark: '开发环境'
      },
      {
        id: 3,
        accountName: '禁用帐号',
        userName: 'disabled1',
        email: 'disabled@example.com',
        phone: '13800138003',
        createTime: '2023-03-10 16:45:12',
        status: 'disabled',
        remark: '已停用'
      }
    ];
    total.value = 25; // 模拟总数
    loading.value = false;
  }, 500);
};

// 搜索处理
const handleSearch = () => {
  currentPage.value = 1;
  fetchAccountList();
};

// 重置搜索
const resetSearch = () => {
  searchKeyword.value = '';
  accountStatus.value = '';
  handleSearch();
};

// 分页处理
const handlePageChange = (page: number) => {
  currentPage.value = page;
  fetchAccountList();
};

// 添加帐号
const handleAddAccount = () => {
  console.log('添加帐号');
  // 这里实现添加帐号的逻辑
};

// 编辑帐号
const handleEditAccount = (row: AccountItem) => {
  console.log('编辑帐号:', row);
  // 这里实现编辑帐号的逻辑
};

// 切换帐号状态
const handleToggleStatus = (row: AccountItem) => {
  console.log('切换帐号状态:', row);
  row.status = row.status === 'normal' ? 'disabled' : 'normal';
  // 这里实现切换帐号状态的逻辑
};

// 删除帐号
const handleDeleteAccount = (row: AccountItem) => {
  console.log('删除帐号:', row);
  // 这里实现删除帐号的逻辑
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
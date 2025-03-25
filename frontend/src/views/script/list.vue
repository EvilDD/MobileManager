<template>
  <div class="main-container">
    <el-card class="box-card">
      <template #header>
        <div class="card-header">
          <span>脚本列表</span>
          <div>
            <el-button type="primary" @click="handleAddScript">新建脚本</el-button>
            <el-button type="success" @click="handleImportScript">导入脚本</el-button>
          </div>
        </div>
      </template>
      <div class="search-bar">
        <el-input v-model="searchKeyword" placeholder="搜索脚本名称或描述" class="search-input" />
        <el-select v-model="scriptType" placeholder="脚本类型" class="filter-select">
          <el-option label="全部" value="" />
          <el-option label="自动化测试" value="test" />
          <el-option label="数据采集" value="collect" />
          <el-option label="系统操作" value="system" />
        </el-select>
        <el-button type="primary" @click="handleSearch">搜索</el-button>
        <el-button @click="resetSearch">重置</el-button>
      </div>
      <div class="card-content">
        <el-table :data="scriptList" style="width: 100%" v-loading="loading">
          <el-table-column prop="scriptName" label="脚本名称" width="180" />
          <el-table-column prop="type" label="类型" width="120">
            <template #default="{ row }">
              <el-tag :type="getTagType(row.type)">
                {{ getTypeName(row.type) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="author" label="创建者" width="120" />
          <el-table-column prop="createTime" label="创建时间" width="180" />
          <el-table-column prop="updateTime" label="更新时间" width="180" />
          <el-table-column prop="description" label="描述" />
          <el-table-column label="操作" width="300">
            <template #default="{ row }">
              <el-button type="primary" size="small" @click="handleEditScript(row)">编辑</el-button>
              <el-button type="success" size="small" @click="handleRunScript(row)">运行</el-button>
              <el-button type="info" size="small" @click="handleCopyScript(row)">复制</el-button>
              <el-button type="danger" size="small" @click="handleDeleteScript(row)">删除</el-button>
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

// 定义脚本类型
interface ScriptItem {
  id: number;
  scriptName: string;
  type: 'test' | 'collect' | 'system';
  author: string;
  createTime: string;
  updateTime: string;
  description: string;
}

// 状态数据
const scriptList = ref<ScriptItem[]>([]);
const loading = ref(false);
const total = ref(0);
const currentPage = ref(1);
const pageSize = ref(10);
const searchKeyword = ref('');
const scriptType = ref('');

// 初始化示例数据
onMounted(() => {
  fetchScriptList();
});

// 获取脚本列表
const fetchScriptList = () => {
  loading.value = true;
  setTimeout(() => {
    // 模拟数据
    scriptList.value = [
      {
        id: 1,
        scriptName: '自动登录脚本',
        type: 'test',
        author: '张三',
        createTime: '2023-10-15 14:20:30',
        updateTime: '2023-11-05 09:45:15',
        description: '自动登录并验证账户状态的测试脚本'
      },
      {
        id: 2,
        scriptName: '数据采集工具',
        type: 'collect',
        author: '李四',
        createTime: '2023-09-20 10:30:45',
        updateTime: '2023-10-25 16:20:10',
        description: '从应用中采集用户行为数据'
      },
      {
        id: 3,
        scriptName: '系统配置脚本',
        type: 'system',
        author: '王五',
        createTime: '2023-08-10 08:15:20',
        updateTime: '2023-09-12 11:40:35',
        description: '自动配置系统参数和服务'
      }
    ];
    total.value = 35; // 模拟总数
    loading.value = false;
  }, 500);
};

// 根据脚本类型获取标签类型
const getTagType = (type: string) => {
  const typeMap: Record<string, string> = {
    'test': 'success',
    'collect': 'primary',
    'system': 'warning'
  };
  return typeMap[type] || 'info';
};

// 根据脚本类型获取类型名称
const getTypeName = (type: string) => {
  const nameMap: Record<string, string> = {
    'test': '自动化测试',
    'collect': '数据采集',
    'system': '系统操作'
  };
  return nameMap[type] || '未知类型';
};

// 搜索处理
const handleSearch = () => {
  currentPage.value = 1;
  fetchScriptList();
};

// 重置搜索
const resetSearch = () => {
  searchKeyword.value = '';
  scriptType.value = '';
  handleSearch();
};

// 分页处理
const handlePageChange = (page: number) => {
  currentPage.value = page;
  fetchScriptList();
};

// 新建脚本
const handleAddScript = () => {
  console.log('新建脚本');
  // 这里实现新建脚本的逻辑
};

// 导入脚本
const handleImportScript = () => {
  console.log('导入脚本');
  // 这里实现导入脚本的逻辑
};

// 编辑脚本
const handleEditScript = (row: ScriptItem) => {
  console.log('编辑脚本:', row);
  // 这里实现编辑脚本的逻辑
};

// 运行脚本
const handleRunScript = (row: ScriptItem) => {
  console.log('运行脚本:', row);
  // 这里实现运行脚本的逻辑
};

// 复制脚本
const handleCopyScript = (row: ScriptItem) => {
  console.log('复制脚本:', row);
  // 这里实现复制脚本的逻辑
};

// 删除脚本
const handleDeleteScript = (row: ScriptItem) => {
  console.log('删除脚本:', row);
  // 这里实现删除脚本的逻辑
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
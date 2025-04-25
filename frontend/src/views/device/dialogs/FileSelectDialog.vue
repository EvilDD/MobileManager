<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue';
import { ElMessage } from 'element-plus';
import { getFileList, type File as FileType } from '@/api/file';

defineOptions({
  name: 'FileSelectDialog'
});

const props = defineProps<{
  visible: boolean
}>();

const emit = defineEmits<{
  (e: 'update:visible', visible: boolean): void
  (e: 'select', file: FileType): void
}>();

// 使用计算属性处理visible属性
const dialogVisible = computed({
  get: () => props.visible,
  set: (value) => emit('update:visible', value)
});

// 文件列表数据
const fileList = ref<FileType[]>([]);
const loading = ref(false);

// 获取文件列表
const loadFileList = async () => {
  console.log('加载文件列表...');
  try {
    loading.value = true;
    const res = await getFileList({
      page: 1,
      pageSize: 100,
      originalName: ""
    });
    
    if (res.code === 0) {
      // 倒序排列
      fileList.value = [...res.data.list].reverse();
      console.log('文件列表加载成功，共', fileList.value.length, '个文件');
    } else {
      ElMessage.error(res.message || "获取文件列表失败");
    }
  } catch (error) {
    console.error("获取文件列表出错:", error);
    ElMessage.error("获取文件列表出错");
  } finally {
    loading.value = false;
  }
};

// 监听visible属性变化，当对话框显示时加载文件列表
watch(() => props.visible, (newVal) => {
  if (newVal) {
    console.log('对话框显示，加载文件列表');
    loadFileList();
  }
});

// 关闭对话框
const handleClose = () => {
  emit('update:visible', false);
};

// 选择文件
const handleFileSelected = (file: FileType) => {
  emit('select', file);
  handleClose();
};

// 格式化文件大小
const formatFileSize = (size: number) => {
  if (size < 1024) {
    return size + 'B';
  } else if (size < 1024 * 1024) {
    return (size / 1024).toFixed(2) + 'KB';
  } else if (size < 1024 * 1024 * 1024) {
    return (size / (1024 * 1024)).toFixed(2) + 'MB';
  } else {
    return (size / (1024 * 1024 * 1024)).toFixed(2) + 'GB';
  }
};

// 获取文件标签类型
const getFileTagType = (fileType: string): 'primary' | 'success' | 'info' | 'warning' | 'danger' => {
  switch (fileType) {
    case 'image':
      return 'success';
    case 'document':
      return 'primary';
    case 'video':
      return 'warning';
    case 'audio':
      return 'info';
    case 'archive':
      return 'danger';
    case 'app':
      return 'warning';
    default:
      return 'info';
  }
};

// 在组件挂载时加载文件列表
onMounted(() => {
  if (props.visible) {
    loadFileList();
  }
});

// 暴露刷新方法
defineExpose({
  refresh: loadFileList
});
</script>

<template>
  <el-dialog
    v-model="dialogVisible"
    title="选择要推送的文件"
    width="600px"
    :close-on-click-modal="false"
    @close="handleClose"
  >
    <el-table
      v-loading="loading"
      :data="fileList"
      style="width: 100%"
      height="400"
      @row-click="handleFileSelected"
    >
      <el-table-column prop="fileName" label="文件名称" width="180" />
      <el-table-column prop="originalName" label="原始文件名" width="180" />
      <el-table-column label="大小" width="100">
        <template #default="{ row }">
          {{ formatFileSize(row.fileSize) }}
        </template>
      </el-table-column>
      <el-table-column prop="fileType" label="类型" width="100">
        <template #default="{ row }">
          <el-tag :type="getFileTagType(row.fileType)">
            {{ row.fileType || '未知' }}
          </el-tag>
        </template>
      </el-table-column>
    </el-table>
  </el-dialog>
</template>

<style scoped>
/* 可以在这里添加特定的样式 */
</style> 
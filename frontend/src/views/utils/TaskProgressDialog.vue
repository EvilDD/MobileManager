<template>
  <el-dialog v-model="dialogVisible" title="任务进度" width="600px">
    <div v-if="taskStatus">
      <el-descriptions :column="2" border>
        <el-descriptions-item label="任务ID">{{ taskStatus.taskId }}</el-descriptions-item>
        <el-descriptions-item label="状态">{{ getTaskStatusText(taskStatus.status) }}</el-descriptions-item>
        <el-descriptions-item label="总设备数">{{ taskStatus.total }}</el-descriptions-item>
        <el-descriptions-item label="已完成">{{ taskStatus.completed }}</el-descriptions-item>
        <el-descriptions-item label="失败数">{{ taskStatus.failed }}</el-descriptions-item>
        <el-descriptions-item label="进度">
          <el-progress 
            :percentage="Math.round(((taskStatus.completed + taskStatus.failed) / taskStatus.total) * 100)" 
            :status="getProgressStatus(taskStatus)"
          />
          <div class="progress-text">
            {{ taskStatus.completed }}/{{ taskStatus.total }} {{ taskStatus.failed > 0 ? `(失败: ${taskStatus.failed})` : '' }}
          </div>
        </el-descriptions-item>
      </el-descriptions>

      <el-divider content-position="center">设备列表</el-divider>

      <el-table :data="taskStatus.results" style="width: 100%">
        <el-table-column prop="deviceId" label="设备ID" width="180" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 'complete' ? 'success' : 'danger'">
              {{ row.status === 'complete' ? '成功' : '失败' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="message" label="信息" />
      </el-table>
    </div>
    <template #footer>
      <span class="dialog-footer">
        <el-button @click="handleClose">关闭</el-button>
        <el-button type="primary" @click="handleRefresh" :loading="loading">
          刷新
        </el-button>
      </span>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, onMounted, watch, onUnmounted } from 'vue';
import { ElMessage } from 'element-plus';
import { getBatchTaskStatus as getFileBatchTaskStatus, type BatchTaskStatus } from '@/api/file';
import { getBatchTaskStatus as getAppBatchTaskStatus } from '@/api/app';

const props = withDefaults(defineProps<{
  visible: boolean;
  taskId: string;
  taskType?: 'app' | 'file'; // 任务类型：app 或 file
  autoRefresh?: boolean;    // 是否自动刷新
  refreshInterval?: number; // 刷新间隔 (毫秒)
}>(), {
  taskType: 'file',         // 默认为文件任务
  autoRefresh: true,
  refreshInterval: 2000
});

const emit = defineEmits<{
  'update:visible': [value: boolean];
  'update:taskStatus': [status: BatchTaskStatus | null];
  'close': [];
  'refresh': [status: BatchTaskStatus | null];
}>();

// 内部状态
const dialogVisible = ref(props.visible);
const taskStatus = ref<BatchTaskStatus | null>(null);
const loading = ref(false);
const refreshTimer = ref<number | null>(null);

// 监听visible属性变化
watch(() => props.visible, (val) => {
  dialogVisible.value = val;
  if (val && props.taskId) {
    fetchTaskStatus();
    if (props.autoRefresh) {
      startAutoRefresh();
    }
  } else {
    stopAutoRefresh();
  }
});

// 监听dialogVisible变化
watch(dialogVisible, (val) => {
  emit('update:visible', val);
  if (!val) {
    stopAutoRefresh();
  }
});

// 监听taskId变化
watch(() => props.taskId, (val) => {
  if (val && dialogVisible.value) {
    fetchTaskStatus();
    if (props.autoRefresh) {
      startAutoRefresh();
    }
  } else {
    stopAutoRefresh();
  }
});

// 组件挂载时
onMounted(() => {
  if (props.visible && props.taskId) {
    fetchTaskStatus();
    if (props.autoRefresh) {
      startAutoRefresh();
    }
  }
});

// 组件卸载时
onUnmounted(() => {
  stopAutoRefresh();
});

// 获取任务状态
const fetchTaskStatus = async () => {
  if (!props.taskId) return;
  
  loading.value = true;
  
  try {
    // 根据任务类型选择不同的API
    const api = props.taskType === 'app' ? getAppBatchTaskStatus : getFileBatchTaskStatus;
    const res = await api(props.taskId);
    
    if (res.code === 0) {
      taskStatus.value = res.data;
      emit('update:taskStatus', res.data);
      
      // 如果任务已完成或失败，停止自动刷新
      if (res.data.status === 'complete' || res.data.status === 'failed') {
        stopAutoRefresh();
      }
    } else {
      ElMessage.error(res.message || "获取任务状态失败");
    }
  } catch (error) {
    console.error("获取任务状态出错:", error);
    ElMessage.error("获取任务状态出错");
    stopAutoRefresh();
  } finally {
    loading.value = false;
  }
};

// 开始自动刷新
const startAutoRefresh = () => {
  stopAutoRefresh(); // 先清除可能存在的定时器
  refreshTimer.value = window.setInterval(fetchTaskStatus, props.refreshInterval);
};

// 停止自动刷新
const stopAutoRefresh = () => {
  if (refreshTimer.value) {
    clearInterval(refreshTimer.value);
    refreshTimer.value = null;
  }
};

// 手动刷新
const handleRefresh = () => {
  fetchTaskStatus();
  emit('refresh', taskStatus.value);
};

// 关闭对话框
const handleClose = () => {
  dialogVisible.value = false;
  emit('update:visible', false);
  emit('close');
  stopAutoRefresh();
};

// 获取任务状态文本
const getTaskStatusText = (status: string) => {
  switch (status) {
    case 'pending':
      return '等待执行';
    case 'running':
      return '执行中';
    case 'complete':
      return '已完成';
    case 'failed':
      return '执行失败';
    default:
      return status;
  }
};

// 获取进度条状态
const getProgressStatus = (task: BatchTaskStatus) => {
  if (task.status === 'failed') return 'exception';
  if (task.status === 'complete') return 'success';
  return '';
};
</script>
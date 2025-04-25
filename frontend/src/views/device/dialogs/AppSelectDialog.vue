<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue';
import { ElMessage } from 'element-plus';
import { getAppList, type App } from '@/api/app';

defineOptions({
  name: 'AppSelectDialog'
});

const props = defineProps<{
  visible: boolean
}>();

const emit = defineEmits<{
  (e: 'update:visible', visible: boolean): void
  (e: 'select', app: App): void
}>();

// 使用计算属性处理visible属性
const dialogVisible = computed({
  get: () => props.visible,
  set: (value) => emit('update:visible', value)
});

// 应用列表数据
const appList = ref<App[]>([]);
const loading = ref(false);

// 获取应用列表
const loadAppList = async () => {
  try {
    loading.value = true;
    const res = await getAppList({
      page: 1,
      pageSize: 100,
      appType: "用户应用"
    });
    if (res.code === 0) {
      // 倒序排列
      appList.value = [...res.data.list].reverse();
      console.log('应用列表加载成功，共', appList.value.length, '个应用');
    } else {
      ElMessage.error(res.message || "获取应用列表失败");
    }
  } catch (error) {
    console.error("获取应用列表失败:", error);
    ElMessage.error("获取应用列表失败");
  } finally {
    loading.value = false;
  }
};

// 监听visible属性变化，当对话框显示时加载应用列表
watch(() => props.visible, (newVal) => {
  if (newVal) {
    console.log('应用对话框显示，加载应用列表');
    loadAppList();
  }
});

// 关闭对话框
const handleClose = () => {
  emit('update:visible', false);
};

// 选择应用
const handleAppSelected = (app: App) => {
  emit('select', app);
  handleClose();
};

// 在组件挂载时加载应用列表
onMounted(() => {
  if (props.visible) {
    loadAppList();
  }
});

// 暴露刷新方法
defineExpose({
  refresh: loadAppList
});
</script>

<template>
  <el-dialog
    v-model="dialogVisible"
    title="选择应用"
    width="600px"
    :close-on-click-modal="false"
    @close="handleClose"
  >
    <el-table
      v-loading="loading"
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
</template>

<style scoped>
/* 可以在这里添加特定的样式 */
</style> 
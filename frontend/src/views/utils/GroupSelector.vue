<template>
  <el-dialog v-model="dialogVisible" :title="title" width="500px">
    <el-form ref="groupFormRef" :model="form" label-width="100px">
      <el-form-item label="选择分组" prop="groupId">
        <el-select v-model="form.groupId" placeholder="请选择分组" style="width: 100%">
          <el-option
            v-for="group in groupList"
            :key="group.id"
            :label="group.name"
            :value="group.id"
          >
            <span>{{ group.name }}</span>
            <span style="float: right; color: #8492a6; font-size: 13px">
              ({{ group.deviceCount }}台设备)
            </span>
          </el-option>
        </el-select>
      </el-form-item>
      <el-form-item label="并发数" prop="maxWorker">
        <el-input-number 
          v-model="form.maxWorker" 
          :min="1" 
          :max="100"
          placeholder="请输入并发数"
          style="width: 100%"
        />
        <div class="form-tip">
          并发数表示同时处理的设备数量，建议设置在 1-20 之间，过大的并发数会被服务器自动调整
        </div>
      </el-form-item>
    </el-form>
    <template #footer>
      <span class="dialog-footer">
        <el-button @click="handleCancel">取消</el-button>
        <el-button type="primary" @click="handleConfirm" :loading="loading">
          确认
        </el-button>
      </span>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, watch } from 'vue';
import { ElMessage } from 'element-plus';
import { getGroupList, type GroupItem } from '@/api/group';

const props = withDefaults(defineProps<{
  title?: string;
  visible: boolean;
  maxWorker?: number;
}>(), {
  title: '选择分组',
  maxWorker: 50
});

const emit = defineEmits<{
  'update:visible': [value: boolean];
  'confirm': [data: {
    groupId: number;
    maxWorker: number;
  }];
  'cancel': [];
}>();

// 内部状态
const dialogVisible = ref(props.visible);
const groupList = ref<GroupItem[]>([]);
const loading = ref(false);

// 表单数据
const form = reactive({
  groupId: 0,
  maxWorker: props.maxWorker
});

// 监听visible属性变化
watch(() => props.visible, (val) => {
  dialogVisible.value = val;
  if (val) {
    // 打开对话框时获取分组列表
    fetchGroupList();
    // 重置表单
    form.groupId = 0;
    form.maxWorker = props.maxWorker;
  }
});

// 监听dialogVisible变化
watch(dialogVisible, (val) => {
  emit('update:visible', val);
});

// 初始化
onMounted(() => {
  dialogVisible.value = props.visible;
  if (props.visible) {
    fetchGroupList();
  }
});

// 获取分组列表
const fetchGroupList = async () => {
  loading.value = true;
  try {
    const res = await getGroupList({
      page: 1,
      pageSize: 1000 // 获取较多分组
    });
    
    if (res.code === 0 && res.data) {
      groupList.value = res.data.list || [];
    } else {
      ElMessage.error(res.message || "获取分组列表失败");
    }
  } catch (error) {
    console.error("获取分组列表出错:", error);
    ElMessage.error("获取分组列表出错");
  } finally {
    loading.value = false;
  }
};

// 确认选择
const handleConfirm = () => {
  if (!form.groupId) {
    ElMessage.warning('请选择分组');
    return;
  }

  emit('confirm', {
    groupId: form.groupId,
    maxWorker: form.maxWorker
  });
};

// 取消选择
const handleCancel = () => {
  dialogVisible.value = false;
  emit('update:visible', false);
  emit('cancel');
};
</script>

<style scoped>
.form-tip {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
  line-height: 1.4;
}
</style> 
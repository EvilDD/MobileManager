<script setup lang="ts">
import { ref, computed } from "vue";
import { ElMessage } from "element-plus";
import { Grid, Search } from "@element-plus/icons-vue";
import GroupBadge from "./GroupBadge.vue";
import type { GroupItem } from "@/api/group";

const props = defineProps({
  visible: {
    type: Boolean,
    default: false
  },
  groups: {
    type: Array as () => GroupItem[],
    default: () => []
  },
  loading: {
    type: Boolean,
    default: false
  }
});

const emit = defineEmits(['update:visible', 'confirm']);

// 双向绑定visible属性
const dialogVisible = computed({
  get: () => props.visible,
  set: (value) => emit('update:visible', value)
});

// 搜索关键词
const searchKeyword = ref("");

// 过滤后的分组列表
const filteredGroups = computed(() => {
  if (!searchKeyword.value) return props.groups;
  
  return props.groups.filter(group => 
    group.name.toLowerCase().includes(searchKeyword.value.toLowerCase())
  );
});

// 选中的分组ID
const selectedGroupId = ref<number>(0);

// 关闭对话框
const handleClose = () => {
  dialogVisible.value = false;
  selectedGroupId.value = 0;
  searchKeyword.value = "";
};

// 确认移动
const handleConfirm = () => {
  if (!selectedGroupId.value) {
    ElMessage.warning('请选择目标分组');
    return;
  }
  
  emit('confirm', selectedGroupId.value);
  handleClose();
};
</script>

<template>
  <el-dialog
    v-model="dialogVisible"
    title="移动到分组"
    width="400px"
    :close-on-click-modal="false"
    @closed="handleClose"
  >
    <div class="search-container">
      <el-input
        v-model="searchKeyword"
        placeholder="搜索分组"
        clearable
        size="default"
        prefix-icon="search"
      >
        <template #prefix>
          <el-icon><Search /></el-icon>
        </template>
      </el-input>
    </div>
    
    <div class="dialog-content" v-loading="loading">
      <el-radio-group v-model="selectedGroupId" class="group-radio-list">
        <el-radio
          v-for="group in filteredGroups"
          :key="group.id"
          :value="group.id"
          class="group-radio-item"
        >
          <div class="group-radio-content">
            <div class="group-icon">
              <GroupBadge v-if="group.id > 0" :group-id="group.id" small />
              <el-icon v-else class="new-device-icon"><Grid /></el-icon>
            </div>
            <div class="group-info">
              <div class="group-name">{{ group.name }}</div>
              <div class="group-count">{{ group.deviceCount }} 台设备</div>
            </div>
          </div>
        </el-radio>
      </el-radio-group>
      
      <!-- 没有找到分组的提示 -->
      <div v-if="filteredGroups.length === 0" class="no-groups-tip">
        未找到匹配的分组
      </div>
    </div>
    <template #footer>
      <span class="dialog-footer">
        <el-button @click="handleClose">取消</el-button>
        <el-button type="primary" @click="handleConfirm" :loading="loading">确定</el-button>
      </span>
    </template>
  </el-dialog>
</template>

<style scoped>
.search-container {
  margin-bottom: 15px;
}

.dialog-content {
  max-height: 400px;
  overflow-y: auto;
}

.group-radio-list {
  display: flex;
  flex-direction: column;
  width: 100%;
}

.group-radio-item {
  margin-right: 0;
  padding: 12px;
  border-radius: 6px;
  transition: background-color 0.3s;
  margin-bottom: 8px;
  border: 1px solid #ebeef5;
  height: auto;
}

.group-radio-item:hover {
  background-color: #f5f7fa;
}

.group-radio-content {
  display: flex;
  align-items: center;
  width: 100%;
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

.new-device-icon {
  color: #909399;
  font-size: 18px;
}

.group-info {
  flex: 1;
  overflow: hidden;
}

.group-name {
  font-weight: 500;
  margin-bottom: 4px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.group-count {
  font-size: 12px;
  color: #909399;
}

/* 选中状态样式 */
.el-radio.is-checked .group-radio-item {
  background-color: #ecf5ff;
  border-color: #409eff;
}

/* 修复 el-radio 的布局问题 */
.el-radio {
  margin-right: 0 !important;
  margin-bottom: 10px;
  width: 100%;
  height: auto;
}

.el-radio :deep(.el-radio__label) {
  padding-left: 10px;
  width: calc(100% - 20px);
}

.no-groups-tip {
  text-align: center;
  padding: 20px;
  color: #909399;
  font-size: 14px;
}
</style> 
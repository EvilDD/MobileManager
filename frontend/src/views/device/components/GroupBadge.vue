<script setup lang="ts">
import { inject, computed } from "vue";

const props = defineProps({
  groupId: {
    type: Number,
    required: true
  },
  small: {
    type: Boolean,
    default: false
  }
});

// 固定的基础颜色数组
const baseColors = [
  "#409EFF", // 蓝色
  "#67C23A", // 绿色
  "#E6A23C", // 黄色
  "#F56C6C", // 红色
  "#909399", // 灰色
  "#8E44AD", // 紫色
  "#1ABC9C", // 青绿色
  "#3498DB", // 天蓝色
  "#F39C12", // 橙色
  "#E74C3C"  // 橘红色
];

// 注入分组列表，如果没有则使用空数组
const groupsList = inject("groupsList", computed(() => []));

// 计算当前分组在列表中的序号和颜色
const badgeInfo = computed(() => {
  if (!props.groupId || props.groupId <= 0) {
    return { index: 0, color: "#909399" };
  }

  // 转换为数组以避免可能的类型错误
  const list = Array.isArray(groupsList.value) ? groupsList.value : [];
  
  // 过滤掉id为0的"新设备"分组
  const userGroups = list.filter(group => group && group.id !== 0);
  
  // 查找当前分组在用户分组列表中的位置
  const index = userGroups.findIndex(group => group && group.id === props.groupId);
  // 从1开始计数
  const displayIndex = index >= 0 ? index + 1 : props.groupId;
  
  // 根据显示索引生成颜色
  const colorIndex = (displayIndex - 1) % baseColors.length;
  const color = baseColors[colorIndex >= 0 ? colorIndex : 0];
  
  return { index: displayIndex, color };
});
</script>

<template>
  <div
    v-if="groupId > 0"
    class="group-badge"
    :class="{ small }"
    :style="{ backgroundColor: badgeInfo.color }"
  >
    {{ badgeInfo.index }}
  </div>
</template>

<style scoped>
.group-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 22px;
  height: 22px;
  border-radius: 4px;
  color: white;
  font-size: 12px;
  font-weight: bold;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.15);
  transition:
    transform 0.2s,
    box-shadow 0.2s;
}

.group-badge:hover {
  transform: scale(1.1);
  box-shadow: 0 3px 6px rgba(0, 0, 0, 0.2);
}

.group-badge.small {
  width: 18px;
  height: 18px;
  font-size: 10px;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.1);
}
</style>

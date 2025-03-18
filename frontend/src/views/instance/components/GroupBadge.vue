<script setup lang="ts">
defineProps({
  groupId: {
    type: Number,
    required: true
  },
  small: {
    type: Boolean,
    default: false
  }
});

// 分组对应的颜色
const groupColors = {
  1: "#409EFF", // 蓝色
  2: "#67C23A", // 绿色
  3: "#E6A23C", // 黄色
  4: "#909399"  // 灰色
};
</script>

<template>
  <div 
    class="group-badge" 
    :class="{ small }"
    :style="{ backgroundColor: groupColors[groupId] || '#909399' }"
    v-if="groupId > 0"
  >
    {{
      groupId === 1 ? "一" : groupId === 2 ? "二" : groupId === 3 ? "三" : "未"
    }}
  </div>
</template>

<style scoped>
.group-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  height: 20px;
  border-radius: 4px;
  color: white;
  font-size: 12px;
  font-weight: bold;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.15);
  transition: transform 0.2s, box-shadow 0.2s;
}

.group-badge:hover {
  transform: scale(1.1);
  box-shadow: 0 3px 6px rgba(0, 0, 0, 0.2);
}

.group-badge.small {
  width: 16px;
  height: 16px;
  font-size: 10px;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.1);
}
</style> 
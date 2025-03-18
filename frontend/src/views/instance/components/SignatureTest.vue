<!-- 签名测试组件 -->
<template>
  <el-card class="mb-4">
    <template #header>
      <div class="flex items-center justify-between">
        <span class="font-medium">签名认证测试</span>
      </div>
    </template>
    <div class="flex items-center space-x-4">
      <el-button
        type="primary"
        :loading="loading"
        @click="testSignatureHandler"
      >
        测试签名
      </el-button>
      <div v-if="result" class="flex-1">
        <p class="text-gray-600">测试结果：</p>
        <pre class="bg-gray-100 p-2 rounded mt-2 overflow-auto">{{
          result
        }}</pre>
      </div>
    </div>
  </el-card>
</template>

<script setup lang="ts">
import { ref } from "vue";
import { ElMessage } from "element-plus";
import { testSignature } from "@/api/signature";

const loading = ref(false);
const result = ref("");

const testSignatureHandler = async () => {
  loading.value = true;
  try {
    const response = await testSignature({
      page: 11,
      size: 11
    });
    result.value = JSON.stringify(response, null, 2);
    ElMessage.success("签名测试成功");
  } catch (error) {
    console.error("签名测试失败:", error);
    ElMessage.error("签名测试失败");
    result.value = JSON.stringify(error, null, 2);
  } finally {
    loading.value = false;
  }
};
</script>

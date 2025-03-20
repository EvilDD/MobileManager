<template>
  <div class="device-screenshot">
    <el-card class="screenshot-card">
      <template #header>
        <div class="card-header">
          <span>设备截图</span>
          <div class="header-controls">
            <el-slider
              v-model="quality"
              :min="1"
              :max="100"
              :step="1"
              class="quality-slider"
              :disabled="loading"
            >
              <template #tooltip>
                <div>图片质量: {{ quality }}%</div>
              </template>
            </el-slider>
            <el-button
              type="primary"
              :loading="loading"
              @click="handleCapture"
              :disabled="!selectedDevices.length"
            >
              截图
            </el-button>
          </div>
        </div>
      </template>

      <el-table
        ref="tableRef"
        :data="deviceList"
        style="width: 100%"
        v-loading="loading"
        height="500"
      >
        <el-table-column type="selection" width="55" />
        <el-table-column prop="deviceId" label="设备ID" width="180" />
        <el-table-column prop="name" label="设备名称" width="180" />
        <el-table-column label="截图预览" min-width="200">
          <template #default="{ row }">
            <div class="preview-container">
              <el-image
                v-if="row.imageData"
                :src="row.imageData"
                :preview-src-list="[row.imageData]"
                fit="contain"
                class="preview-image"
              >
                <template #error>
                  <div class="image-error">
                    <el-icon><Picture /></el-icon>
                    <span>暂无截图</span>
                  </div>
                </template>
              </el-image>
              <div v-else class="no-image">
                <el-icon><Picture /></el-icon>
                <span>暂无截图</span>
              </div>
              <el-tag
                v-if="row.error"
                type="danger"
                class="error-tag"
              >
                {{ row.error }}
              </el-tag>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="100" fixed="right">
          <template #default="{ row }">
            <el-button
              v-if="row.imageData"
              type="primary"
              link
              @click="handleDownload(row)"
            >
              下载
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { Picture } from '@element-plus/icons-vue'
import { captureScreenshots } from '@/api/screenshot'
import type { DeviceScreenshot } from '@/api/screenshot'

interface Device {
  deviceId: string
  name: string
  imageData?: string
  error?: string
}

const props = defineProps<{
  devices: Device[]
}>()

const loading = ref(false)
const quality = ref(80)
const tableRef = ref()

// 设备列表数据
const deviceList = ref<Device[]>([])

// 监听设备列表变化
watch(() => props.devices, (newDevices) => {
  deviceList.value = newDevices.map(device => ({
    ...device,
    imageData: undefined,
    error: undefined
  }))
}, { immediate: true })

// 选中的设备
const selectedDevices = computed(() => {
  return tableRef.value?.getSelectionRows() || []
})

// 处理截图
const handleCapture = async () => {
  if (selectedDevices.value.length === 0) {
    ElMessage.warning('请选择需要截图的设备')
    return
  }

  loading.value = true
  try {
    const response = await captureScreenshots({
      deviceIds: selectedDevices.value.map(device => device.deviceId),
      quality: quality.value
    })

    // 更新设备截图数据
    response.screenshots.forEach(screenshot => {
      const device = deviceList.value.find(d => d.deviceId === screenshot.deviceId)
      if (device) {
        device.imageData = screenshot.imageData
        device.error = screenshot.error
      }
    })

    ElMessage.success('截图完成')
  } catch (error: any) {
    ElMessage.error('截图失败：' + (error.message || '未知错误'))
  } finally {
    loading.value = false
  }
}

// 处理下载
const handleDownload = (device: Device) => {
  if (!device.imageData) return
  
  // 创建下载链接
  const link = document.createElement('a')
  link.href = device.imageData
  link.download = `${device.name || device.deviceId}_screenshot.jpg`
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
}
</script>

<style scoped>
.device-screenshot {
  padding: 20px;
}

.screenshot-card {
  width: 100%;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-controls {
  display: flex;
  align-items: center;
  gap: 20px;
}

.quality-slider {
  width: 200px;
}

.preview-container {
  position: relative;
  width: 100%;
  height: 150px;
  display: flex;
  justify-content: center;
  align-items: center;
  background-color: #f5f7fa;
  border-radius: 4px;
}

.preview-image {
  max-height: 150px;
  object-fit: contain;
}

.no-image,
.image-error {
  display: flex;
  flex-direction: column;
  align-items: center;
  color: #909399;
  font-size: 14px;
}

.no-image .el-icon,
.image-error .el-icon {
  font-size: 24px;
  margin-bottom: 8px;
}

.error-tag {
  position: absolute;
  bottom: 8px;
  left: 8px;
}
</style> 
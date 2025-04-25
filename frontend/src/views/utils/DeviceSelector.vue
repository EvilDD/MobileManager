<template>
  <el-dialog v-model="dialogVisible" :title="title" width="500px">
    <el-form ref="deviceFormRef" :model="form" label-width="100px">
      <el-form-item :label="multiSelect ? '选择设备' : '选择设备'" :prop="multiSelect ? 'deviceIds' : 'deviceId'">
        <el-select 
          v-if="multiSelect"
          v-model="form.deviceIds" 
          multiple 
          placeholder="请选择设备" 
          style="width: 100%"
        >
          <el-option
            v-for="device in deviceList"
            :key="device.deviceId"
            :label="device.name"
            :value="device.deviceId"
          >
            <span>{{ device.name }}</span>
            <span style="float: right; color: #8492a6; font-size: 13px">
              {{ device.deviceId }}
            </span>
          </el-option>
        </el-select>
        <el-select 
          v-else
          v-model="form.deviceId" 
          placeholder="请选择设备" 
          style="width: 100%"
        >
          <el-option
            v-for="device in deviceList"
            :key="device.deviceId"
            :label="device.name"
            :value="device.deviceId"
          >
            <span>{{ device.name }}</span>
            <span style="float: right; color: #8492a6; font-size: 13px">
              {{ device.deviceId }}
            </span>
          </el-option>
        </el-select>
      </el-form-item>
      <el-form-item v-if="multiSelect" label="最大并发" prop="maxWorker">
        <el-input-number v-model="form.maxWorker" :min="1" :max="100" />
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
import { getDeviceList, type Device } from '@/api/device';

const props = withDefaults(defineProps<{
  title?: string;
  visible: boolean;
  multiSelect?: boolean; // 是否多选
  maxWorker?: number;
}>(), {
  title: '选择设备',
  multiSelect: true, 
  maxWorker: 50
});

const emit = defineEmits<{
  'update:visible': [value: boolean];
  'confirm': [data: {
    deviceId?: string;
    deviceIds?: string[];
    maxWorker?: number;
  }];
  'cancel': [];
}>();

// 内部状态
const dialogVisible = ref(props.visible);
const deviceList = ref<Device[]>([]);
const loading = ref(false);

// 表单数据
const form = reactive({
  deviceId: '',
  deviceIds: [] as string[],
  maxWorker: props.maxWorker
});

// 监听visible属性变化
watch(() => props.visible, (val) => {
  console.log('DeviceSelector - props.visible变化:', val);
  dialogVisible.value = val;
  if (val) {
    // 打开对话框时获取设备列表
    fetchDeviceList();
    // 重置表单
    form.deviceId = '';
    form.deviceIds = [];
    form.maxWorker = props.maxWorker;
  }
});

// 监听dialogVisible变化
watch(dialogVisible, (val) => {
  console.log('DeviceSelector - dialogVisible变化:', val);
  emit('update:visible', val);
});

// 初始化
onMounted(() => {
  console.log('DeviceSelector - 组件挂载, visible:', props.visible);
  dialogVisible.value = props.visible;
  if (props.visible) {
    fetchDeviceList();
  }
});

// 获取设备列表
const fetchDeviceList = async () => {
  console.log('DeviceSelector - 获取设备列表');
  loading.value = true;
  try {
    const res = await getDeviceList({
      page: 1,
      pageSize: 100, // 获取较多设备
    });
    
    if (res.code === 0) {
      deviceList.value = res.data.list;
      console.log('DeviceSelector - 获取到设备列表:', deviceList.value.length);
    } else {
      ElMessage.error(res.message || "获取设备列表失败");
    }
  } catch (error) {
    console.error("获取设备列表出错:", error);
    ElMessage.error("获取设备列表出错");
  } finally {
    loading.value = false;
  }
};

// 确认选择
const handleConfirm = () => {
  console.log('DeviceSelector - 确认选择');
  if (props.multiSelect && form.deviceIds.length === 0) {
    ElMessage.warning('请选择至少一个设备');
    return;
  }

  if (!props.multiSelect && !form.deviceId) {
    ElMessage.warning('请选择设备');
    return;
  }

  const data = props.multiSelect 
    ? { deviceIds: form.deviceIds, maxWorker: form.maxWorker } 
    : { deviceId: form.deviceId };

  console.log('DeviceSelector - 提交数据:', data);
  emit('confirm', data);
};

// 取消选择
const handleCancel = () => {
  console.log('DeviceSelector - 取消选择');
  dialogVisible.value = false;
  emit('update:visible', false);
  emit('cancel');
};
</script> 
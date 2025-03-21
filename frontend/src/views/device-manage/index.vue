<template>
  <div class="device-manage">
    <el-card class="box-card">
      <template #header>
        <div class="card-header">
          <el-row :gutter="20" class="mb-4">
            <el-col :span="6">
              <el-input
                v-model="searchKeyword"
                placeholder="请输入设备名称搜索"
                clearable
                @keyup.enter="handleSearch"
              >
                <template #prefix>
                  <el-icon><Search /></el-icon>
                </template>
              </el-input>
            </el-col>
            <el-col :span="18" class="text-right">
              <el-button type="primary" @click="handleAdd">
                <el-icon><Plus /></el-icon>新增设备
              </el-button>
            </el-col>
          </el-row>
        </div>
      </template>

      <el-table :data="tableData" v-loading="loading" border style="width: 100%">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="name" label="设备名称" />
        <el-table-column prop="deviceId" label="设备ID" />
        <el-table-column prop="status" label="状态">
          <template #default="{ row }">
            <el-tag :type="row.status === 'online' ? 'success' : 'danger'">
              {{ row.status === 'online' ? '在线' : '离线' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="createdAt" label="创建时间" />
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button-group>
              <el-button type="primary" link @click="handleEdit(row)">
                编辑
              </el-button>
              <el-button type="danger" link @click="handleDelete(row)">
                删除
              </el-button>
            </el-button-group>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-container">
        <el-pagination
          v-model:current-page="currentPage"
          v-model:page-size="pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="total"
          layout="total, sizes, prev, pager, next"
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
        />
      </div>
    </el-card>

    <!-- 新增/编辑设备对话框 -->
    <el-dialog
      :title="dialogTitle"
      v-model="dialogVisible"
      width="500px"
      destroy-on-close
    >
      <el-form
        ref="formRef"
        :model="formData"
        :rules="rules"
        label-width="100px"
        style="max-width: 460px"
      >
        <el-form-item label="设备名称" prop="name">
          <el-input v-model="formData.name" />
        </el-form-item>
        <el-form-item label="设备ID" prop="deviceId">
          <el-input v-model="formData.deviceId" />
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-select v-model="formData.status" style="width: 100%">
            <el-option label="在线" value="online" />
            <el-option label="离线" value="offline" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="dialogVisible = false">取消</el-button>
          <el-button type="primary" @click="handleSubmit">
            确定
          </el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { getDeviceList, saveDevice, updateDevice, deleteDevice } from '@/api/device'
import type { Device } from '@/api/device'
import { Search, Plus } from '@element-plus/icons-vue'

const loading = ref(false)
const searchKeyword = ref('')
const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)
const tableData = ref<Device[]>([])

// 表单相关
const dialogVisible = ref(false)
const dialogTitle = ref('')
const formRef = ref<FormInstance>()
const formData = ref({
  id: undefined as number | undefined,
  name: '',
  deviceId: '',
  status: 'offline'
})

const rules: FormRules = {
  name: [{ required: true, message: '请输入设备名称', trigger: 'blur' }],
  deviceId: [{ required: true, message: '请输入设备ID', trigger: 'blur' }],
  status: [{ required: true, message: '请选择状态', trigger: 'change' }]
}

// 获取设备列表
const fetchData = async () => {
  loading.value = true
  try {
    console.log('正在获取设备列表，参数:', {
      page: currentPage.value,
      pageSize: pageSize.value,
      keyword: searchKeyword.value
    })
    
    const res = await getDeviceList({
      page: currentPage.value,
      pageSize: pageSize.value,
      keyword: searchKeyword.value
    })
    
    console.log('获取设备列表结果:', res)
    
    if (res.code === 0) {
      tableData.value = res.data.list
      total.value = res.data.total
    } else {
      ElMessage.error(res.message || '获取设备列表失败')
    }
  } catch (error) {
    console.error('获取设备列表失败:', error)
    ElMessage.error('获取设备列表失败')
  } finally {
    loading.value = false
  }
}

// 搜索
const handleSearch = () => {
  currentPage.value = 1
  fetchData()
}

// 分页
const handleSizeChange = (val: number) => {
  pageSize.value = val
  fetchData()
}

const handleCurrentChange = (val: number) => {
  currentPage.value = val
  fetchData()
}

// 新增设备
const handleAdd = () => {
  dialogTitle.value = '新增设备'
  formData.value = {
    id: undefined,
    name: '',
    deviceId: '',
    status: 'offline'
  }
  dialogVisible.value = true
}

// 编辑设备
const handleEdit = (row: Device) => {
  dialogTitle.value = '编辑设备'
  formData.value = { ...row }
  dialogVisible.value = true
}

// 删除设备
const handleDelete = (row: Device) => {
  ElMessageBox.confirm('确认删除该设备吗？', '提示', {
    type: 'warning'
  })
    .then(async () => {
      try {
        const res = await deleteDevice(row.id)
        if (res.code === 0) {
          ElMessage.success('删除成功')
          fetchData()
        }
      } catch (error) {
        console.error('删除设备失败:', error)
        ElMessage.error('删除设备失败')
      }
    })
    .catch(() => {})
}

// 提交表单
const handleSubmit = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (valid) {
      try {
        const submitData = { ...formData.value }
        let res;
        
        if (formData.value.id) {
          res = await updateDevice({
            id: formData.value.id,
            name: formData.value.name,
            deviceId: formData.value.deviceId,
            status: formData.value.status
          });
        } else {
          res = await saveDevice({
            name: formData.value.name,
            deviceId: formData.value.deviceId,
            status: formData.value.status
          });
        }
        
        if (res.code === 0) {
          ElMessage.success(formData.value.id ? '更新成功' : '创建成功')
          dialogVisible.value = false
          fetchData()
        } else {
          ElMessage.error(res.message || '保存失败')
        }
      } catch (error) {
        console.error('保存设备失败:', error)
        ElMessage.error('保存设备失败')
      }
    }
  })
}

onMounted(() => {
  fetchData()
})
</script>

<style scoped>
.device-manage {
  padding: 20px;
}

.text-right {
  text-align: right;
}

.mb-4 {
  margin-bottom: 16px;
}

.pagination-container {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}
</style> 
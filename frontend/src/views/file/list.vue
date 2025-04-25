<template>
  <div class="main-container">
    <el-card class="box-card">
      <template #header>
        <div class="card-header">
          <span>文件列表</span>
          <div>
            <el-button type="success" @click="(e: MouseEvent) => openUploadDialog()">上传文件</el-button>
          </div>
        </div>
      </template>
      <div class="search-bar">
        <el-input v-model="searchKeyword" placeholder="搜索原始文件名" class="search-input" />
        <el-button type="primary" @click="handleSearch">搜索</el-button>
        <el-button @click="resetSearch">重置</el-button>
      </div>
      <div class="card-content">
        <el-table :data="fileList" style="width: 100%" v-loading="loading">
          <el-table-column prop="fileName" label="文件名称" width="160" />
          <el-table-column prop="originalName" label="原始文件名" width="160" />
          <el-table-column label="大小" width="100">
            <template #default="{ row }">
              {{ formatSize(row.fileSize) }}
            </template>
          </el-table-column>
          <el-table-column prop="fileType" label="类型" width="120">
            <template #default="{ row }">
              <el-tag :type="getTagType(row.fileType)">
                {{ row.fileType || '未知' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作">
            <template #default="{ row }">
              <div class="operation-buttons">
                <el-button type="danger" size="small" @click="handleDeleteFile(row)">删除</el-button>
                <el-dropdown>
                  <el-button type="primary" size="small">
                    设备操作<el-icon class="el-icon--right"><arrow-down /></el-icon>
                  </el-button>
                  <template #dropdown>
                    <el-dropdown-menu>
                      <el-dropdown-item @click="openDeviceDialog(row)">推送到设备</el-dropdown-item>
                    </el-dropdown-menu>
                  </template>
                </el-dropdown>
              </div>
            </template>
          </el-table-column>
        </el-table>
        <div class="pagination-container">
          <el-pagination
            background
            layout="total, sizes, prev, pager, next"
            :total="total"
            :current-page="currentPage"
            :page-size="pageSize"
            :page-sizes="[10, 20, 50, 100]"
            @current-change="handlePageChange"
            @size-change="handleSizeChange"
          />
        </div>
      </div>
    </el-card>

    <!-- 上传文件对话框 -->
    <el-dialog v-model="uploadDialogVisible" title="上传文件" width="500px">
      <el-form ref="uploadFormRef" :model="uploadForm" label-width="100px">
        <!-- 文件上传 -->
        <el-form-item label="选择文件">
          <el-upload
            class="upload-demo"
            drag
            action="#"
            :auto-upload="false"
            :on-change="handleFileChange"
            :file-list="fileList"
          >
            <el-icon class="el-icon--upload"><upload-filled /></el-icon>
            <div class="el-upload__text">
              拖拽文件到此处或 <em>点击上传</em>
            </div>
          </el-upload>
        </el-form-item>

        <!-- 文件信息显示 -->
        <template v-if="uploadFiles.length > 0">
          <div class="file-info">
            <p><strong>文件名：</strong>{{ uploadFiles[0].name }}</p>
            <p><strong>大小：</strong>{{ formatSize(uploadFiles[0].size || 0) }}</p>
          </div>
          
          <!-- 上传进度条 -->
          <el-progress 
            v-if="uploading" 
            :percentage="uploadProgress" 
            :status="uploadProgress === 100 ? 'success' : 'exception'"
          />
        </template>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="cancelUpload">取消</el-button>
          <el-button type="primary" @click="submitUpload" :loading="uploading">
            上传
          </el-button>
        </span>
      </template>
    </el-dialog>

    <!-- 设备选择对话框 -->
    <el-dialog v-model="deviceDialogVisible" title="选择推送设备" width="500px">
      <el-form ref="deviceFormRef" :model="deviceForm" label-width="100px">
        <el-form-item label="选择设备" prop="deviceIds">
          <el-select 
            v-model="deviceForm.deviceIds" 
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
        </el-form-item>
        <el-form-item label="并发数" prop="maxWorker">
          <el-slider
            v-model="deviceForm.maxWorker"
            :min="1"
            :max="10"
            :step="1"
            show-stops
            show-input
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="deviceDialogVisible = false">取消</el-button>
          <el-button type="primary" @click="submitPushToDevice" :loading="actionLoading">
            确认
          </el-button>
        </span>
      </template>
    </el-dialog>

    <!-- 批量任务进度对话框 -->
    <el-dialog v-model="taskDialogVisible" title="任务进度" width="600px">
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
              :status="taskStatus.status === 'complete' ? 'success' : 'exception'"
            />
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
          <el-button @click="taskDialogVisible = false">关闭</el-button>
          <el-button type="primary" @click="refreshTaskStatus" :loading="taskLoading">
            刷新
          </el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive } from "vue";
import { ArrowDown, UploadFilled } from "@element-plus/icons-vue";
import { ElMessage, ElMessageBox, UploadFile } from "element-plus";
import { getFileList, deleteFile, uploadFile, batchPushByDevices, getBatchTaskStatus, type File, type BatchTaskStatus } from "@/api/file";
import { getDeviceList, type Device } from "@/api/device";

// 文件列表数据
const fileList = ref<File[]>([]);
const total = ref(0);
const currentPage = ref(1);
const pageSize = ref(10);
const loading = ref(false);
const searchKeyword = ref("");

// 上传对话框
const uploadDialogVisible = ref(false);
const uploadForm = reactive({
  file: null as File | null,
});
const uploadFiles = ref<UploadFile[]>([]);
const uploading = ref(false);
const uploadProgress = ref(0);

// 设备选择对话框
const deviceDialogVisible = ref(false);
const deviceList = ref<Device[]>([]);
const deviceForm = reactive({
  fileId: 0,
  deviceIds: [] as string[],
  maxWorker: 3,
});
const actionLoading = ref(false);

// 批量任务对话框
const taskDialogVisible = ref(false);
const taskStatus = ref<BatchTaskStatus | null>(null);
const currentTaskId = ref("");
const taskLoading = ref(false);

// 获取文件列表
const fetchFileList = async () => {
  loading.value = true;
  try {
    const res = await getFileList({
      page: currentPage.value,
      pageSize: pageSize.value,
      originalName: searchKeyword.value
    });
    
    if (res.code === 0) {
      fileList.value = res.data.list;
      total.value = res.data.total;
    } else {
      ElMessage.error(res.message || "获取文件列表失败");
    }
  } catch (error) {
    console.error("获取文件列表出错:", error);
    ElMessage.error("获取文件列表出错");
  } finally {
    loading.value = false;
  }
};

// 处理搜索
const handleSearch = () => {
  currentPage.value = 1;
  fetchFileList();
};

// 重置搜索
const resetSearch = () => {
  searchKeyword.value = "";
  currentPage.value = 1;
  fetchFileList();
};

// 处理分页
const handlePageChange = (page: number) => {
  currentPage.value = page;
  fetchFileList();
};

// 处理每页数量变化
const handleSizeChange = (size: number) => {
  pageSize.value = size;
  currentPage.value = 1;
  fetchFileList();
};

// 打开上传对话框
const openUploadDialog = () => {
  uploadDialogVisible.value = true;
  uploadFiles.value = [];
  uploadProgress.value = 0;
};

// 处理文件选择变化
const handleFileChange = (file: UploadFile) => {
  uploadFiles.value = [file];
};

// 取消上传
const cancelUpload = () => {
  uploadDialogVisible.value = false;
  uploadFiles.value = [];
};

// 提交上传
const submitUpload = async () => {
  if (uploadFiles.value.length === 0) {
    ElMessage.warning("请选择要上传的文件");
    return;
  }

  uploading.value = true;
  uploadProgress.value = 0;

  try {
    // 模拟上传进度
    const progressInterval = setInterval(() => {
      if (uploadProgress.value < 90) {
        uploadProgress.value += 10;
      }
    }, 300);

    const file = uploadFiles.value[0].raw;
    if (!file) {
      throw new Error("文件对象无效");
    }

    const res = await uploadFile(file);
    
    clearInterval(progressInterval);
    uploadProgress.value = 100;
    
    if (res.code === 0) {
      ElMessage.success("文件上传成功");
      uploadDialogVisible.value = false;
      fetchFileList(); // 刷新文件列表
    } else {
      ElMessage.error(res.message || "文件上传失败");
    }
  } catch (error) {
    console.error("文件上传出错:", error);
    ElMessage.error("文件上传出错");
  } finally {
    uploading.value = false;
  }
};

// 删除文件
const handleDeleteFile = (file: File) => {
  ElMessageBox.confirm(
    `确定要删除文件 "${file.fileName}" 吗？`,
    "警告",
    {
      confirmButtonText: "确定",
      cancelButtonText: "取消",
      type: "warning",
    }
  ).then(async () => {
    try {
      const res = await deleteFile(file.fileId);
      
      if (res.code === 0) {
        ElMessage.success("文件删除成功");
        fetchFileList(); // 刷新文件列表
      } else {
        ElMessage.error(res.message || "文件删除失败");
      }
    } catch (error) {
      console.error("删除文件出错:", error);
      ElMessage.error("删除文件出错");
    }
  }).catch(() => {
    // 取消删除操作
  });
};

// 打开设备选择对话框
const openDeviceDialog = (file: File) => {
  deviceForm.fileId = file.fileId;
  deviceForm.deviceIds = [];
  deviceDialogVisible.value = true;
  fetchDeviceList();
};

// 获取设备列表
const fetchDeviceList = async () => {
  try {
    const res = await getDeviceList({
      page: 1,
      pageSize: 100, // 获取较多设备
    });
    
    if (res.code === 0) {
      deviceList.value = res.data.list;
    } else {
      ElMessage.error(res.message || "获取设备列表失败");
    }
  } catch (error) {
    console.error("获取设备列表出错:", error);
    ElMessage.error("获取设备列表出错");
  }
};

// 提交推送到设备
const submitPushToDevice = async () => {
  if (deviceForm.deviceIds.length === 0) {
    ElMessage.warning("请选择至少一个设备");
    return;
  }

  actionLoading.value = true;
  
  try {
    const res = await batchPushByDevices({
      fileId: deviceForm.fileId,
      deviceIds: deviceForm.deviceIds,
      maxWorker: deviceForm.maxWorker,
    });
    
    if (res.code === 0) {
      ElMessage.success("推送任务已创建");
      deviceDialogVisible.value = false;
      
      // 显示任务进度对话框
      currentTaskId.value = res.data.taskId;
      await fetchTaskStatus();
      taskDialogVisible.value = true;
      
      // 定时刷新任务状态
      const statusInterval = setInterval(async () => {
        await fetchTaskStatus();
        if (taskStatus.value?.status === "complete" || taskStatus.value?.status === "failed") {
          clearInterval(statusInterval);
        }
      }, 2000);
    } else {
      ElMessage.error(res.message || "推送任务创建失败");
    }
  } catch (error) {
    console.error("推送任务创建出错:", error);
    ElMessage.error("推送任务创建出错");
  } finally {
    actionLoading.value = false;
  }
};

// 获取任务状态
const fetchTaskStatus = async () => {
  if (!currentTaskId.value) return;
  
  taskLoading.value = true;
  
  try {
    const res = await getBatchTaskStatus(currentTaskId.value);
    
    if (res.code === 0) {
      taskStatus.value = res.data;
    } else {
      ElMessage.error(res.message || "获取任务状态失败");
    }
  } catch (error) {
    console.error("获取任务状态出错:", error);
    ElMessage.error("获取任务状态出错");
  } finally {
    taskLoading.value = false;
  }
};

// 刷新任务状态
const refreshTaskStatus = () => {
  fetchTaskStatus();
};

// 格式化文件大小
const formatSize = (size: number) => {
  if (size < 1024) {
    return size + 'B';
  } else if (size < 1024 * 1024) {
    return (size / 1024).toFixed(2) + 'KB';
  } else if (size < 1024 * 1024 * 1024) {
    return (size / (1024 * 1024)).toFixed(2) + 'MB';
  } else {
    return (size / (1024 * 1024 * 1024)).toFixed(2) + 'GB';
  }
};

// 获取标签类型
const getTagType = (fileType: string) => {
  switch (fileType) {
    case 'image':
      return 'success';
    case 'document':
      return 'primary';
    case 'video':
      return 'warning';
    case 'audio':
      return 'info';
    case 'archive':
      return 'danger';
    case 'app':
      return 'warning';
    default:
      return '';
  }
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

// 页面初始化
onMounted(() => {
  fetchFileList();
});
</script>

<style scoped>
.main-container {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.search-bar {
  display: flex;
  align-items: center;
  margin-bottom: 20px;
  gap: 10px;
}

.search-input {
  width: 200px;
}

.filter-select {
  width: 150px;
}

.operation-buttons {
  display: flex;
  gap: 8px;
}

.pagination-container {
  margin-top: 20px;
  text-align: right;
}

.file-info {
  margin: 15px 0;
  padding: 10px;
  background-color: #f5f7fa;
  border-radius: 4px;
}

.upload-demo {
  width: 100%;
}
</style> 
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
            :file-list="uploadFiles"
            :on-remove="() => { uploadFiles.value = []; }"
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
    <DeviceSelector
      v-model:visible="deviceDialogVisible"
      title="选择推送设备" 
      :multi-select="true"
      @confirm="handleDeviceConfirm"
    />

    <!-- 批量任务进度对话框 -->
    <TaskProgressDialog
      v-model:visible="taskDialogVisible"
      :taskId="currentTaskId"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive } from "vue";
import { ArrowDown, UploadFilled } from "@element-plus/icons-vue";
import { ElMessage, ElMessageBox, UploadFile } from "element-plus";
import { getFileList, deleteFile, uploadFile, type File, type BatchTaskStatus } from "@/api/file";
import DeviceSelector from '@/views/utils/DeviceSelector.vue';
import TaskProgressDialog from '@/views/utils/TaskProgressDialog.vue';
import { pushFileToDevices } from '@/views/utils/DevicePushService';

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
const currentFileId = ref<number>(0);
const actionLoading = ref(false);

// 批量任务对话框
const taskDialogVisible = ref(false);
const currentTaskId = ref("");

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
  uploading.value = false;
};

// 处理文件选择变化
const handleFileChange = (file: UploadFile) => {
  uploadFiles.value = [file];
};

// 取消上传
const cancelUpload = () => {
  uploadDialogVisible.value = false;
  uploadFiles.value = [];
  uploading.value = false;
  uploadProgress.value = 0;
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
      uploadFiles.value = []; // 清除文件列表
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
  console.log('打开设备选择对话框', file);
  if (!file || !file.fileId) {
    console.error('文件信息无效', file);
    ElMessage.error('文件信息无效');
    return;
  }
  
  currentFileId.value = file.fileId;
  // 确保先设置好ID再显示对话框
  setTimeout(() => {
    deviceDialogVisible.value = true;
    console.log('设置对话框可见性：', deviceDialogVisible.value);
  }, 0);
};

// 处理设备选择确认
const handleDeviceConfirm = (data: { deviceIds?: string[], maxWorker?: number }) => {
  console.log('处理设备选择确认:', data);
  if (!data.deviceIds || data.deviceIds.length === 0) {
    ElMessage.warning("请选择至少一个设备");
    return;
  }
  
  actionLoading.value = true;
  console.log('开始推送文件, 文件ID:', currentFileId.value);
  
  pushFileToDevices(
    currentFileId.value, 
    data.deviceIds,
    data.maxWorker || 50
  ).then(taskId => {
    console.log('推送任务创建成功, 任务ID:', taskId);
    deviceDialogVisible.value = false;
    
    // 显示任务进度对话框
    currentTaskId.value = taskId;
    setTimeout(() => {
      taskDialogVisible.value = true;
      console.log('显示任务进度对话框:', taskDialogVisible.value);
    }, 0);
  }).catch(error => {
    console.error("推送任务创建出错:", error);
  }).finally(() => {
    actionLoading.value = false;
  });
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
const getTagType = (fileType: string): 'primary' | 'success' | 'info' | 'warning' | 'danger' => {
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
      return 'info';
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

.progress-text {
  margin-top: 5px;
  font-size: 14px;
  color: #606266;
}
</style> 
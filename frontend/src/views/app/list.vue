<template>
  <div class="main-container">
    <el-card class="box-card">
      <template #header>
        <div class="card-header">
          <span>应用列表</span>
          <div>
            <!-- <el-button type="primary" @click="openAppDialog(null)">添加应用</el-button> -->
            <el-button type="success" @click="(e: MouseEvent) => openUploadDialog()">导入应用</el-button>
          </div>
        </div>
      </template>
      <div class="search-bar">
        <el-input v-model="searchKeyword" placeholder="搜索应用名称" class="search-input" />
        <el-select v-model="appType" placeholder="应用类型" class="filter-select">
          <el-option label="全部" value="" />
          <el-option label="系统应用" :value="AppTypeSystem" />
          <el-option label="用户应用" :value="AppTypeUser" />
          <el-option label="系统设置" :value="AppTypeSettings" />
        </el-select>
        <el-button type="primary" @click="handleSearch">搜索</el-button>
        <el-button @click="resetSearch">重置</el-button>
      </div>
      <div class="card-content">
        <el-table :data="appList" style="width: 100%" v-loading="loading">
          <el-table-column prop="name" label="应用名称" width="180" />
          <el-table-column prop="packageName" label="包名" width="220" />
          <el-table-column prop="version" label="版本" width="120" />
          <el-table-column label="大小" width="120">
            <template #default="{ row }">
              {{ formatSize(row.size) }}
            </template>
          </el-table-column>
          <el-table-column prop="appType" label="类型" width="120">
            <template #default="{ row }">
              <el-tag :type="getTagType(row.appType)">
                {{ row.appType }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作">
            <template #default="{ row }">
              <div class="operation-buttons">
                <el-button type="danger" size="small" @click="handleDeleteApp(row)">删除</el-button>
                <el-dropdown>
                  <el-button type="primary" size="small">
                    设备操作<el-icon class="el-icon--right"><arrow-down /></el-icon>
                  </el-button>
                  <template #dropdown>
                    <el-dropdown-menu>
                      <el-dropdown-item @click="openDeviceDialog('install', row)">安装到设备</el-dropdown-item>
                      <el-dropdown-item @click="openDeviceDialog('uninstall', row)">从设备卸载</el-dropdown-item>
                      <el-dropdown-item @click="openDeviceDialog('start', row)">在设备上启动</el-dropdown-item>
                    </el-dropdown-menu>
                  </template>
                </el-dropdown>
                <el-dropdown>
                  <el-button type="success" size="small">
                    分组操作<el-icon class="el-icon--right"><arrow-down /></el-icon>
                  </el-button>
                  <template #dropdown>
                    <el-dropdown-menu>
                      <el-dropdown-item @click="openGroupDialog('install', row)">批量安装到分组</el-dropdown-item>
                      <el-dropdown-item @click="openGroupDialog('uninstall', row)">批量从分组卸载</el-dropdown-item>
                      <el-dropdown-item @click="openGroupDialog('start', row)">批量在分组启动</el-dropdown-item>
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

    <!-- 添加/编辑应用对话框 -->
    <el-dialog
      v-model="appDialogVisible"
      :title="dialogType === 'create' ? '添加应用' : '编辑应用'"
      width="500px"
    >
      <el-form ref="appFormRef" :model="appForm" label-width="100px" :rules="appFormRules">
        <el-form-item label="应用名称" prop="name">
          <el-input v-model="appForm.name" placeholder="请输入应用名称" />
        </el-form-item>
        <el-form-item label="包名" prop="packageName">
          <el-input v-model="appForm.packageName" placeholder="请输入包名" />
        </el-form-item>
        <el-form-item label="版本" prop="version">
          <el-input v-model="appForm.version" placeholder="请输入版本号" />
        </el-form-item>
        <el-form-item label="应用类型" prop="appType">
          <el-select v-model="appForm.appType" placeholder="请选择应用类型" style="width: 100%">
            <el-option :label="AppTypeSystem" :value="AppTypeSystem" />
            <el-option :label="AppTypeUser" :value="AppTypeUser" />
            <el-option :label="AppTypeSettings" :value="AppTypeSettings" />
          </el-select>
        </el-form-item>
        <el-form-item label="APK路径" prop="apkPath">
          <el-input v-model="appForm.apkPath" placeholder="请上传APK文件获取路径" disabled>
            <template #append>
              <el-button @click="openUploadDialog(true)">上传</el-button>
            </template>
          </el-input>
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="appDialogVisible = false">取消</el-button>
          <el-button type="primary" @click="submitAppForm" :loading="submitting">
            确认
          </el-button>
        </span>
      </template>
    </el-dialog>

    <!-- 上传APK对话框 -->
    <el-dialog v-model="uploadDialogVisible" title="上传APK文件" width="500px">
      <el-form ref="uploadFormRef" :model="uploadForm" label-width="100px">
        <!-- APK文件上传 -->
        <el-form-item label="APK文件">
          <el-upload
            class="upload-demo"
            drag
            action="#"
            :auto-upload="false"
            :on-change="handleFileChange"
            :file-list="fileList"
            accept=".apk"
          >
            <el-icon class="el-icon--upload"><upload-filled /></el-icon>
            <div class="el-upload__text">
              拖拽文件到此处或 <em>点击上传</em>
            </div>
            <template #tip>
              <div class="el-upload__tip">只能上传APK文件</div>
            </template>
          </el-upload>
        </el-form-item>

        <!-- 文件信息显示 -->
        <template v-if="fileList.length > 0">
          <div class="file-info">
            <p><strong>文件名：</strong>{{ fileList[0].name }}</p>
            <p><strong>大小：</strong>{{ formatSize(fileList[0].size || 0) }}</p>
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
      :title="getDeviceDialogTitle"
      :multi-select="true"
      @confirm="handleDeviceConfirm"
    />

    <!-- 分组选择对话框 -->
    <GroupSelector
      v-model:visible="groupDialogVisible"
      :title="getGroupDialogTitle"
      @confirm="handleGroupConfirm"
    />

    <!-- 任务状态对话框 -->
    <TaskProgressDialog
      v-model:visible="taskStatusDialogVisible"
      :taskId="currentTaskId"
      taskType="app"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from "vue";
import { ElMessage, ElMessageBox, UploadUserFile } from "element-plus";
import { ArrowDown, UploadFilled } from "@element-plus/icons-vue";
import { 
  getAppList,
  createApp,
  deleteApp,
  uploadApk,
  installApp,
  uninstallApp,
  startApp,
  batchInstallApp,
  batchUninstallApp,
  batchStartApp,
  getBatchTaskStatus,
  type App,
  type BatchTaskStatus
} from "@/api/app";
import { getDeviceList, type Device } from "@/api/device";
import DeviceSelector from '@/views/utils/DeviceSelector.vue';
import GroupSelector from '@/views/utils/GroupSelector.vue';
import TaskProgressDialog from '@/views/utils/TaskProgressDialog.vue';
import { 
  uploadAppFile, 
  performDeviceAppAction, 
  performGroupAppAction,
  formatSize,
  getTaskStatusType,
  getTaskStatusText,
  getProgressStatus
} from '@/views/utils/AppService';

// 应用类型常量
const AppTypeSystem = "系统应用";
const AppTypeUser = "用户应用";
const AppTypeSettings = "系统设置";

// 状态数据
const appList = ref<App[]>([]);
const loading = ref(false);
const total = ref(0);
const currentPage = ref(1);
const pageSize = ref(50);
const searchKeyword = ref('');
const appType = ref('');

// 设备列表
const deviceList = ref<Device[]>([]);

// 对话框状态
const appDialogVisible = ref(false);
const uploadDialogVisible = ref(false);
const deviceDialogVisible = ref(false);
const dialogType = ref<'create'>('create');
const uploading = ref(false);

// 设备操作类型
const deviceActionType = ref<'install' | 'uninstall' | 'start'>('install');
const currentApp = ref<App | null>(null);

// 表单数据
const appFormRef = ref<FormInstance>();
const appForm = ref({
  name: '',
  packageName: '',
  version: '',
  size: 0,
  appType: AppTypeUser,
  apkPath: ''
});

// 设备表单
const deviceFormRef = ref<FormInstance>();
const deviceForm = ref({
  deviceId: ''
});

// 上传文件列表
const fileList = ref<UploadUserFile[]>([]);

// 表单校验规则
const appFormRules = {
  name: [{ required: true, message: '请输入应用名称', trigger: 'blur' }],
  packageName: [{ required: true, message: '请输入包名', trigger: 'blur' }],
  version: [{ required: true, message: '请输入版本号', trigger: 'blur' }],
  appType: [{ required: true, message: '请选择应用类型', trigger: 'change' }],
  apkPath: [{ required: true, message: '请上传APK文件', trigger: 'blur' }]
};

// 上传表单数据
const uploadFormRef = ref<FormInstance>();
const uploadForm = ref({
  file: null as File | null
});

// 上传表单校验规则
const uploadFormRules = {
  name: [{ required: true, message: '请输入应用名称', trigger: 'blur' }],
  packageName: [{ required: true, message: '请输入包名', trigger: 'blur' }],
  version: [{ required: true, message: '请输入版本号', trigger: 'blur' }],
  appType: [{ required: true, message: '请选择应用类型', trigger: 'change' }],
  file: [{ required: true, message: '请选择APK文件', trigger: 'change' }]
};

// 上传进度
const uploadProgress = ref(0);

// 设备对话框标题
const getDeviceDialogTitle = computed(() => {
  switch (deviceActionType.value) {
    case 'install':
      return '安装到设备';
    case 'uninstall':
      return '从设备卸载';
    case 'start':
      return '在设备上启动';
    default:
      return '选择设备';
  }
});

// 获取标签类型
const getTagType = (type: string): 'success' | 'warning' | 'info' | 'primary' | 'danger' => {
  switch (type) {
    case AppTypeSystem:
      return 'info';
    case AppTypeUser:
      return 'success';
    case AppTypeSettings:
      return 'warning';
    default:
      return 'info';
  }
};

// 分组相关
const groupDialogVisible = ref(false);
const groupActionType = ref<'install' | 'uninstall' | 'start'>('install');
const groupActionLoading = ref(false);

// 任务状态相关
const taskStatusDialogVisible = ref(false);
const currentTaskId = ref<string>('');

// 初始化数据
onMounted(() => {
  fetchAppList();
  fetchDeviceList();
});

// 获取应用列表
const fetchAppList = () => {
  loading.value = true;
  getAppList({
    page: currentPage.value,
    pageSize: pageSize.value,
    keyword: searchKeyword.value || undefined,
    appType: appType.value || undefined
  }).then(res => {
    if (res.code === 0) {
      appList.value = res.data.list;
      total.value = res.data.total;
    } else {
      ElMessage.error(res.message || '获取应用列表失败');
    }
  }).catch(err => {
    console.error('获取应用列表出错:', err);
    ElMessage.error('获取应用列表出错');
  }).finally(() => {
    loading.value = false;
  });
};

// 获取设备列表
const fetchDeviceList = () => {
  getDeviceList({
    page: 1,
    pageSize: 100
  }).then(res => {
    if (res.code === 0) {
      deviceList.value = res.data.list;
    }
  }).catch(err => {
    console.error('获取设备列表出错:', err);
  });
};

// 搜索处理
const handleSearch = () => {
  currentPage.value = 1;
  fetchAppList();
};

// 重置搜索
const resetSearch = () => {
  searchKeyword.value = '';
  appType.value = '';
  handleSearch();
};

// 分页处理
const handlePageChange = (page: number) => {
  currentPage.value = page;
  fetchAppList();
};

// 每页条数变化
const handleSizeChange = (size: number) => {
  pageSize.value = size;
  currentPage.value = 1;
  fetchAppList();
};

// 打开应用对话框
const openAppDialog = (app: App | null) => {
  dialogType.value = 'create';
  resetAppForm();
  if (app) {
    // 如果后续需要编辑功能，在这里设置表单数据
  }
  appDialogVisible.value = true;
};

// 重置应用表单
const resetAppForm = () => {
  appForm.value = {
    name: '',
    packageName: '',
    version: '',
    size: 0,
    appType: AppTypeUser,
    apkPath: ''
  };
  if (appFormRef.value) {
    appFormRef.value.resetFields();
  }
};

// 打开上传对话框
const openUploadDialog = (forApp = false) => {
  uploadDialogVisible.value = true;
  fileList.value = [];
  uploadProgress.value = 0;
};

// 处理文件变化
const handleFileChange = (file: UploadUserFile) => {
  console.log('文件变化:', file);
  fileList.value = [file];
};

// 取消上传
const cancelUpload = () => {
  uploadDialogVisible.value = false;
  fileList.value = [];
  uploadProgress.value = 0;
  uploading.value = false;
};

// 提交上传
const submitUpload = async () => {
  if (fileList.value.length === 0) {
    ElMessage.warning('请选择要上传的APK文件');
    return;
  }

  const file = fileList.value[0].raw;
  if (!file) {
    ElMessage.warning('文件数据无效');
    return;
  }

  uploading.value = true;
  uploadProgress.value = 0;

  // 模拟上传进度
  const progressInterval = setInterval(() => {
    if (uploadProgress.value < 90) {
      uploadProgress.value += 10;
    }
  }, 300);

  try {
    await uploadAppFile(file);
    uploadProgress.value = 100;
    fetchAppList();
    setTimeout(() => {
      cancelUpload();
    }, 1000);
  } catch (error) {
    console.error('上传出错:', error);
  } finally {
    clearInterval(progressInterval);
    uploading.value = false;
  }
};

// 提交应用表单
const submitAppForm = () => {
  if (!appFormRef.value) return;
  
  appFormRef.value.validate((valid) => {
    if (valid) {
      submitting.value = true;
      
      createApp({
        name: appForm.value.name,
        packageName: appForm.value.packageName,
        version: appForm.value.version,
        size: appForm.value.size,
        appType: appForm.value.appType,
        apkPath: appForm.value.apkPath
      }).then(res => {
        if (res.code === 0) {
          ElMessage.success('应用创建成功');
          appDialogVisible.value = false;
          fetchAppList();
        } else {
          ElMessage.error(res.message || '创建失败');
        }
      }).catch(err => {
        console.error('创建应用出错:', err);
        ElMessage.error('创建应用出错');
      }).finally(() => {
        submitting.value = false;
      });
    }
  });
};

// 删除应用
const handleDeleteApp = (row: App) => {
  ElMessageBox.confirm(
    `确定要删除应用 "${row.name}" 吗？`,
    '删除确认',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    }
  ).then(() => {
    deleteApp(row.id).then(res => {
      if (res.code === 0) {
        ElMessage.success('删除成功');
        fetchAppList();
      } else {
        ElMessage.error(res.message || '删除失败');
      }
    }).catch(err => {
      console.error('删除应用出错:', err);
      ElMessage.error('删除应用出错');
    });
  }).catch(() => {});
};

// 打开设备选择对话框
const openDeviceDialog = (action: 'install' | 'uninstall' | 'start', app: App) => {
  deviceActionType.value = action;
  currentApp.value = app;
  deviceForm.value.deviceId = '';
  deviceDialogVisible.value = true;
};

// 处理设备选择确认
const handleDeviceConfirm = async (data: { deviceId?: string; deviceIds?: string[], maxWorker?: number }) => {
  if (!currentApp.value || (!data.deviceId && (!data.deviceIds || data.deviceIds.length === 0))) {
    ElMessage.warning('请选择至少一个设备');
    return;
  }

  // 如果是多选模式
  if (data.deviceIds && data.deviceIds.length > 0) {
    groupActionLoading.value = true;
    try {
      const taskId = await performGroupAppAction(
        deviceActionType.value,
        currentApp.value.id,
        0, // 使用0表示不是通过分组而是通过设备列表
        data.maxWorker || 50,
        data.deviceIds
      );
      
      deviceDialogVisible.value = false;
      
      // 显示任务进度对话框
      currentTaskId.value = taskId;
      setTimeout(() => {
        taskStatusDialogVisible.value = true;
      }, 0);
    } catch (error) {
      console.error('批量设备操作失败:', error);
    } finally {
      groupActionLoading.value = false;
    }
  } 
  // 单选模式 (向后兼容)
  else if (data.deviceId) {
    try {
      await performDeviceAppAction(deviceActionType.value, currentApp.value.id, data.deviceId);
      deviceDialogVisible.value = false;
    } catch (error) {
      console.error('设备操作失败:', error);
    }
  }
};

// 打开分组操作对话框
const openGroupDialog = (action: 'install' | 'uninstall' | 'start', app: App) => {
  groupActionType.value = action;
  currentApp.value = app;
  groupDialogVisible.value = true;
};

// 获取分组对话框标题
const getGroupDialogTitle = computed(() => {
  const actionText = {
    install: '批量安装',
    uninstall: '批量卸载',
    start: '批量启动'
  }[groupActionType.value];
  return `${actionText}应用到分组设备`;
});

// 处理分组选择确认
const handleGroupConfirm = async (data: { groupId: number; maxWorker: number }) => {
  if (!currentApp.value) {
    ElMessage.warning('应用信息无效');
    return;
  }

  groupActionLoading.value = true;
  try {
    const taskId = await performGroupAppAction(
      groupActionType.value,
      currentApp.value.id,
      data.groupId,
      data.maxWorker
    );
    
    groupDialogVisible.value = false;
    
    // 显示任务进度对话框
    currentTaskId.value = taskId;
    setTimeout(() => {
      taskStatusDialogVisible.value = true;
    }, 0);
  } catch (error) {
    console.error('批量操作失败:', error);
  } finally {
    groupActionLoading.value = false;
  }
};
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
  margin-bottom: 20px;
  display: flex;
  gap: 10px;
}

.search-input {
  width: 250px;
}

.filter-select {
  width: 140px;
}

.card-content {
  margin-top: 20px;
}

.pagination-container {
  margin-top: 20px;
  display: flex;
  justify-content: center;
}

.upload-demo {
  width: 100%;
}

.operation-buttons {
  display: flex;
  gap: 8px;
}

.file-info {
  width: 400px;
  margin: 16px auto;
  padding: 16px 24px;
  background-color: var(--el-fill-color-light);
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
}

.file-info p {
  display: flex;
  align-items: center;
  margin: 8px 0;
}

.file-info strong {
  width: 80px;
  color: var(--el-text-color-secondary);
  text-align: right;
  padding-right: 12px;
}

.task-status {
  .el-descriptions {
    margin-bottom: 20px;
  }
}

.task-results {
  max-height: 400px;
  overflow-y: auto;
}

.form-tip {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
  line-height: 1.4;
}
</style> 
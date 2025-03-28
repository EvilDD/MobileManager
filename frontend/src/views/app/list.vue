<template>
  <div class="main-container">
    <el-card class="box-card">
      <template #header>
        <div class="card-header">
          <span>应用列表</span>
          <div>
            <el-button type="primary" @click="openAppDialog(null)">添加应用</el-button>
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
                    操作<el-icon class="el-icon--right"><arrow-down /></el-icon>
                  </el-button>
                  <template #dropdown>
                    <el-dropdown-menu>
                      <el-dropdown-item @click="openDeviceDialog('install', row)">安装到设备</el-dropdown-item>
                      <el-dropdown-item @click="openDeviceDialog('uninstall', row)">从设备卸载</el-dropdown-item>
                      <el-dropdown-item @click="openDeviceDialog('start', row)">在设备上启动</el-dropdown-item>
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
    <el-dialog v-model="deviceDialogVisible" :title="getDeviceDialogTitle" width="500px">
      <el-form ref="deviceFormRef" :model="deviceForm" label-width="100px">
        <el-form-item label="选择设备" prop="deviceId">
          <el-select v-model="deviceForm.deviceId" placeholder="请选择设备" style="width: 100%">
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
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="deviceDialogVisible = false">取消</el-button>
          <el-button type="primary" @click="submitDeviceAction" :loading="actionLoading">
            确认
          </el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from "vue";
import { ElMessage, ElMessageBox, FormInstance, UploadUserFile } from "element-plus";
import { ArrowDown, UploadFilled } from "@element-plus/icons-vue";
import { 
  getAppList,
  createApp,
  deleteApp,
  uploadApk,
  installApp,
  uninstallApp,
  startApp,
  type App
} from "@/api/app";
import { getDeviceList, type Device } from "@/api/device";

// 应用类型常量
const AppTypeSystem = "系统应用";
const AppTypeUser = "用户应用";
const AppTypeSettings = "系统设置";

// 状态数据
const appList = ref<App[]>([]);
const loading = ref(false);
const total = ref(0);
const currentPage = ref(1);
const pageSize = ref(10);
const searchKeyword = ref('');
const appType = ref('');

// 设备列表
const deviceList = ref<Device[]>([]);

// 对话框状态
const appDialogVisible = ref(false);
const uploadDialogVisible = ref(false);
const deviceDialogVisible = ref(false);
const dialogType = ref<'create'>('create');
const uploadForApp = ref(false);
const submitting = ref(false);
const uploading = ref(false);
const actionLoading = ref(false);

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
const getTagType = (type: string) => {
  switch (type) {
    case AppTypeSystem:
      return 'info';
    case AppTypeUser:
      return 'success';
    case AppTypeSettings:
      return 'warning';
    default:
      return '';
  }
};

// 格式化文件大小
const formatSize = (size: number) => {
  if (size < 1024) {
    return size + 'KB';
  } else if (size < 1024 * 1024) {
    return (size / 1024).toFixed(2) + 'MB';
  } else {
    return (size / (1024 * 1024)).toFixed(2) + 'GB';
  }
};

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
const openUploadDialog = (isForApp = false) => {
  uploadForApp.value = isForApp;
  fileList.value = [];
  uploadDialogVisible.value = true;
};

// 处理文件变化
const handleFileChange = (file: UploadUserFile) => {
  fileList.value = [file];
};

// 取消上传
const cancelUpload = () => {
  fileList.value = [];
  uploadDialogVisible.value = false;
};

// 提交上传
const submitUpload = () => {
  if (fileList.value.length === 0) {
    ElMessage.warning('请选择要上传的APK文件');
    return;
  }

  const rawFile = fileList.value[0].raw;
  if (!rawFile) {
    ElMessage.error('获取文件失败');
    return;
  }

  console.log('准备上传文件:', rawFile.name, '类型:', rawFile.type, '大小:', rawFile.size);
  
  uploading.value = true;
  uploadApk(rawFile).then(res => {
    if (res.code === 0) {
      ElMessage.success('上传成功');
      
      // 如果是为应用表单上传，则设置路径
      if (uploadForApp.value) {
        appForm.value.apkPath = res.data.filePath;
        appForm.value.size = res.data.fileSize;
        uploadDialogVisible.value = false;
      } else {
        uploadDialogVisible.value = false;
      }
      
      // 刷新应用列表
      fetchAppList();
    } else {
      ElMessage.error(res.message || '上传失败');
    }
  }).catch(err => {
    console.error('上传出错:', err);
    ElMessage.error('上传出错');
  }).finally(() => {
    uploading.value = false;
  });
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

// 提交设备操作
const submitDeviceAction = () => {
  if (!currentApp.value || !deviceForm.value.deviceId) {
    ElMessage.warning('请选择设备');
    return;
  }

  actionLoading.value = true;
  
  const actionData = {
    id: currentApp.value.id,
    deviceId: deviceForm.value.deviceId
  };

  let actionPromise;
  let actionName = '';

  switch (deviceActionType.value) {
    case 'install':
      actionPromise = installApp(actionData);
      actionName = '安装';
      break;
    case 'uninstall':
      actionPromise = uninstallApp(actionData);
      actionName = '卸载';
      break;
    case 'start':
      actionPromise = startApp(actionData);
      actionName = '启动';
      break;
  }

  actionPromise.then(res => {
    if (res.code === 0) {
      ElMessage.success(`${actionName}成功`);
      deviceDialogVisible.value = false;
    } else {
      ElMessage.error(res.message || `${actionName}失败`);
    }
  }).catch(err => {
    console.error(`${actionName}出错:`, err);
    ElMessage.error(`${actionName}出错`);
  }).finally(() => {
    actionLoading.value = false;
  });
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
  align-items: center;
}
</style> 
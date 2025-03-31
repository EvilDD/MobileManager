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

    <!-- 分组操作对话框 -->
    <el-dialog v-model="groupDialogVisible" :title="getGroupDialogTitle" width="500px">
      <el-form ref="groupFormRef" :model="groupForm" label-width="100px">
        <el-form-item label="选择分组" prop="groupId">
          <el-select v-model="groupForm.groupId" placeholder="请选择分组" style="width: 100%">
            <el-option
              v-for="group in groupList"
              :key="group.id"
              :label="group.name"
              :value="group.id"
            >
              <span>{{ group.name }}</span>
              <span style="float: right; color: #8492a6; font-size: 13px">
                ({{ group.deviceCount }}台设备)
              </span>
            </el-option>
          </el-select>
        </el-form-item>
        <el-form-item label="并发数" prop="maxWorker">
          <el-input-number 
            v-model="groupForm.maxWorker" 
            :min="1" 
            :max="50" 
            placeholder="请输入并发数"
            style="width: 100%"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="groupDialogVisible = false">取消</el-button>
          <el-button type="primary" @click="submitGroupAction" :loading="groupActionLoading">
            确认
          </el-button>
        </span>
      </template>
    </el-dialog>

    <!-- 任务状态对话框 -->
    <el-dialog v-model="taskStatusDialogVisible" title="任务执行状态" width="800px">
      <div v-if="currentTask" class="task-status">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="任务ID">{{ currentTask.taskId }}</el-descriptions-item>
          <el-descriptions-item label="任务状态">
            <el-tag :type="getTaskStatusType(currentTask.status)">
              {{ getTaskStatusText(currentTask.status) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="总设备数">{{ currentTask.total }}</el-descriptions-item>
          <el-descriptions-item label="执行进度">
            <el-progress 
              :percentage="Math.round(((currentTask.completed + currentTask.failed) / currentTask.total) * 100)"
              :status="getProgressStatus(currentTask)"
            >
              <template #default>
                {{ currentTask.completed }}/{{ currentTask.total }}
                <span v-if="currentTask.failed > 0" style="color: #f56c6c">
                  (失败: {{ currentTask.failed }})
                </span>
              </template>
            </el-progress>
          </el-descriptions-item>
        </el-descriptions>

        <div class="task-results" style="margin-top: 20px">
          <el-table :data="currentTask.results" style="width: 100%">
            <el-table-column prop="deviceId" label="设备ID" width="180" />
            <el-table-column label="状态" width="120">
              <template #default="{ row }">
                <el-tag :type="row.status === 'complete' ? 'success' : 'danger'">
                  {{ row.status === 'complete' ? '成功' : '失败' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="message" label="执行结果" />
          </el-table>
        </div>
      </div>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="taskStatusDialogVisible = false">关闭</el-button>
          <el-button 
            type="primary" 
            @click="refreshTaskStatus" 
            :loading="taskStatusLoading"
            v-if="currentTask?.status === 'running' || currentTask?.status === 'pending'"
          >
            刷新
          </el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed, onUnmounted } from "vue";
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
  batchInstallApp,
  batchUninstallApp,
  batchStartApp,
  getBatchTaskStatus,
  type App,
  type BatchTaskStatus
} from "@/api/app";
import { getDeviceList, type Device } from "@/api/device";
import { getGroupList, type GroupItem } from "@/api/group";

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

// 格式化文件大小
const formatSize = (size: number) => {
  const KB = 1024;
  const MB = KB * 1024;
  const GB = MB * 1024;
  
  if (size < KB) {
    return size + 'B';
  } else if (size < MB) {
    return (size / KB).toFixed(2) + 'KB';
  } else if (size < GB) {
    return (size / MB).toFixed(2) + 'MB';
  } else {
    return (size / GB).toFixed(2) + 'GB';
  }
};

// 分组相关
const groupList = ref<GroupItem[]>([]);
const groupDialogVisible = ref(false);
const groupActionType = ref<'install' | 'uninstall' | 'start'>('install');
const groupActionLoading = ref(false);

interface GroupDialogForm {
  groupId: number;
  maxWorker: number;
}

const groupForm = ref<GroupDialogForm>({
  groupId: 0,
  maxWorker: 10
});

// 任务状态相关
const taskStatusDialogVisible = ref(false);
const taskStatusLoading = ref(false);
const currentTask = ref<BatchTaskStatus | null>(null);
const currentTaskId = ref<string>('');
const taskStatusTimer = ref<number | null>(null);

// 初始化数据
onMounted(() => {
  fetchAppList();
  fetchDeviceList();
  fetchGroupList();
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

// 获取分组列表
const fetchGroupList = async () => {
  try {
    const res = await getGroupList({
      page: 1,
      pageSize: 1000 // 获取所有分组
    });
    if (res.code === 0 && res.data) {
      groupList.value = res.data.list || [];
    } else {
      console.error('获取分组列表失败:', res.message);
      ElMessage.error(res.message || '获取分组列表失败');
    }
  } catch (error) {
    console.error('获取分组列表失败:', error);
    ElMessage.error('获取分组列表失败');
  }
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
  uploadForApp.value = forApp;
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

  try {
    const res = await uploadApk(file);

    if (res.code === 0) {
      ElMessage.success('应用导入成功');
      if (uploadForApp.value) {
        appForm.value.apkPath = res.data.filePath;
        uploadForApp.value = false;
      }
      fetchAppList();
      cancelUpload();
    } else {
      ElMessage.error(res.message || '上传失败');
    }
  } catch (error) {
    console.error('上传出错:', error);
    ElMessage.error('上传出错');
  } finally {
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

// 打开分组操作对话框
const openGroupDialog = (action: 'install' | 'uninstall' | 'start', app: App) => {
  groupActionType.value = action;
  currentApp.value = app;
  groupForm.value = {
    groupId: 0,
    maxWorker: 10
  };
  groupDialogVisible.value = true;
  fetchGroupList();
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

// 提交分组操作
const submitGroupAction = async () => {
  if (!currentApp.value || !groupForm.value.groupId) {
    ElMessage.warning('请选择分组');
    return;
  }

  groupActionLoading.value = true;
  try {
    const data = {
      id: currentApp.value.id,
      groupId: groupForm.value.groupId,
      maxWorker: groupForm.value.maxWorker
    };

    let res;
    switch (groupActionType.value) {
      case 'install':
        res = await batchInstallApp(data);
        break;
      case 'uninstall':
        res = await batchUninstallApp(data);
        break;
      case 'start':
        res = await batchStartApp(data);
        break;
    }

    groupDialogVisible.value = false;
    ElMessage.success('批量操作已开始执行');
    
    // 打开任务状态对话框
    currentTaskId.value = res.data.taskId;
    await refreshTaskStatus();
    taskStatusDialogVisible.value = true;
    startTaskStatusPolling();

  } catch (error) {
    console.error('批量操作失败:', error);
    ElMessage.error('批量操作失败');
  } finally {
    groupActionLoading.value = false;
  }
};

// 刷新任务状态
const refreshTaskStatus = async () => {
  if (!currentTaskId.value) return;

  taskStatusLoading.value = true;
  try {
    const res = await getBatchTaskStatus(currentTaskId.value);
    currentTask.value = res.data;

    // 如果任务已完成或失败，停止轮询
    if (res.data.status === 'complete' || res.data.status === 'failed') {
      stopTaskStatusPolling();
    }
  } catch (error) {
    console.error('获取任务状态失败:', error);
    ElMessage.error('获取任务状态失败');
    stopTaskStatusPolling();
  } finally {
    taskStatusLoading.value = false;
  }
};

// 开始任务状态轮询
const startTaskStatusPolling = () => {
  stopTaskStatusPolling(); // 先清除可能存在的定时器
  taskStatusTimer.value = window.setInterval(refreshTaskStatus, 2000);
};

// 停止任务状态轮询
const stopTaskStatusPolling = () => {
  if (taskStatusTimer.value) {
    clearInterval(taskStatusTimer.value);
    taskStatusTimer.value = null;
  }
};

// 获取任务状态类型
const getTaskStatusType = (status: string) => {
  switch (status) {
    case 'complete': return 'success';
    case 'failed': return 'danger';
    case 'running': return 'warning';
    default: return 'info';
  }
};

// 获取任务状态文本
const getTaskStatusText = (status: string) => {
  switch (status) {
    case 'pending': return '等待执行';
    case 'running': return '执行中';
    case 'complete': return '执行完成';
    case 'failed': return '执行失败';
    default: return status;
  }
};

// 获取进度条状态
const getProgressStatus = (task: BatchTaskStatus) => {
  if (task.status === 'failed') return 'exception';
  if (task.status === 'complete') return 'success';
  return '';
};

// 组件卸载时清理
onUnmounted(() => {
  stopTaskStatusPolling();
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
</style> 
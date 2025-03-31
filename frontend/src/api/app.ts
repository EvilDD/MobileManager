import { http } from "@/utils/http";

export interface App {
  id: number;
  name: string;
  packageName: string;
  version: string;
  size: number;
  appType: string;
  apkPath: string;
  createdAt: string;
  updatedAt: string;
}

export interface AppListResult {
  code: number;
  message: string;
  data: {
    list: App[];
    page: number;
    pageSize: number;
    total: number;
  };
}

/** 获取应用列表 */
export function getAppList(params: {
  page: number;
  pageSize: number;
  keyword?: string;
  appType?: string;
}) {
  return http.request<AppListResult>("get", "/api/apps/list", { params });
}

/** 创建应用 */
export function createApp(data: {
  name: string;
  packageName: string;
  version: string;
  size: number;
  appType: string;
  apkPath: string;
}) {
  return http.request<{ code: number; message: string; data: Record<string, unknown> }>(
    "post",
    "/api/apps/create",
    { data }
  );
}

/** 上传APK文件 */
export function uploadApk(file: File) {
  const formData = new FormData();
  formData.append("file", file);

  console.log("准备上传文件:", file.name, "类型:", file.type, "大小:", file.size);

  return http.request<{
    code: number;
    message: string;
    data: {
      fileName: string;
      fileSize: number;
      filePath: string;
    };
  }>("post", "/api/apps/upload", 
  { 
    data: formData,
    headers: {
      // 显式设置为undefined，确保axios不添加默认的Content-Type
      'Content-Type': undefined
    }
  });
}

/** 删除应用 */
export function deleteApp(id: number) {
  return http.request<{ code: number; message: string; data: Record<string, unknown> }>(
    "delete",
    "/api/apps/delete",
    { data: { id } }
  );
}

/** 安装应用到设备 */
export function installApp(data: {
  id: number;
  deviceId: string;
}) {
  return http.request<{ code: number; message: string; data: Record<string, unknown> }>(
    "post",
    "/api/apps/install",
    { data }
  );
}

/** 从设备卸载应用 */
export function uninstallApp(data: {
  id: number;
  deviceId: string;
}) {
  return http.request<{ code: number; message: string; data: Record<string, unknown> }>(
    "post",
    "/api/apps/uninstall",
    { data }
  );
}

/** 启动设备上的应用 */
export function startApp(data: {
  id: number;
  deviceId: string;
}) {
  return http.request<{ code: number; message: string; data: Record<string, unknown> }>(
    "post",
    "/api/apps/start",
    { data }
  );
}

/** 批量操作任务结果 */
export interface BatchTaskResult {
  deviceId: string;
  status: string;
  message: string;
}

/** 批量操作任务状态 */
export interface BatchTaskStatus {
  taskId: string;
  status: string;
  total: number;
  completed: number;
  failed: number;
  results: BatchTaskResult[];
}

/** 批量操作响应 */
export interface BatchOperationResult {
  code: number;
  message: string;
  data: {
    taskId: string;
    total: number;
    deviceIds: string[];
  };
}

/** 批量任务状态响应 */
export interface BatchTaskStatusResult {
  code: number;
  message: string;
  data: BatchTaskStatus;
}

/** 批量安装应用到分组设备 */
export function batchInstallApp(data: {
  id: number;
  groupId: number;
  maxWorker: number;
}) {
  return http.request<BatchOperationResult>(
    "post",
    "/api/apps/batch-install",
    { data }
  );
}

/** 批量从分组设备卸载应用 */
export function batchUninstallApp(data: {
  id: number;
  groupId: number;
  maxWorker: number;
}) {
  return http.request<BatchOperationResult>(
    "post",
    "/api/apps/batch-uninstall",
    { data }
  );
}

/** 批量启动分组设备上的应用 */
export function batchStartApp(data: {
  id: number;
  groupId: number;
  maxWorker: number;
}) {
  return http.request<BatchOperationResult>(
    "post",
    "/api/apps/batch-start",
    { data }
  );
}

/** 查询批量操作任务状态 */
export function getBatchTaskStatus(taskId: string) {
  return http.request<BatchTaskStatusResult>(
    "get",
    "/api/apps/batch-task-status",
    { params: { taskId } }
  );
}

/** 按设备ID批量安装应用 */
export function batchInstallByDevices(data: {
  id: number;
  deviceIds: string[];
  maxWorker: number;
}) {
  return http.request<BatchOperationResult>(
    "post",
    "/api/apps/batch-install-by-devices",
    { data }
  );
}

/** 按设备ID批量卸载应用 */
export function batchUninstallByDevices(data: {
  id: number;
  deviceIds: string[];
  maxWorker: number;
}) {
  return http.request<BatchOperationResult>(
    "post",
    "/api/apps/batch-uninstall-by-devices",
    { data }
  );
}

/** 按设备ID批量启动应用 */
export function batchStartByDevices(data: {
  id: number;
  deviceIds: string[];
  maxWorker: number;
}) {
  return http.request<BatchOperationResult>(
    "post",
    "/api/apps/batch-start-by-devices",
    { data }
  );
}

/** 按设备ID批量停止应用 */
export function batchStopByDevices(data: {
  id: number;
  deviceIds: string[];
  maxWorker: number;
}) {
  return http.request<BatchOperationResult>(
    "post",
    "/api/apps/batch-stop-by-devices",
    { data }
  );
}
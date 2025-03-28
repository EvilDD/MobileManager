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
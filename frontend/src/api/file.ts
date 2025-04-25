import { http } from "@/utils/http";

export interface File {
  fileId: number;
  fileName: string;
  originalName: string;
  fileType: string;
  fileSize: number;
  filePath: string;
  md5: string;
  createdAt: string;
  updatedAt: string;
}

export interface FileListParams {
  page: number;
  pageSize: number;
  fileName?: string;
  fileType?: string;
  originalName?: string;
}

export interface FileListResult {
  code: number;
  message: string;
  data: {
    list: File[];
    total: number;
    page: number;
    pageSize: number;
  }
}

export interface BatchTaskResult {
  deviceId: string;
  status: string;
  message: string;
}

export interface BatchTaskStatus {
  taskId: string;
  status: string;
  total: number;
  completed: number;
  failed: number;
  results: BatchTaskResult[];
}

export interface BatchOperationResult {
  code: number;
  message: string;
  data: {
    taskId: string;
    total: number;
    deviceIds: string[];
  }
}

export interface BatchTaskStatusResult {
  code: number;
  message: string;
  data: BatchTaskStatus;
}

export interface BatchPushParams {
  fileId: number;
  deviceIds: string[];
  maxWorker: number;
}

// 获取文件列表
export function getFileList(params: FileListParams) {
  return http.request<FileListResult>(
    'get',
    '/api/files/list', 
    { params }
  );
}

// 上传文件
export function uploadFile(file: globalThis.File) {
  const formData = new FormData();
  formData.append('file', file);
  
  return http.request<{ code: number; message: string; data: { fileId: number } }>(
    'post',
    '/api/files/upload',
    { 
      data: formData,
      headers: {
        'Content-Type': undefined
      }
    }
  );
}

// 批量推送文件到设备
export function batchPushByDevices(data: BatchPushParams) {
  return http.request<BatchOperationResult>(
    'post',
    '/api/files/batch-push-by-devices',
    { data }
  );
}

// 获取批量任务状态
export function getBatchTaskStatus(taskId: string) {
  return http.request<BatchTaskStatusResult>(
    'get',
    '/api/files/batch-task-status',
    { params: { taskId } }
  );
}

// 删除文件
export function deleteFile(fileId: number) {
  return http.request<{ code: number; message: string; data: null }>(
    'delete',
    '/api/files/delete',
    { data: { fileId } }
  );
}
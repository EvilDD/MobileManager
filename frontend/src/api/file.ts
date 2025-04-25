import request from '@/utils/request';
import type { BaseResponse } from './types';

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
}

export interface FileListResult {
  list: File[];
  total: number;
  page: number;
  pageSize: number;
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

export interface BatchPushParams {
  fileId: number;
  deviceIds: string[];
  maxWorker: number;
}

// 获取文件列表
export function getFileList(params: FileListParams) {
  return request<BaseResponse<FileListResult>>({
    url: '/api/v1/file/list',
    method: 'GET',
    params,
  });
}

// 上传文件
export function uploadFile(file: globalThis.File) {
  const formData = new FormData();
  formData.append('file', file);
  
  return request<BaseResponse<{ fileId: number }>>({
    url: '/api/v1/file/upload',
    method: 'POST',
    data: formData,
    headers: {
      'Content-Type': 'multipart/form-data',
    },
  });
}

// 批量推送文件到设备
export function batchPushByDevices(data: BatchPushParams) {
  return request<BaseResponse<{ taskId: string }>>({
    url: '/api/v1/file/batch-push',
    method: 'POST',
    data,
  });
}

// 获取批量任务状态
export function getBatchTaskStatus(taskId: string) {
  return request<BaseResponse<BatchTaskStatus>>({
    url: `/api/v1/file/batch-task-status/${taskId}`,
    method: 'GET',
  });
}

// 删除文件
export function deleteFile(fileId: number) {
  return request<BaseResponse<null>>({
    url: `/api/v1/file/delete/${fileId}`,
    method: 'DELETE',
  });
}

// 创建文件夹（空实现，根据需要扩展）
export function createFolder(name: string, parentId?: number) {
  return request<BaseResponse<{ fileId: number }>>({
    url: '/api/v1/file/create-folder',
    method: 'POST',
    data: {
      name,
      parentId,
    },
  });
}

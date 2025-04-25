import { 
  installApp, 
  uninstallApp, 
  startApp, 
  batchInstallApp, 
  batchUninstallApp, 
  batchStartApp, 
  getBatchTaskStatus, 
  uploadApk,
  type BatchTaskStatus,
  batchInstallByDevices,
  batchUninstallByDevices,
  batchStartByDevices
} from "@/api/app";
import { ElMessage } from 'element-plus';

/**
 * 上传APK文件
 * @param file 要上传的文件
 * @returns 返回上传结果
 */
export async function uploadAppFile(file: File): Promise<any> {
  try {
    const res = await uploadApk(file);
    
    if (res.code === 0) {
      ElMessage.success("应用导入成功");
      return res.data;
    } else {
      ElMessage.error(res.message || "上传失败");
      throw new Error(res.message || "上传失败");
    }
  } catch (error) {
    console.error("上传出错:", error);
    ElMessage.error("上传出错");
    throw error;
  }
}

/**
 * 在单个设备上执行应用操作
 * @param action 操作类型：install, uninstall, start
 * @param appId 应用ID
 * @param deviceId 设备ID
 * @returns 操作结果
 */
export async function performDeviceAppAction(
  action: 'install' | 'uninstall' | 'start',
  appId: number,
  deviceId: string
): Promise<any> {
  try {
    const actionData = { id: appId, deviceId };
    let actionPromise;
    let actionName = '';

    switch (action) {
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

    const res = await actionPromise;
    if (res.code === 0) {
      ElMessage.success(`${actionName}成功`);
      return res.data;
    } else {
      ElMessage.error(res.message || `${actionName}失败`);
      throw new Error(res.message || `${actionName}失败`);
    }
  } catch (error) {
    console.error(`操作出错:`, error);
    ElMessage.error(`操作出错`);
    throw error;
  }
}

/**
 * 执行批量应用操作
 * @param action 操作类型：install, uninstall, start
 * @param appId 应用ID
 * @param groupId 分组ID (如果为0则使用deviceIds)
 * @param maxWorker 最大并发数
 * @param deviceIds 可选的设备ID列表，当指定时直接对这些设备操作，而不通过分组
 * @returns 返回包含taskId的结果
 */
export async function performGroupAppAction(
  action: 'install' | 'uninstall' | 'start',
  appId: number,
  groupId: number,
  maxWorker: number = 50,
  deviceIds?: string[]
): Promise<string> {
  try {
    const data: any = {
      id: appId,
      maxWorker
    };
    
    // 如果提供了设备ID列表，使用设备ID模式，否则使用分组模式
    if (deviceIds && deviceIds.length > 0) {
      data.deviceIds = deviceIds;
    } else {
      data.groupId = groupId;
    }

    let res;
    switch (action) {
      case 'install':
        res = deviceIds 
          ? await batchInstallByDevices(data) 
          : await batchInstallApp(data);
        break;
      case 'uninstall':
        res = deviceIds 
          ? await batchUninstallByDevices(data) 
          : await batchUninstallApp(data);
        break;
      case 'start':
        res = deviceIds 
          ? await batchStartByDevices(data) 
          : await batchStartApp(data);
        break;
    }

    if (res.code === 0) {
      ElMessage.success('批量操作已开始执行');
      return res.data.taskId;
    } else {
      ElMessage.error(res.message || "批量操作失败");
      throw new Error(res.message || "批量操作失败");
    }
  } catch (error) {
    console.error('批量操作失败:', error);
    ElMessage.error('批量操作失败');
    throw error;
  }
}

/**
 * 获取任务状态
 * @param taskId 任务ID
 * @returns 任务状态
 */
export async function getTaskStatus(taskId: string): Promise<BatchTaskStatus> {
  try {
    const res = await getBatchTaskStatus(taskId);
    
    if (res.code === 0) {
      return res.data;
    } else {
      ElMessage.error(res.message || "获取任务状态失败");
      throw new Error(res.message || "获取任务状态失败");
    }
  } catch (error) {
    console.error("获取任务状态出错:", error);
    ElMessage.error("获取任务状态出错");
    throw error;
  }
}

/**
 * 获取任务状态类型
 * @param status 状态
 * @returns 状态类型
 */
export function getTaskStatusType(status: string): 'success' | 'danger' | 'warning' | 'info' {
  switch (status) {
    case 'complete': return 'success';
    case 'failed': return 'danger';
    case 'running': return 'warning';
    default: return 'info';
  }
}

/**
 * 获取任务状态文本
 * @param status 状态
 * @returns 状态文本
 */
export function getTaskStatusText(status: string): string {
  switch (status) {
    case 'pending': return '等待执行';
    case 'running': return '执行中';
    case 'complete': return '执行完成';
    case 'failed': return '执行失败';
    default: return status;
  }
}

/**
 * 获取进度条状态
 * @param task 任务状态对象
 * @returns 进度条状态
 */
export function getProgressStatus(task: BatchTaskStatus): '' | 'success' | 'exception' {
  if (task.status === 'failed') return 'exception';
  if (task.status === 'complete') return 'success';
  return '';
}

/**
 * 格式化文件大小
 * @param size 文件大小（字节）
 * @returns 格式化后的文件大小
 */
export function formatSize(size: number): string {
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
} 
import { batchPushByDevices, getBatchTaskStatus, type BatchTaskStatus } from '@/api/file';
import { ElMessage } from 'element-plus';

/**
 * 推送文件到设备
 * @param fileId 文件ID
 * @param deviceIds 设备ID数组
 * @param maxWorker 最大并发数
 * @returns 返回包含taskId的Promise
 */
export async function pushFileToDevices(fileId: number, deviceIds: string[], maxWorker: number = 50): Promise<string> {
  try {
    const res = await batchPushByDevices({
      fileId,
      deviceIds,
      maxWorker,
    });
    
    if (res.code === 0) {
      ElMessage.success("推送任务已创建");
      return res.data.taskId;
    } else {
      ElMessage.error(res.message || "推送任务创建失败");
      throw new Error(res.message || "推送任务创建失败");
    }
  } catch (error) {
    console.error("推送任务创建出错:", error);
    ElMessage.error("推送任务创建出错");
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
 * 格式化任务状态文本
 * @param status 状态
 * @returns 格式化后的状态文本
 */
export function formatTaskStatusText(status: string): string {
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
}

/**
 * 获取进度条状态
 * @param status 任务状态
 * @returns 进度条状态
 */
export function getProgressStatus(status: string): '' | 'success' | 'exception' {
  if (status === 'failed') return 'exception';
  if (status === 'complete') return 'success';
  return '';
}
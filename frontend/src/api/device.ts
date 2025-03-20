import { http } from "@/utils/http";

export interface InfoEntity {
  id: number;
  inst_type: number;
  inner_ip_v_6: string;
  telecom_ip_v_6: string;
  unicom_ip_v_6: string;
  mobile_ip_v_6: string;
  inner_ip_v_4: string;
  telecom_ip_v_4: string;
  unicom_ip_v_4: string;
  mobile_ip_v_4: string;
  status: string;
  desc: string;
  group_id?: number;
}

export interface InfoRealTime {
  status_code: number;
  status: string;
  alloced: boolean;
  room_id: string;
  room_role_num: number;
}

export interface Device {
  id: number;
  name: string;
  deviceId: string;
  groupId: number;
  status: string;
  createdAt: string;
  updatedAt: string;
}

export interface DeviceListResult {
  code: number;
  message: string;
  data: {
    list: Device[];
    page: number;
    pageSize: number;
    total: number;
  };
}

/** 获取设备列表 */
export function getDeviceList(params: { 
  page: number; 
  pageSize: number;
  keyword?: string;
  groupId?: number;
}) {
  return http.request<DeviceListResult>("get", "/api/devices/list", { params });
}

/** 保存设备 */
export function saveDevice(data: { 
  name: string;
  deviceId: string;
  groupId?: number;
  status: string;
}) {
  return http.request<{ code: number; message: string; data: Record<string, unknown> }>(
    "post",
    "/api/devices/create",
    { data }
  );
}

/** 设备截图响应结构 */
export interface ScreenshotResponse {
  code: number;
  message: string;
  data: {
    deviceId: string;
    success: boolean;
    imageData?: string; // base64编码的图片数据
    error?: string;     // 错误信息
  };
}

/** 后端直接返回的截图响应结构 */
interface ScreenshotBackendResponse {
  code: number;
  message: string;
  data: {
    deviceId: string;
    success: boolean;
    imageData?: string;
    error?: string;
  } | ScreenshotRes; // 可能是直接返回的结构
}

/** 后端可能直接返回的截图数据结构 */
interface ScreenshotRes {
  deviceId: string;
  success: boolean;
  imageData?: string;
  error?: string;
}

/** 获取设备截图 
 * @param data.deviceId 设备ID
 * @param data.quality 图片质量，1-100之间的整数，默认80
 * @returns 截图响应
 */
export function captureDeviceScreenshot(data: {
  deviceId: string;
  quality?: number;
}) {
  return http.request<ScreenshotBackendResponse>(
    "post",
    "/api/screenshot/capture",
    { data }
  ).then(response => {
    // 检查响应结构，进行适配转换
    if (response.code === 0) {
      // 检查data是否符合ScreenshotRes结构(直接返回的结构)
      const resData = response.data;
      if (!resData.deviceId && 'deviceId' in resData) {
        // 调整为标准化的响应格式
        return {
          code: response.code,
          message: response.message,
          data: response.data as ScreenshotRes
        };
      }
    }
    // 返回原始响应
    return response as unknown as ScreenshotResponse;
  });
} 
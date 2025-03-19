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
  return http.request<{ code: number; message: string; data: {} }>(
    "post",
    "/api/devices/create",
    { data }
  );
} 
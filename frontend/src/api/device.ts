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
  info_entity: InfoEntity;
  info_real_time: InfoRealTime;
}

export interface DeviceListResult {
  code: number;
  message: string;
  data: {
    page: number;
    size: number;
    total: number;
    devices: Device[];
  };
}

/** 获取设备列表 */
export function getDeviceList(params: { page: number; size: number }) {
  return http.request<DeviceListResult>("get", "/api/devices/list", { params });
}

/** 保存设备 */
export function saveDevice(data: { device: InfoEntity }) {
  return http.request<{ code: number; message: string; data: {} }>(
    "post",
    "/api/devices/save",
    { data }
  );
} 
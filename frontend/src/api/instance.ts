import { http } from "@/utils/http";

// 添加一个函数，用于在发起请求前修改JSON解析器
const configureJsonParsing = () => {
  // 保存原始的JSON.parse方法
  const originalJSONParse = JSON.parse;

  // 重写JSON.parse方法，添加对大整数的处理
  JSON.parse = function (text, reviver) {
    // 使用正则表达式匹配可能是大整数的字符串
    // 以下正则匹配形如 "room_id":1234567890123456789 的模式
    const bigIntRegex = /"(room_id)":(\d{15,})/g;
    // 将大整数转为字符串表示
    const safeText = text.replace(bigIntRegex, '"$1":"$2"');

    // console.log("原始JSON:", text);
    // console.log("处理后JSON:", safeText);

    // 使用原始方法解析修改后的文本
    return originalJSONParse(safeText, reviver);
  };
};

// 初始化时配置JSON解析
configureJsonParsing();

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
}

export interface InfoRealTime {
  status_code: number;
  status: string;
  alloced: boolean;
  room_id: string;
  room_role_num: number;
}

export interface Instance {
  info_entity: InfoEntity;
  info_real_time: InfoRealTime;
}

export interface InstanceListResult {
  code: number;
  message: string;
  data: {
    page: number;
    size: number;
    total: number;
    instances: Instance[];
  };
}

/** 获取实例列表 */
export const getInstanceList = (params: { page: number; size: number }) => {
  return http.request<InstanceListResult>("get", "/api/inst/list", { params });
};

/** 保存实例 */
export const saveInstance = (data: { instance: InfoEntity }) => {
  return http.request<{ code: number; message: string; data: {} }>(
    "post",
    "/api/inst/save",
    { data }
  );
};

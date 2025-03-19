import { http } from "@/utils/http";

export interface GroupItem {
  id: number;
  name: string;
  description: string;
  deviceCount: number;
  createdAt: string;
  updatedAt: string;
}

export interface GroupListReq {
  page: number;
  pageSize: number;
  keyword?: string;
}

export interface GroupListRes {
  code: number;
  message: string;
  data: {
    list: GroupItem[];
    total: number;
    page: number;
    pageSize: number;
  };
}

export interface GroupCreateReq {
  name: string;
  description?: string;
}

export interface GroupUpdateReq {
  id: number;
  name: string;
  description?: string;
}

export interface GroupDeleteReq {
  id: number;
}

// 获取分组列表
export function getGroupList(params: GroupListReq) {
  return http.request<GroupListRes>("get", "/api/groups/list", { params });
}

// 创建分组
export function createGroup(data: GroupCreateReq) {
  return http.request("post", "/api/groups/create", { data });
}

// 更新分组
export function updateGroup(data: GroupUpdateReq) {
  return http.request("put", "/api/groups/update", { data });
}

// 删除分组
export function deleteGroup(data: GroupDeleteReq) {
  return http.request("delete", "/api/groups/delete", { data });
}

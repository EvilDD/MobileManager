import { http } from "@/utils/http";

/** 签名测试接口返回类型 */
export interface SignatureTestResult {
  code: number;
  message: string;
  data: any;
}

/** 签名测试参数类型 */
export interface SignatureTestParams {
  page: number;
  size: number;
}

/** 签名测试 */
export const testSignature = (params: SignatureTestParams) => {
  return http.request<SignatureTestResult>("post", "/api/signature/test", {
    data: params
  });
};

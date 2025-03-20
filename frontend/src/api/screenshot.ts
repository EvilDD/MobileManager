import { http } from '@/utils/http'

export interface DeviceScreenshot {
  deviceId: string
  success: boolean
  imageData?: string
  error?: string
}

export interface ScreenshotRequest {
  deviceIds: string[]
  quality?: number
}

export interface ScreenshotResponse {
  screenshots: DeviceScreenshot[]
}

export const captureScreenshots = (data: ScreenshotRequest) => {
  return http.post<ScreenshotResponse, ScreenshotRequest>('/screenshot/capture', data)
} 
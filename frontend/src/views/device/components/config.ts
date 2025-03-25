// 设备流窗口的默认配置
export const STREAM_WINDOW_CONFIG = {
  // 默认尺寸
  DEFAULT_WIDTH: 480,  // 手机屏幕的一半宽度
  DEFAULT_HEIGHT: 990, // 手机屏幕的一半高度
  
  // 最小尺寸限制
  MIN_WIDTH: 400,
  MIN_HEIGHT: 600,
  
  // 最大尺寸限制（等于默认手机屏幕尺寸）
  MAX_WIDTH: 1080,
  MAX_HEIGHT: 1920
} as const; 
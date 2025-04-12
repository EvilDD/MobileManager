// 设备流窗口的默认配置
export const STREAM_WINDOW_CONFIG = {
  // 竖屏模式默认尺寸
  PORTRAIT: {
    WIDTH: 540,   // 默认宽度
    HEIGHT: 960,  // 默认高度
  },
  
  // 横屏模式默认尺寸
  LANDSCAPE: {
    WIDTH: 960,   // 横屏宽度等于竖屏高度
    HEIGHT: 540,  // 横屏高度等于竖屏宽度
  },
  
  // 控制按钮相关配置
  BUTTON: {
    HEIGHT: 52,   // 按钮高度
  },
  
  // 尺寸限制
  LIMITS: {
    MIN_WIDTH: 400,
    MIN_HEIGHT: 600,
    MAX_WIDTH: 1080,
    MAX_HEIGHT: 1920
  }
} as const; 
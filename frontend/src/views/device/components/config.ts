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
  
  // device-stream-container 固定尺寸
  CANVAS: {
    WIDTH: 960,   // 固定宽度，不区分横竖屏
    HEIGHT: 960,  // 固定高度，不区分横竖屏
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

export const DEVICE_CONFIG = {
  // 云机同步设备尺寸配置
  SYNC: {
    CANVAS_PORTRAIT: {
      WIDTH: 432,
      HEIGHT: 960
    },
    CANVAS_LANDSCAPE: {
      WIDTH: 870,
      HEIGHT: 480
    }
  }
} as const;

// 导出类型
export type DeviceConfig = typeof DEVICE_CONFIG; 
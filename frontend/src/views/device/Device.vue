// 打开串流窗口
const openStreamWindow = (device) => {
  // 更新当前选中设备
  currentDevice.value = device;
  
  // 根据当前环境选择合适的串流服务器地址
  const serverUrl = import.meta.env.VITE_WSCRCPY_SERVER_URL || 'http://localhost:8000';
  
  // 设置串流窗口参数
  streamDialogVisible.value = true;
  streamDeviceId.value = device.udid;
  streamServerUrl.value = serverUrl;
}; 
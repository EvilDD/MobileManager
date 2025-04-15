import { defineStore } from 'pinia'
import type { Device } from '@/api/device'

interface CloudPhoneState {
  activeGroupId: number;
  isLandscape: boolean;
  selectedDevices: Device[];
}

export const useCloudPhoneStore = defineStore({
  id: 'cloudphone',
  state: (): CloudPhoneState => ({
    activeGroupId: Number(localStorage.getItem('cloudphone-active-group')) || 0,
    isLandscape: localStorage.getItem('cloudphone-landscape') === 'true',
    selectedDevices: getStoredSelectedDevices(),
  }),
  actions: {
    setActiveGroup(groupId: number) {
      this.activeGroupId = groupId
      localStorage.setItem('cloudphone-active-group', String(groupId))
    },
    toggleOrientation() {
      this.isLandscape = !this.isLandscape;
      localStorage.setItem('cloudphone-landscape', String(this.isLandscape));
    },
    setSelectedDevices(devices: Device[]) {
      this.selectedDevices = devices;
      try {
        localStorage.setItem('cloudphone-selected-devices', JSON.stringify(devices));
      } catch (e) {
        console.error('Failed to save selected devices to localStorage:', e);
      }
    },
    clearSelectedDevices() {
      this.selectedDevices = [];
      localStorage.removeItem('cloudphone-selected-devices');
    },
    // 从localStorage恢复状态 - 不再需要，状态会在初始化时自动恢复
    restoreState() {
      console.log('State restoration not needed, already done in store initialization');
    }
  }
})

// 从localStorage获取已存储的设备列表
function getStoredSelectedDevices(): Device[] {
  try {
    const storedDevices = localStorage.getItem('cloudphone-selected-devices');
    if (storedDevices) {
      return JSON.parse(storedDevices);
    }
  } catch (e) {
    console.error('Failed to parse selected devices from localStorage:', e);
  }
  return [];
} 
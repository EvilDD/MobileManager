import { defineStore } from 'pinia'

interface CloudPhoneState {
  activeGroupId: number;
  isLandscape: boolean; // 添加横竖屏状态
}

export const useCloudPhoneStore = defineStore({
  id: 'cloudphone',
  state: (): CloudPhoneState => ({
    activeGroupId: 0,
    isLandscape: false, // 默认竖屏
  }),
  persist: {
    enabled: true,
    strategies: [
      {
        storage: localStorage,
        paths: ['activeGroupId', 'isLandscape'], // 添加 isLandscape 到持久化列表
      },
    ],
  },

  actions: {
    setActiveGroup(groupId: number) {
      this.activeGroupId = groupId
      localStorage.setItem('cloudphone-active-group', String(groupId))
    },
    toggleOrientation() { // 添加切换方向的 action
      this.isLandscape = !this.isLandscape;
    },
  },
}) 
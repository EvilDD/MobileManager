import { defineStore } from 'pinia'

export const useCloudPhoneStore = defineStore('cloudphone', {
  state: () => ({
    activeGroupId: Number(localStorage.getItem('cloudphone-active-group')) || 0, // 从 localStorage 获取，默认为新设备分组
  }),

  actions: {
    setActiveGroup(groupId: number) {
      this.activeGroupId = groupId
      localStorage.setItem('cloudphone-active-group', String(groupId))
    },
  },
}) 
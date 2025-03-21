const Layout = () => import("@/layout/index.vue");

export default {
  path: "/device-manage",
  name: "DeviceManage",
  component: Layout,
  meta: {
    icon: "ri:device-line",
    title: "设备管理",
    rank: 8,
    alwaysShow: true
  },
  children: [
    {
      path: "/device-manage/list",
      name: "DeviceManageList",
      component: () => import("@/views/device-manage/index.vue"),
      meta: {
        title: "设备列表",
        icon: "ri:list-check",
        showParent: true
      }
    },
    {
      path: "/device-manage/empty",
      name: "DeviceManageEmpty",
      component: () => import("@/views/device-manage/index.vue"),
      meta: {
        title: "空路由",
        showLink: false
      }
    }
  ]
} satisfies RouteConfigsTable;
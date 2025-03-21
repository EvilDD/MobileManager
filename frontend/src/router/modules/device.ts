const Layout = () => import("@/layout/index.vue");

export default {
  path: "/device",
  name: "Device",
  component: Layout,
  meta: {
    icon: "ri:smartphone-line",
    title: "手机控制台",
    rank: 9,
    alwaysShow: true
  },
  children: [
    {
      path: "/device/cloudphone",
      name: "CloudPhone",
      component: () => import("@/views/device/cloudphone.vue"),
      meta: {
        title: "分组手机",
        icon: "ri:smartphone-line",
        showParent: true
      }
    },
    {
      path: "/device/empty",
      name: "DeviceEmpty",
      component: () => import("@/views/device/cloudphone.vue"),
      meta: {
        title: "空路由",
        showLink: false
      }
    }
  ]
} satisfies RouteConfigsTable;

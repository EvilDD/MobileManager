export default {
  path: "/device",
  redirect: "/device/cloudphone",
  meta: {
    icon: "ri:smartphone-line",
    title: "云手机管理",
    rank: 9
    // showLink: false
  },
  children: [
    {
      path: "/device/cloudphone",
      name: "CloudPhone",
      component: () => import("@/views/device/cloudphone.vue"),
      meta: {
        title: "云手机管理",
        icon: "ri:smartphone-line"
      }
    }
  ]
} satisfies RouteConfigsTable;

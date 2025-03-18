export default {
  path: "/instance",
  redirect: "/instance/cloudphone",
  meta: {
    icon: "ri:smartphone-line",
    title: "云手机管理",
    rank: 9
    // showLink: false
  },
  children: [
    {
      path: "/instance/cloudphone",
      name: "CloudPhone",
      component: () => import("@/views/instance/cloudphone.vue"),
      meta: {
        title: "云手机管理",
        icon: "ri:smartphone-line"
      }
    }
  ]
} satisfies RouteConfigsTable;

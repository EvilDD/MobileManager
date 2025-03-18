export default {
  path: "/instance",
  redirect: "/instance/list",
  meta: {
    icon: "ri:server-line",
    title: "实例管理",
    rank: 9
    // showLink: false
  },
  children: [
    {
      path: "/instance/list",
      name: "InstanceList",
      component: () => import("@/views/instance/list.vue"),
      meta: {
        title: "实例列表"
        // showLink: true
      }
    },
    {
      path: "/instance/list",
      name: "InstanceList",
      component: () => import("@/views/instance/list.vue"),
      meta: {
        title: "实例列表",
        showLink: false
      }
    }
  ]
} satisfies RouteConfigsTable;

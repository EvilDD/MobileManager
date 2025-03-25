const Layout = () => import("@/layout/index.vue");

export default {
  path: "/app",
  name: "AppManage",
  component: Layout,
  meta: {
    icon: "ri:apps-line",
    title: "应用管理",
    rank: 10,
    alwaysShow: true
  },
  children: [
    {
      path: "/app/list",
      name: "AppList",
      component: () => import("@/views/app/list.vue"),
      meta: {
        title: "应用列表",
        icon: "ri:list-check",
        showParent: true
      }
    },
    {
      path: "/app/empty",
      name: "AppEmpty",
      component: () => import("@/views/app/list.vue"),
      meta: {
        title: "空路由",
        showLink: false
      }
    }
  ]
} satisfies RouteConfigsTable; 
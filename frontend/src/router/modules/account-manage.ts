const Layout = () => import("@/layout/index.vue");

export default {
  path: "/account",
  name: "AccountManage",
  component: Layout,
  meta: {
    icon: "ri:user-settings-line",
    title: "帐号管理",
    rank: 21,
    alwaysShow: true,
    showLink: false
  },
  children: [
    {
      path: "/account/list",
      name: "AccountList",
      component: () => import("@/views/account/list.vue"),
      meta: {
        title: "帐号列表",
        icon: "ri:user-line",
        showParent: true
      }
    },
    {
      path: "/account/empty",
      name: "AccountEmpty",
      component: () => import("@/views/account/list.vue"),
      meta: {
        title: "空路由",
        showLink: false
      }
    }
  ]
} satisfies RouteConfigsTable; 
const Layout = () => import("@/layout/index.vue");

export default {
  path: "/file",
  name: "FileManage",
  component: Layout,
  meta: {
    icon: "ri:file-list-line",
    title: "文件管理",
    rank: 11,
    alwaysShow: true
  },
  children: [
    {
      path: "/file/list",
      name: "FileList",
      component: () => import("@/views/file/list.vue"),
      meta: {
        title: "文件列表",
        icon: "ri:file-list-line",
        showParent: true
      }
    },
    {
      path: "/file/empty",
      name: "FileEmpty",
      component: () => import("@/views/file/list.vue"),
      meta: {
        title: "空路由",
        showLink: false
      }
    }
  ]
} satisfies RouteConfigsTable;

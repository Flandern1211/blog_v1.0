const Layout = () => import('@/layout/index.vue')

export default {
  name: 'System',
  path: '/setting',
  component: Layout,
  redirect: '/setting/website',
  meta: {
    title: '系统管理',
    icon: 'ion:md-settings',
    order: 6,
  },
  children: [
    {
      name: 'Website',
      path: 'website',
      component: () => import('./website/index.vue'),
      meta: {
        title: '网站管理',
        icon: 'el:website',
        order: 1,
        keepAlive: true,
      },
    },
    {
      name: 'Page',
      path: 'page',
      component: () => import('./page/index.vue'),
      meta: {
        title: '页面管理',
        icon: 'iconoir:journal-page',
        order: 2,
        keepAlive: true,
      },
    },
    {
      name: 'About',
      path: 'about',
      component: () => import('./about/index.vue'),
      meta: {
        title: '关于我',
        icon: 'cib:about-me',
        order: 4,
        keepAlive: true,
      },
    },
  ],
}

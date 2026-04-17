export default [
  {
    path: '/roles',
    name: 'Roles',
    component: () => import('../../views/Roles.vue'),
    meta: { requiresAuth: true, title: '角色管理' }
  },
  {
    path: '/menus',
    name: 'Menus',
    component: () => import('../../views/Menus.vue'),
    meta: { requiresAuth: true, title: '菜单管理' }
  },
  {
    path: '/apis',
    name: 'Apis',
    component: () => import('../../views/Apis.vue'),
    meta: { requiresAuth: true, title: 'API管理' }
  }
]
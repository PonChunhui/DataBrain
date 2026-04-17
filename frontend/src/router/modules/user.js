export default [
  {
    path: '/users',
    name: 'Users',
    component: () => import('../../views/Users.vue'),
    meta: { requiresAuth: true, title: '用户管理' }
  },
  {
    path: '/profile',
    name: 'Profile',
    component: () => import('../../views/Profile.vue'),
    meta: { requiresAuth: true, title: '个人信息' }
  }
]
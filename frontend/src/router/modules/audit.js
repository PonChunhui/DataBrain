export default [
  {
    path: '/audit',
    name: 'AuditLogs',
    component: () => import('../../views/AuditLogs.vue'),
    meta: { requiresAuth: true, title: '审计日志' }
  }
]
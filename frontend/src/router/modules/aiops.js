export default [
  {
    path: '/aiops',
    name: 'AIOPS',
    redirect: '/aiops/llm-config',
    meta: { requiresAuth: true, title: 'AIOPS智能运维' },
    children: [
      {
        path: '/aiops/llm-config',
        name: 'LLMConfig',
        component: () => import('../../views/aiops/LLMConfig.vue'),
        meta: { requiresAuth: true, title: 'LLM模型配置' }
      },
      {
        path: '/aiops/history',
        name: 'DiagnosticHistory',
        component: () => import('../../views/aiops/DiagnosticHistory.vue'),
        meta: { requiresAuth: true, title: '诊断历史记录' }
      },
      {
        path: '/aiops/report/:id',
        name: 'DiagnosticReportDetail',
        component: () => import('../../views/aiops/DiagnosticReportDetail.vue'),
        meta: { requiresAuth: true, title: '诊断报告详情' }
      }
    ]
  }
]
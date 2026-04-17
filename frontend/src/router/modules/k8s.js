export default [
  {
    path: '/k8s/auth',
    name: 'K8sAuth',
    component: () => import('../../views/k8s/K8sAuth.vue'),
    meta: { requiresAuth: true, title: 'K8s授权管理' }
  },
  {
    path: '/k8s/node/:name',
    name: 'K8sNodeDetail',
    component: () => import('../../views/k8s/K8sNodeDetail.vue'),
    meta: { requiresAuth: true, title: '节点详情' }
  },
  {
    path: '/k8s/nodes',
    name: 'K8sNodes',
    component: () => import('../../views/k8s/K8sNodes.vue'),
    meta: { requiresAuth: true, title: '集群节点' }
  },
  {
    path: '/k8s/cluster/:alias',
    name: 'K8sClusterDetail',
    component: () => import('../../views/k8s/K8sClusterDetail.vue'),
    meta: { requiresAuth: true, title: '集群详情' }
  },
  {
    path: '/k8s/clusters',
    name: 'K8sClusters',
    component: () => import('../../views/k8s/K8sClusters.vue'),
    meta: { requiresAuth: true, title: '集群管理' }
  },
  {
    path: '/k8s/deployments',
    name: 'K8sDeployments',
    component: () => import('../../views/k8s/K8sDeployments.vue'),
    meta: { requiresAuth: true, title: 'Deployment管理' }
  },
  {
    path: '/k8s/deployment/create',
    name: 'K8sDeploymentCreate',
    component: () => import('../../views/k8s/K8sDeploymentCreate.vue'),
    meta: { requiresAuth: true, title: '创建Deployment' }
  },
  {
    path: '/k8s/deployment/:name',
    name: 'K8sDeploymentDetail',
    component: () => import('../../views/k8s/K8sDeploymentDetail.vue'),
    meta: { requiresAuth: true, title: 'Deployment详情' }
  },
  {
    path: '/k8s/deployment/:name/edit',
    name: 'K8sDeploymentEdit',
    component: () => import('../../views/k8s/K8sDeploymentEdit.vue'),
    meta: { requiresAuth: true, title: '编辑Deployment' }
  },
  {
    path: '/k8s/pods',
    name: 'K8sPods',
    component: () => import('../../views/k8s/K8sPods.vue'),
    meta: { requiresAuth: true, title: 'Pod管理' }
  },
  {
    path: '/k8s/pod/:name',
    name: 'K8sPodDetail',
    component: () => import('../../views/k8s/K8sPodDetail.vue'),
    meta: { requiresAuth: true, title: 'Pod详情' }
  },
  {
    path: '/k8s/services',
    name: 'K8sServices',
    component: () => import('../../views/k8s/K8sServices.vue'),
    meta: { requiresAuth: true, title: 'Service管理' }
  },
  {
    path: '/k8s/service/:name',
    name: 'K8sServiceDetail',
    component: () => import('../../views/k8s/K8sServiceDetail.vue'),
    meta: { requiresAuth: true, title: 'Service详情' }
  },
  {
    path: '/k8s/configmaps',
    name: 'K8sConfigMaps',
    component: () => import('../../views/k8s/K8sConfigMaps.vue'),
    meta: { requiresAuth: true, title: 'ConfigMap管理' }
  },
  {
    path: '/k8s/secrets',
    name: 'K8sSecrets',
    component: () => import('../../views/k8s/K8sSecrets.vue'),
    meta: { requiresAuth: true, title: 'Secret管理' }
  },
  {
    path: '/k8s/ingress',
    name: 'K8sIngress',
    component: () => import('../../views/k8s/K8sIngress.vue'),
    meta: { requiresAuth: true, title: 'Ingress管理' }
  },
  {
    path: '/k8s/ingress/:name',
    name: 'K8sIngressDetail',
    component: () => import('../../views/k8s/K8sIngressDetail.vue'),
    meta: { requiresAuth: true, title: 'Ingress详情' }
  },
  {
    path: '/k8s/ingress/:name/edit',
    name: 'K8sIngressEdit',
    component: () => import('../../views/k8s/K8sIngressEdit.vue'),
    meta: { requiresAuth: true, title: '编辑Ingress' }
  }
]
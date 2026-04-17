-- 更新K8s子菜单的父级ID

-- 1. 先查找K8s管理菜单的ID
SELECT id FROM menus WHERE path = '/k8s' AND name = 'K8s管理';

-- 2. 更新子菜单的parent_id (假设K8s管理菜单的ID是6，请根据实际情况调整)
UPDATE menus SET parent_id = 6 WHERE path IN (
  '/k8s/clusters',
  '/k8s/deployments',
  '/k8s/pods',
  '/k8s/services'
);

-- 或者使用子查询自动获取ID
UPDATE menus m1
SET parent_id = (SELECT id FROM (SELECT id FROM menus WHERE path = '/k8s') m2)
WHERE m1.path IN (
  '/k8s/clusters',
  '/k8s/deployments',
  '/k8s/pods',
  '/k8s/services'
);

-- 3. 验证结果
SELECT id, name, path, parent_id FROM menus WHERE path LIKE '/k8s%' ORDER BY path;
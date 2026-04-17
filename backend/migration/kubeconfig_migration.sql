-- 数据库迁移脚本：将Token和CaCert字段改为Kubeconfig
-- 注意：执行前请备份数据库！

-- 方式1: 直接重建表结构（推荐，如果数据不重要）
-- DROP TABLE IF EXISTS k8s_clusters;
-- 然后重启后端程序，会自动创建新表结构

-- 方式2: 保留数据并迁移（需要手动构造kubeconfig）

-- 1. 添加kubeconfig字段
ALTER TABLE k8s_clusters ADD COLUMN kubeconfig TEXT AFTER name;

-- 2. 迁移现有数据（需要根据api_server、token、ca_cert构造kubeconfig）
-- 示例：为每个集群生成kubeconfig格式
UPDATE k8s_clusters SET kubeconfig = CONCAT(
  'apiVersion: v1\n',
  'kind: Config\n',
  'clusters:\n',
  '- cluster:\n',
  '    certificate-authority-data: ', ca_cert, '\n',
  '    server: ', api_server, '\n',
  '  name: default-cluster\n',
  'contexts:\n',
  '- context:\n',
  '    cluster: default-cluster\n',
  '    user: default-user\n',
  '    namespace: ', IFNULL(namespace, 'default'), '\n',
  '  name: default-context\n',
  'current-context: default-context\n',
  'users:\n',
  '- name: default-user\n',
  '  user:\n',
  '    token: ', token, '\n'
) WHERE kubeconfig IS NULL OR kubeconfig = '';

-- 3. 删除旧字段
ALTER TABLE k8s_clusters DROP COLUMN api_server;
ALTER TABLE k8s_clusters DROP COLUMN token;
ALTER TABLE k8s_clusters DROP COLUMN ca_cert;

-- 4. 设置kubeconfig字段为必填
ALTER TABLE k8s_clusters MODIFY kubeconfig TEXT NOT NULL;

-- 验证结果
SELECT id, name, LEFT(kubeconfig, 100) as kubeconfig_preview FROM k8s_clusters;

-- 注意事项：
-- 1. ca_cert字段可能已经是base64编码，如果不是需要手动转换
-- 2. 如果ca_cert为空，kubeconfig可能无法正常工作，建议手动编辑
-- 3. 迁移后建议重新添加集群，使用完整的kubeconfig文件
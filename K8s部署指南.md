# K8s多集群管理模块部署指南

## 1. 集群配置准备

### 方式1: 使用现有kubeconfig（推荐）

直接复制 `~/.kube/config` 文件内容：

```bash
# 查看kubeconfig内容
cat ~/.kube/config

# 或生成最小化配置
kubectl config view --minify --flatten
```

### 方式2: 创建专用kubeconfig

```bash
# 1. 创建ServiceAccount
kubectl create sa devops-admin -n default

# 2. 创建ClusterRoleBinding（管理员权限）
kubectl create clusterrolebinding devops-admin-binding \
  --clusterrole=cluster-admin \
  --serviceaccount=default:devops-admin

# 3. 获取Token
kubectl create token devops-admin -n default --duration=87600h

# 4. 创建kubeconfig文件
cat > devops-kubeconfig.yaml <<EOF
apiVersion: v1
kind: Config
clusters:
- cluster:
    certificate-authority-data: <base64-encoded-ca>
    server: https://<api-server>:6443
  name: devops-cluster
contexts:
- context:
    cluster: devops-cluster
    user: devops-user
    namespace: default
  name: devops-context
current-context: devops-context
users:
- name: devops-user
  user:
    token: <token-from-step-3>
EOF
```

### 方式3: 从多个集群合并kubeconfig

```bash
# 合并多个kubeconfig
KUBECONFIG=~/.kube/config:~/.kube/config-cluster2 kubectl config view --flatten > merged-config.yaml

# 查看所有context
kubectl config get-contexts --kubeconfig merged-config.yaml
```

## 2. 添加集群到平台

1. 登录平台 → K8s管理 → 集群管理
2. 点击"新增集群"
3. 填写信息：
   - **集群名称**: 如"生产环境"
   - **Kubeconfig**: 粘贴kubeconfig文件内容
   - **默认命名空间**: `default`

4. 点击"测试连接"验证kubeconfig有效性
5. 保存后即可使用

## 3. 使用功能

### Deployment管理
- 查看所有Deployment及其副本状态
- 扩缩容: 调整副本数(0-100)
- 重启: 滚动重启所有Pod
- 删除: 删除Deployment

### Pod管理
- 查看Pod列表、状态、IP、节点
- **查看日志**: 
  - 选择容器
  - 设置日志行数(100-5000)
  - 实时查看容器输出
- **查看事件**: Pod生命周期事件(创建、调度、启动等)
- **终端**: 提示kubectl exec命令(需WebSocket支持)
- 删除Pod

### Service管理
- 查看Service列表、类型(ClusterIP/NodePort等)
- 端口映射展示
- 删除Service

## 4. 权限配置

系统内置K8s相关权限：

| 权限代码 | 说明 | 菜单 |
|---------|------|------|
| cluster:create | 新增集群 | 集群管理 |
| cluster:edit | 编辑集群 | 集群管理 |
| cluster:delete | 删除集群 | 集群管理 |
| deployment:scale | 扩缩容 | Deployment |
| deployment:restart | 重启 | Deployment |
| deployment:delete | 删除Deployment | Deployment |
| pod:logs | 查看日志 | Pod管理 |
| pod:events | 查看事件 | Pod管理 |
| pod:terminal | 终端 | Pod管理 |
| pod:delete | 删除Pod | Pod管理 |
| service:create | 新增Service | Service管理 |
| service:delete | 删除Service | Service管理 |

管理员角色默认拥有所有权限，可在角色管理中分配。

## 5. API接口说明

所有K8s接口都在 `/api/k8s/` 路径下，需要JWT认证。

**通用参数**:
- `cluster_id`: 集群ID(必填)
- `namespace`: 命名空间(可选，默认default)

**示例请求**:
```bash
# 获取集群列表
curl -H "Authorization: Bearer <token>" \
  "http://localhost:8080/api/k8s/cluster"

# 测试kubeconfig连接
curl -X POST -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"kubeconfig":"<kubeconfig-content>"}' \
  "http://localhost:8080/api/k8s/cluster/test"

# 获取Deployment列表
curl -H "Authorization: Bearer <token>" \
  "http://localhost:8080/api/k8s/deployment?cluster_id=1&namespace=default"

# 扩缩容
curl -X POST -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"replicas":3}' \
  "http://localhost:8080/api/k8s/deployment/nginx-deployment/scale?cluster_id=1&namespace=default"

# 获取Pod日志
curl -H "Authorization: Bearer <token>" \
  "http://localhost:8080/api/k8s/pod/nginx-pod/logs?cluster_id=1&namespace=default&container=nginx&tail_lines=200"
```

## 6. 故障排查

| 问题 | 解决方案 |
|-----|---------|
| kubeconfig连接失败 | 检查kubeconfig格式、API Server可达性、证书有效性 |
| 命名空间获取失败 | 确认kubeconfig有足够权限(cluster-admin) |
| Pod日志空白 | 检查容器是否运行、日志行数设置 |
| Deployment重启无效果 | 确认Deployment存在、检查Pod状态 |
| kubeconfig格式错误 | 使用 `kubectl config view --flatten` 重新生成 |

### kubeconfig调试

```bash
# 验证kubeconfig格式
kubectl config view --kubeconfig=<your-kubeconfig>

# 测试连接
kubectl cluster-info --kubeconfig=<your-kubeconfig>

# 查看当前context
kubectl config current-context --kubeconfig=<your-kubeconfig>
```

## 7. 安全建议

1. **kubeconfig管理**: 
   - 使用最小化kubeconfig（只包含必要context）
   - 定期更换Token
   - 使用只读权限（ReadOnly权限）而非cluster-admin
2. **权限控制**: 生产环境使用更细粒度的RBAC
3. **网络隔离**: 平台与K8s API Server之间使用VPN/专线
4. **审计日志**: 启用K8s审计日志，记录所有操作
5. **敏感信息**: 
   - Kubeconfig内容不返回给前端（列表查询时自动隐藏）
   - 建议使用专用ServiceAccount而非个人kubeconfig

### 创建只读权限kubeconfig

```bash
# 创建只读ClusterRole
kubectl create clusterrole devops-reader \
  --verb=get,list,watch \
  --resource=pods,deployments,services,namespaces

# 创建ServiceAccount
kubectl create sa devops-reader -n default

# 创建ClusterRoleBinding
kubectl create clusterrolebinding devops-reader-binding \
  --clusterrole=devops-reader \
  --serviceaccount=default:devops-reader

# 获取Token
kubectl create token devops-reader -n default --duration=87600h
```
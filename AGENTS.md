# AGENTS.md

## 项目状态
✅ **已完成标准分层架构重构** - 按照后端开发指南.md重构为标准目录结构

## 默认账号
系统启动时自动创建管理员账号：
- **用户名**: `admin`
- **密码**: `admin123`

## 后端目录结构（标准架构）
```
backend/
├── api/v1          # API层(控制器)
├── config          # 配置结构体
├── core            # 核心组件初始化(Viper/Zap)
├── docs            # Swagger文档目录
├── global          # 全局对象(DB/Config/Log)
├── initialize      # 初始化层(GORM/Router)
│   └── internal    # 内部初始化函数
├── middleware      # 中间件层(JWT认证)
├── model           # 模型层
│   ├── request     # 入参结构体
│   └── response    # 出参结构体
├── resource        # 静态资源
├── router          # 路由层
├── service         # 服务层(业务逻辑)
├── source          # 初始化数据函数
├── utils           # 工具包
├── main.go         # 入口文件
└── config.yaml     # 配置文件
```

## 开发命令

```bash
# 后端开发
cd backend
go mod tidy                 # 整理依赖
go run main.go              # 启动服务
go build -o devops-backend  # 编译生产版本

# 前端开发
cd frontend
npm install                 # 安装依赖
npm run dev                 # 开发模式
npm run build               # 生产构建

# Docker部署
docker-compose up -d        # 启动完整环境
docker-compose down         # 停止环境
```

## 技术栈
- **后端**: Go 1.20 + Gin + GORM v2 + JWT + Zap + Viper
- **前端**: Vue 3 + Element Plus + Vue Router 4
- **数据库**: MySQL 8.0

## 核心架构特点

### 配置管理
- 使用`config.yaml`管理所有配置
- Viper自动读取并映射到结构体
- 支持MySQL、JWT、Server、Log配置

### 分层架构
- **API层**: 处理HTTP请求,参数验证
- **Service层**: 业务逻辑处理
- **Model层**: 数据模型定义
- **Router层**: 路由注册和中间件绑定

### 数据库准备
```bash
# 使用Docker启动MySQL
docker run -d --name devops-mysql \
  -e MYSQL_ROOT_PASSWORD=password \
  -e MYSQL_DATABASE=devopsdb \
  -p 3306:3306 mysql:8.0

# 或使用已有MySQL
mysql -uroot -p -e "CREATE DATABASE devopsdb;"
```

## 关键配置文件
- `backend/config.yaml` - 后端配置中心
- `backend/global/global.go` - 全局对象定义
- `backend/initialize/gorm.go` - GORM初始化
- `backend/router/router.go` - 路由注册
- `frontend/src/router/index.js` - 前端路由配置

### API文档
### 用户相关
- `POST /api/user/login` - 登录(无需认证)
- `POST /api/user/register` - 注册(无需认证)
- `GET /api/user` - 用户列表(需认证)
- `GET /api/user/:id` - 获取用户信息(需认证)
- `POST /api/user` - 创建用户(需认证)
- `PUT /api/user/:id` - 更新用户(需认证)
- `DELETE /api/user/:id` - 删除用户(需认证)
- `POST /api/user/:id/role` - 分配角色(需认证)
- `GET /api/user/:id/roles` - 获取用户角色(需认证)
- `PUT /api/user/:id/password` - 修改密码(需认证)

### 角色相关
- `GET /api/role` - 角色列表
- `POST /api/role` - 创建角色
- `PUT /api/role/:id` - 更新角色
- `DELETE /api/role/:id` - 删除角色
- `POST /api/role/:id/button` - 分配菜单按钮
- `POST /api/role/:id/menus` - 分配菜单
- `GET /api/role/:id/menus` - 获取角色菜单
- `POST /api/role/:id/apis` - 分配API
- `GET /api/role/:id/apis` - 获取角色API

### 用户权限相关
- `GET /api/user/:id/menus` - 获取用户菜单树(需认证)
- `GET /api/user/:id/apis` - 获取用户API列表(需认证)

### 菜单按钮权限相关
- `GET /api/menu-button` - 菜单按钮列表
- `GET /api/menu-button/menu/:menu_id` - 获取菜单按钮
- `POST /api/menu-button` - 创建菜单按钮
- `PUT /api/menu-button/:id` - 更新菜单按钮
- `DELETE /api/menu-button/:id` - 删除菜单按钮

### 菜单相关
- `GET /api/menu` - 菜单列表(需认证)
- `GET /api/menu/tree` - 菜单树(需认证)
- `POST /api/menu` - 创建菜单(需认证)
- `PUT /api/menu/:id` - 更新菜单(需认证)
- `DELETE /api/menu/:id` - 删除菜单(需认证)

### API管理相关
- `GET /api/api` - API列表(需认证,支持分页和搜索)
  - 参数：path(路径), method(方法), group(分组), status(状态), page(页码), pageSize(每页数量)
- `POST /api/api` - 创建API(需认证)
- `PUT /api/api/:id` - 更新API(需认证)
- `DELETE /api/api/:id` - 删除API(需认证)

## 环境变量管理
使用`config.yaml`配置,无需环境变量:
```yaml
mysql:
  path: 127.0.0.1
  port: 3306
  username: root
  password: password
  db-name: devopsdb

jwt:
  signing-key: your-secret-key
  expires-time: 86400
```

## 常见问题

| 问题 | 解决方案 |
|------|----------|
| 后端启动失败 | 检查config.yaml中的MySQL配置 |
| 前端无法连接后端 | 检查前端组件中API地址(默认 http://localhost:8080) |
| 登录401错误 | 确认用户已注册或密码正确 |
| 编译失败 | 运行`go mod tidy`整理依赖 |
| 配置读取失败 | 确保config.yaml在backend目录下 |
| JWT认证失败(Bearer prefix) | middleware/jwt.go已修复，支持"Bearer "前缀剥离 |

## 已实现功能
- ✅ 标准分层架构(API/Service/Model)
- ✅ 配置文件管理(Viper)
- ✅ 日志系统(Zap)
- ✅ JWT认证中间件
- ✅ 用户管理 CRUD
- ✅ 角色管理 CRUD
- ✅ 菜单按钮权限管理
- ✅ 菜单管理 CRUD(含按钮权限配置)
- ✅ API管理 CRUD
- ✅ RBAC关联
- ✅ 动态侧边栏菜单(支持多级子菜单)
- ✅ 前端按钮权限控制
- ✅ 前端完整界面
- ✅ K8s多集群管理模块
- ✅ Deployment管理(列表/扩缩容/重启)
- ✅ Pod管理(列表/日志/事件/删除)
- ✅ Service管理(列表/删除)

## 菜单结构

```
首页
用户管理
角色管理
菜单管理
API管理
K8s管理
  ├── 集群监控
  ├── 集群管理
  ├── Deployment
  ├── Pod管理
  └── Service管理
```

## 新增功能

### 1. 密码加密存储
- 使用bcrypt加密
- 登录验证使用`CheckPassword`
- 用户创建/更新自动加密

### 2. 权限验证中间件
- `middleware/api_auth.go`
- 基于RoleApi表验证API访问权限
- 可选择性应用于需要权限控制的路由

### 3. Deployment创建功能
- 前端表单：`/k8s/deployment/create`
- 支持配置：镜像、副本数、端口、环境变量、资源限制、标签

### 4. K8s集群监控Dashboard
- API: `GET /api/k8s/stats/:cluster_id`
- 显示：节点列表、Pod统计（运行/等待/失败）、命名空间数量

### 5. Swagger文档集成
- 使用swaggo/swag
- 访问：`http://localhost:8080/swagger/index.html`
- 自动生成API文档

### 6. Redis缓存
- 配置：`config.yaml`添加redis配置
- 用途：用户菜单缓存、API权限缓存

### 7. 单元测试
- 测试目录：`backend/tests/`
- 覆盖：用户服务、权限验证、K8s客户端

### 8. Pod日志WebSocket实时流
- 后端：`api/v1/pod_log_ws.go` - WebSocket handler
- 前端：`components/PodLogTerminal.vue` - xterm.js集成
- API：
  - `GET /api/k8s/pod/log/stream` - WebSocket日志流
  - `POST /api/k8s/pod/log/stop` - 停止流
- 特性：
  - 实时输出日志（follow模式）
  - 支持容器选择
  - 终端主题（GitHub暗色）
  - 清屏/重连/停止控制
  - 自动重连（最多3次，Succeeded状态不重连）
  - xterm.js终端模拟（50000行scrollback）
  - 搜索功能、下载日志、全屏模式
  - 自动滚动、连接时长显示

### 9. Pod终端WebSocket交互
- 后端：`api/v1/pod_exec_ws.go` - WebSocket exec handler
- 前端：`components/PodExecTerminal.vue` - xterm.js交互终端
- API：
  - `GET /api/k8s/pod/exec/stream` - WebSocket终端流
  - `POST /api/k8s/pod/exec/stop` - 停止终端
- 特性：
  - 实时命令交互（类似kubectl exec）
  - 支持Shell选择（/bin/sh, /bin/bash）
  - 终端窗口自适应
  - 全屏模式
  - GitHub暗色主题

### 10. 操作审计日志系统
- 后端：
  - `model/model.go` - AuditLog模型
  - `service/audit.go` - 审计日志服务
  - `api/v1/audit.go` - 审计日志API
  - `middleware/audit.go` - 自动审计中间件
- 前端：`views/AuditLogs.vue` - 审计日志页面
- API：
  - `GET /api/audit` - 审计日志列表（支持分页筛选）
  - `GET /api/audit/:id` - 审计日志详情
- 特性：
  - 自动记录所有POST/PUT/DELETE操作
  - 记录用户、操作类型、资源、IP、状态等
  - 敏感信息脱敏（密码字段显示***）
  - 支持时间范围筛选
  - JSON详情展示

### 11. ConfigMap/Secret管理
- 后端：
  - `api/v1/k8s_configmap.go` - ConfigMap控制器
  - `api/v1/k8s_secret.go` - Secret控制器
  - `service/k8s_configmap.go` - ConfigMap服务
  - `service/k8s_secret.go` - Secret服务
- 前端：
  - `views/K8sConfigMaps.vue` - ConfigMap管理页面
  - `views/K8sSecrets.vue` - Secret管理页面
- API：
  - ConfigMap: GET/POST/PUT/DELETE `/api/k8s/configmap`
  - Secret: GET/POST/PUT/DELETE `/api/k8s/secret`
- 特性：
  - 键值对编辑器
  - Secret类型选择（Opaque/TLS等）
  - Base64编码/解码显示
  - YAML导入导出

### 12. Deployment YAML编辑器
- 后端：
  - `api/v1/k8s_deployment.go` - 新增YAML API
  - `service/k8s_deployment.go` - YAML服务方法
- 前端：`views/K8sDeploymentEdit.vue` - YAML编辑页面
- API：
  - `GET /api/k8s/deployment/:name/yaml` - 获取YAML
  - `PUT /api/k8s/deployment/:name/yaml` - 更新YAML
- 特性：
  - YAML语法验证
  - 实时编辑保存
  - 重置恢复原始配置
  - 显示当前Deployment状态

### 13. Ingress路由管理
- 后端：
  - `api/v1/k8s_ingress.go` - Ingress控制器
  - `service/k8s_ingress.go` - Ingress服务
- 前端：`views/K8sIngress.vue` - Ingress管理页面
- API：
  - `GET /api/k8s/ingress` - Ingress列表
  - `GET /api/k8s/ingress/:name` - Ingress详情
  - `POST /api/k8s/ingress` - 创建Ingress
  - `PUT /api/k8s/ingress/:name` - 更新Ingress
  - `DELETE /api/k8s/ingress/:name` - 删除Ingress
- 特性：
  - 规则配置（主机、路径、后端服务）
  - 多规则支持
  - 规则详情展示
  - YAML导入导出

### 14. Pipeline可视化编辑器
- 后端：
  - `model/pipeline.go` - PipelineConfig字段（JSON存储）
  - `model/request/pipeline.go` - PipelineConfig请求字段
  - `api/v1/pipeline.go` - PipelineConfig处理
- 前端：
  - `components/pipeline/PipelineEditor.vue` - 主编辑器组件
  - `components/pipeline/AgentConfig.vue` - Agent/K8s Pod配置
  - `components/pipeline/EnvironmentConfig.vue` - 环境变量配置
  - `components/pipeline/OptionsConfig.vue` - Pipeline选项配置
  - `components/pipeline/PostConfig.vue` - Post处理配置
  - `components/pipeline/WhenConfig.vue` - 条件判断配置
  - `components/pipeline/steps/` - 步骤组件目录：
    - ShellStep.vue - Shell脚本执行
    - GitStep.vue - Git仓库操作
    - ArchiveStep.vue - 归档文件
    - EmailStep.vue - 发送邮件
    - ScriptStep.vue - 自定义脚本
    - DockerStep.vue - Docker操作（build/push/run）
    - KubectlStep.vue - Kubernetes操作（apply/delete/custom）
    - TimeoutStep.vue - 超时控制
    - RetryStep.vue - 重试控制
    - WaitStep.vue - 等待控制
  - `utils/pipeline/pipelineConverter.js` - JSON→Groovy脚本转换器
- 特性：
  - 双模式编辑：可视化编辑 + 脚本编辑切换
  - Agent配置：any/docker/kubernetes
  - Kubernetes Pod支持：
    - 多容器配置
    - Volume挂载（PVC/HostPath/Secret/ConfigMap/emptyDir）
    - 资源限制
    - container()包装（K8s环境下自动为Step添加）
  - Stage管理：
    - 拖拽排序（vuedraggable）
    - 环境变量配置
    - 条件判断
    - 人工确认
    - 并行阶段（parallel）
  - 步骤类型：
    - 基础：Shell/Git/Script
    - 控制：Timeout/Retry/Wait
    - 通知：Archive/Email
    - 容器：Docker/Kubectl
  - Post处理：always/success/failure
  - 实时脚本预览
  - JSON配置存储到数据库
- 数据结构：
```javascript
{
  agent: { type: 'kubernetes', containers: [], volumes: [] },
  environment: {},
  options: [],
  stages: [{ 
    name, 
    steps: [], 
    environment, 
    when, 
    hasInput, 
    input, 
    isParallel, 
    parallel 
  }],
  post: { always: [], success: [], failure: [] }
}
```

## K8s管理模块

### 集群管理
- `GET /api/k8s/cluster` - 集群列表（不返回kubeconfig）
- `GET /api/k8s/cluster/:id` - 集群详情（返回完整kubeconfig，用于编辑）
- `GET /api/k8s/cluster/:id/info` - 集群信息（API Server/Context，不含kubeconfig）
- `POST /api/k8s/cluster` - 创建集群(需kubeconfig文件内容)
- `PUT /api/k8s/cluster/:id` - 更新集群
- `DELETE /api/k8s/cluster/:id` - 删除集群
- `GET /api/k8s/cluster/:id/namespaces` - 获取命名空间列表
- `POST /api/k8s/cluster/test` - 测试kubeconfig连接

### 集群配置说明
1. **Kubeconfig**: 直接粘贴 `~/.kube/config` 文件内容
2. 系统会自动提取：
   - API Server地址
   - 当前Context
   - CA证书和认证信息
3. **默认命名空间**: 默认操作的命名空间

### 获取Kubeconfig
```bash
# 方式1: 使用现有kubeconfig
cat ~/.kube/config

# 方式2: 从远程集群获取
kubectl config view --raw

# 方式3: 创建专用kubeconfig
kubectl config set-credentials devops-user --token=<token>
kubectl config set-cluster devops-cluster --server=<api-server> --certificate-authority=<ca-file>
kubectl config set-context devops-context --cluster=devops-cluster --user=devops-user --namespace=default
kubectl config use-context devops-context
kubectl config view --minify --flatten
```
- `GET /api/k8s/deployment` - Deployment列表(参数: cluster_id, namespace)
- `POST /api/k8s/deployment/:name/scale` - 扩缩容(参数: replicas)
- `POST /api/k8s/deployment/:name/restart` - 重启Deployment
- `DELETE /api/k8s/deployment/:name` - 删除Deployment

### Pod管理
- `GET /api/k8s/pod` - Pod列表(参数: cluster_id, namespace)
- `GET /api/k8s/pod/:name/logs` - 获取日志(参数: container, tail_lines)
- `GET /api/k8s/pod/:name/events` - 获取Pod事件
- `DELETE /api/k8s/pod/:name` - 删除Pod

### Service管理
- `GET /api/k8s/service` - Service列表(参数: cluster_id, namespace)
- `POST /api/k8s/service` - 创建Service
- `DELETE /api/k8s/service/:name` - 删除Service

### 集群配置说明
1. **API Server**: K8s集群API地址，如 `https://k8s-api.example.com:6443`
2. **Token**: ServiceAccount Token，用于认证
3. **CA证书**: 集群CA证书(Base64编码或原文)
4. **默认命名空间**: 默认操作的命名空间

## 已完成优化
- ✅ 密码加密存储(bcrypt)
- ✅ 权限验证中间件(ApiAuth)
- ✅ Swagger文档集成
- ✅ 单元测试框架
- ✅ Redis缓存配置
- ✅ Deployment创建功能(完整表单)
- ✅ K8s资源监控Dashboard(节点/Pod统计)
- ✅ Pod日志WebSocket实时流(xterm.js)
- ✅ Pod终端WebSocket交互(kubectl exec)
- ✅ 操作审计日志系统(自动记录)
- ✅ ConfigMap管理(键值对编辑)
- ✅ Secret管理(Base64编码)
- ✅ Deployment YAML编辑器
- ✅ Ingress路由管理(规则配置)
- ✅ AIOPS智能运维平台(LLM故障诊断/集群巡检)

### AIOPS智能运维模块

#### LLM模型配置管理
- `GET /api/aiops/llm-config` - LLM配置列表
- `POST /api/aiops/llm-config` - 创建LLM配置
- `PUT /api/aiops/llm-config/:id` - 更新LLM配置
- `DELETE /api/aiops/llm-config/:id` - 删除LLM配置
- `POST /api/aiops/llm-config/:id/test` - 测试LLM连接
- `POST /api/aiops/llm-config/:id/set-default` - 设置默认配置

#### Pod智能诊断(SSE流式)
- `POST /api/aiops/diagnostic/stream` - SSE流式诊断
- `POST /api/aiops/diagnostic/chat` - 继续对话
- `GET /api/aiops/diagnostic/:id` - 获取诊断详情
- `GET /api/aiops/history` - 诊断历史列表
- `GET /api/aiops/history/stats` - 诊断统计

#### 集群智能巡检(SSE流式)
- `POST /api/aiops/inspection/stream` - SSE流式巡检
- `POST /api/aiops/inspection/chat` - 继续对话

#### Prometheus指标集成
- 集成kube-state-metrics、node-exporter、cAdvisor
- 30分钟指标采集窗口
- 聚合统计(avg/max/min/trend)供LLM分析
- Pod CPU/内存/网络/磁盘IO指标
- 节点资源使用率、集群整体健康度
- 容器throttling检测、OOM风险评估

#### 技术实现
- langchaingo框架集成(支持OpenAI兼容API)
- SSE实时流式响应
- 支持DeepSeek/Claude/Qwen/Gemini/OpenAI等模型
- 自动数据采集(Pod状态/事件/日志/节点信息/Prometheus指标)

## 待扩展功能
- [ ] HPA自动扩缩容
- [ ] CronJob定时任务
- [ ] 告警规则配置
- [ ] 多租户隔离
- [ ] PV/PVC存储管理
- [ ] NetworkPolicy网络策略
- [ ] LimitRange资源限制
- [ ] ResourceQuota配额管理
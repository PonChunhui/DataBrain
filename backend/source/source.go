package source

import (
	"devops-backend/global"
	"devops-backend/model"
	"devops-backend/utils"
	"go.uber.org/zap"
)

func InitData() {
	var user model.User
	err := global.GVA_DB.Where("username = ?", "admin").First(&user).Error
	if err != nil {
		hashedPassword, _ := utils.HashPassword("admin123")
		admin := model.User{
			Username: "admin",
			Password: hashedPassword,
			Email:    "admin@devops.com",
		}
		if err := global.GVA_DB.Create(&admin).Error; err != nil {
			global.GVA_LOG.Error("创建默认管理员失败", zap.Error(err))
		} else {
			global.GVA_LOG.Info("默认管理员账号创建成功", zap.String("username", "admin"), zap.String("password", "admin123"))
		}
	} else {
		global.GVA_LOG.Info("管理员账号已存在", zap.String("username", "admin"))

		if user.Password == "admin123" || len(user.Password) < 20 {
			hashedPassword, _ := utils.HashPassword("admin123")
			user.Password = hashedPassword
			if err := global.GVA_DB.Save(&user).Error; err != nil {
				global.GVA_LOG.Error("更新管理员密码失败", zap.Error(err))
			} else {
				global.GVA_LOG.Info("管理员密码已更新为bcrypt加密格式")
			}
		}
	}

	// 创建默认角色
	var role model.Role
	err = global.GVA_DB.Where("name = ?", "管理员").First(&role).Error
	if err != nil {
		adminRole := model.Role{
			Name:        "管理员",
			Description: "系统管理员角色，拥有所有权限",
		}
		if err := global.GVA_DB.Create(&adminRole).Error; err != nil {
			global.GVA_LOG.Error("创建默认角色失败", zap.Error(err))
		} else {
			global.GVA_LOG.Info("默认角色创建成功", zap.String("role", "管理员"))
		}
	}

	// 为管理员分配角色
	var adminUser model.User
	global.GVA_DB.Where("username = ?", "admin").First(&adminUser)

	var adminRole model.Role
	global.GVA_DB.Where("name = ?", "管理员").First(&adminRole)

	menus := []model.Menu{
		{Name: "首页", Path: "/dashboard", Icon: "HomeFilled", Sort: 1, ParentID: 0, IsShow: true, Component: "Dashboard"},
		{Name: "用户管理", Path: "/users", Icon: "User", Sort: 2, ParentID: 0, IsShow: true, Component: "Users"},
		{Name: "角色管理", Path: "/roles", Icon: "UserFilled", Sort: 3, ParentID: 0, IsShow: true, Component: "Roles"},
		{Name: "菜单管理", Path: "/menus", Icon: "Menu", Sort: 4, ParentID: 0, IsShow: true, Component: "Menus"},
		{Name: "API管理", Path: "/apis", Icon: "Document", Sort: 5, ParentID: 0, IsShow: true, Component: "Apis"},
		{Name: "审计日志", Path: "/audit", Icon: "DocumentChecked", Sort: 6, ParentID: 0, IsShow: true, Component: "AuditLogs"},
		{Name: "K8s管理", Path: "/k8s", Icon: "CloudServer", Sort: 7, ParentID: 0, IsShow: true, Component: ""},
		{Name: "AIOPS智能运维", Path: "/aiops", Icon: "Cpu", Sort: 8, ParentID: 0, IsShow: true, Component: ""},
	}

	for i := range menus {
		var existMenu model.Menu
		err := global.GVA_DB.Where("path = ?", menus[i].Path).First(&existMenu).Error
		if err != nil {
			if err := global.GVA_DB.Create(&menus[i]).Error; err != nil {
				global.GVA_LOG.Error("创建默认菜单失败", zap.Error(err))
			}
		} else {
			menus[i].ID = existMenu.ID
		}
	}

	var k8sMenu model.Menu
	global.GVA_DB.Where("path = ?", "/k8s").First(&k8sMenu)

	k8sSubMenus := []model.Menu{
		{Name: "集群节点", Path: "/k8s/nodes", Icon: "DataLine", Sort: 1, ParentID: k8sMenu.ID, IsShow: true, Component: "K8sNodes"},
		{Name: "集群管理", Path: "/k8s/clusters", Icon: "Connection", Sort: 2, ParentID: k8sMenu.ID, IsShow: true, Component: "K8sClusters"},
		{Name: "K8s授权管理", Path: "/k8s/auth", Icon: "Lock", Sort: 3, ParentID: k8sMenu.ID, IsShow: true, Component: "K8sAuth"},
		{Name: "Deployment", Path: "/k8s/deployments", Icon: "Box", Sort: 4, ParentID: k8sMenu.ID, IsShow: true, Component: "K8sDeployments"},
		{Name: "Pod管理", Path: "/k8s/pods", Icon: "Grid", Sort: 5, ParentID: k8sMenu.ID, IsShow: true, Component: "K8sPods"},
		{Name: "Service管理", Path: "/k8s/services", Icon: "Share", Sort: 6, ParentID: k8sMenu.ID, IsShow: true, Component: "K8sServices"},
		{Name: "ConfigMap", Path: "/k8s/configmaps", Icon: "Files", Sort: 7, ParentID: k8sMenu.ID, IsShow: true, Component: "K8sConfigMaps"},
		{Name: "Secret", Path: "/k8s/secrets", Icon: "Lock", Sort: 8, ParentID: k8sMenu.ID, IsShow: true, Component: "K8sSecrets"},
		{Name: "Ingress", Path: "/k8s/ingress", Icon: "Link", Sort: 9, ParentID: k8sMenu.ID, IsShow: true, Component: "K8sIngress"},
	}

	for i := range k8sSubMenus {
		var existMenu model.Menu
		err := global.GVA_DB.Where("path = ?", k8sSubMenus[i].Path).First(&existMenu).Error
		if err != nil {
			if err := global.GVA_DB.Create(&k8sSubMenus[i]).Error; err != nil {
				global.GVA_LOG.Error("创建K8s子菜单失败", zap.Error(err))
			}
		} else {
			if existMenu.ParentID != k8sMenu.ID {
				existMenu.ParentID = k8sMenu.ID
				global.GVA_DB.Save(&existMenu)
			}
			k8sSubMenus[i].ID = existMenu.ID
		}
	}

	var aiopsMenu model.Menu
	global.GVA_DB.Where("path = ?", "/aiops").First(&aiopsMenu)

	aiopsSubMenus := []model.Menu{
		{Name: "LLM模型配置", Path: "/aiops/llm-config", Icon: "Setting", Sort: 1, ParentID: aiopsMenu.ID, IsShow: true, Component: "LLMConfig"},
		{Name: "诊断历史记录", Path: "/aiops/history", Icon: "Document", Sort: 2, ParentID: aiopsMenu.ID, IsShow: true, Component: "DiagnosticHistory"},
	}

	for i := range aiopsSubMenus {
		var existMenu model.Menu
		err := global.GVA_DB.Where("path = ?", aiopsSubMenus[i].Path).First(&existMenu).Error
		if err != nil {
			if err := global.GVA_DB.Create(&aiopsSubMenus[i]).Error; err != nil {
				global.GVA_LOG.Error("创建AIOPS子菜单失败", zap.Error(err))
			}
		} else {
			if existMenu.ParentID != aiopsMenu.ID {
				existMenu.ParentID = aiopsMenu.ID
				global.GVA_DB.Save(&existMenu)
			}
			aiopsSubMenus[i].ID = existMenu.ID
		}
	}

	for i := range menus {
		if menus[i].Path == "/k8s" {
			menus[i].Children = k8sSubMenus
			break
		}
	}

	for i := range menus {
		if menus[i].Path == "/aiops" {
			menus[i].Children = aiopsSubMenus
			break
		}
	}

	menus = append(menus, k8sSubMenus...)
	menus = append(menus, aiopsSubMenus...)

	global.GVA_LOG.Info("默认菜单初始化完成")

	apis := []model.Api{
		{Path: "/api/user/login", Method: "POST", Description: "用户登录", Group: "用户管理", Status: true},
		{Path: "/api/user", Method: "GET", Description: "获取用户列表", Group: "用户管理", Status: true},
		{Path: "/api/user", Method: "POST", Description: "创建用户", Group: "用户管理", Status: true},
		{Path: "/api/user/:id", Method: "PUT", Description: "更新用户", Group: "用户管理", Status: true},
		{Path: "/api/user/:id", Method: "DELETE", Description: "删除用户", Group: "用户管理", Status: true},
		{Path: "/api/user/:id/role", Method: "POST", Description: "分配用户角色", Group: "用户管理", Status: true},
		{Path: "/api/role", Method: "GET", Description: "获取角色列表", Group: "角色管理", Status: true},
		{Path: "/api/role", Method: "POST", Description: "创建角色", Group: "角色管理", Status: true},
		{Path: "/api/role/:id", Method: "PUT", Description: "更新角色", Group: "角色管理", Status: true},
		{Path: "/api/role/:id", Method: "DELETE", Description: "删除角色", Group: "角色管理", Status: true},
		{Path: "/api/role/:id/button", Method: "POST", Description: "分配菜单按钮", Group: "角色管理", Status: true},
		{Path: "/api/role/:id/menus", Method: "POST", Description: "分配菜单", Group: "角色管理", Status: true},
		{Path: "/api/role/:id/menus", Method: "GET", Description: "获取角色菜单", Group: "角色管理", Status: true},
		{Path: "/api/role/:id/apis", Method: "POST", Description: "分配API", Group: "角色管理", Status: true},
		{Path: "/api/role/:id/apis", Method: "GET", Description: "获取角色API", Group: "角色管理", Status: true},
		{Path: "/api/menu-button", Method: "GET", Description: "获取菜单按钮列表", Group: "菜单按钮", Status: true},
		{Path: "/api/menu-button/menu/:menu_id", Method: "GET", Description: "获取菜单按钮", Group: "菜单按钮", Status: true},
		{Path: "/api/menu-button", Method: "POST", Description: "创建菜单按钮", Group: "菜单按钮", Status: true},
		{Path: "/api/menu-button/:id", Method: "PUT", Description: "更新菜单按钮", Group: "菜单按钮", Status: true},
		{Path: "/api/menu-button/:id", Method: "DELETE", Description: "删除菜单按钮", Group: "菜单按钮", Status: true},
		{Path: "/api/menu", Method: "GET", Description: "获取菜单列表", Group: "菜单管理", Status: true},
		{Path: "/api/menu/tree", Method: "GET", Description: "获取菜单树", Group: "菜单管理", Status: true},
		{Path: "/api/menu", Method: "POST", Description: "创建菜单", Group: "菜单管理", Status: true},
		{Path: "/api/menu/:id", Method: "PUT", Description: "更新菜单", Group: "菜单管理", Status: true},
		{Path: "/api/menu/:id", Method: "DELETE", Description: "删除菜单", Group: "菜单管理", Status: true},
		{Path: "/api/api", Method: "GET", Description: "获取API列表", Group: "API管理", Status: true},
		{Path: "/api/api", Method: "POST", Description: "创建API", Group: "API管理", Status: true},
		{Path: "/api/api/:id", Method: "PUT", Description: "更新API", Group: "API管理", Status: true},
		{Path: "/api/api/:id", Method: "DELETE", Description: "删除API", Group: "API管理", Status: true},
		{Path: "/api/audit", Method: "GET", Description: "获取审计日志列表", Group: "审计日志", Status: true},
		{Path: "/api/audit/:id", Method: "GET", Description: "获取审计日志详情", Group: "审计日志", Status: true},
		{Path: "/api/k8s/cluster", Method: "GET", Description: "获取集群列表", Group: "K8s管理", Status: true},
		{Path: "/api/k8s/cluster", Method: "POST", Description: "创建集群", Group: "K8s管理", Status: true},
		{Path: "/api/k8s/cluster/:id", Method: "PUT", Description: "更新集群", Group: "K8s管理", Status: true},
		{Path: "/api/k8s/cluster/:id", Method: "DELETE", Description: "删除集群", Group: "K8s管理", Status: true},
		{Path: "/api/k8s/cluster/:id/namespaces", Method: "GET", Description: "获取命名空间", Group: "K8s管理", Status: true},
		{Path: "/api/k8s/deployment", Method: "GET", Description: "获取Deployment列表", Group: "K8s管理", Status: true},
		{Path: "/api/k8s/deployment/:name/scale", Method: "POST", Description: "扩缩容Deployment", Group: "K8s管理", Status: true},
		{Path: "/api/k8s/deployment/:name/restart", Method: "POST", Description: "重启Deployment", Group: "K8s管理", Status: true},
		{Path: "/api/k8s/deployment/:name", Method: "DELETE", Description: "删除Deployment", Group: "K8s管理", Status: true},
		{Path: "/api/k8s/pod", Method: "GET", Description: "获取Pod列表", Group: "K8s管理", Status: true},
		{Path: "/api/k8s/pod/:name/logs", Method: "GET", Description: "获取Pod日志", Group: "K8s管理", Status: true},
		{Path: "/api/k8s/pod/:name/events", Method: "GET", Description: "获取Pod事件", Group: "K8s管理", Status: true},
		{Path: "/api/k8s/pod/:name", Method: "DELETE", Description: "删除Pod", Group: "K8s管理", Status: true},
		{Path: "/api/k8s/service", Method: "GET", Description: "获取Service列表", Group: "K8s管理", Status: true},
		{Path: "/api/k8s/service", Method: "POST", Description: "创建Service", Group: "K8s管理", Status: true},
		{Path: "/api/k8s/service/:name", Method: "DELETE", Description: "删除Service", Group: "K8s管理", Status: true},
		{Path: "/api/k8s/ingress", Method: "GET", Description: "获取Ingress列表", Group: "K8s管理", Status: true},
		{Path: "/api/k8s/ingress", Method: "POST", Description: "创建Ingress", Group: "K8s管理", Status: true},
		{Path: "/api/k8s/ingress/:name", Method: "GET", Description: "获取Ingress详情", Group: "K8s管理", Status: true},
		{Path: "/api/k8s/ingress/:name", Method: "PUT", Description: "更新Ingress", Group: "K8s管理", Status: true},
		{Path: "/api/k8s/ingress/:name", Method: "DELETE", Description: "删除Ingress", Group: "K8s管理", Status: true},
		{Path: "/api/aiops/llm-config", Method: "GET", Description: "获取LLM配置列表", Group: "AIOPS智能运维", Status: true},
		{Path: "/api/aiops/llm-config/:id", Method: "GET", Description: "获取LLM配置详情", Group: "AIOPS智能运维", Status: true},
		{Path: "/api/aiops/llm-config", Method: "POST", Description: "创建LLM配置", Group: "AIOPS智能运维", Status: true},
		{Path: "/api/aiops/llm-config/:id", Method: "PUT", Description: "更新LLM配置", Group: "AIOPS智能运维", Status: true},
		{Path: "/api/aiops/llm-config/:id", Method: "DELETE", Description: "删除LLM配置", Group: "AIOPS智能运维", Status: true},
		{Path: "/api/aiops/llm-config/:id/test", Method: "POST", Description: "测试LLM配置", Group: "AIOPS智能运维", Status: true},
		{Path: "/api/aiops/llm-config/:id/set-default", Method: "POST", Description: "设置默认LLM配置", Group: "AIOPS智能运维", Status: true},
		{Path: "/api/aiops/diagnostic", Method: "POST", Description: "启动诊断", Group: "AIOPS智能运维", Status: true},
		{Path: "/api/aiops/diagnostic/:id", Method: "GET", Description: "获取诊断结果", Group: "AIOPS智能运维", Status: true},
		{Path: "/api/aiops/history", Method: "GET", Description: "获取诊断历史", Group: "AIOPS智能运维", Status: true},
		{Path: "/api/aiops/history/stats", Method: "GET", Description: "获取诊断统计", Group: "AIOPS智能运维", Status: true},
	}

	for i := range apis {
		var existApi model.Api
		err := global.GVA_DB.Where("path = ? AND method = ?", apis[i].Path, apis[i].Method).First(&existApi).Error
		if err != nil {
			if err := global.GVA_DB.Create(&apis[i]).Error; err != nil {
				global.GVA_LOG.Error("创建默认API失败", zap.Error(err))
			}
		} else {
			apis[i].ID = existApi.ID
		}
	}
	global.GVA_LOG.Info("默认API初始化完成")

	if adminUser.ID > 0 && adminRole.ID > 0 {
		var userRole model.UserRole
		err := global.GVA_DB.Where("user_id = ? AND role_id = ?", adminUser.ID, adminRole.ID).First(&userRole).Error
		if err != nil {
			userRole = model.UserRole{
				UserID: adminUser.ID,
				RoleID: adminRole.ID,
			}
			if err := global.GVA_DB.Create(&userRole).Error; err != nil {
				global.GVA_LOG.Error("分配管理员角色失败", zap.Error(err))
			} else {
				global.GVA_LOG.Info("管理员角色分配成功")
			}
		}

		var roleMenus []model.RoleMenu
		global.GVA_DB.Where("role_id = ?", adminRole.ID).Find(&roleMenus)
		if len(roleMenus) == 0 {
			for _, menu := range menus {
				roleMenu := model.RoleMenu{
					RoleID: adminRole.ID,
					MenuID: menu.ID,
				}
				if err := global.GVA_DB.Create(&roleMenu).Error; err != nil {
					global.GVA_LOG.Error("分配管理员菜单失败", zap.Error(err))
				}
			}
			global.GVA_LOG.Info("管理员菜单分配成功")
		}

		var roleApis []model.RoleApi
		global.GVA_DB.Where("role_id = ?", adminRole.ID).Find(&roleApis)

		existingApiIDs := make(map[uint]bool)
		for _, ra := range roleApis {
			existingApiIDs[ra.ApiID] = true
		}

		for _, api := range apis {
			if !existingApiIDs[api.ID] {
				roleApi := model.RoleApi{
					RoleID: adminRole.ID,
					ApiID:  api.ID,
				}
				if err := global.GVA_DB.Create(&roleApi).Error; err != nil {
					global.GVA_LOG.Error("分配管理员API失败", zap.Error(err))
				}
			}
		}
		if len(roleApis) == 0 {
			global.GVA_LOG.Info("管理员API分配成功")
		} else {
			global.GVA_LOG.Info("管理员API权限已更新，新增缺失的API")
		}
	}

	menuButtons := []struct {
		Path    string
		Buttons []model.MenuButton
	}{
		{
			Path: "/users",
			Buttons: []model.MenuButton{
				{Code: "user:create", Name: "新增用户", Description: "创建新用户"},
				{Code: "user:edit", Name: "编辑用户", Description: "编辑用户信息"},
				{Code: "user:delete", Name: "删除用户", Description: "删除用户"},
			},
		},
		{
			Path: "/roles",
			Buttons: []model.MenuButton{
				{Code: "role:create", Name: "新增角色", Description: "创建新角色"},
				{Code: "role:edit", Name: "编辑角色", Description: "编辑角色信息"},
				{Code: "role:delete", Name: "删除角色", Description: "删除角色"},
			},
		},
		{
			Path: "/menus",
			Buttons: []model.MenuButton{
				{Code: "menu:create", Name: "新增菜单", Description: "创建新菜单"},
				{Code: "menu:edit", Name: "编辑菜单", Description: "编辑菜单信息"},
				{Code: "menu:delete", Name: "删除菜单", Description: "删除菜单"},
				{Code: "button:create", Name: "新增按钮", Description: "配置按钮权限"},
				{Code: "button:edit", Name: "编辑按钮", Description: "编辑按钮权限"},
				{Code: "button:delete", Name: "删除按钮", Description: "删除按钮权限"},
			},
		},
		{
			Path: "/apis",
			Buttons: []model.MenuButton{
				{Code: "api:create", Name: "新增API", Description: "创建新API"},
				{Code: "api:edit", Name: "编辑API", Description: "编辑API信息"},
				{Code: "api:delete", Name: "删除API", Description: "删除API"},
			},
		},
		{
			Path: "/k8s/clusters",
			Buttons: []model.MenuButton{
				{Code: "cluster:create", Name: "新增集群", Description: "添加K8s集群"},
				{Code: "cluster:edit", Name: "编辑集群", Description: "编辑集群配置"},
				{Code: "cluster:delete", Name: "删除集群", Description: "删除集群"},
			},
		},
		{
			Path: "/k8s/deployments",
			Buttons: []model.MenuButton{
				{Code: "deployment:scale", Name: "扩缩容", Description: "调整副本数"},
				{Code: "deployment:restart", Name: "重启", Description: "重启Deployment"},
				{Code: "deployment:delete", Name: "删除", Description: "删除Deployment"},
			},
		},
		{
			Path: "/k8s/pods",
			Buttons: []model.MenuButton{
				{Code: "pod:logs", Name: "查看日志", Description: "查看容器日志"},
				{Code: "pod:events", Name: "查看事件", Description: "查看Pod事件"},
				{Code: "pod:terminal", Name: "终端", Description: "进入容器终端"},
				{Code: "pod:delete", Name: "删除", Description: "删除Pod"},
			},
		},
		{
			Path: "/k8s/auth",
			Buttons: []model.MenuButton{
				{Code: "auth:create", Name: "新增授权", Description: "创建K8s授权"},
				{Code: "auth:edit", Name: "编辑授权", Description: "编辑K8s授权"},
				{Code: "auth:delete", Name: "删除授权", Description: "删除K8s授权"},
			},
		},
		{
			Path: "/k8s/services",
			Buttons: []model.MenuButton{
				{Code: "service:create", Name: "新增Service", Description: "创建新Service"},
				{Code: "service:delete", Name: "删除Service", Description: "删除Service"},
			},
		},
	}

	for _, item := range menuButtons {
		var menu model.Menu
		if err := global.GVA_DB.Where("path = ?", item.Path).First(&menu).Error; err == nil {
			for _, button := range item.Buttons {
				button.MenuID = menu.ID
				var existButton model.MenuButton
				err := global.GVA_DB.Where("menu_id = ? AND code = ?", menu.ID, button.Code).First(&existButton).Error
				if err != nil {
					if err := global.GVA_DB.Create(&button).Error; err != nil {
						global.GVA_LOG.Error("创建菜单按钮权限失败", zap.Error(err), zap.String("path", item.Path), zap.String("code", button.Code))
					}
				}
			}
		}
	}
	global.GVA_LOG.Info("默认菜单按钮权限初始化完成")

	// 为管理员角色分配所有菜单按钮权限
	if adminUser.ID > 0 && adminRole.ID > 0 {
		var roleMenuButtons []model.RoleMenuButton
		global.GVA_DB.Where("role_id = ?", adminRole.ID).Find(&roleMenuButtons)

		if len(roleMenuButtons) == 0 {
			var allButtons []model.MenuButton
			global.GVA_DB.Find(&allButtons)

			for _, button := range allButtons {
				roleMenuButton := model.RoleMenuButton{
					RoleID:       adminRole.ID,
					MenuButtonID: button.ID,
				}
				if err := global.GVA_DB.Create(&roleMenuButton).Error; err != nil {
					global.GVA_LOG.Error("分配管理员菜单按钮权限失败", zap.Error(err))
				}
			}
			global.GVA_LOG.Info("管理员菜单按钮权限分配成功")
		}
	}
}

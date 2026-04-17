package service

import (
	"errors"

	"devops-backend/global"
	"devops-backend/model"
	"devops-backend/model/request"
	"devops-backend/utils"
	"go.uber.org/zap"
)

type UserService struct{}

func (s *UserService) Login(req request.LoginRequest) (string, *model.User, error) {
	var user model.User
	if err := global.GVA_DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		global.GVA_LOG.Error("用户不存在", zap.Error(err))
		return "", nil, errors.New("用户不存在")
	}

	if !utils.CheckPassword(req.Password, user.Password) {
		return "", nil, errors.New("密码错误")
	}

	j := utils.NewJWT()
	token, err := j.CreateToken(user.ID)
	if err != nil {
		return "", nil, err
	}

	return token, &user, nil
}

func (s *UserService) GetUserRoles(userID uint) ([]model.Role, error) {
	var userRoles []model.UserRole
	if err := global.GVA_DB.Where("user_id = ?", userID).Find(&userRoles).Error; err != nil {
		return nil, err
	}

	if len(userRoles) == 0 {
		return []model.Role{}, nil
	}

	roleIDs := make([]uint, len(userRoles))
	for i, ur := range userRoles {
		roleIDs[i] = ur.RoleID
	}

	var roles []model.Role
	if err := global.GVA_DB.Where("id IN ?", roleIDs).Find(&roles).Error; err != nil {
		return nil, err
	}

	return roles, nil
}

func (s *UserService) ChangePassword(userID uint, oldPassword, newPassword string) error {
	var user model.User
	if err := global.GVA_DB.First(&user, userID).Error; err != nil {
		return err
	}

	if !utils.CheckPassword(oldPassword, user.Password) {
		return errors.New("原密码错误")
	}

	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		return errors.New("密码加密失败")
	}

	user.Password = hashedPassword
	return global.GVA_DB.Save(&user).Error
}

func (s *UserService) GetUserMenus(userID uint) ([]model.Menu, error) {
	var userRoles []model.UserRole
	if err := global.GVA_DB.Where("user_id = ?", userID).Find(&userRoles).Error; err != nil {
		return nil, err
	}

	if len(userRoles) == 0 {
		return []model.Menu{}, nil
	}

	roleIDs := make([]uint, len(userRoles))
	for i, ur := range userRoles {
		roleIDs[i] = ur.RoleID
	}

	var roleMenus []model.RoleMenu
	if err := global.GVA_DB.Where("role_id IN ?", roleIDs).Find(&roleMenus).Error; err != nil {
		return nil, err
	}

	if len(roleMenus) == 0 {
		return []model.Menu{}, nil
	}

	menuIDs := make([]uint, 0)
	menuIDMap := make(map[uint]bool)
	for _, rm := range roleMenus {
		if !menuIDMap[rm.MenuID] {
			menuIDMap[rm.MenuID] = true
			menuIDs = append(menuIDs, rm.MenuID)
		}
	}

	var menus []model.Menu
	if err := global.GVA_DB.Where("id IN ?", menuIDs).Where("is_show = ?", true).Order("sort asc, id asc").Find(&menus).Error; err != nil {
		return nil, err
	}

	var roleMenuButtons []model.RoleMenuButton
	if err := global.GVA_DB.Where("role_id IN ?", roleIDs).Find(&roleMenuButtons).Error; err != nil {
		return nil, err
	}

	menuButtonIDs := make(map[uint][]uint)
	for _, rmb := range roleMenuButtons {
		menuButtonIDs[rmb.MenuButtonID] = append(menuButtonIDs[rmb.MenuButtonID], rmb.RoleID)
	}

	var allButtons []model.MenuButton
	global.GVA_DB.Find(&allButtons)

	buttonMap := make(map[uint]model.MenuButton)
	for _, btn := range allButtons {
		buttonMap[btn.ID] = btn
	}

	userButtons := make(map[uint][]model.MenuButton)
	for btnID := range menuButtonIDs {
		if btn, ok := buttonMap[btnID]; ok {
			userButtons[btn.MenuID] = append(userButtons[btn.MenuID], btn)
		}
	}

	for i := range menus {
		if buttons, ok := userButtons[menus[i].ID]; ok {
			menus[i].Buttons = buttons
		}
	}

	parentMenus := make([]model.Menu, 0)
	for i := range menus {
		if menus[i].ParentID == 0 {
			s.buildMenuTreeWithButtons(&menus[i], menus, userButtons)
			parentMenus = append(parentMenus, menus[i])
		}
	}

	return parentMenus, nil
}

func (s *UserService) buildMenuTreeWithButtons(menu *model.Menu, allMenus []model.Menu, userButtons map[uint][]model.MenuButton) {
	children := make([]model.Menu, 0)
	for i := range allMenus {
		if allMenus[i].ParentID == menu.ID {
			s.buildMenuTreeWithButtons(&allMenus[i], allMenus, userButtons)
			children = append(children, allMenus[i])
		}
	}
	if len(children) > 0 {
		menu.Children = children
	}
	if buttons, ok := userButtons[menu.ID]; ok {
		menu.Buttons = buttons
	}
}

func (s *UserService) buildMenuTree(menu *model.Menu, allMenus []model.Menu) {
	children := make([]model.Menu, 0)
	for i := range allMenus {
		if allMenus[i].ParentID == menu.ID {
			s.buildMenuTree(&allMenus[i], allMenus)
			children = append(children, allMenus[i])
		}
	}
	if len(children) > 0 {
		menu.Children = children
	}
}

func (s *UserService) GetUserApis(userID uint) ([]uint, error) {
	var userRoles []model.UserRole
	if err := global.GVA_DB.Where("user_id = ?", userID).Find(&userRoles).Error; err != nil {
		return nil, err
	}

	if len(userRoles) == 0 {
		return []uint{}, nil
	}

	roleIDs := make([]uint, len(userRoles))
	for i, ur := range userRoles {
		roleIDs[i] = ur.RoleID
	}

	var roleApis []model.RoleApi
	if err := global.GVA_DB.Where("role_id IN ?", roleIDs).Find(&roleApis).Error; err != nil {
		return nil, err
	}

	apiIDs := make([]uint, 0)
	apiIDMap := make(map[uint]bool)
	for _, ra := range roleApis {
		if !apiIDMap[ra.ApiID] {
			apiIDMap[ra.ApiID] = true
			apiIDs = append(apiIDs, ra.ApiID)
		}
	}

	return apiIDs, nil
}

func (s *UserService) GetUserList() ([]model.User, error) {
	var users []model.User
	if err := global.GVA_DB.Find(&users).Error; err != nil {
		return nil, err
	}

	for i := range users {
		var userRoles []model.UserRole
		global.GVA_DB.Where("user_id = ?", users[i].ID).Find(&userRoles)

		if len(userRoles) > 0 {
			roleIDs := make([]uint, len(userRoles))
			for j, ur := range userRoles {
				roleIDs[j] = ur.RoleID
			}

			var roles []model.Role
			global.GVA_DB.Where("id IN ?", roleIDs).Find(&roles)
			users[i].Roles = roles
		}
	}

	return users, nil
}

func (s *UserService) GetUserByID(userID uint) (*model.User, error) {
	var user model.User
	if err := global.GVA_DB.First(&user, userID).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *UserService) CreateUser(req request.UserRequest) error {
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return errors.New("密码加密失败")
	}

	user := model.User{
		Username: req.Username,
		RealName: req.RealName,
		Phone:    req.Phone,
		Email:    req.Email,
		Password: hashedPassword,
	}
	return global.GVA_DB.Create(&user).Error
}

func (s *UserService) UpdateUser(id uint, req request.UserRequest) error {
	var user model.User
	if err := global.GVA_DB.First(&user, id).Error; err != nil {
		return err
	}

	user.Username = req.Username
	user.RealName = req.RealName
	user.Phone = req.Phone
	user.Email = req.Email
	if req.Password != "" {
		hashedPassword, err := utils.HashPassword(req.Password)
		if err != nil {
			return errors.New("密码加密失败")
		}
		user.Password = hashedPassword
	}

	return global.GVA_DB.Save(&user).Error
}

func (s *UserService) DeleteUser(id uint) error {
	return global.GVA_DB.Delete(&model.User{}, id).Error
}

func (s *UserService) AssignRole(userID uint, roleID uint) error {
	userRole := model.UserRole{
		UserID: userID,
		RoleID: roleID,
	}
	return global.GVA_DB.Create(&userRole).Error
}

func (s *UserService) AssignRoles(userID uint, roleIDs []uint) error {
	if err := global.GVA_DB.Where("user_id = ?", userID).Delete(&model.UserRole{}).Error; err != nil {
		return err
	}

	if len(roleIDs) == 0 {
		return nil
	}

	userRoles := make([]model.UserRole, 0, len(roleIDs))
	for _, roleID := range roleIDs {
		userRoles = append(userRoles, model.UserRole{
			UserID: userID,
			RoleID: roleID,
		})
	}

	return global.GVA_DB.Create(&userRoles).Error
}

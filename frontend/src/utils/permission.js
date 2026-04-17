import request from './request'

const permissionsCache = new Map()

export async function fetchUserPermissions() {
  try {
    const menusStr = localStorage.getItem('menus')
    if (!menusStr) return []
    
    const menus = JSON.parse(menusStr)
    const permissions = []
    
    function extractButtons(menuList) {
      for (const menu of menuList) {
        if (menu.buttons && menu.buttons.length > 0) {
          permissions.push(...menu.buttons)
        }
        if (menu.children && menu.children.length > 0) {
          extractButtons(menu.children)
        }
      }
    }
    
    extractButtons(menus)
    
    permissions.forEach(p => {
      permissionsCache.set(p.code, true)
    })
    
    localStorage.setItem('permissions', JSON.stringify(permissions))
    return permissions
  } catch (error) {
    console.error('获取权限失败:', error)
    return []
  }
}

export function hasPermission(code) {
  if (!code) return true
  
  const cached = permissionsCache.get(code)
  if (cached !== undefined) return cached
  
  const permissions = JSON.parse(localStorage.getItem('permissions') || '[]')
  return permissions.some(p => p.code === code)
}

export function checkPermission(code) {
  return hasPermission(code)
}

export function clearPermissionsCache() {
  permissionsCache.clear()
  localStorage.removeItem('permissions')
}

export function setPermissions(permissions) {
  localStorage.setItem('permissions', JSON.stringify(permissions))
  permissions.forEach(p => {
    permissionsCache.set(p.code, true)
  })
}

export default {
  hasPermission,
  checkPermission,
  fetchUserPermissions,
  clearPermissionsCache,
  setPermissions
}
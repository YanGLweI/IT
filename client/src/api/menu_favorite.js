import request from './request'

// 获取当前用户的菜单收藏列表
export function getMenuFavorites() {
  return request.get('/menu-favorites')
}

// 切换菜单收藏状态
export function toggleMenuFavorite(data) {
  return request.put('/menu-favorites/toggle', data)
}

// 批量更新收藏排序
export function reorderMenuFavorites(menuIndices) {
  return request.put('/menu-favorites/reorder', { menu_indices: menuIndices })
}

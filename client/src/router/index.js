import Vue from 'vue'
import VueRouter from 'vue-router'
import Layout from '../views/Layout.vue'

Vue.use(VueRouter)

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('../views/login/Login.vue'),
    meta: { title: '登录', public: true }
  },
  {
    path: '/',
    component: Layout,
    redirect: '/dashboard',
    children: [
      { path: 'dashboard', name: 'Dashboard', component: () => import('../views/dashboard/Dashboard.vue'), meta: { title: '看板' } },
      { path: 'regions', name: 'Regions', component: () => import('../views/region/RegionList.vue'), meta: { title: '区域管理' } },
      { path: 'os-types', name: 'OSTypes', component: () => import('../views/ostype/OSTypeList.vue'), meta: { title: '操作系统管理' } },
      { path: 'assets', name: 'Assets', component: () => import('../views/asset/AssetList.vue'), meta: { title: '资产管理' } },
      { path: 'policies', name: 'Policies', component: () => import('../views/policy/PolicyList.vue'), meta: { title: 'IT政策' } },
      { path: 'topology', name: 'Topology', component: () => import('../views/topology/TopologyView.vue'), meta: { title: '网络拓扑图' } },
      { path: 'permissions', name: 'Permissions', component: () => import('../views/permission/PermissionList.vue'), meta: { title: '岗位权限' } }
    ]
  }
]

const router = new VueRouter({
  mode: 'history',
  routes
})

// 路由守卫：检查登录状态
router.beforeEach((to, from, next) => {
  // 设置页面标题
  document.title = to.meta.title ? `${to.meta.title} - IT管理平台` : 'IT管理平台'

  // 公开页面直接放行
  if (to.meta.public) {
    next()
    return
  }

  // 检查是否已登录
  const token = localStorage.getItem('token')
  if (!token) {
    next('/login')
    return
  }

  next()
})

export default router

import Vue from 'vue'
import VueRouter from 'vue-router'
import Layout from '../views/Layout.vue'

Vue.use(VueRouter)

const routes = [
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
      { path: 'topology', name: 'Topology', component: () => import('../views/topology/TopologyView.vue'), meta: { title: '网络拓扑图' } }
    ]
  }
]

const router = new VueRouter({
  mode: 'history',
  routes
})

export default router

<template>
  <el-container style="height: 100vh">
    <el-aside width="200px" style="background-color: #304156">
      <div style="padding: 20px; text-align: center; color: #fff; font-size: 18px; font-weight: bold">
        IT管理平台
      </div>
      <el-menu
        :default-active="activeMenu"
        background-color="#304156"
        text-color="#bfcbd9"
        active-text-color="#409EFF"
        router
      >
        <el-menu-item index="/dashboard">
          <i class="el-icon-data-board"></i>
          <span slot="title">看板</span>
        </el-menu-item>
        <el-submenu index="asset">
          <template slot="title">
            <i class="el-icon-monitor"></i>
            <span>资产管理</span>
          </template>
          <el-menu-item index="/assets">
            <i class="el-icon-s-order"></i>
            <span>资产列表</span>
          </el-menu-item>
          <el-menu-item index="/regions">
            <i class="el-icon-place"></i>
            <span>区域管理</span>
          </el-menu-item>
          <el-menu-item index="/os-types">
            <i class="el-icon-s-platform"></i>
            <span>操作系统管理</span>
          </el-menu-item>
        </el-submenu>
        <el-menu-item index="/policies">
          <i class="el-icon-document"></i>
          <span slot="title">IT政策</span>
        </el-menu-item>
        <el-menu-item index="/topology">
          <i class="el-icon-share"></i>
          <span slot="title">网络拓扑图</span>
        </el-menu-item>
      </el-menu>
    </el-aside>
    <el-container>
      <el-header style="background: #fff; border-bottom: 1px solid #e6e6e6; display: flex; align-items: center; padding: 0 20px">
        <h3 style="margin: 0; color: #333">{{ $route.meta.title }}</h3>
      </el-header>
      <el-main style="background: #f0f2f5; padding: 20px">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script>
export default {
  name: 'Layout',
  computed: {
    activeMenu() {
      return this.$route.path
    }
  }
}
</script>

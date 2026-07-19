const { defineConfig } = require('@vue/cli-service')

module.exports = defineConfig({
  transpileDependencies: true,
  lintOnSave: false,
  devServer: {
    port: 8081,
    proxy: {
      '/api': {
        target: 'https://localhost:8080',
        changeOrigin: true,
        secure: false
      },
      '/uploads/it_guide_media': {
        target: 'https://localhost:8080',
        changeOrigin: true,
        secure: false
      },
      '/uploads/dedicated_lines': {
        target: 'https://localhost:8080',
        changeOrigin: true,
        secure: false
      },
      '/uploads/ipsec_vpn': {
        target: 'https://localhost:8080',
        changeOrigin: true,
        secure: false
      }
    }
  }
})

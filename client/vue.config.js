const { defineConfig } = require('@vue/cli-service')
const path = require('path')

module.exports = defineConfig({
  transpileDependencies: true,
  lintOnSave: false,
  outputDir: 'dist',
  assetsDir: 'static/file-viewer',
  configureWebpack: {
    resolve: {
      fallback: {
        "path": require.resolve("path-browserify"),
        "fs": false,
        "zlib": require.resolve("browserify-zlib"),
        "util": require.resolve("util")
      }
    },
    module: {
      rules: [{ test: /\.wasm$/, type: "webassembly/async" }]
    },
    experiments: {
      asyncWebAssembly: true
    }
  },
  devServer: {
    port: 8081,
    static: {
      directory: path.join(__dirname, 'node_modules/@file-viewer')
    },
    proxy: {
      '/api': {
        target: 'https://localhost:9080',
        changeOrigin: true,
        secure: false
      },
      '/uploads': {
        target: 'https://localhost:9080',
        changeOrigin: true,
        secure: false
      },
      '/uploads/it_guide_media': {
        target: 'https://localhost:9080',
        changeOrigin: true,
        secure: false
      },
      '/uploads/dedicated_lines': {
        target: 'https://localhost:9080',
        changeOrigin: true,
        secure: false
      },
      '/uploads/ipsec_vpn': {
        target: 'https://localhost:9080',
        changeOrigin: true,
        secure: false
      }
    }
  }
})

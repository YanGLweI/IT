const { defineConfig } = require('@vue/cli-service')
const path = require('path')

// 增加 EventEmitter 最大监听器限制，避免 MaxListenersExceededWarning
// 这是在开发服务器启动前设置的，用于处理 http-proxy-middleware 等库的升级监听器
if (typeof process !== 'undefined' && typeof process.setMaxListeners === 'function') {
  // 使用 setMaxListeners 而不是 eventEmitter listenerCount 来避免副作用
  process.setMaxListeners(50)
}

// 全局事件发射器最大监听器限制
const { EventEmitter } = require('events');
if (typeof EventEmitter !== 'undefined' && typeof EventEmitter.defaultMaxListeners !== 'undefined') {
  EventEmitter.defaultMaxListeners = 50;
}

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
        "util": require.resolve("util"),
        "process": require.resolve("process/browser")
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

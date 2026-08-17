import Vue from 'vue'
import ElementUI from 'element-ui'
import 'element-ui/lib/theme-chalk/index.css'
import './styles/dialog-theme.css'
import './styles/table-theme.css'
import './styles/filter-bar.css'
import './styles/header-theme.css'
import './styles/sidebar-theme.css'
import './styles/fonts.css'
import App from './App.vue'
import router from './router'
import { setDefaultFileViewerAssetBaseUrl } from '@file-viewer/core'

// file-viewer 运行时资源（pdf worker/cmaps/标准字体/wasm 等）统一存放在 public/file-viewer 下。
// 库的自动基址推断在本项目会误判为 /（产物目录 static/file-viewer + SPA 路由），
// 导致 /vendor/pdf/pdf.worker.mjs、/vendor/pdf/cmaps/*.bcmap 等 404：
// worker 降级为 fake worker，未内嵌 CJK 字体的 PDF 表格文字丢失。显式指定基址修复。
setDefaultFileViewerAssetBaseUrl('/file-viewer/')

// 修复 ResizeObserver loop 错误（file-viewer 1:1 缩放时容器尺寸快速变化触发）
// 这是浏览器已知的 ResizeObserver 限制，通过 requestAnimationFrame 包装回调解决
if (typeof ResizeObserver !== 'undefined') {
  const _OrigResizeObserver = ResizeObserver
  window.ResizeObserver = class extends _OrigResizeObserver {
    constructor(callback) {
      let rafId = null
      super((entries, observer) => {
        if (rafId) cancelAnimationFrame(rafId)
        rafId = requestAnimationFrame(() => {
          rafId = null
          callback(entries, observer)
        })
      })
    }
  }
}

Vue.use(ElementUI)
Vue.config.productionTip = false

new Vue({
  router,
  render: h => h(App)
}).$mount('#app')

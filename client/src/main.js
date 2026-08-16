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

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

// file-viewer 运行时资源（pdf worker/cmaps、标准字体/wasm 等）统一存放在 public/file-viewer 下。
// 库的自动基址推断在本项目会误判为 /（产物目录 static/file-viewer + SPA 路由），
// 导致 /vendor/pdf/pdf.worker.mjs、/vendor/pdf/cmaps/*.bcmap 等 404：
// worker 降级为 fake worker，未内嵌 CJK 字体的 PDF 表格文字丢失。显式指定基址修复。
setDefaultFileViewerAssetBaseUrl('/file-viewer/')

// 修复 ResizeObserver loop 错误（file-viewer 1:1 缩放时容器尺寸快速变化触发）
// 这是浏览器已知的 ResizeObserver 限制，通过 requestAnimationFrame 包装回调解决
// 仅作用于 file-viewer 容器，避免影响第三方库（如 echarts、cytoscape）的回调时序
if (typeof ResizeObserver !== 'undefined') {
  const _OrigResizeObserver = ResizeObserver
  window.ResizeObserver = class extends _OrigResizeObserver {
    constructor(callback) {
      let rafId = null
      super((entries, observer) => {
        // 仅对 file-viewer 容器的回调使用 rAF 包装，其余走原生
        const hasFvContainer = entries.some(entry =>
          entry.target && (entry.target.closest ? entry.target.closest('.fv-container, .file-viewer-web-shell, .file-viewer') : false)
        )
        if (hasFvContainer) {
          if (rafId) cancelAnimationFrame(rafId)
          rafId = requestAnimationFrame(() => {
            rafId = null
            callback(entries, observer)
          })
        } else {
          callback(entries, observer)
        }
      })
    }
  }
}

// 修复 file-viewer 图片预览放大后左边界无法滚动的问题
// 背景：内置 image 渲染器用 flex 居中（.image-stage{justify-content:center}），
// 图片放大超出容器宽度后负剩余空间被均分到左右两侧，而 overflow 容器的
// 可滚动区域只向右/下延伸，导致图片左侧永远滚动不到。
// 查看器内容渲染在 open Shadow DOM 中（styleIsolation 默认 auto），
// 文档级样式无法穿透 shadow 边界，只能把覆盖规则注入 shadow root 内部。
// 改为 flex-start 放置 + img auto 边距：未溢出时仍居中显示（视觉不变），
// 溢出时边距归零、图片贴左，左右边界均可滚动到达。
const FV_IMAGE_SCROLL_FIX_CSS = '.image-stage{justify-content:flex-start!important}.image-stage img{margin:auto!important}'
function applyFileViewerImageScrollFix(host) {
  // shadow root 由适配器 mounted 时的 mountViewer 同步挂载
  const shadowRoot = host.shadowRoot
  if (!shadowRoot || shadowRoot.querySelector('style[data-fv-image-scroll-fix]')) return
  const style = document.createElement('style')
  style.setAttribute('data-fv-image-scroll-fix', '')
  style.textContent = FV_IMAGE_SCROLL_FIX_CSS
  shadowRoot.appendChild(style)
}
if (typeof MutationObserver !== 'undefined') {
  const scanFileViewerHosts = root => {
    if (root.classList && root.classList.contains('ff-file-viewer-vue27')) applyFileViewerImageScrollFix(root)
    if (root.querySelectorAll) root.querySelectorAll('.ff-file-viewer-vue27').forEach(applyFileViewerImageScrollFix)
  }
  new MutationObserver(mutations => {
    for (const mutation of mutations) {
      mutation.addedNodes.forEach(node => {
        if (node.nodeType === Node.ELEMENT_NODE) scanFileViewerHosts(node)
      })
    }
  }).observe(document.body, { childList: true, subtree: true })
}

Vue.use(ElementUI)
Vue.config.productionTip = false

new Vue({
  router,
  render: h => h(App)
}).$mount('#app')

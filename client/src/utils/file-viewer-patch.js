import { setDefaultFileViewerAssetBaseUrl } from '@file-viewer/core'

// file-viewer 运行时资源（pdf worker/cmaps、标准字体/wasm 等）统一存放在 public/file-viewer 下。
// 库的自动基址推断在本项目会误判为 /（产物目录 static/file-viewer + SPA 路由），
// 导致 /vendor/pdf/pdf.worker.mjs、/vendor/pdf/cmaps/*.bcmap 等 404：
// worker 降级为 fake worker，未内嵌 CJK 字体的 PDF 表格文字丢失。显式指定基址修复。
setDefaultFileViewerAssetBaseUrl('/file-viewer/')

// 修复 ResizeObserver loop 错误（file-viewer 1:1 缩放时容器尺寸快速变化触发）
// 这是浏览器已知的 ResizeObserver 限制，通过 requestAnimationFrame 包装回调解决
// 仅作用于 file-viewer 容器，避免影响第三方库（如 echarts、cytoscape）的回调时序
function patchResizeObserverLoop() {
  if (typeof ResizeObserver === 'undefined') return
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

// file-viewer 图片预览交互补丁（内容在 open Shadow DOM 内，只能注入 shadow root）
// 1) 修复放大后左边界无法滚动：内置 image 渲染器用 justify-content:center 居中，
//    图片溢出后负剩余空间均分两侧，而 overflow 容器可滚动区域不向左延伸。
//    改为 flex-start + img auto 边距：未溢出时仍居中，溢出时贴左、左右均可滚动到达。
// 2) 屏蔽点击/回车打开内置 lightbox 遮罩层：库把 click/keydown 监听绑在 img 上，
//    在 shadowRoot 捕获阶段 stopPropagation 即可阻止其触发，多余遮罩层不再出现。
// 3) 按住图片拖拽平移画布：mousedown 记录 .image-viewer 滚动原点，
//    window mousemove 按位移改 scrollLeft/scrollTop（移出弹窗也不中断），放大后免用滚动条。
const FV_IMAGE_FIX_CSS = [
  '.image-stage{justify-content:flex-start!important}',
  '.image-stage img{margin:auto!important;cursor:grab!important}',
  '.image-stage img:active{cursor:grabbing!important}'
].join('')

function patchImagePreviewInteractions() {
  if (typeof MutationObserver === 'undefined') return
  const dragState = { current: null }
  window.addEventListener('mousemove', event => {
    const drag = dragState.current
    if (!drag) return
    drag.viewer.scrollLeft = drag.left - (event.clientX - drag.x)
    drag.viewer.scrollTop = drag.top - (event.clientY - drag.y)
  })
  window.addEventListener('mouseup', () => { dragState.current = null })

  function applyFileViewerImageFix(host) {
    // shadow root 由适配器 mounted 时的 mountViewer 同步挂载
    const shadowRoot = host.shadowRoot
    if (!shadowRoot || shadowRoot.querySelector('style[data-fv-image-fix]')) return
    const style = document.createElement('style')
    style.setAttribute('data-fv-image-fix', '')
    style.textContent = FV_IMAGE_FIX_CSS
    shadowRoot.appendChild(style)

    const isStageImage = target => target && target.closest && target.closest('.image-stage img')
    shadowRoot.addEventListener('click', event => {
      if (isStageImage(event.target)) {
        event.stopPropagation()
        event.preventDefault()
      }
    }, true)
    shadowRoot.addEventListener('keydown', event => {
      if (isStageImage(event.target) && (event.key === 'Enter' || event.key === ' ')) {
        event.stopPropagation()
        event.preventDefault()
      }
    }, true)
    shadowRoot.addEventListener('mousedown', event => {
      if (event.button !== 0 || !isStageImage(event.target)) return
      const viewer = event.target.closest('.image-viewer')
      if (!viewer) return
      dragState.current = {
        viewer,
        x: event.clientX,
        y: event.clientY,
        left: viewer.scrollLeft,
        top: viewer.scrollTop
      }
      // 阻止浏览器原生图片拖拽与文字选中
      event.preventDefault()
    })
  }

  const scanFileViewerHosts = root => {
    if (root.classList && root.classList.contains('ff-file-viewer-vue27')) applyFileViewerImageFix(root)
    if (root.querySelectorAll) root.querySelectorAll('.ff-file-viewer-vue27').forEach(applyFileViewerImageFix)
  }
  new MutationObserver(mutations => {
    for (const mutation of mutations) {
      mutation.addedNodes.forEach(node => {
        if (node.nodeType === Node.ELEMENT_NODE) scanFileViewerHosts(node)
      })
    }
  }).observe(document.body, { childList: true, subtree: true })
}

// 在应用挂载前统一初始化全部 file-viewer 补丁
export function setupFileViewerPatches() {
  patchResizeObserverLoop()
  patchImagePreviewInteractions()
}

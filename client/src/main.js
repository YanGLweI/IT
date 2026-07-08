import { createApp } from 'vue'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import './styles/dialog-theme.css'
import './styles/header-theme.css'
import App from './App.vue'
import router from './router'

const app = createApp(App)

// 全局注册所有 ElementPlus 图标组件
for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component)
}

// 挂载 $message/$confirm 到 globalProperties，保持 Options API 兼容
app.config.globalProperties.$message = ElMessage
app.config.globalProperties.$confirm = ElMessageBox.confirm

app.use(ElementPlus)
app.use(router)
app.mount('#app')

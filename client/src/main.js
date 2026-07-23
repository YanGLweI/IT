import Vue from 'vue'
import ElementUI from 'element-ui'
import 'element-ui/lib/theme-chalk/index.css'
import './styles/dialog-theme.css'
import './styles/table-theme.css'
import './styles/header-theme.css'
import './styles/sidebar-theme.css'
import './styles/fonts.css'
import App from './App.vue'
import router from './router'

Vue.use(ElementUI)
Vue.config.productionTip = false

new Vue({
  router,
  render: h => h(App)
}).$mount('#app')

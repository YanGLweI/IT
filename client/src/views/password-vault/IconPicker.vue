<template>
  <div class="icon-picker">
    <div v-for="group in iconGroups" :key="group.label" class="icon-group">
      <div class="icon-group-label">{{ group.label }}</div>
      <div class="icon-grid">
        <div
          v-for="icon in group.icons"
          :key="icon.name"
          class="icon-item"
          :class="{ active: value === icon.name }"
          @click="$emit('input', icon.name)"
        >
          <svg-icon :name="icon.name" :size="24" />
          <span class="icon-label">{{ icon.label }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import SvgIcon from '@/components/SvgIcon.vue'

export default {
  name: 'IconPicker',
  components: { SvgIcon },
  props: {
    value: { type: String, default: '' }
  },
  data() {
    return {
      iconGroups: [
        {
          label: '服务器类',
          icons: [
            { name: 'server', label: '服务器' },
            { name: 'database', label: '数据库' },
            { name: 'cloud', label: '云服务' }
          ]
        },
        {
          label: '网络类',
          icons: [
            { name: 'router', label: '路由器' },
            { name: 'switch', label: '交换机' },
            { name: 'firewall', label: '防火墙' }
          ]
        },
        {
          label: '应用类',
          icons: [
            { name: 'app', label: '应用系统' },
            { name: 'browser', label: '浏览器' },
            { name: 'terminal', label: '终端' }
          ]
        },
        {
          label: '工具类',
          icons: [
            { name: 'tool', label: '工具' },
            { name: 'key', label: '密钥' },
            { name: 'monitor', label: '监控' }
          ]
        },
        {
          label: '通用类',
          icons: [
            { name: 'desktop', label: '台式机' },
            { name: 'laptop', label: '笔记本' },
            { name: 'printer', label: '打印机' },
            { name: 'wifi', label: '无线网络' }
          ]
        }
      ]
    }
  },

}
</script>

<style scoped>
.icon-picker {
  max-height: 350px;
  overflow-y: auto;
}
.icon-group {
  margin-bottom: 16px;
}
.icon-group-label {
  font-size: 11px;
  color: #94a3b8;
  margin-bottom: 8px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}
.icon-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 8px;
}
.icon-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 10px 6px;
  border-radius: 10px;
  cursor: pointer;
  border: 1px solid transparent;
  transition: all 0.2s ease;
  color: #64748b;
}
.icon-item:hover {
  background: #f7f8fa;
  color: #3b82f6;
}
.icon-item.active {
  background: linear-gradient(135deg, #e8f4fd 0%, #f0f7ff 100%);
  border-color: #3b82f6;
  color: #3b82f6;
}
.icon-label {
  font-size: 11px;
  margin-top: 6px;
  white-space: nowrap;
  font-weight: 500;
}
</style>

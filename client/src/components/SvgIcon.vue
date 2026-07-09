<template>
  <svg
    class="menu-icon"
    :width="size"
    :height="size"
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    stroke-width="2"
    stroke-linecap="round"
    stroke-linejoin="round"
  >
    <template v-for="(path, i) in iconPaths">
      <path v-if="path.tag === 'path'" :key="i" :d="path.d" />
      <circle v-else-if="path.tag === 'circle'" :key="i" :cx="path.cx" :cy="path.cy" :r="path.r" />
      <line v-else-if="path.tag === 'line'" :key="i" :x1="path.x1" :y1="path.y1" :x2="path.x2" :y2="path.y2" />
      <polyline v-else-if="path.tag === 'polyline'" :key="i" :points="path.points" />
      <rect v-else-if="path.tag === 'rect'" :key="i" :x="path.x" :y="path.y" :width="path.w" :height="path.h" :rx="path.rx" />
      <ellipse v-else-if="path.tag === 'ellipse'" :key="i" :cx="path.cx" :cy="path.cy" :rx="path.rx" :ry="path.ry" />
      <polygon v-else-if="path.tag === 'polygon'" :key="i" :points="path.points" />
    </template>
  </svg>
</template>

<script>
const icons = {
  'bar-chart-2': [
    { tag: 'line', x1: 18, y1: 20, x2: 18, y2: 10 },
    { tag: 'line', x1: 12, y1: 20, x2: 12, y2: 4 },
    { tag: 'line', x1: 6, y1: 20, x2: 6, y2: 14 }
  ],
  'file-text': [
    { tag: 'path', d: 'M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z' },
    { tag: 'polyline', points: '14 2 14 8 20 8' },
    { tag: 'line', x1: 16, y1: 13, x2: 8, y2: 13 },
    { tag: 'line', x1: 16, y1: 17, x2: 8, y2: 17 },
    { tag: 'polyline', points: '10 9 9 9 8 9' }
  ],
  'monitor': [
    { tag: 'rect', x: 2, y: 3, w: 20, h: 14, rx: 2 },
    { tag: 'line', x1: 8, y1: 21, x2: 16, y2: 21 },
    { tag: 'line', x1: 12, y1: 17, x2: 12, y2: 21 }
  ],
  'list': [
    { tag: 'line', x1: 8, y1: 6, x2: 21, y2: 6 },
    { tag: 'line', x1: 8, y1: 12, x2: 21, y2: 12 },
    { tag: 'line', x1: 8, y1: 18, x2: 21, y2: 18 },
    { tag: 'line', x1: 3, y1: 6, x2: 3.01, y2: 6 },
    { tag: 'line', x1: 3, y1: 12, x2: 3.01, y2: 12 },
    { tag: 'line', x1: 3, y1: 18, x2: 3.01, y2: 18 }
  ],
  'map-pin': [
    { tag: 'path', d: 'M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z' },
    { tag: 'circle', cx: 12, cy: 10, r: 3 }
  ],
  'layers': [
    { tag: 'polygon', points: '12 2 2 7 12 12 22 7 12 2' },
    { tag: 'polyline', points: '2 17 12 22 22 17' },
    { tag: 'polyline', points: '2 12 12 17 22 12' }
  ],
  'shield': [
    { tag: 'path', d: 'M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z' }
  ],
  'network': [
    { tag: 'rect', x: 16, y: 16, w: 6, h: 6, rx: 1 },
    { tag: 'rect', x: 2, y: 16, w: 6, h: 6, rx: 1 },
    { tag: 'rect', x: 9, y: 2, w: 6, h: 6, rx: 1 },
    { tag: 'path', d: 'M5 16v-3a1 1 0 0 1 1-1h12a1 1 0 0 1 1 1v3' },
    { tag: 'line', x1: 12, y1: 12, x2: 12, y2: 8 }
  ],
  'git-branch': [
    { tag: 'line', x1: 6, y1: 3, x2: 6, y2: 15 },
    { tag: 'circle', cx: 18, cy: 6, r: 3 },
    { tag: 'circle', cx: 6, cy: 18, r: 3 },
    { tag: 'path', d: 'M18 9a9 9 0 0 1-9 9' }
  ],
  'search': [
    { tag: 'circle', cx: 11, cy: 11, r: 8 },
    { tag: 'line', x1: 21, y1: 21, x2: 16.65, y2: 16.65 }
  ],
  'crosshair': [
    { tag: 'circle', cx: 12, cy: 12, r: 10 },
    { tag: 'line', x1: 22, y1: 12, x2: 18, y2: 12 },
    { tag: 'line', x1: 6, y1: 12, x2: 2, y2: 12 },
    { tag: 'line', x1: 12, y1: 6, x2: 12, y2: 2 },
    { tag: 'line', x1: 12, y1: 22, x2: 12, y2: 18 }
  ],
  'shield-check': [
    { tag: 'path', d: 'M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z' },
    { tag: 'polyline', points: '9 12 11 14 15 10' }
  ],
  'pencil': [
    { tag: 'path', d: 'M17 3a2.828 2.828 0 1 1 4 4L7.5 20.5 2 22l1.5-5.5L17 3z' }
  ],
  'settings': [
    { tag: 'circle', cx: 12, cy: 12, r: 3 },
    { tag: 'path', d: 'M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06A1.65 1.65 0 0 0 9 4.68a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06A1.65 1.65 0 0 0 19.4 9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z' }
  ],
  'hammer': [
    { tag: 'path', d: 'M15 12l-8.5 8.5c-.83.83-2.17.83-3 0 0 0 0 0 0 0a2.12 2.12 0 0 1 0-3L12 9' },
    { tag: 'path', d: 'M17.64 15L22 10.64' },
    { tag: 'path', d: 'M20.91 11.7l-1.25-1.25c-.6-.6-.93-1.4-.93-2.25V6.5l-3.5-3.5H13.5c-.85 0-1.65-.33-2.25-.93L10 1.31' }
  ],
  'download': [
    { tag: 'path', d: 'M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4' },
    { tag: 'polyline', points: '7 10 12 15 17 10' },
    { tag: 'line', x1: 12, y1: 15, x2: 12, y2: 3 }
  ],
  'shield-alert': [
    { tag: 'path', d: 'M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z' },
    { tag: 'line', x1: 12, y1: 8, x2: 12, y2: 12 },
    { tag: 'line', x1: 12, y1: 16, x2: 12.01, y2: 16 }
  ],
  'database': [
    { tag: 'ellipse', cx: 12, cy: 5, rx: 9, ry: 3 },
    { tag: 'path', d: 'M21 12c0 1.66-4 3-9 3s-9-1.34-9-3' },
    { tag: 'path', d: 'M3 5v14c0 1.66 4 3 9 3s9-1.34 9-3V5' }
  ],
  'users': [
    { tag: 'path', d: 'M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2' },
    { tag: 'circle', cx: 9, cy: 7, r: 4 },
    { tag: 'path', d: 'M23 21v-2a4 4 0 0 0-3-3.87' },
    { tag: 'path', d: 'M16 3.13a4 4 0 0 1 0 7.75' }
  ],
  'key': [
    { tag: 'path', d: 'M21 2l-2 2m-7.61 7.61a5.5 5.5 0 1 1-7.778 7.778 5.5 5.5 0 0 1 7.777-7.777zm0 0L15.5 7.5m0 0l3 3L22 7l-3-3m-3.5 3.5L19 4' }
  ],
  'user-check': [
    { tag: 'path', d: 'M16 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2' },
    { tag: 'circle', cx: 8.5, cy: 7, r: 4 },
    { tag: 'polyline', points: '17 11 19 13 23 9' }
  ],
  'terminal': [
    { tag: 'polyline', points: '4 17 10 11 4 5' },
    { tag: 'line', x1: 12, y1: 19, x2: 20, y2: 19 }
  ],
  'calendar-check': [
    { tag: 'rect', x: 3, y: 4, w: 18, h: 18, rx: 2 },
    { tag: 'line', x1: 16, y1: 2, x2: 16, y2: 6 },
    { tag: 'line', x1: 8, y1: 2, x2: 8, y2: 6 },
    { tag: 'line', x1: 3, y1: 10, x2: 21, y2: 10 },
    { tag: 'polyline', points: '9 16 11 18 15 14' }
  ],
  'user-cog': [
    { tag: 'path', d: 'M16 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2' },
    { tag: 'circle', cx: 8.5, cy: 7, r: 4 },
    { tag: 'circle', cx: 19, cy: 17, r: 3 },
    { tag: 'path', d: 'M21.5 19.5l-1.1-1.1' },
    { tag: 'path', d: 'M19 14.5V13' },
    { tag: 'path', d: 'M22 17h-1.5' },
    { tag: 'path', d: 'M16.5 19.5l1.1-1.1' },
    { tag: 'path', d: 'M19 20.5V22' },
    { tag: 'path', d: 'M22 17h1.5' },
    { tag: 'path', d: 'M21.5 14.5l-1.1 1.1' },
    { tag: 'path', d: 'M16.5 14.5l1.1 1.1' }
  ],
  'layout-grid': [
    { tag: 'rect', x: 3, y: 3, w: 7, h: 7 },
    { tag: 'rect', x: 14, y: 3, w: 7, h: 7 },
    { tag: 'rect', x: 14, y: 14, w: 7, h: 7 },
    { tag: 'rect', x: 3, y: 14, w: 7, h: 7 }
  ],
  'check-circle': [
    { tag: 'path', d: 'M22 11.08V12a10 10 0 1 1-5.93-9.14' },
    { tag: 'polyline', points: '22 4 12 14.01 9 11.01' }
  ],
  'table': [
    { tag: 'path', d: 'M9 3H5a2 2 0 0 0-2 2v4m6-6h10a2 2 0 0 1 2 2v4M9 3v18m0 0h10a2 2 0 0 0 2-2V9M9 21H5a2 2 0 0 1-2-2V9m0 0h18' }
  ],
  'calendar': [
    { tag: 'rect', x: 3, y: 4, w: 18, h: 18, rx: 2 },
    { tag: 'line', x1: 16, y1: 2, x2: 16, y2: 6 },
    { tag: 'line', x1: 8, y1: 2, x2: 8, y2: 6 },
    { tag: 'line', x1: 3, y1: 10, x2: 21, y2: 10 }
  ],
  'scroll-text': [
    { tag: 'path', d: 'M8 21h12a2 2 0 0 0 2-2v-2H10v2a2 2 0 1 1-4 0V5a2 2 0 1 0-4 0v3h4' },
    { tag: 'path', d: 'M19 17V5a2 2 0 0 0-2-2H4' },
    { tag: 'line', x1: 8, y1: 9, x2: 16, y2: 9 },
    { tag: 'line', x1: 8, y1: 13, x2: 14, y2: 13 }
  ],
  'log-in': [
    { tag: 'path', d: 'M15 3h4a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2h-4' },
    { tag: 'polyline', points: '10 17 15 12 10 7' },
    { tag: 'line', x1: 15, y1: 12, x2: 3, y2: 12 }
  ],
  'file-search': [
    { tag: 'path', d: 'M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z' },
    { tag: 'polyline', points: '14 2 14 8 20 8' },
    { tag: 'circle', cx: 11.5, cy: 14.5, r: 2.5 },
    { tag: 'line', x1: 13.3, y1: 16.3, x2: 15, y2: 18 }
  ],
  'upload-cloud': [
    { tag: 'path', d: 'M16 16l-4-4-4 4' },
    { tag: 'path', d: 'M12 12v9' },
    { tag: 'path', d: 'M20.39 18.39A5 5 0 0 0 18 9h-1.26A8 8 0 1 0 3 16.3' },
    { tag: 'path', d: 'M16 16l-4-4-4 4' }
  ]
}

export default {
  name: 'SvgIcon',
  props: {
    name: { type: String, required: true },
    size: { type: [Number, String], default: 20 }
  },
  computed: {
    iconPaths() {
      return icons[this.name] || []
    }
  }
}
</script>

<style scoped>
.menu-icon {
  display: inline-block;
  vertical-align: middle;
  flex-shrink: 0;
}
</style>

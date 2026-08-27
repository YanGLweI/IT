<template>
  <div class="topology-graph" ref="pageRoot">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-left">
        <h2 class="page-title">拓扑图生成</h2>
        <p class="page-subtitle">根据防火墙树结构与区域挂靠关系自动生成网络拓扑图</p>
      </div>
      <div class="header-actions">
        <el-button size="small" icon="el-icon-rank" @click="expandAll">全部展开</el-button>
        <el-button size="small" icon="el-icon-refresh-left" @click="resetView">重置视图</el-button>
        <el-button size="small" icon="el-icon-refresh" @click="fetchAll" :loading="loading">刷新</el-button>
        <el-button size="small" icon="el-icon-download" @click="exportPNG" :loading="exportingPNG">导出PNG</el-button>
        <el-button size="small" icon="el-icon-download" @click="exportPDF" :loading="exporting">导出PDF</el-button>
      </div>
    </div>

    <!-- 图例 -->
    <div class="legend-bar" v-if="hasData">
      <span class="legend-item"><img class="legend-icon" :src="legendIcons.internet" alt="">互联网</span>
      <span class="legend-item"><img class="legend-icon" :src="legendIcons.firewall" alt="">防火墙</span>
      <span class="legend-item"><img class="legend-icon" :src="legendIcons.region2" alt="">二级区域</span>
      <span class="legend-item"><img class="legend-icon" :src="legendIcons.region3" alt="">三级区域</span>
      <span class="legend-item"><img class="legend-icon" :src="legendIcons.region4" alt="">四级区域</span>
      <span class="legend-item"><img class="legend-icon" :src="legendIcons.region5" alt="">高安全区域</span>
      <span class="legend-item"><img class="legend-icon" :src="legendIcons.linux" alt="">Linux资产</span>
      <span class="legend-item"><img class="legend-icon" :src="legendIcons.windows" alt="">Windows资产</span>
      <span class="legend-item"><img class="legend-icon" :src="legendIcons.vmware" alt="">VMware资产</span>
      <span class="legend-item"><i class="legend-dot" style="background: #d4a574"></i>其他资产</span>
      <span class="legend-tip">提示：点击区域节点可展开/收起其下资产，点击防火墙节点折叠子树，右侧按钮或滚轮缩放、拖动平移</span>
    </div>

    <div class="chart-wrap" v-show="hasData">
      <div ref="chartRef" class="chart-container"></div>
      <!-- 缩放控制：模拟滚轮事件，与滚轮缩放完全一致的整体缩放 -->
      <div class="zoom-controls">
        <el-tooltip content="放大" placement="left">
          <button class="zoom-btn" @click="zoomIn">+</button>
        </el-tooltip>
        <el-tooltip content="缩小" placement="left">
          <button class="zoom-btn" @click="zoomOut">−</button>
        </el-tooltip>
      </div>
    </div>
    <el-empty v-if="!loading && !hasData" description="暂无拓扑数据，请先在《拓扑关系配置》中配置防火墙与区域关系" />
  </div>
</template>

<script>
import * as echarts from 'echarts'
import * as zrender from 'zrender'
import { jsPDF } from 'jspdf'
import { getFirewallTopologyTree } from '@/api/firewall_topology'

// 互联网云朵形状：云轮廓 + 地球经纬线图标（1195x1024 viewBox，多段子路径拼接）
const CLOUD_PATH_D = [
  'M935.112817 810.777735H919.029795a21.5972 21.5972 0 0 1 0-42.734885h17.921081a213.214913 213.214913 0 0 0 216.891032-209.079279 209.998308 209.998308 0 0 0-146.125737-197.591406 20.218655 20.218655 0 0 1-14.244962-22.51623v-29.408953a261.004462 261.004462 0 0 0-34.463618-129.123686 21.5972 21.5972 0 0 1 5.973694-28.949439 22.056715 22.056715 0 0 1 28.949439 7.811753 303.739347 303.739347 0 0 1 39.977796 150.261372v17.461566a251.814164 251.814164 0 0 1 160.830214 232.055023 256.409313 256.409313 0 0 1-259.625917 251.814164zM275.708939 810.777735h-16.083022a261.004462 261.004462 0 0 1-158.073125-51.925183 21.137685 21.137685 0 0 1-4.595149-29.868469 22.056715 22.056715 0 0 1 29.868469-4.595149 220.107636 220.107636 0 0 0 132.799805 45.95149H275.708939a21.5972 21.5972 0 0 1 0 42.734886zm-219.188607-109.82406A20.218655 20.218655 0 0 1 38.599251 689.465802 243.083381 243.083381 0 0 1 0 558.963571a249.976104 249.976104 0 0 1 158.992155-229.757449 178.29178 178.29178 0 0 1 180.129839-175.53469 183.805959 183.805959 0 0 1 89.145891 22.975744 317.065279 317.065279 0 0 1 115.338239-127.285626 328.093637 328.093637 0 0 1 321.660428-13.785447 21.137685 21.137685 0 0 1 9.190298 28.489924 21.5972 21.5972 0 0 1-28.949438 9.190298 281.682632 281.682632 0 0 0-128.664172-30.327984A275.708939 275.708939 0 0 0 459.514898 216.165458a21.137685 21.137685 0 0 1-16.083022 13.785447 21.5972 21.5972 0 0 1-18.840111-3.67612 137.854469 137.854469 0 0 0-85.469771-28.949438 135.09738 135.09738 0 0 0-137.854469 132.34029 108.445516 108.445516 0 0 0 0 11.487873 20.67817 20.67817 0 0 1-14.244962 21.5972 209.998308 209.998308 0 0 0-144.287678 196.212861 204.943644 204.943644 0 0 0 31.706528 109.364546 19.299626 19.299626 0 0 1 2.75709 15.623506 22.51623 22.51623 0 0 1-9.190298 13.785447 25.273319 25.273319 0 0 1-11.487873 3.216605zM924.543974 124.721993a21.5972 21.5972 0 0 1-21.5972-21.5972 22.51623 22.51623 0 0 1 5.973694-14.244962 22.51623 22.51623 0 0 1 30.327983 0 21.5972 21.5972 0 0 1 6.433208 15.163992 21.137685 21.137685 0 0 1-21.137685 22.056715zM597.369367 523.121409a250.435619 250.435619 0 1 0 250.435619 249.976105A250.435619 250.435619 0 0 0 597.369367 523.121409zm0 464.569562a214.593457 214.593457 0 1 1 214.593457-214.593457A214.593457 214.593457 0 0 1 597.369367 987.690971z',
  'M597.369367 558.963571a212.295883 212.295883 0 0 0-70.765294 7.811754 280.304088 280.304088 0 0 0-79.036563 175.53469 275.708939 275.708939 0 0 0 106.147942 242.623866h15.163991a171.858572 171.858572 0 0 0 57.439363 0 229.757449 229.757449 0 0 1-142.909134-241.245321A235.731143 235.731143 0 0 1 597.369367 558.963571z',
  'M597.369367 558.963571a212.295883 212.295883 0 0 1 70.765294 7.811754 280.304088 280.304088 0 0 1 79.036563 175.53469 275.708939 275.708939 0 0 1-106.147942 242.623866h-15.163991a171.858572 171.858572 0 0 1-57.439363 0 229.757449 229.757449 0 0 0 142.909134-241.245321A235.731143 235.731143 0 0 0 597.369367 558.963571z',
  'M597.369367 666.030542a249.057075 249.057075 0 0 0 147.504282-48.708579 205.403159 205.403159 0 0 0-28.949438-22.975745 214.133942 214.133942 0 0 1-237.109688 0 205.403159 205.403159 0 0 0-28.949438 22.975745A249.057075 249.057075 0 0 0 597.369367 666.030542zM597.369367 881.083515a248.59756 248.59756 0 0 1 147.504282 45.951489 179.21081 179.21081 0 0 1-28.949438 23.43526 214.133942 214.133942 0 0 0-237.109688 0 179.21081 179.21081 0 0 1-28.949438-23.43526 248.59756 248.59756 0 0 1 147.504282-45.951489zM359.800165 755.635948h474.678889v35.842162H359.800165z'
].join(' ')
// 线稿图标下垫白色遮罩（云内白色椭圆 + 地球内白色圆，均不超出轮廓），遮挡穿过节点背后的连接线
const CLOUD_SYMBOL = 'image://data:image/svg+xml;charset=utf-8,' + encodeURIComponent(
  '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1195 1024">' +
  '<ellipse cx="597" cy="512" rx="535" ry="245" fill="#ffffff"/>' +
  '<circle cx="597" cy="773" r="214" fill="#ffffff"/>' +
  '<path d="' + CLOUD_PATH_D + '" fill="#334155"/>' +
  '</svg>'
)
// 防火墙砖墙图标：圆角方块 + 镂空砖块（1024x1024 viewBox）
const FIREWALL_PATH_D = 'M896 85.333333a42.666667 42.666667 0 0 1 42.666667 42.666667v768a42.666667 42.666667 0 0 1-42.666667 42.666667H128a42.666667 42.666667 0 0 1-42.666667-42.666667V128a42.666667 42.666667 0 0 1 42.666667-42.666667h768zM362.666667 682.666667H170.666667v170.666666h192v-170.666666zm490.666666 0H448v170.666666h405.333333v-170.666666zM576 426.666667H170.666667v170.666666h405.333333v-170.666666zm277.333333 0h-192v170.666666h192v-170.666666zM362.666667 170.666667H170.666667v170.666666h192V170.666667zm490.666666 0H448v170.666666h405.333333V170.666667z'
// 线稿图标下垫与外轮廓同尺寸的白色圆角矩形遮罩，遮挡穿过节点背后的连接线（砖缝填白）
const FIREWALL_SYMBOL = 'image://data:image/svg+xml;charset=utf-8,' + encodeURIComponent(
  '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1024 1024">' +
  '<rect x="85.333333" y="85.333333" width="853.333334" height="853.333334" rx="42.666667" fill="#ffffff"/>' +
  '<path d="' + FIREWALL_PATH_D + '" fill="#f87171"/>' +
  '</svg>'
)
// 区域图标：容器 + 交换箭头（1024x1024 viewBox），填充色按区域等级对应现有配色
const REGION_PATH_D = [
  'M789.311488 720.384A72.192 72.192 0 1 0 861.759488 793.6a72.192 72.192 0 0 0-72.448-73.216zm0 97.024a25.6 25.6 0 1 1 25.6-25.6 25.6 25.6 0 0 1-25.6 25.6zM356.159488 730.624a23.808 23.808 0 0 0-23.552 23.808v76.8a23.808 23.808 0 1 0 47.36 0v-76.8a23.808 23.808 0 0 0-23.808-23.808z',
  'M1024.319488 658.688L935.487488 118.272a104.704 104.704 0 0 0-25.6-53.76 79.616 79.616 0 0 0-57.6-23.808H173.375488a84.736 84.736 0 0 0-59.904 25.6 87.552 87.552 0 0 0-24.064 57.088L0.319488 655.104A43.264 43.264 0 0 0 0.319488 665.6v227.584a95.744 95.744 0 0 0 71.68 87.808H929.855488A94.976 94.976 0 0 0 1024.319488 888.32V665.6a23.808 23.808 0 0 0 0-6.912zM166.719488 131.584a29.44 29.44 0 0 0 0-6.656 5.376 5.376 0 0 1 1.792-3.84 6.912 6.912 0 0 1 4.096-2.048h678.912a20.992 20.992 0 0 1 5.12 11.52l72.704 470.272a36.352 36.352 0 0 0-7.168 0H99.903488a35.072 35.072 0 0 0-7.168 0zm763.136 772.608H94.527488a15.616 15.616 0 0 1-15.36-14.08v-194.56a16.128 16.128 0 0 1 16.384-15.36h834.56a14.336 14.336 0 0 1 11.008 4.864 15.104 15.104 0 0 1 4.864 10.496v192.768a15.872 15.872 0 0 1-16.128 15.872z',
  'M182.079488 730.624a23.808 23.808 0 0 0-23.552 23.808v76.8a24.064 24.064 0 0 0 23.552 25.6 25.6 25.6 0 0 0 24.064-25.6v-76.8a24.064 24.064 0 0 0-24.064-23.808zM272.447488 273.408l151.552 79.36a47.104 47.104 0 0 0 17.664 1.536c8.96-1.536 8.704-6.4 8.704-6.4V286.72h147.456s116.224 12.032 116.224 44.8c0 0-1.792-102.4-119.808-102.4h-143.872V162.816s0-3.584-8.704-5.12a25.6 25.6 0 0 0-14.592 2.56l-150.528 81.92s-14.08 5.376-14.08 16.128a17.408 17.408 0 0 0 9.984 15.104zM269.375488 730.624a23.808 23.808 0 0 0-23.552 23.808v76.8a23.552 23.552 0 1 0 47.104 0v-76.8a23.808 23.808 0 0 0-23.552-23.808zM612.415488 365.056c-10.752 2.304-11.264 6.4-11.264 6.4v61.184h-148.224s-116.224-12.032-116.224-45.056c0 0 2.048 103.68 120.064 103.68h144.384v65.536s1.536 3.84 11.264 4.864a18.944 18.944 0 0 0 11.52-2.56l150.784-81.664s14.08-5.632 14.08-16.64a17.152 17.152 0 0 0-9.984-14.848l-151.552-79.616a27.648 27.648 0 0 0-14.848-1.28z'
].join(' ')
// 白色遮罩：容器两处镂空区（不超出外轮廓），遮挡穿过节点背后的连接线
const REGION_MASK_D = [
  'M166.719488 131.584a29.44 29.44 0 0 0 0-6.656 5.376 5.376 0 0 1 1.792-3.84 6.912 6.912 0 0 1 4.096-2.048h678.912a20.992 20.992 0 0 1 5.12 11.52l72.704 470.272a36.352 36.352 0 0 0-7.168 0H99.903488a35.072 35.072 0 0 0-7.168 0z',
  'M929.855488 904.192H94.527488a15.616 15.616 0 0 1-15.36-14.08v-194.56a16.128 16.128 0 0 1 16.384-15.36h834.56a14.336 14.336 0 0 1 11.008 4.864 15.104 15.104 0 0 1 4.864 10.496v192.768a15.872 15.872 0 0 1-16.128 15.872z'
].join(' ')
// 按颜色生成并缓存区域 image symbol
const regionSymbolCache = {}
function regionSymbol(color) {
  if (!regionSymbolCache[color]) {
    regionSymbolCache[color] = 'image://data:image/svg+xml;charset=utf-8,' + encodeURIComponent(
      '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1024 1024">' +
      '<path d="' + REGION_MASK_D + '" fill="#ffffff"/>' +
      '<path d="' + REGION_PATH_D + '" fill="' + color + '"/>' +
      '</svg>'
    )
  }
  return regionSymbolCache[color]
}
// Windows 资产图标：四格方块（1024x1024 viewBox），下垫白色矩形遮罩挡住背后连线，填充色为现有 Windows 淡蓝 #93c5fd
const WINDOWS_SYMBOL = 'image://data:image/svg+xml;charset=utf-8,' + encodeURIComponent(
  '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1024 1024">' +
  '<rect x="96" y="96" width="832" height="832" fill="#ffffff"/>' +
  '<path d="M96 96h396.16v396.16H96V96zm435.84 0h396.16v396.16H531.84V96zM96 531.84h396.16v396.16H96V531.84zm435.84 0h396.16v396.16H531.84V531.84z" fill="#93c5fd"/>' +
  '</svg>'
)
// Linux 资产图标：红帽图标（1024x1024 viewBox），使用 SVG 自带配色（红帽 + 黑色阴影），下垫白色椭圆遮罩挡住背后连线
const LINUX_SYMBOL = 'image://data:image/svg+xml;charset=utf-8,' + encodeURIComponent(
  '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1024 1024">' +
  '<ellipse cx="520" cy="420" rx="225" ry="195" fill="#ffffff"/>' +
  '<ellipse cx="512" cy="610" rx="380" ry="180" fill="#ffffff"/>' +
  '<path d="M640.512 556.672c50.486857 0 123.574857-10.514286 123.574857-70.966857a57.179429 57.179429 0 0 0-1.28-13.897143l-30.098286-131.492571c-6.930286-28.946286-13.001143-42.057143-63.451428-67.456C630.089143 252.708571 544.768 219.428571 519.552 219.428571c-23.478857 0-30.336 30.482286-58.331429 30.482286-28.013714 0-46.976-22.747429-72.192-22.747428-25.234286 0-40.027429 16.603429-52.187428 50.797714q-33.938286 96.365714-38.326857 110.336c-0.694857 2.56-0.987429 5.248-0.896 7.899428 0 37.449143 146.541714 160.292571 342.875428 160.292572" fill="#F00001"/>' +
  '<path d="M771.803429 510.244571c6.985143 33.28 6.985143 36.754286 6.985142 41.161143 0 56.886857-63.524571 88.466286-147.053714 88.466286-188.708571 0.128-354.066286-111.213714-354.066286-184.795429 0-10.24 2.048-20.388571 6.089143-29.805714-67.84 3.419429-155.757714 15.634286-155.757714 93.714286C128 646.912 429.092571 804.571429 667.465143 804.571429c182.802286 0 228.882286-83.218286 228.882286-148.900572 0-51.730286-44.416-110.372571-124.452572-145.408" fill="#F00001"/>' +
  '<path d="M763.501714 475.867429v0zM763.538286 475.922286l-0.018286-0.054857v0.036571c0 0.164571 0.237714 2.249143 0.146286 5.650286 0-1.828571 0-3.620571-0.146286-5.485715a65.462857 65.462857 0 0 1-23.186286 52.205715c-88.045714 63.122286-232.064 8.045714-320.292571-33.371429-48.091429-23.332571-94.482286-54.308571-122.404572-101.12-90.532571 132.754286 127.616 210.870857 224.438858 235.593143 82.322286 21.522286 217.965714 43.209143 253.165714-59.52 8.685714-31.579429-0.603429-64.457143-11.702857-93.933714zm0.109714 0.457143c2.066286 7.296 3.657143 14.189714 5.028571 20.900571l-5.028571-20.900571z" fill="#000000"/>' +
  '</svg>'
)
// VMware 资产图标：蓝橙双圆角方块（1064x1024 viewBox），使用 SVG 自带配色，下垫两个白色圆角矩形遮罩挡住背后连线
const VMWARE_SYMBOL = 'image://data:image/svg+xml;charset=utf-8,' + encodeURIComponent(
  '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1064 1024">' +
  '<rect x="265.05216" y="23.38816" width="762.72" height="768.08" rx="120" fill="#ffffff"/>' +
  '<rect x="8.3968" y="247.3984" width="765.34" height="765.42" rx="120" fill="#ffffff"/>' +
  '<path d="M385.06496 23.38816a120.0128 120.0128 0 0 0-120.0128 120.0128v528.05632a120.0128 120.0128 0 0 0 120.0128 120.0128h522.6496a120.0128 120.0128 0 0 0 120.05376-120.0128V143.36a120.0128 120.0128 0 0 0-120.0128-120.0128H385.10592zM894.5664 584.66304c0 38.2976-31.04768 69.34528-69.34528 69.34528h-357.5808a69.34528 69.34528 0 0 1-69.38624-69.34528V230.15424c0-38.25664 31.04768-69.30432 69.34528-69.30432h357.5808c38.2976 0 69.34528 31.04768 69.34528 69.30432v354.5088z" fill="#068BEF"/>' +
  '<path d="M653.7216 247.3984H128.4096a120.0128 120.0128 0 0 0-120.0128 120.0128v525.39392a120.0128 120.0128 0 0 0 120.0128 120.0128h525.39392a120.0128 120.0128 0 0 0 120.0128-120.0128V367.4112a120.0128 120.0128 0 0 0-120.0128-120.0128zm-14.49984 559.9232c0 38.2976-31.04768 69.38624-69.34528 69.38624h-357.5808A69.34528 69.34528 0 0 1 142.9504 807.3216v-354.46784c0-38.2976 31.04768-69.34528 69.34528-69.34528h357.5808c38.2976 0 69.34528 31.00672 69.34528 69.30432v354.5088z" fill="#FF8B02"/>' +
  '<path d="M633.2416 654.00832h143.89248v137.4208h-143.93344z" fill="#068BEF"/>' +
  '<path d="M398.25408 383.50848L265.05216 460.3904V383.50848z" fill="#057EAF"/>' +
  '<path d="M639.22176 791.42912l134.5536 72.74496v-72.74496z" fill="#D77D06"/>' +
  '</svg>'
)
// 导出图图例：生成与节点图标一致的 graphic 子元素（图标 + 文字），其他资产仍为圆点
function buildLegendElements() {
  const t = (text, x) => ({ type: 'text', style: { text, x, y: 10, font: '12px sans-serif', fill: '#64748b' } })
  const img = (src, x, w) => ({ type: 'image', style: { image: src.replace(/^image:\/\//, ''), x, y: 0, width: w || 14, height: 14 } })
  return [
    img(CLOUD_SYMBOL, 0, 16), t('互联网', 22),
    img(FIREWALL_SYMBOL, 70), t('防火墙', 88),
    img(regionSymbol('#86efac'), 140), t('二级区域', 158),
    img(regionSymbol('#fde047'), 210), t('三级区域', 228),
    img(regionSymbol('#fdba74'), 280), t('四级区域', 298),
    img(regionSymbol('#f87171'), 350), t('高安全区域', 368),
    img(LINUX_SYMBOL, 440), t('Linux 资产', 458),
    img(WINDOWS_SYMBOL, 530), t('Windows 资产', 548),
    img(VMWARE_SYMBOL, 640), t('VMware 资产', 658),
    { type: 'rect', shape: { x: 750, y: 1, width: 12, height: 12 }, style: { fill: '#d4a574' } }, t('其他资产', 768)
  ]
}

export default {
  name: 'TopologyGraph',
  data() {
    return {
      loading: false,
      exporting: false,
      exportingPNG: false,
      nodes: [],
      chart: null
    }
  },
  computed: {
    hasData() {
      return this.nodes.some(n => !n.invalid)
    },
    // 页面图例图标：复用节点 image symbol（去掉 image:// 前缀作为 img src）
    legendIcons() {
      const url = s => s.replace(/^image:\/\//, '')
      return {
        internet: url(CLOUD_SYMBOL),
        firewall: url(FIREWALL_SYMBOL),
        region2: url(regionSymbol('#86efac')),
        region3: url(regionSymbol('#fde047')),
        region4: url(regionSymbol('#fdba74')),
        region5: url(regionSymbol('#f87171')),
        linux: url(LINUX_SYMBOL),
        windows: url(WINDOWS_SYMBOL),
        vmware: url(VMWARE_SYMBOL)
      }
    }
  },
  mounted() {
    this.fetchAll()
    this.bindExternalRoam()
    window.addEventListener('resize', this.handleResize)
  },
  beforeDestroy() {
    this.unbindExternalRoam()
    window.removeEventListener('resize', this.handleResize)
    if (this.chart) {
      this.chart.dispose()
      this.chart = null
    }
  },
  methods: {
    async fetchAll() {
      this.loading = true
      try {
        const res = await getFirewallTopologyTree()
        this.nodes = res.data || []
        if (this.hasData) {
          this.$nextTick(() => this.renderChart())
        }
      } catch (e) {
        console.error(e)
      } finally {
        this.loading = false
      }
    },
    // 组装 ECharts 树数据：互联网 -> 防火墙树 -> 区域 -> 资产
    buildTreeData() {
      const fwNode = (n) => {
        const name = n.asset.computer_name
        const ip = n.asset.ip_address
        const children = []
        // 挂靠区域（虚拟机挂到所属虚拟主机节点下，非虚拟机直接挂区域下）
        ;(n.regions || []).forEach(r => {
          const assetNode = a => ({
            id: 'asset:' + a.id,
            name: a.computer_name,
            nodeType: 'asset',
            osName: a.os_name || '',
            tooltip: `资产：${a.computer_name}<br/>IP：${a.ip_address || '无'}`,
            children: []
          })
          // 第一遍：非虚拟机构建节点并建立 id 映射
          const hostById = {}
          const regionChildren = (r.assets || [])
            .filter(a => !a.is_virtual_machine)
            .map(a => {
              const node = assetNode(a)
              hostById[a.id] = node
              return node
            })
          // 第二遍：虚拟机挂到同区域的主机节点下；主机不在本区域则兜底挂区域层并提示
          ;(r.assets || []).filter(a => a.is_virtual_machine).forEach(a => {
            const host = hostById[a.host_asset_id]
            const vmNode = assetNode(a)
            vmNode.tooltip += host
              ? `<br/>所属虚拟主机：${host.name}`
              : '<br/>所属虚拟主机：不在本区域或不存在'
            if (host) {
              host.children.push(vmNode)
            } else {
              regionChildren.push(vmNode)
            }
          })
          // 主机有虚拟机时，tooltip 追加虚拟机数（含兜底未挂上的情况已覆盖）
          regionChildren.forEach(h => {
            if (h.children.length > 0) {
              h.tooltip += `<br/>虚拟机数：${h.children.length}`
            }
          })
          children.push({
            // 稳定 id 供 ECharts 数据 diff 精确匹配，避免更新时重建节点导致抖动
            id: 'region:' + r.id,
            name: r.name,
            nodeType: 'region',
            networkLevel: r.network_level || 0,
            subnet: r.subnet || '',
            tooltip: `区域：${r.name}<br/>资产数：${(r.assets || []).length}${r.subnet ? '<br/>子网：' + r.subnet : ''}`,
            children: regionChildren
          })
        })
        // 下级防火墙
        this.nodes
          .filter(c => c.parent_id === n.id && !c.invalid)
          .sort((a, b) => a.sort_order - b.sort_order || a.id - b.id)
          .forEach(c => children.push(fwNode(c)))
        return {
          id: 'fw:' + n.id,
          name,
          nodeType: 'firewall',
          tooltip: `防火墙：${name}<br/>IP：${ip || '无'}${n.remark ? '<br/>备注：' + n.remark : ''}`,
          children
        }
      }
      const roots = this.nodes
        .filter(n => n.parent_id === 0 && !n.invalid)
        .sort((a, b) => a.sort_order - b.sort_order || a.id - b.id)
        .map(fwNode)
      return {
        id: 'internet',
        name: '互联网',
        nodeType: 'internet',
        tooltip: '互联网（虚拟根节点）',
        children: roots
      }
    },
    // 根据节点类型设置外观（自定义形状 + 胶囊标签，保证文字与背景对比度）
    decorateNode(data) {
      if (data.nodeType === 'internet') {
        // 互联网：云 + 地球线稿图标（image symbol 内置白色遮罩挡住背后连线），文字置于图标下方胶囊标签
        data.symbol = CLOUD_SYMBOL
        data.symbolSize = [84, 72]
        data.label = {
          position: 'bottom',
          distance: 6,
          color: '#334155',
          fontSize: 12,
          fontWeight: 600,
          backgroundColor: '#ffffff',
          borderColor: '#cbd5e1',
          borderWidth: 1,
          borderRadius: 4,
          padding: [3, 6]
        }
      } else if (data.nodeType === 'firewall') {
        // 防火墙：砖墙图标（image symbol 内置白色遮罩挡住背后连线，淡红 #f87171），整块图标均可点击，淡红胶囊标签
        data.symbol = FIREWALL_SYMBOL
        data.symbolSize = [38, 38]
        data.label = {
          position: 'left',
          distance: 8,
          color: '#b91c1c',
          fontSize: 12,
          fontWeight: 600,
          backgroundColor: '#fee2e2',
          borderColor: '#fca5a5',
          borderWidth: 1,
          borderRadius: 4,
          padding: [4, 8]
        }
      } else if (data.nodeType === 'region') {
        // 区域：根据网络等级设置不同配色（二级淡绿、三级淡黄、四级淡橙、五级淡红）
        const levelColors = {
          2: { dot: '#86efac', border: '#4ade80', label: '#14532d', bg: '#dcfce7', labelBorder: '#86efac' },
          3: { dot: '#fde047', border: '#facc15', label: '#713f12', bg: '#fef9c3', labelBorder: '#fde047' },
          4: { dot: '#fdba74', border: '#fb923c', label: '#7c2d12', bg: '#ffedd5', labelBorder: '#fdba74' },
          5: { dot: '#f87171', border: '#ef4444', label: '#b91c1c', bg: '#fee2e2', labelBorder: '#fca5a5' }
        }
        const colors = levelColors[data.networkLevel] || levelColors[2]
        // 区域：容器 + 交换箭头图标（image symbol 内置白色遮罩挡住背后连线），颜色对应网络等级
        data.symbol = regionSymbol(colors.dot)
        data.symbolSize = 26
        data.label = {
          position: 'left',
          distance: 6,
          color: colors.label,
          fontSize: 12,
          fontWeight: 500,
          backgroundColor: colors.bg,
          borderColor: colors.labelBorder,
          borderWidth: 1,
          borderRadius: 4,
          padding: [3, 6]
        }
      } else {
        // 资产：根据操作系统类型设置不同颜色（RHEL=Linux 灰色、Windows 淡蓝色、其他淡灰色）
        const osName = (data.osName || '').toLowerCase()
        let assetColor = '#d4a574' // 默认淡棕色（其他）
        if (osName.startsWith('rhel')) {
          assetColor = '#64748b' // Linux 灰色
          data.symbol = LINUX_SYMBOL // Linux 资产：红帽图标（SVG 自带配色）
        } else if (osName.startsWith('windows')) {
          assetColor = '#93c5fd' // Windows 淡蓝色
          data.symbol = WINDOWS_SYMBOL // Windows 资产：四格方块图标（内置白色遮罩）
        } else if (osName.startsWith('vmware')) {
          data.symbol = VMWARE_SYMBOL // VMware 资产：蓝橙双方块图标（SVG 自带配色）
        }
        // ESXi 虚拟主机标签置于图标左侧（参考区域标签设置：position left + 右对齐避免遮住图标），VCSA 及其余资产仍为右侧
        const isEsxiHost = osName.includes('esxi')
        data.symbolSize = 14
        data.itemStyle = { color: assetColor, borderColor: assetColor }
        data.label = {
          position: isEsxiHost ? 'left' : 'right',
          distance: isEsxiHost ? 6 : 5,
          color: '#1e293b',
          fontSize: 10,
          lineHeight: 14,
          align: isEsxiHost ? 'right' : 'left',
          backgroundColor: 'rgba(241, 245, 249, 0.95)',
          borderColor: '#e2e8f0',
          borderWidth: 1,
          borderRadius: 3,
          padding: [2, 4],
          formatter: () => data.name
        }
        const m = data.tooltip && data.tooltip.match(/IP：(.+?)(<br\/\>|$)/)
        if (m && m[1] !== '无') {
          data.label.formatter = () => `${data.name}  ${m[1]}`
        }
      }
      ;(data.children || []).forEach(c => this.decorateNode(c))
      return data
    },
    // 导出拓扑图为 PNG（离屏渲染高清大图，确保全部资产完整显示不遮挡）
    async exportPNG() {
      if (!this.treeData || this.exportingPNG) return
      this.exportingPNG = true
      try {
        // 创建离屏容器
        const container = document.createElement('div')
        container.style.cssText = 'position:fixed;left:-9999px;top:0;width:3200px;height:2200px;'
        document.body.appendChild(container)

        // 初始化临时 ECharts 实例
        const tmpChart = echarts.init(container, null, { renderer: 'canvas' })

        // 深拷贝树数据并重新 decorate（避免修改主图表数据）
        const exportData = JSON.parse(JSON.stringify(this.treeData))
        this.decorateNode(exportData)

        // 渲染配置：展开全部层级、关闭动画、放大符号与字体
        tmpChart.setOption({
          graphic: {
            elements: [
              {
                type: 'group',
                left: 'center',
                top: 20,
                children: buildLegendElements()
              }
            ]
          },
          series: [{
            type: 'tree',
            data: [exportData],
            orient: 'LR',
            top: '12%',
            bottom: '5%',
            left: '5%',
            right: '10%',
            roam: false,
            expandAndCollapse: false,
            initialTreeDepth: 10,
            symbol: 'circle',
            lineStyle: { color: '#cbd5e1', width: 1.5, curveness: 0.3 },
            emphasis: { focus: 'descendant' },
            animation: false
          }]
        }, true)

        // 等待渲染完成
        await new Promise(resolve => setTimeout(resolve, 500))

        // 在"防火墙→区域"连接线上绘制子网标签（导出图字号放大适配 3200px 画布）
        const tmpView = tmpChart.getViewOfSeriesModel(tmpChart.getModel().getSeriesByIndex(0))
        const subnetGroup = this.buildSubnetLabelGroup(tmpView, 18, 0.3)
        if (subnetGroup && tmpView._mainGroup) {
          tmpView._mainGroup.add(subnetGroup)
          tmpChart.getZr().refresh()
        }
        await new Promise(resolve => setTimeout(resolve, 200))

        // 获取高清 PNG（pixelRatio: 1.5 平衡清晰度与文件大小）
        const imgData = tmpChart.getDataURL({ type: 'png', pixelRatio: 1.5, backgroundColor: '#fff' })

        // 销毁临时图表和容器
        tmpChart.dispose()
        document.body.removeChild(container)

        // 创建下载链接
        const link = document.createElement('a')
        link.href = imgData
        const today = new Date()
        const dateStr = `${today.getFullYear()}-${String(today.getMonth() + 1).padStart(2, '0')}-${String(today.getDate()).padStart(2, '0')}`
        link.download = `网络拓扑图-${dateStr}.png`
        document.body.appendChild(link)
        link.click()
        document.body.removeChild(link)
      } catch (e) {
        console.error('exportPNG error:', e)
        this.$message.error('导出失败，请重试')
      } finally {
        this.exportingPNG = false
      }
    },
    // 导出拓扑图为 PDF（离屏渲染高清大图，确保全部资产完整显示不遮挡）
    async exportPDF() {
      if (!this.treeData || this.exporting) return
      this.exporting = true
      try {
        // 创建离屏容器
        const container = document.createElement('div')
        container.style.cssText = 'position:fixed;left:-9999px;top:0;width:3200px;height:2200px;'
        document.body.appendChild(container)

        // 初始化临时 ECharts 实例
        const tmpChart = echarts.init(container, null, { renderer: 'canvas' })

        // 深拷贝树数据并重新 decorate（避免修改主图表数据）
        const exportData = JSON.parse(JSON.stringify(this.treeData))
        this.decorateNode(exportData)

        // 渲染配置：展开全部层级、关闭动画、放大符号与字体
        tmpChart.setOption({
          graphic: {
            elements: [
              {
                type: 'group',
                left: 'center',
                top: 20,
                children: buildLegendElements()
              }
            ]
          },
          series: [{
            type: 'tree',
            data: [exportData],
            orient: 'LR',
            top: '12%',
            bottom: '5%',
            left: '5%',
            right: '10%',
            roam: false,
            expandAndCollapse: false,
            initialTreeDepth: 10,
            symbol: 'circle',
            lineStyle: { color: '#cbd5e1', width: 1.5, curveness: 0.3 },
            emphasis: { focus: 'descendant' },
            animation: false
          }]
        }, true)

        // 等待渲染完成
        await new Promise(resolve => setTimeout(resolve, 500))

        // 在“防火墙→区域”连接线上绘制子网标签（导出图字号放大适配 3200px 画布）
        const tmpView = tmpChart.getViewOfSeriesModel(tmpChart.getModel().getSeriesByIndex(0))
        const subnetGroup = this.buildSubnetLabelGroup(tmpView, 18, 0.3)
        if (subnetGroup && tmpView._mainGroup) {
          tmpView._mainGroup.add(subnetGroup)
          tmpChart.getZr().refresh()
        }
        await new Promise(resolve => setTimeout(resolve, 200))

        // 获取高清 PNG（pixelRatio: 1.5 平衡清晰度与文件大小）
        const imgData = tmpChart.getDataURL({ type: 'png', pixelRatio: 1.5, backgroundColor: '#fff' })

        // 销毁临时图表和容器
        tmpChart.dispose()
        document.body.removeChild(container)

        // 创建横向 A4 PDF
        const pdf = new jsPDF({ orientation: 'landscape', unit: 'mm', format: 'a4' })
        const pageWidth = pdf.internal.pageSize.getWidth()
        const pageHeight = pdf.internal.pageSize.getHeight()
        const margin = 10
        const availWidth = pageWidth - margin * 2
        const availHeight = pageHeight - margin * 2

        // 解析图片尺寸，等比缩放适配页面
        const imgProps = pdf.getImageProperties(imgData)
        const imgRatio = imgProps.width / imgProps.height
        const pageRatio = availWidth / availHeight
        let drawWidth, drawHeight
        if (imgRatio > pageRatio) {
          drawWidth = availWidth
          drawHeight = availWidth / imgRatio
        } else {
          drawHeight = availHeight
          drawWidth = availHeight * imgRatio
        }
        const x = (pageWidth - drawWidth) / 2
        const y = (pageHeight - drawHeight) / 2

        pdf.addImage(imgData, 'PNG', x, y, drawWidth, drawHeight)

        // 生成文件名：网络拓扑图-YYYY-MM-DD.pdf
        const today = new Date()
        const dateStr = `${today.getFullYear()}-${String(today.getMonth() + 1).padStart(2, '0')}-${String(today.getDate()).padStart(2, '0')}`
        pdf.save(`网络拓扑图-${dateStr}.pdf`)
      } catch (e) {
        console.error('exportPDF error:', e)
        this.$message.error('导出失败，请重试')
      } finally {
        this.exporting = false
      }
    },
    renderChart() {
      const el = this.$refs.chartRef
      if (!el) return
      if (!this.chart) {
        this.chart = echarts.init(el)
      }
      this.treeData = this.decorateNode(this.buildTreeData())
      this.chart.setOption({
        tooltip: {
          trigger: 'item',
          formatter: params => params.data && params.data.tooltip ? params.data.tooltip : params.name
        },
        series: [{
          type: 'tree',
          data: [this.treeData],
          orient: 'LR',
          top: '5%',
          bottom: '5%',
          left: '8%',
          right: '15%',
          roam: true,
          expandAndCollapse: true,
          // 默认展开到区域层，点击区域节点再展开其下资产，避免大量资产标签拥挤重叠
          initialTreeDepth: 2,
          symbol: 'circle',
          lineStyle: { color: '#cbd5e1', width: 1.5, curveness: 0.4 },
          emphasis: { focus: 'descendant' },
          animationDuration: 300,
          animationDurationUpdate: 300
        }]
      }, true)
      // 放行 roam 指针检查：使图表容器外部的代理滚轮/拖拽也能触发平移缩放
      this.allowExternalRoam()
      // 监听 finished 事件：每次渲染/动画完成后重新放行（节点展开/收起会重建视图重置 pointerChecker）并重绘子网标签
      this.chart.off('finished', this._onFinished)
      this._onFinished = () => {
        this.allowExternalRoam()
        this.renderSubnetLabels()
      }
      this.chart.on('finished', this._onFinished)
    },
    // 在“防火墙→区域”连接线上生成子网标签组（添加到 _mainGroup 内，随 roam 变换同步移动缩放）
    buildSubnetLabelGroup(view, fontSize, curveness) {
      if (!view || !view._data || !view._data.tree) return null
      const tree = view._data.tree
      const c = curveness || 0.4
      const group = new zrender.Group()
      tree.root.eachNode(node => {
        const parent = node.parentNode
        if (!parent || !node.hostTree.data) return
        const raw = node.hostTree.data.getRawDataItem(node.dataIndex)
        const pRaw = parent.hostTree.data.getRawDataItem(parent.dataIndex)
        if (!raw || !pRaw || raw.nodeType !== 'region' || pRaw.nodeType !== 'firewall') return
        const subnet = raw.subnet
        if (!subnet) return
        const l = node.getLayout()
        const pl = parent.getLayout()
        if (!l || !pl) return
        // 与 ECharts 边绘制同公式的三次贝塞尔曲线（LR 布局），取 t=0.5 精确落在曲线上
        const x1 = pl.x, y1 = pl.y, x2 = l.x, y2 = l.y
        const cpx1 = x1 + (x2 - x1) * c, cpy1 = y1
        const cpx2 = x2 + (x1 - x2) * c, cpy2 = y2
        // P(0.5) = 0.125·P0 + 0.375·P1 + 0.375·P2 + 0.125·P3
        const mx = 0.125 * x1 + 0.375 * cpx1 + 0.375 * cpx2 + 0.125 * x2
        const my = 0.125 * y1 + 0.375 * cpy1 + 0.375 * cpy2 + 0.125 * y2
        // 计算 t=0.5 处的切线方向（贝塞尔曲线导数）
        // B'(0.5) = 0.75·(P3 + P2 - P1 - P0)
        const dx = 0.75 * (x2 + cpx2 - cpx1 - x1)
        const dy = 0.75 * (y2 + cpy2 - cpy1 - y1)
        const angle = Math.atan2(dy, dx)
        // 标签贴着线，垂直偏移 8px（根据连接线方向决定上下）
        const offset = 8
        const nx = -Math.sin(angle) * offset
        const ny = Math.cos(angle) * offset
        const labelX = mx + nx
        const labelY = my + ny
        group.add(new zrender.Text({
          style: {
            text: subnet,
            fontSize,
            fill: '#475569',
            textAlign: 'center',
            textVerticalAlign: 'middle'
          },
          x: labelX,
          y: labelY,
          rotation: -angle
        }))
      })
      return group
    },
    // 主图表：重绘子网标签（先移除旧组再添加新组，避免重复叠加）
    renderSubnetLabels() {
      if (!this.chart) return
      try {
        const seriesModel = this.chart.getModel().getSeriesByIndex(0)
        const view = this.chart.getViewOfSeriesModel(seriesModel)
        if (!view) return
        if (this._subnetGroup) {
          view._mainGroup.remove(this._subnetGroup)
          this._subnetGroup = null
        }
        const g = this.buildSubnetLabelGroup(view, 11, 0.4)
        if (g) {
          // 标签加入 _mainGroup（与节点同一坐标系，随 roam 变换同步移动缩放）
          view._mainGroup.add(g)
          this._subnetGroup = g
          // 直接操作 zrender 元素后需手动触发重绘
          this.chart.getZr().refresh()
        }
      } catch (e) {
        console.warn('renderSubnetLabels failed', e)
      }
    },
    // 放行 RoamController 的指针命中检查（内部 API，失败不影响主功能）
    allowExternalRoam() {
      if (!this.chart) return
      try {
        const seriesModel = this.chart.getModel().getSeriesByIndex(0)
        const view = this.chart.getViewOfSeriesModel(seriesModel)
        if (view && view._controller) {
          view._controller.setPointerChecker(() => true)
        }
      } catch (e) {
        console.warn('allowExternalRoam failed', e)
      }
    },
    // 图表容器外（图例/头部/空白区域）的滚轮与拖拽代理：转投到图表，等效于在图内操作
    bindExternalRoam() {
      const root = this.$refs.pageRoot
      if (!root || this._roamBound) return
      this._roamBound = true
      this._onProxyWheel = e => this.proxyWheel(e)
      this._onProxyDown = e => this.proxyDown(e)
      this._onProxyMove = e => this.proxyMove(e)
      this._onProxyUp = e => this.proxyUp(e)
      root.addEventListener('wheel', this._onProxyWheel, { passive: false })
      root.addEventListener('mousedown', this._onProxyDown)
      document.addEventListener('mousemove', this._onProxyMove)
      document.addEventListener('mouseup', this._onProxyUp)
    },
    unbindExternalRoam() {
      const root = this.$refs.pageRoot
      if (root && this._onProxyWheel) {
        root.removeEventListener('wheel', this._onProxyWheel)
        root.removeEventListener('mousedown', this._onProxyDown)
        document.removeEventListener('mousemove', this._onProxyMove)
        document.removeEventListener('mouseup', this._onProxyUp)
      }
      this._roamBound = false
      this._proxyDragging = false
    },
    proxyWheel(e) {
      const chartEl = this.$refs.chartRef
      if (!this.chart || !chartEl || chartEl.contains(e.target) || this._dispatching) return
      e.preventDefault()
      this.zoomByWheel(e.deltaY < 0 ? 120 : -120)
    },
    proxyDown(e) {
      const chartEl = this.$refs.chartRef
      if (!this.chart || !chartEl || chartEl.contains(e.target) || e.button !== 0 || this._dispatching) return
      this._proxyDragging = true
      this.dispatchMouseToChart(chartEl, 'mousedown', e)
    },
    proxyMove(e) {
      if (!this._proxyDragging || !this.chart) return
      this.dispatchMouseToChart(this.$refs.chartRef, 'mousemove', e)
    },
    proxyUp(e) {
      if (!this._proxyDragging || !this.chart) return
      this._proxyDragging = false
      this.dispatchMouseToChart(this.$refs.chartRef, 'mouseup', e)
    },
    // 把原生鼠标事件按相同坐标转投到图表 DOM，走 zrender 完整事件管线
    dispatchMouseToChart(chartEl, type, e) {
      if (!chartEl || this._dispatching) return
      this._dispatching = true
      chartEl.dispatchEvent(new MouseEvent(type, {
        clientX: e.clientX,
        clientY: e.clientY,
        button: e.button,
        buttons: e.buttons,
        bubbles: true,
        cancelable: true
      }))
      this._dispatching = false
    },
    // 向画布中心派发合成 wheel 事件，走 ECharts 滚轮漫游的同一套处理逻辑，实现与滚轮完全一致的整体缩放
    zoomByWheel(deltaY) {
      if (!this.chart || this._dispatching) return
      this._dispatching = true
      const canvas = this.chart.getZr().painter.getViewportRoot()
      const rect = canvas.getBoundingClientRect()
      canvas.dispatchEvent(new WheelEvent('wheel', {
        deltaY,
        deltaMode: 0,
        clientX: rect.left + rect.width / 2,
        clientY: rect.top + rect.height / 2,
        bubbles: true,
        cancelable: true
      }))
      this._dispatching = false
    },
    // 放大（合成 wheel 事件经 Chrome 的 wheelDelta 兼容填充后符号与直觉相反，实测 deltaY>0 为放大）
    zoomIn() {
      this.zoomByWheel(120)
    },
    // 缩小
    zoomOut() {
      this.zoomByWheel(-120)
    },
    // 全部展开：所有节点 collapsed 置 false 后合并渲染（数据 diff 按稳定 id 精确匹配、瞬时更新，不会抖动）
    expandAll() {
      if (!this.chart || !this.treeData) return
      const expand = d => {
        d.collapsed = false
        ;(d.children || []).forEach(expand)
      }
      expand(this.treeData)
      this.chart.setOption({
        series: [{
          data: [this.treeData],
          // 瞬时更新，避免节点位置过渡动画引起抖动
          animationDurationUpdate: 0
        }]
      })
      // 恢复常规更新动画时长，保持后续节点折叠/展开动画
      this.chart.setOption({ series: [{ animationDurationUpdate: 300 }] })
    },
    // 重置视图：重新渲染以复位漫游/折叠状态
    resetView() {
      if (this.hasData) this.renderChart()
    },
    handleResize() {
      if (this.chart) this.chart.resize()
    }
  }
}
</script>

<style scoped>
.topology-graph {
  background: #fff;
  border-radius: 14px;
  border: 1px solid #e2e8f0;
  margin: 20px;
  padding: 24px;
  height: calc(100% - 85px);
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}
.page-title {
  font-size: 20px;
  font-weight: 600;
  color: #1e293b;
  margin: 0;
}
.page-subtitle {
  font-size: 13px;
  color: #64748b;
  margin: 4px 0 0;
}
.header-actions {
  display: flex;
  align-items: center;
  gap: 10px;
}

.legend-bar {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 10px 18px;
  font-size: 13px;
  color: #475569;
  margin-bottom: 10px;
}
.legend-item {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  white-space: nowrap;
}
.legend-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  display: inline-block;
}
.legend-icon {
  width: 16px;
  height: 16px;
  object-fit: contain;
  display: inline-block;
}
.legend-tip {
  margin-left: auto;
  color: #94a3b8;
  font-size: 12px;
}

.chart-wrap {
  position: relative;
  flex: 1;
  min-height: 400px;
}
.chart-container {
  position: absolute;
  inset: 0;
  border: 1px dashed #e2e8f0;
  border-radius: 10px;
  background: linear-gradient(180deg, #f8fafc 0%, #ffffff 100%);
}

/* 右侧缩放控制 */
.zoom-controls {
  position: absolute;
  right: 14px;
  top: 50%;
  transform: translateY(-50%);
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  background: rgba(255, 255, 255, 0.92);
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  padding: 6px;
  box-shadow: 0 2px 8px rgba(15, 23, 42, 0.08);
  z-index: 10;
}
.zoom-btn {
  width: 30px;
  height: 30px;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  background: #ffffff;
  color: #334155;
  font-size: 18px;
  line-height: 1;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.15s;
}
.zoom-btn:hover:not(:disabled) {
  background: #3b82f6;
  border-color: #3b82f6;
  color: #ffffff;
}
.zoom-btn:disabled {
  color: #cbd5e1;
  cursor: not-allowed;
}
</style>

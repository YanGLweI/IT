<template>
  <div id="app">
    <router-view />
  </div>
</template>

<script>
export default {
  name: 'App'
}
</script>

<style>
html, body, #app {
  margin: 0;
  padding: 0;
  height: 100%;
}
body {
  font-family: 'Inter', 'Noto Sans SC', -apple-system, BlinkMacSystemFont,
    'PingFang SC', 'Microsoft YaHei', sans-serif;
}
</style>

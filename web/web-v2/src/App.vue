<template>
  <div id="app" class="navicat-app">
    <router-view />
  </div>
</template>

<script setup>
// Navicat 风格主应用容器
</script>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

html, body {
  height: 100%;
  overflow: hidden;
}

#app {
  height: 100vh;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}

.navicat-app {
  height: 100%;
  display: flex;
  flex-direction: column;
  background: #f5f5f5;
}
</style>


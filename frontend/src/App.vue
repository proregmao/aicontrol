<template>
  <div id="app">
    <router-view />
  </div>
</template>

<script setup lang="ts">
// 应用根组件
</script>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

html, body {
  height: 100%;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
  background-color: #f5f5f5;
}

#app {
  min-height: 100vh;
  background-color: #f5f5f5;
  display: flex;
  align-items: flex-start;
  justify-content: center;
  padding-top: 0;
}

/* 全局样式 */
.page-container {
  max-width: 1400px; /* 最大宽度限制 */
  width: 100%; /* 响应式宽度 */
  min-width: 1200px; /* 最小宽度保证 */
  height: 1080px;
  margin: 0 auto;
  background: white;
  box-shadow: 0 0 20px rgba(0, 0, 0, 0.1);
  border-radius: 8px;
  overflow: hidden;
}

.main-layout {
  display: flex;
  height: 100%;
}

.sidebar {
  width: 144px;
  background: #001529;
  color: white;
  flex-shrink: 0;
}

.content-area {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.breadcrumb-nav {
  height: 60px;
  background: white;
  border-bottom: 1px solid #e8e8e8;
  display: flex;
  align-items: center;
  padding: 0 20px;
}

.main-content {
  flex: 1;
  padding: 15px; /* 减少padding */
  overflow-y: auto;
  overflow-x: hidden; /* 防止水平溢出 */
  max-width: calc(1400px - 144px - 30px); /* 容器宽度 - 侧边栏宽度 - padding */
}
</style>

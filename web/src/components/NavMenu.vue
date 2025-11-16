<template>
  <el-menu
    :default-active="activeIndex"
    router
    class="nav-menu"
  >
    <el-menu-item index="/connection">
      <el-icon><Connection /></el-icon>
      <span>连接管理</span>
    </el-menu-item>
    <el-menu-item index="/database-select">
      <el-icon><Grid /></el-icon>
      <span>选择数据库</span>
    </el-menu-item>
    <el-menu-item index="/tasks">
      <el-icon><List /></el-icon>
      <span>任务管理</span>
    </el-menu-item>
  </el-menu>
</template>

<style scoped>
.nav-menu {
  border: none;
  height: 100%;
}
</style>

<script setup>
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { Connection, Grid, List } from '@element-plus/icons-vue'

const route = useRoute()
const activeIndex = computed(() => route.path)
</script>


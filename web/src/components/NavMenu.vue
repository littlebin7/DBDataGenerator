<template>
  <el-menu
    :default-active="activeIndex"
    router
    class="nav-menu"
    :default-openeds="['connection', 'task', 'data', 'monitor']"
  >
    <el-sub-menu index="connection">
      <template #title>
        <el-icon><Connection /></el-icon>
        <span>连接管理</span>
      </template>
      <el-menu-item index="/connection">
        <el-icon><Setting /></el-icon>
        <span>连接配置</span>
      </el-menu-item>
      <el-menu-item index="/database-select">
        <el-icon><Grid /></el-icon>
        <span>选择数据库</span>
      </el-menu-item>
    </el-sub-menu>
    
    <el-sub-menu index="task">
      <template #title>
        <el-icon><List /></el-icon>
        <span>任务管理</span>
      </template>
      <el-menu-item index="/tasks">
        <el-icon><Operation /></el-icon>
        <span>任务列表</span>
      </el-menu-item>
      <el-menu-item index="/task-history">
        <el-icon><Clock /></el-icon>
        <span>任务历史</span>
      </el-menu-item>
      <el-menu-item index="/schedule">
        <el-icon><Timer /></el-icon>
        <span>定时任务</span>
      </el-menu-item>
    </el-sub-menu>

    <el-sub-menu index="data">
      <template #title>
        <el-icon><DataAnalysis /></el-icon>
        <span>数据管理</span>
      </template>
      <el-menu-item index="/cascade">
        <el-icon><Share /></el-icon>
        <span>级联生成</span>
      </el-menu-item>
      <el-menu-item index="/quality">
        <el-icon><DocumentChecked /></el-icon>
        <span>质量检查</span>
      </el-menu-item>
      <el-menu-item index="/template">
        <el-icon><Files /></el-icon>
        <span>模板管理</span>
      </el-menu-item>
      <el-menu-item index="/rollback">
        <el-icon><RefreshLeft /></el-icon>
        <span>数据回滚</span>
      </el-menu-item>
    </el-sub-menu>

    <el-sub-menu index="monitor">
      <template #title>
        <el-icon><Monitor /></el-icon>
        <span>监控</span>
      </template>
      <el-menu-item index="/monitor">
        <el-icon><DataLine /></el-icon>
        <span>系统监控</span>
      </el-menu-item>
      <el-menu-item index="/pool">
        <el-icon><Connection /></el-icon>
        <span>连接池监控</span>
      </el-menu-item>
    </el-sub-menu>
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
import {
  Connection,
  Grid,
  List,
  Clock,
  Setting,
  Operation,
  Timer,
  DataAnalysis,
  Share,
  DocumentChecked,
  Monitor,
  DataLine,
  Files,
  RefreshLeft
} from '@element-plus/icons-vue'

const route = useRoute()
const activeIndex = computed(() => route.path)
</script>


import { createRouter, createWebHistory } from 'vue-router'
import ConnectionView from '../views/ConnectionView.vue'
import DatabaseSelectView from '../views/DatabaseSelectView.vue'
import TaskView from '../views/TaskView.vue'
import TaskConfigView from '../views/TaskConfigView.vue'

const routes = [
  {
    path: '/',
    redirect: '/connection'
  },
  {
    path: '/connection',
    name: 'Connection',
    component: ConnectionView
  },
  {
    path: '/database-select',
    name: 'DatabaseSelect',
    component: DatabaseSelectView
  },
  {
    path: '/tasks',
    name: 'Tasks',
    component: TaskView
  },
  {
    path: '/task/config',
    name: 'TaskConfig',
    component: TaskConfigView
  },
  {
    path: '/task-history',
    name: 'TaskHistory',
    component: () => import('../views/TaskHistoryView.vue')
  },
  {
    path: '/monitor',
    name: 'Monitor',
    component: () => import('../views/MonitorView.vue')
  },
  {
    path: '/pool',
    name: 'Pool',
    component: () => import('../views/PoolView.vue')
  },
  {
    path: '/quality',
    name: 'Quality',
    component: () => import('../views/QualityView.vue')
  },
  {
    path: '/schedule',
    name: 'Schedule',
    component: () => import('../views/ScheduleView.vue')
  },
  {
    path: '/cascade',
    name: 'Cascade',
    component: () => import('../views/CascadeView.vue')
  },
  {
    path: '/template',
    name: 'Template',
    component: () => import('../views/TemplateView.vue')
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

export default router


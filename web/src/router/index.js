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
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

export default router


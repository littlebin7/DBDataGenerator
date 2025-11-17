<template>
  <el-tag :type="statusType" :size="size">
    {{ statusText }}
  </el-tag>
</template>

<script setup>
import { computed } from 'vue'
import { TASK_STATUS_TYPE, TASK_STATUS_TEXT } from '../utils/constants'

const props = defineProps({
  status: {
    type: String,
    required: true
  },
  size: {
    type: String,
    default: 'default'
  }
})

const statusType = computed(() => {
  return TASK_STATUS_TYPE[props.status] || 'info'
})

const statusText = computed(() => {
  return TASK_STATUS_TEXT[props.status] || props.status
})
</script>


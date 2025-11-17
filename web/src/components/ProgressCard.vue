<template>
  <el-card :shadow="shadow" :body-style="{ padding: '15px' }">
    <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 10px">
      <span style="font-weight: 500">{{ title }}</span>
      <el-tag v-if="status" :type="statusType" size="small">{{ statusText }}</el-tag>
    </div>
    <el-progress 
      :percentage="percentage" 
      :status="progressStatus"
      :stroke-width="strokeWidth"
    />
    <div v-if="showInfo" style="margin-top: 10px; font-size: 12px; color: #909399">
      <slot name="info">
        <span v-if="current !== undefined && total !== undefined">
          {{ formatNumber(current) }} / {{ formatNumber(total) }}
        </span>
      </slot>
    </div>
    <div v-if="showExtra" style="margin-top: 8px; font-size: 12px">
      <slot name="extra"></slot>
    </div>
  </el-card>
</template>

<script setup>
import { computed } from 'vue'
import { TASK_STATUS_TYPE, TASK_STATUS_TEXT } from '../utils/constants'
import { formatNumber } from '../utils/formatters'

const props = defineProps({
  title: {
    type: String,
    required: true
  },
  percentage: {
    type: Number,
    default: 0
  },
  status: {
    type: String,
    default: ''
  },
  current: {
    type: Number,
    default: undefined
  },
  total: {
    type: Number,
    default: undefined
  },
  showInfo: {
    type: Boolean,
    default: true
  },
  showExtra: {
    type: Boolean,
    default: false
  },
  shadow: {
    type: String,
    default: 'hover'
  },
  strokeWidth: {
    type: Number,
    default: 8
  }
})

const statusType = computed(() => {
  return props.status ? (TASK_STATUS_TYPE[props.status] || 'info') : ''
})

const statusText = computed(() => {
  return props.status ? (TASK_STATUS_TEXT[props.status] || props.status) : ''
})

const progressStatus = computed(() => {
  if (props.percentage >= 100) return 'success'
  if (props.status === 'error') return 'exception'
  return undefined
})
</script>


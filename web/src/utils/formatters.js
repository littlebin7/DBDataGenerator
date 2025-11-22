/**
 * 格式化工具函数
 */

/**
 * 格式化时间
 * @param {string|Date|number} timeStr - 时间字符串、Date 对象或时间戳（秒或毫秒）
 * @returns {string} 格式化后的时间字符串
 */
export function formatTime(timeStr) {
  if (!timeStr) return '-'
  
  let date
  // 如果是数字，直接作为时间戳处理
  if (typeof timeStr === 'number') {
    // 判断是秒还是毫秒（大于 10^12 的是毫秒，否则是秒）
    date = new Date(timeStr > 1e12 ? timeStr : timeStr * 1000)
  } else if (typeof timeStr === 'string') {
    // 如果是字符串，先尝试解析为数字（Unix 时间戳）
    const timestamp = parseInt(timeStr, 10)
    if (!isNaN(timestamp) && timestamp > 0) {
      // 判断是秒还是毫秒（大于 10^12 的是毫秒，否则是秒）
      date = new Date(timestamp > 1e12 ? timestamp : timestamp * 1000)
    } else {
      // 如果不是数字字符串，尝试直接解析为日期
      date = new Date(timeStr)
    }
  } else {
    // 如果是 Date 对象，直接使用
    date = timeStr
  }
  
  if (isNaN(date.getTime())) return '-'
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  })
}

/**
 * 格式化时长（秒）
 * @param {number} seconds - 秒数
 * @returns {string} 格式化后的时长字符串
 */
export function formatDuration(seconds) {
  if (!seconds && seconds !== 0) return '-'
  if (seconds < 60) {
    return `${Math.round(seconds)}秒`
  } else if (seconds < 3600) {
    const minutes = Math.floor(seconds / 60)
    const secs = Math.round(seconds % 60)
    return secs > 0 ? `${minutes}分钟${secs}秒` : `${minutes}分钟`
  } else {
    const hours = Math.floor(seconds / 3600)
    const minutes = Math.round((seconds % 3600) / 60)
    return minutes > 0 ? `${hours}小时${minutes}分钟` : `${hours}小时`
  }
}

/**
 * 格式化速度（行/秒）
 * @param {number} speed - 速度（行/秒）
 * @returns {string} 格式化后的速度字符串
 */
export function formatSpeed(speed) {
  if (!speed || speed <= 0) return '0 行/秒'
  if (speed < 1000) {
    return `${Math.round(speed)} 行/秒`
  } else if (speed < 1000000) {
    return `${(speed / 1000).toFixed(1)}K 行/秒`
  } else {
    return `${(speed / 1000000).toFixed(2)}M 行/秒`
  }
}

/**
 * 格式化文件大小
 * @param {number} bytes - 字节数
 * @returns {string} 格式化后的文件大小字符串
 */
export function formatFileSize(bytes) {
  if (!bytes || bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return `${(bytes / Math.pow(k, i)).toFixed(2)} ${sizes[i]}`
}

/**
 * 格式化百分比
 * @param {number} value - 数值
 * @param {number} total - 总数
 * @param {number} decimals - 小数位数
 * @returns {string} 格式化后的百分比字符串
 */
export function formatPercent(value, total, decimals = 2) {
  if (!total || total === 0) return '0%'
  const percent = (value / total) * 100
  return `${percent.toFixed(decimals)}%`
}

/**
 * 格式化数字（添加千分位）
 * @param {number} num - 数字
 * @returns {string} 格式化后的数字字符串
 */
export function formatNumber(num) {
  if (num === null || num === undefined) return '0'
  return num.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ',')
}


<template>
    <div>
      <el-card>
        <template #header>
          <div style="display: flex; justify-content: space-between; align-items: center">
            <span>配置造数规则 - {{ tableName }}</span>
            <div>
              <el-button @click="showLoadTemplateDialog = true" :disabled="actionLoading.get('loadTemplate')">从模板加载</el-button>
              <el-button @click="showSaveTemplateDialog = true" :disabled="actionLoading.get('saveTemplate')">保存为模板</el-button>
              <el-button type="success" @click="handlePreviewData" :loading="actionLoading.get('preview')" :disabled="actionLoading.get('preview')">预览数据</el-button>
              <el-button @click="handleBack" :disabled="actionLoading.get('back')">返回</el-button>
            </div>
          </div>
        </template>
  
        <!-- 基本信息配置（可折叠） -->
        <el-collapse v-model="showBasicConfig" style="margin-bottom: 20px">
          <el-collapse-item name="basic" :title="'基本信息配置'">
            <el-form :model="taskConfig" label-width="150px">
              <el-row :gutter="20">
                <el-col :span="12">
                  <el-form-item label="任务名称">
                    <el-input v-model="taskConfig.name" placeholder="给任务起个名字" />
                  </el-form-item>
                </el-col>
                <el-col :span="12">
                  <el-form-item label="生成数量">
                    <el-input-number 
                      v-model="taskConfig.totalRows" 
                      :min="1" 
                      :max="100000000"
                      style="width: 100%"
                    />
                  </el-form-item>
                </el-col>
              </el-row>
              <el-row :gutter="20">
                <el-col :span="12">
                  <el-form-item label="批次大小">
                    <el-input-number 
                      v-model="taskConfig.batchSize" 
                      :min="1" 
                      :max="10000"
                      style="width: 100%"
                    />
                  </el-form-item>
                </el-col>
                <el-col :span="12">
                  <el-form-item label="线程数">
                    <el-input-number 
                      v-model="taskConfig.threadCount" 
                      :min="1" 
                      :max="10"
                      style="width: 100%"
                    />
                    <div style="font-size: 12px; color: #909399; margin-top: 4px">
                      最多支持10个线程，主键将统一生成以避免冲突
                    </div>
                  </el-form-item>
                </el-col>
              </el-row>
            </el-form>
          </el-collapse-item>
        </el-collapse>
  
        <!-- 字段规则配置 -->
        <el-divider>字段生成规则配置</el-divider>
        
        <!-- 搜索和批量操作 -->
        <div style="margin-bottom: 15px; display: flex; gap: 10px; align-items: center">
          <el-input
            v-model="fieldSearchText"
            placeholder="搜索字段名"
            style="width: 200px"
            clearable
            :prefix-icon="Search"
          />
          <el-button size="small" @click="showBatchDialog = true">批量设置</el-button>
          <el-button size="small" @click="resetAllRules">重置所有</el-button>
        </div>
  
        <el-table :data="filteredFieldRules" border style="width: 100%" :max-height="tableMaxHeight">
          <el-table-column prop="fieldName" label="字段名" width="150" />
          <el-table-column prop="fieldType" label="字段类型" width="120" />
          <el-table-column label="规则类型" width="150">
            <template #default="scope">
              <el-select 
                :model-value="scope.row.ruleType" 
                @update:model-value="(val) => { scope.row.ruleType = val; onRuleTypeChange(scope.row); }"
                placeholder="选择规则"
              >
                <el-option 
                  v-for="rule in getAvailableRules(scope.row)" 
                  :key="rule.value"
                  :label="rule.label" 
                  :value="rule.value" 
                />
              </el-select>
            </template>
          </el-table-column>
          <el-table-column label="规则配置" min-width="300">
            <template #default="scope">
              <!-- 随机字符串配置 -->
              <div v-if="scope.row.ruleType === 'random_string'" style="display: flex; flex-direction: column; gap: 10px">
                <!-- 长度模式选择 -->
                <div style="display: flex; gap: 10px; align-items: center">
                  <span style="font-size: 12px; color: #606266; width: 80px;">长度模式:</span>
                  <el-radio-group v-model="scope.row.config.lengthMode" size="small" @change="() => {
                    // 确保 lengthMode 有默认值
                    if (!scope.row.config.lengthMode) {
                      scope.row.config.lengthMode = 'random'
                    }
                    // 切换模式时，清空另一个模式的配置
                    if (scope.row.config.lengthMode === 'fixed') {
                      scope.row.config.minLength = 0
                      scope.row.config.maxLength = 0
                    } else {
                      scope.row.config.fixedLength = 0
                    }
                    validateFieldRule(scope.row)
                  }">
                    <el-radio-button label="random">随机长度</el-radio-button>
                    <el-radio-button label="fixed">固定长度</el-radio-button>
                  </el-radio-group>
                </div>
                <!-- 基础配置 -->
                <div style="display: flex; gap: 10px; flex-wrap: wrap; align-items: center">
                  <!-- 随机长度模式 -->
                  <template v-if="!scope.row.config.lengthMode || scope.row.config.lengthMode === 'random'">
                    <el-input-number 
                      v-model="scope.row.config.minLength" 
                      :min="1" 
                      :max="scope.row.maxLength || 1000"
                      placeholder="最小长度"
                      size="small"
                      style="width: 100px"
                      @change="() => validateFieldRule(scope.row)"
                    />
                    <el-input-number 
                      v-model="scope.row.config.maxLength" 
                      :min="1" 
                      :max="scope.row.maxLength || 1000"
                      placeholder="最大长度"
                      size="small"
                      style="width: 100px"
                      @change="() => validateFieldRule(scope.row)"
                    />
                  </template>
                  <!-- 固定长度模式 -->
                  <template v-else>
                    <el-input-number 
                      v-model="scope.row.config.fixedLength" 
                      :min="1" 
                      :max="scope.row.maxLength || 1000"
                      placeholder="固定长度"
                      size="small"
                      style="width: 100px"
                      @change="() => validateFieldRule(scope.row)"
                    />
                  </template>
                  <el-select 
                    v-model="scope.row.config.charSet" 
                    multiple
                    collapse-tags
                    collapse-tags-tooltip
                    size="small" 
                    style="width: 200px"
                    placeholder="选择字符类型"
                  >
                    <el-option label="字母" value="letters" />
                    <el-option label="数字" value="numbers" />
                    <el-option label="中文" value="chinese" />
                    <el-option label="特殊字符" value="special" />
                  </el-select>
                  <el-tooltip
                    v-if="getFieldBoundaryInfo(scope.row) || validateFieldRule(scope.row).valid === false"
                    :content="validateFieldRule(scope.row).valid === false ? validateFieldRule(scope.row).message : getFieldBoundaryInfo(scope.row)"
                    placement="top"
                  >
                    <el-icon 
                      :style="{ color: validateFieldRule(scope.row).valid === false ? '#F56C6C' : '#909399', cursor: 'pointer', fontSize: '16px' }"
                    >
                      <InfoFilled v-if="validateFieldRule(scope.row).valid !== false" />
                      <WarningFilled v-else />
                    </el-icon>
                  </el-tooltip>
                </div>
                <!-- 粒度配置（可折叠） -->
                <el-collapse v-model="scope.row.showGranularity" style="border: none;">
                  <el-collapse-item :name="scope.row.fieldName + '_granularity'" style="border: none;">
                    <template #title>
                      <span style="font-size: 12px; color: #909399;">高级配置（粒度控制）</span>
                    </template>
                    <div style="display: flex; flex-direction: column; gap: 10px; padding: 10px; background: #f5f7fa; border-radius: 4px;">
                      <!-- 大小写 -->
                      <div style="display: flex; gap: 10px; align-items: center">
                        <span style="width: 80px; font-size: 12px;">大小写:</span>
                        <el-select v-model="scope.row.config.case" size="small" style="width: 150px">
                          <el-option label="混合（默认）" value="mixed" />
                          <el-option label="全小写" value="lower" />
                          <el-option label="全大写" value="upper" />
                        </el-select>
                      </div>
                      <!-- 数字位置 -->
                      <div style="display: flex; gap: 10px; align-items: center">
                        <span style="width: 80px; font-size: 12px;">数字位置:</span>
                        <el-select v-model="scope.row.config.numberPosition" size="small" style="width: 150px">
                          <el-option label="无数字" value="none" />
                          <el-option label="开头" value="start" />
                          <el-option label="结尾" value="end" />
                          <el-option label="随机" value="random" />
                        </el-select>
                      </div>
                      <!-- 前缀和后缀 -->
                      <div style="display: flex; gap: 10px; align-items: center">
                        <span style="width: 80px; font-size: 12px;">前缀:</span>
                        <el-input 
                          v-model="scope.row.config.prefix" 
                          placeholder="前缀（可选）"
                          size="small"
                          style="width: 120px"
                        />
                        <span style="font-size: 12px; color: #909399;">后缀:</span>
                        <el-input 
                          v-model="scope.row.config.suffix" 
                          placeholder="后缀（可选）"
                          size="small"
                          style="width: 120px"
                        />
                      </div>
                      <!-- 自定义字符集 -->
                      <div style="display: flex; gap: 10px; align-items: center">
                        <span style="width: 80px; font-size: 12px;">自定义字符:</span>
                        <el-input 
                          v-model="scope.row.config.customChars" 
                          placeholder="自定义字符集（可选）"
                          size="small"
                          style="flex: 1"
                        />
                      </div>
                    </div>
                  </el-collapse-item>
                </el-collapse>
              </div>
  
              <!-- 随机数字配置 -->
              <div v-if="scope.row.ruleType === 'random_number'" style="display: flex; flex-direction: column; gap: 10px">
                <!-- 基础配置 -->
                <div style="display: flex; gap: 10px; flex-wrap: wrap; align-items: center">
                  <div style="display: flex; gap: 5px; align-items: center;">
                    <span style="font-size: 12px; color: #606266; white-space: nowrap;">最小值:</span>
                    <el-input-number 
                      v-model="scope.row.config.min" 
                      :precision="scope.row.config.isInt ? 0 : 2"
                      :min="getNumericFieldRange(scope.row).min"
                      :max="getNumericFieldRange(scope.row).max"
                      placeholder="最小值"
                      size="small"
                      style="width: 120px"
                      @change="() => { 
                        const range = getNumericFieldRange(scope.row)
                        if (scope.row.config.min < range.min) scope.row.config.min = range.min
                        if (scope.row.config.min > range.max) scope.row.config.min = range.max
                        if (scope.row.config.max !== undefined && scope.row.config.max < scope.row.config.min) {
                          scope.row.config.max = scope.row.config.min
                        }
                        if (isIntType(scope.row)) scope.row.config.isInt = true
                        validateFieldRule(scope.row) 
                      }"
                    />
                  </div>
                  <div style="display: flex; gap: 5px; align-items: center;">
                    <span style="font-size: 12px; color: #606266; white-space: nowrap;">最大值:</span>
                    <el-input-number 
                      v-model="scope.row.config.max" 
                      :precision="scope.row.config.isInt ? 0 : 2"
                      :min="getNumericFieldRange(scope.row).min"
                      :max="getNumericFieldRange(scope.row).max"
                      placeholder="最大值"
                      size="small"
                      style="width: 120px"
                      @change="() => { 
                        const range = getNumericFieldRange(scope.row)
                        if (scope.row.config.max < range.min) scope.row.config.max = range.min
                        if (scope.row.config.max > range.max) scope.row.config.max = range.max
                        if (scope.row.config.min !== undefined && scope.row.config.min > scope.row.config.max) {
                          scope.row.config.min = scope.row.config.max
                        }
                        if (isIntType(scope.row)) scope.row.config.isInt = true
                        validateFieldRule(scope.row) 
                      }"
                    />
                  </div>
                  <el-checkbox 
                    v-model="scope.row.config.isInt" 
                    size="small" 
                    :disabled="isIntType(scope.row)"
                    @change="() => validateFieldRule(scope.row)"
                  >
                    整数
                  </el-checkbox>
                  <span v-if="isIntType(scope.row)" style="font-size: 12px; color: #909399;">(INT类型固定为整数)</span>
                  <el-tooltip
                    v-if="getFieldBoundaryInfo(scope.row) || validateFieldRule(scope.row).valid === false"
                    :content="validateFieldRule(scope.row).valid === false ? validateFieldRule(scope.row).message : getFieldBoundaryInfo(scope.row)"
                    placement="top"
                  >
                    <el-icon 
                      :style="{ color: validateFieldRule(scope.row).valid === false ? '#F56C6C' : '#909399', cursor: 'pointer', fontSize: '16px' }"
                    >
                      <InfoFilled v-if="validateFieldRule(scope.row).valid !== false" />
                      <WarningFilled v-else />
                    </el-icon>
                  </el-tooltip>
                </div>
                <!-- 粒度配置（可折叠） -->
                <el-collapse v-model="scope.row.showGranularity" style="border: none;">
                  <el-collapse-item :name="scope.row.fieldName + '_granularity'" style="border: none;">
                    <template #title>
                      <span style="font-size: 12px; color: #909399;">高级配置（粒度控制）</span>
                    </template>
                    <div style="display: flex; flex-direction: column; gap: 10px; padding: 10px; background: #f5f7fa; border-radius: 4px;">
                      <!-- 精度和小数位数（仅浮点类型且选择小数时显示） -->
                      <div v-if="isFloatType(scope.row) && !scope.row.config.isInt" style="display: flex; gap: 10px; align-items: center">
                        <span style="width: 80px; font-size: 12px;">精度:</span>
                        <el-input-number 
                          v-model="scope.row.config.precision" 
                          :min="0"
                          placeholder="总位数"
                          size="small"
                          style="width: 120px"
                          @change="() => validateFieldRule(scope.row)"
                        />
                        <span style="font-size: 12px; color: #909399;">小数位:</span>
                        <el-input-number 
                          v-model="scope.row.config.scale" 
                          :min="0"
                          :max="10"
                          placeholder="小数位数"
                          size="small"
                          style="width: 120px"
                          @change="() => validateFieldRule(scope.row)"
                        />
                      </div>
                      <!-- 固定长度（用于字符串类型的数字） -->
                      <div v-if="scope.row.fieldType.toLowerCase().includes('varchar') || scope.row.fieldType.toLowerCase().includes('char')" style="display: flex; gap: 10px; align-items: center">
                        <span style="width: 80px; font-size: 12px;">固定长度:</span>
                        <el-input-number 
                          v-model="scope.row.config.fixedLength" 
                          :min="0"
                          :max="scope.row.maxLength || 1000"
                          placeholder="固定长度（0=不固定）"
                          size="small"
                          style="width: 120px"
                          @change="() => {
                            if (scope.row.config.fixedLength === null || scope.row.config.fixedLength === undefined) {
                              scope.row.config.fixedLength = 0
                            }
                            validateFieldRule(scope.row)
                          }"
                        />
                      </div>
                      <!-- 分布类型 -->
                      <div style="display: flex; gap: 10px; align-items: center">
                        <span style="width: 80px; font-size: 12px;">分布类型:</span>
                        <el-select v-model="scope.row.config.distribution" size="small" style="width: 150px" @change="() => validateFieldRule(scope.row)">
                          <el-option label="均匀分布" value="uniform" />
                          <el-option label="正态分布" value="normal" />
                          <el-option label="指数分布" value="exponential" />
                        </el-select>
                      </div>
                      <!-- 正态分布参数 -->
                      <div v-if="scope.row.config.distribution === 'normal'" style="display: flex; gap: 10px; align-items: center">
                        <span style="width: 80px; font-size: 12px;">均值:</span>
                        <el-input-number 
                          v-model="scope.row.config.mean" 
                          :precision="2"
                          placeholder="均值"
                          size="small"
                          style="width: 120px"
                        />
                        <span style="font-size: 12px; color: #909399;">标准差:</span>
                        <el-input-number 
                          v-model="scope.row.config.stdDev" 
                          :precision="2"
                          :min="0"
                          placeholder="标准差"
                          size="small"
                          style="width: 120px"
                        />
                      </div>
                      <!-- 指数分布参数 -->
                      <div v-if="scope.row.config.distribution === 'exponential'" style="display: flex; gap: 10px; align-items: center">
                        <span style="width: 80px; font-size: 12px;">Lambda:</span>
                        <el-input-number 
                          v-model="scope.row.config.lambda" 
                          :precision="4"
                          :min="0.0001"
                          placeholder="Lambda参数"
                          size="small"
                          style="width: 120px"
                        />
                      </div>
                    </div>
                  </el-collapse-item>
                </el-collapse>
              </div>
  
              <!-- 随机日期/日期时间配置 -->
              <div v-if="scope.row.ruleType === 'random_date'" style="display: flex; flex-direction: column; gap: 10px">
                <!-- 根据字段类型显示不同的标题 -->
                <div v-if="isTimestampType(scope.row)" style="font-size: 12px; color: #909399; margin-bottom: 5px;">
                  日期时间配置（包含日期和时间）
                </div>
                <div v-else-if="isDateOnlyType(scope.row)" style="font-size: 12px; color: #909399; margin-bottom: 5px;">
                  随机日期配置（仅日期）
                </div>
                <!-- 基础配置 -->
                <div style="display: flex; flex-direction: column; gap: 10px">
                  <!-- 日期范围 -->
                  <div style="display: flex; gap: 10px; flex-wrap: wrap; align-items: center">
                    <span style="width: 80px; font-size: 12px;">开始日期:</span>
                    <el-date-picker
                      v-model="scope.row.config.startDate"
                      type="date"
                      placeholder="开始日期"
                      size="small"
                      style="width: 180px"
                      value-format="YYYY-MM-DD"
                      @change="() => validateFieldRule(scope.row)"
                    />
                    <span style="width: 80px; font-size: 12px;">结束日期:</span>
                    <el-date-picker
                      v-model="scope.row.config.endDate"
                      type="date"
                      placeholder="结束日期"
                      size="small"
                      style="width: 180px"
                      value-format="YYYY-MM-DD"
                      @change="() => validateFieldRule(scope.row)"
                    />
                    <el-input 
                      v-model="scope.row.config.format" 
                      placeholder="日期格式（可选）"
                      size="small"
                      style="width: 150px"
                      @blur="() => validateFieldRule(scope.row)"
                    />
                    <el-tooltip
                      v-if="getFieldBoundaryInfo(scope.row) || validateFieldRule(scope.row).valid === false"
                      :content="validateFieldRule(scope.row).valid === false ? validateFieldRule(scope.row).message : getFieldBoundaryInfo(scope.row)"
                      placement="top"
                    >
                      <el-icon 
                        :style="{ color: validateFieldRule(scope.row).valid === false ? '#F56C6C' : '#909399', cursor: 'pointer', fontSize: '16px' }"
                      >
                        <InfoFilled v-if="validateFieldRule(scope.row).valid !== false" />
                        <WarningFilled v-else />
                      </el-icon>
                    </el-tooltip>
                  </div>
                  <!-- 时间配置（仅日期时间类型显示） -->
                  <div v-if="isTimestampType(scope.row)" style="display: flex; flex-direction: column; gap: 10px; padding: 10px; background: #f5f7fa; border-radius: 4px;">
                    <div style="display: flex; gap: 10px; align-items: center; margin-bottom: 5px;">
                      <span style="font-size: 12px; color: #606266; font-weight: 500;">时间配置:</span>
                      <el-checkbox 
                        v-model="scope.row.config.fullDay" 
                        size="small"
                        @change="() => validateFieldRule(scope.row)"
                      >
                        一整天（00:00:00 - 23:59:59）
                      </el-checkbox>
                    </div>
                    <div v-if="!scope.row.config.fullDay" style="display: flex; gap: 10px; flex-wrap: wrap; align-items: center">
                      <span style="width: 80px; font-size: 12px;">时间区间:</span>
                      <el-time-picker
                        v-model="scope.row.config.startTime"
                        placeholder="开始时间"
                        size="small"
                        style="width: 180px"
                        value-format="HH:mm:ss"
                        @change="() => validateFieldRule(scope.row)"
                      />
                      <span style="font-size: 12px; color: #909399;">至</span>
                      <el-time-picker
                        v-model="scope.row.config.endTime"
                        placeholder="结束时间"
                        size="small"
                        style="width: 180px"
                        value-format="HH:mm:ss"
                        @change="() => validateFieldRule(scope.row)"
                      />
                    </div>
                  </div>
                  <!-- 星期配置 -->
                  <div style="display: flex; flex-direction: column; gap: 10px">
                    <div style="display: flex; gap: 10px; align-items: center">
                      <span style="width: 80px; font-size: 12px;">星期:</span>
                      <el-radio-group 
                        v-model="scope.row.config.weekdayMode" 
                        size="small"
                        @change="() => {
                          if (scope.row.config.weekdayMode === 'all') {
                            scope.row.config.weekdayList = []
                          } else if (scope.row.config.weekdayMode === 'weekdays') {
                            scope.row.config.weekdayList = [1, 2, 3, 4, 5]
                          } else if (scope.row.config.weekdayMode === 'custom') {
                            scope.row.config.weekdayList = scope.row.config.weekdayCustom || [1, 2, 3, 4, 5]
                          }
                          validateFieldRule(scope.row)
                        }"
                      >
                        <el-radio label="all">全部</el-radio>
                        <el-radio label="weekdays">工作日</el-radio>
                        <el-radio label="custom">自定义</el-radio>
                      </el-radio-group>
                    </div>
                    <!-- 自定义星期复选框 -->
                    <div v-if="scope.row.config.weekdayMode === 'custom'" style="display: flex; gap: 10px; align-items: center; margin-left: 80px;">
                      <el-checkbox-group 
                        v-model="scope.row.config.weekdayCustom"
                        @change="() => {
                          scope.row.config.weekdayList = scope.row.config.weekdayCustom || []
                          validateFieldRule(scope.row)
                        }"
                      >
                        <el-checkbox :label="1" size="small">星期一</el-checkbox>
                        <el-checkbox :label="2" size="small">星期二</el-checkbox>
                        <el-checkbox :label="3" size="small">星期三</el-checkbox>
                        <el-checkbox :label="4" size="small">星期四</el-checkbox>
                        <el-checkbox :label="5" size="small">星期五</el-checkbox>
                        <el-checkbox :label="6" size="small">星期六</el-checkbox>
                        <el-checkbox :label="0" size="small">星期日</el-checkbox>
                      </el-checkbox-group>
                    </div>
                  </div>
                </div>
              </div>
  
              <!-- 固定值配置 -->
              <div v-if="scope.row.ruleType === 'fixed'" style="display: flex; gap: 8px; align-items: center">
                <span style="font-size: 12px; color: #606266; white-space: nowrap;">固定值:</span>
                <el-input 
                  v-model="scope.row.config.value" 
                  placeholder="输入固定值"
                  size="small"
                  style="flex: 1"
                  @blur="() => validateFieldRule(scope.row)"
                />
                <el-tooltip
                  v-if="getFieldBoundaryInfo(scope.row) || validateFieldRule(scope.row).valid === false"
                  :content="validateFieldRule(scope.row).valid === false ? validateFieldRule(scope.row).message : getFieldBoundaryInfo(scope.row)"
                  placement="top"
                >
                  <el-icon 
                    :style="{ color: validateFieldRule(scope.row).valid === false ? '#F56C6C' : '#909399', cursor: 'pointer', fontSize: '16px' }"
                  >
                    <InfoFilled v-if="validateFieldRule(scope.row).valid !== false" />
                    <WarningFilled v-else />
                  </el-icon>
                </el-tooltip>
              </div>
  
              <!-- 递增配置 -->
              <div v-if="scope.row.ruleType === 'increment'" style="display: flex; flex-direction: column; gap: 10px">
                <!-- 基础配置 -->
                <div style="display: flex; gap: 10px; flex-wrap: wrap; align-items: center">
                  <div style="display: flex; gap: 5px; align-items: center;">
                    <span style="font-size: 12px; color: #606266; white-space: nowrap;">起始值:</span>
                    <el-input-number 
                      v-model="scope.row.config.startValue" 
                      placeholder="起始值"
                      size="small"
                      style="width: 120px"
                      @change="() => validateFieldRule(scope.row)"
                    />
                  </div>
                  <div style="display: flex; gap: 5px; align-items: center;">
                    <span style="font-size: 12px; color: #606266; white-space: nowrap;">步长:</span>
                    <el-input-number 
                      v-model="scope.row.config.step" 
                      placeholder="步长"
                      size="small"
                      style="width: 120px"
                      @change="() => validateFieldRule(scope.row)"
                    />
                  </div>
                  <div style="display: flex; gap: 5px; align-items: center;">
                    <span style="font-size: 12px; color: #606266; white-space: nowrap;">最大值:</span>
                    <el-input-number 
                      v-model="scope.row.config.maxValue" 
                      placeholder="最大值（循环时使用）"
                      size="small"
                      style="width: 120px"
                      @change="() => validateFieldRule(scope.row)"
                    />
                  </div>
                  <el-checkbox v-model="scope.row.config.cycle" size="small">循环</el-checkbox>
                  <el-tooltip
                    v-if="getFieldBoundaryInfo(scope.row) || validateFieldRule(scope.row).valid === false"
                    :content="validateFieldRule(scope.row).valid === false ? validateFieldRule(scope.row).message : getFieldBoundaryInfo(scope.row)"
                    placement="top"
                  >
                    <el-icon 
                      :style="{ color: validateFieldRule(scope.row).valid === false ? '#F56C6C' : '#909399', cursor: 'pointer', fontSize: '16px' }"
                    >
                      <InfoFilled v-if="validateFieldRule(scope.row).valid !== false" />
                      <WarningFilled v-else />
                    </el-icon>
                  </el-tooltip>
                </div>
                <!-- 格式（用于字符串类型） -->
                <div v-if="scope.row.fieldType.toLowerCase().includes('varchar') || scope.row.fieldType.toLowerCase().includes('char')" style="display: flex; gap: 10px; align-items: center">
                  <span style="font-size: 12px; color: #606266; white-space: nowrap;">格式:</span>
                  <el-input 
                    v-model="scope.row.config.format" 
                    placeholder="格式（如 USER_{:05d} 表示 USER_00001）"
                    size="small"
                    style="width: 300px"
                    @blur="() => validateFieldRule(scope.row)"
                  />
                </div>
              </div>
  
              <!-- 列表配置 -->
              <div v-if="scope.row.ruleType === 'list'" style="display: flex; flex-direction: column; gap: 10px">
                <!-- 基础配置 -->
                <div style="display: flex; flex-direction: column; gap: 8px">
                  <div style="display: flex; gap: 8px; align-items: center">
                    <span style="width: 60px; font-size: 12px;">值:</span>
                    <el-input 
                      v-model="scope.row.config.valuesText" 
                      type="textarea"
                      :rows="6"
                      placeholder="输入值列表，每行一个值"
                      size="small"
                      style="flex: 1"
                      @blur="() => { parseListValues(scope.row); validateFieldRule(scope.row) }"
                    />
                    <el-tooltip
                      v-if="getFieldBoundaryInfo(scope.row) || validateFieldRule(scope.row).valid === false"
                      :content="validateFieldRule(scope.row).valid === false ? validateFieldRule(scope.row).message : getFieldBoundaryInfo(scope.row)"
                      placement="top"
                    >
                      <el-icon 
                        :style="{ color: validateFieldRule(scope.row).valid === false ? '#F56C6C' : '#909399', cursor: 'pointer', fontSize: '16px' }"
                      >
                        <InfoFilled v-if="validateFieldRule(scope.row).valid !== false" />
                        <WarningFilled v-else />
                      </el-icon>
                    </el-tooltip>
                  </div>
                </div>
                <!-- 粒度配置（可折叠） -->
                <el-collapse v-model="scope.row.showGranularity" style="border: none;">
                  <el-collapse-item :name="scope.row.fieldName + '_granularity'" style="border: none;">
                    <template #title>
                      <span style="font-size: 12px; color: #909399;">高级配置（粒度控制）</span>
                    </template>
                    <div style="display: flex; flex-direction: column; gap: 10px; padding: 10px; background: #f5f7fa; border-radius: 4px;">
                      <!-- 允许重复 -->
                      <div style="display: flex; gap: 10px; align-items: center">
                        <el-checkbox v-model="scope.row.config.allowRepeat" size="small" @change="() => validateFieldRule(scope.row)">允许重复选择</el-checkbox>
                      </div>
                      <!-- 权重配置 -->
                      <div style="display: flex; flex-direction: column; gap: 5px">
                        <span style="font-size: 12px; color: #606266;">权重配置（可选，控制选择概率）:</span>
                        <el-input 
                          v-model="scope.row.config.weightsText" 
                          placeholder="输入权重列表，用逗号分隔，与值列表一一对应（如：1,2,1）"
                          size="small"
                          @blur="() => parseListWeights(scope.row)"
                        />
                        <div v-if="scope.row.config.weights && scope.row.config.weights.length > 0" style="font-size: 11px; color: #909399; margin-top: 5px;">
                          当前权重: {{ scope.row.config.weights.join(', ') }}
                        </div>
                      </div>
                    </div>
                  </el-collapse-item>
                </el-collapse>
              </div>
  
              <!-- 正则表达式配置 -->
              <div v-if="scope.row.ruleType === 'regex'" style="display: flex; flex-direction: column; gap: 8px">
                <div style="display: flex; gap: 5px; align-items: center;">
                  <span style="font-size: 12px; color: #606266; white-space: nowrap;">预设:</span>
                  <el-select 
                    v-model="scope.row.config.presetPattern" 
                    placeholder="选择常用正则表达式"
                    size="small"
                    clearable
                    style="flex: 1"
                    @change="handleRegexPresetChange(scope.row)"
                  >
                    <el-option label="IP地址" value="ip" />
                    <el-option label="邮箱地址" value="email" />
                    <el-option label="手机号（中国）" value="phone_cn" />
                    <el-option label="身份证号（中国）" value="idcard_cn" />
                    <el-option label="URL地址" value="url" />
                    <el-option label="日期（YYYY-MM-DD）" value="date" />
                    <el-option label="时间（HH:MM:SS）" value="time" />
                    <el-option label="邮政编码（中国）" value="postcode_cn" />
                    <el-option label="车牌号（中国）" value="license_plate_cn" />
                    <el-option label="自定义" value="custom" />
                  </el-select>
                </div>
                <div style="display: flex; gap: 5px; align-items: center;">
                  <span style="font-size: 12px; color: #606266; white-space: nowrap;">正则:</span>
                  <el-input 
                    v-model="scope.row.config.pattern" 
                    placeholder="输入正则表达式（选择预设会自动填充）"
                    size="small"
                    style="flex: 1"
                    @input="handleRegexPatternInput(scope.row)"
                  />
                </div>
              </div>
  
              <!-- 函数配置 -->
              <div v-if="scope.row.ruleType === 'function'" style="display: flex; flex-direction: column; gap: 10px">
                <!-- UUID 配置选项（当选择 UUID 时，不显示函数选择器，直接显示 UUID 配置） -->
                <template v-if="scope.row.config.funcName === 'UUID'">
                  <div style="display: flex; flex-direction: column; gap: 8px;">
                    <div style="display: flex; gap: 10px; align-items: center; flex-wrap: wrap;">
                      <div style="display: flex; gap: 5px; align-items: center;">
                        <span style="font-size: 12px; color: #606266; white-space: nowrap;">版本:</span>
                        <el-select 
                          v-model="scope.row.config.version" 
                          size="small" 
                          style="width: 120px"
                          placeholder="选择版本"
                        >
                          <el-option label="V1 (时间戳+MAC)" value="v1" />
                          <el-option label="V3 (MD5哈希)" value="v3" />
                          <el-option label="V4 (随机)" value="v4" />
                          <el-option label="V5 (SHA-1哈希)" value="v5" />
                        </el-select>
                      </div>
                      <el-select 
                        v-model="scope.row.config.case" 
                        size="small" 
                        style="width: 120px"
                        placeholder="大小写"
                      >
                        <el-option label="小写" value="lower" />
                        <el-option label="大写" value="upper" />
                        <el-option label="混合" value="mixed" />
                      </el-select>
                      <el-checkbox 
                        v-model="scope.row.config.withHyphen" 
                        size="small"
                      >
                        带连字符(-)
                      </el-checkbox>
                    </div>
                    <div style="font-size: 12px; color: #909399;">
                      <span v-if="scope.row.config.version === 'v1'">V1: 基于时间戳和 MAC 地址，包含时间信息，36 字符</span>
                      <span v-else-if="scope.row.config.version === 'v3'">V3: 基于命名空间和名称的 MD5 哈希，确定性生成，36 字符</span>
                      <span v-else-if="scope.row.config.version === 'v4'">V4: 完全随机生成，36 字符，格式: xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx</span>
                      <span v-else-if="scope.row.config.version === 'v5'">V5: 基于命名空间和名称的 SHA-1 哈希，确定性生成，36 字符</span>
                      <span v-else>V4: 完全随机生成，36 字符</span>
                    </div>
                  </div>
                </template>
                <!-- 其他函数配置（当不是 UUID 时，显示函数选择器） -->
                <template v-else>
                  <el-select 
                    v-model="scope.row.config.funcName" 
                    size="small" 
                    style="width: 150px"
                  >
                    <el-option label="NOW()" value="NOW" />
                    <el-option label="TODAY()" value="TODAY" />
                    <el-option label="UUID()" value="UUID" />
                    <el-option label="RAND()" value="RAND" />
                    <el-option label="RAND_INT(min, max)" value="RAND_INT" />
                    <el-option label="CONCAT(...)" value="CONCAT" />
                  </el-select>
                </template>
              </div>
  
              <!-- 模板配置 -->
              <el-input 
                v-if="scope.row.ruleType === 'template'"
                v-model="scope.row.config.template" 
                placeholder="模板，支持 {name}, {date}, {number}, {uuid}"
                size="small"
              />
  
              <!-- 空值配置 -->
              <div v-if="scope.row.ruleType === 'null'" style="display: flex; align-items: center; gap: 8px; padding: 8px; background: #f5f7fa; border-radius: 4px;">
                <span style="font-size: 12px; color: #606266;">生成 NULL 值</span>
              </div>
  
              <!-- 引用字段配置 -->
              <div v-if="scope.row.ruleType === 'reference'" style="display: flex; flex-direction: column; gap: 10px">
                <el-input 
                  v-model="scope.row.config.expression" 
                  placeholder="表达式，如 {first_name}_{last_name} 或 {price} * {quantity}"
                  size="small"
                />
                <el-input 
                  v-model="scope.row.config.fieldsText" 
                  placeholder="引用的字段列表（可选，用逗号分隔）"
                  size="small"
                  @blur="parseReferenceFields(scope.row)"
                />
              </div>
  
              <!-- 地理数据配置 -->
              <div v-if="scope.row.ruleType === 'geographic'" style="display: flex; gap: 10px; flex-wrap: wrap; align-items: center">
                <el-select v-model="scope.row.config.type" size="small" style="width: 150px">
                  <el-option label="城市" value="city" />
                  <el-option label="国家" value="country" />
                  <el-option label="地址" value="address" />
                  <el-option label="坐标" value="coordinates" />
                  <el-option label="纬度" value="latitude" />
                  <el-option label="经度" value="longitude" />
                  <el-option label="邮编" value="postal_code" />
                </el-select>
                <el-input 
                  v-model="scope.row.config.country" 
                  placeholder="国家代码（可选，如 CN, US）"
                  size="small"
                  style="width: 200px"
                />
              </div>
  
              <!-- 从文件读取配置 -->
              <div v-if="scope.row.ruleType === 'file'" style="display: flex; flex-direction: column; gap: 10px">
                <div style="display: flex; gap: 8px; align-items: center">
                  <el-input 
                    v-model="scope.row.config.filePath" 
                    placeholder="文件路径（绝对路径或相对路径）"
                    size="small"
                    style="flex: 1"
                  />
                </div>
                <div style="display: flex; gap: 10px; flex-wrap: wrap">
                  <el-select v-model="scope.row.config.fileType" size="small" style="width: 120px">
                    <el-option label="CSV" value="csv" />
                    <el-option label="TXT" value="txt" />
                    <el-option label="JSON" value="json" />
                  </el-select>
                  <el-input-number 
                    v-model="scope.row.config.columnIndex" 
                    :min="0"
                    placeholder="列索引（CSV/JSON使用）"
                    size="small"
                    style="width: 150px"
                  />
                  <el-checkbox v-model="scope.row.config.loop" size="small">循环读取</el-checkbox>
                </div>
              </div>
  
              <!-- 二进制/图片配置 -->
              <div v-if="scope.row.ruleType === 'binary'" style="display: flex; flex-direction: column; gap: 10px">
                <div style="display: flex; gap: 8px; align-items: center">
                  <el-select v-model="scope.row.config.mode" size="small" style="width: 150px">
                    <el-option label="生成图片" value="generate" />
                    <el-option label="从文件夹读取" value="folder" />
                  </el-select>
                </div>
                
                <!-- 生成图片模式 -->
                <template v-if="scope.row.config.mode === 'generate'">
                  <div style="display: flex; gap: 10px; flex-wrap: wrap">
                    <el-input-number 
                      v-model="scope.row.config.width" 
                      :min="1"
                      :max="10000"
                      placeholder="宽度"
                      size="small"
                      style="width: 120px"
                    />
                    <el-input-number 
                      v-model="scope.row.config.height" 
                      :min="1"
                      :max="10000"
                      placeholder="高度"
                      size="small"
                      style="width: 120px"
                    />
                    <el-select v-model="scope.row.config.format" size="small" style="width: 120px">
                      <el-option label="PNG" value="png" />
                      <el-option label="JPEG" value="jpeg" />
                      <el-option label="GIF" value="gif" />
                    </el-select>
                  </div>
                </template>
                
                <!-- 从文件夹读取模式 -->
                <template v-if="scope.row.config.mode === 'folder'">
                  <el-input 
                    v-model="scope.row.config.folderPath" 
                    placeholder="文件夹路径（绝对路径或相对路径）"
                    size="small"
                  />
                  <el-input 
                    v-model="scope.row.config.extensionsText" 
                    placeholder="文件扩展名过滤（可选，用逗号分隔，如 jpg,png,gif）"
                    size="small"
                    @blur="parseBinaryExtensions(scope.row)"
                  />
                  <el-checkbox v-model="scope.row.config.loop" size="small">循环读取</el-checkbox>
                </template>
              </div>
  
              <!-- 外键引用配置 -->
              <div v-if="scope.row.ruleType === 'foreign'" style="display: flex; flex-direction: column; gap: 10px">
                <!-- 模式（Schema） -->
                <div style="display: flex; gap: 8px; align-items: center">
                  <span style="width: 80px; font-size: 12px; color: #606266;">模式:</span>
                  <el-select 
                    v-model="scope.row.config.foreignDatabase" 
                    placeholder="选择模式"
                    size="small"
                    style="flex: 1"
                    :loading="foreignDatabasesLoading[scope.row.fieldName]"
                    @visible-change="(visible) => { if (visible && !foreignDatabases[scope.row.fieldName]?.length) loadForeignDatabases(scope.row.fieldName) }"
                    @change="onForeignDatabaseChange(scope.row)"
                  >
                    <el-option 
                      v-for="db in foreignDatabases[scope.row.fieldName] || []"
                      :key="db"
                      :label="db" 
                      :value="db" 
                    />
                  </el-select>
                </div>
                
                <!-- 表（Table） -->
                <div style="display: flex; gap: 8px; align-items: center">
                  <span style="width: 80px; font-size: 12px; color: #606266;">表:</span>
                  <el-select 
                    v-model="scope.row.config.foreignTable" 
                    placeholder="选择表"
                    size="small"
                    style="flex: 1"
                    :loading="foreignTablesLoading[scope.row.fieldName]"
                    :disabled="!scope.row.config.foreignDatabase"
                    @visible-change="(visible) => { 
                      if (visible && scope.row.config.foreignDatabase && !foreignTables[scope.row.fieldName]?.length) {
                        loadForeignTables(scope.row.fieldName, scope.row.config.foreignDatabase)
                      }
                    }"
                    @change="(val) => {
                      onForeignTableChange(scope.row)
                    }"
                  >
                    <el-option 
                      v-for="table in foreignTables[scope.row.fieldName] || []"
                      :key="table"
                      :label="table" 
                      :value="table" 
                    />
                  </el-select>
                </div>
                
                <!-- 字段（Field） -->
                <div style="display: flex; gap: 8px; align-items: center">
                  <span style="width: 80px; font-size: 12px; color: #606266;">字段:</span>
                  <el-select 
                    v-model="scope.row.config.foreignField" 
                    placeholder="选择字段"
                    size="small"
                    style="flex: 1"
                    :loading="foreignFieldsLoading[scope.row.fieldName]"
                    :disabled="!scope.row.config.foreignTable"
                    @visible-change="(visible) => { if (visible && scope.row.config.foreignDatabase && scope.row.config.foreignTable && !foreignFields[scope.row.fieldName]?.length) loadForeignFields(scope.row.fieldName, scope.row.config.foreignDatabase, scope.row.config.foreignTable) }"
                  >
                    <el-option 
                      v-for="field in foreignFields[scope.row.fieldName] || []"
                      :key="field.name"
                      :label="field.name" 
                      :value="field.name" 
                    />
                  </el-select>
                </div>
                
                <!-- 生成模式（Generation Mode） -->
                <div style="display: flex; flex-direction: column; gap: 8px">
                  <span style="font-size: 12px; color: #606266; font-weight: 500;">生成模式:</span>
                  <el-radio-group 
                    v-model="scope.row.config.generationMode" 
                    size="small"
                    @change="onForeignGenerationModeChange(scope.row)"
                  >
                    <el-radio label="random">随机</el-radio>
                    <el-radio label="non_repeating">不重复</el-radio>
                    <el-radio label="repeat">重复每个值</el-radio>
                  </el-radio-group>
                  
                  <!-- 重复次数配置（仅当选择"重复每个值"时显示） -->
                  <div v-if="scope.row.config.generationMode === 'repeat'" style="display: flex; gap: 8px; align-items: center; margin-left: 20px;">
                    <el-input-number 
                      v-model="scope.row.config.repeatMin" 
                      :min="1" 
                      :max="scope.row.config.repeatMax || 100"
                      size="small"
                      style="width: 80px"
                    />
                    <span style="font-size: 12px; color: #606266;">到</span>
                    <el-input-number 
                      v-model="scope.row.config.repeatMax" 
                      :min="scope.row.config.repeatMin || 1" 
                      :max="100"
                      size="small"
                      style="width: 80px"
                    />
                    <span style="font-size: 12px; color: #606266;">次</span>
                  </div>
                </div>
              </div>
              
              <!-- 示例值显示（所有规则类型通用） -->
              <div v-if="scope.row.ruleType && scope.row.ruleType !== 'null'" style="margin-top: 8px; padding-top: 8px; border-top: 1px solid #e4e7ed;">
                <div style="display: flex; align-items: center; gap: 8px; padding: 6px 8px; background: #f5f7fa; border-radius: 4px; font-size: 12px;">
                  <span style="color: #909399; font-weight: 500;">示例值:</span>
                  <span style="color: #606266; flex: 1; overflow: hidden; text-overflow: ellipsis; white-space: nowrap;">
                    {{ fieldExamples.get(scope.row.fieldName) || '-' }}
                  </span>
                  <!-- 失败状态显示红色叹号 -->
                  <el-icon 
                    v-if="fieldExamples.get(scope.row.fieldName) === '生成失败' || fieldExamples.get(scope.row.fieldName) === '配置无效'"
                    style="color: #F56C6C; font-size: 16px; flex-shrink: 0;"
                    :title="fieldExamples.get(scope.row.fieldName)"
                  >
                    <WarningFilled />
                  </el-icon>
                  <el-button
                    :icon="Refresh"
                    size="small"
                    text
                    :loading="fieldExampleLoading.get(scope.row.fieldName)"
                    @click="() => generateFieldExample(scope.row)"
                    style="padding: 0; min-height: auto; flex-shrink: 0;"
                  />
                </div>
              </div>
            </template>
          </el-table-column>
          <el-table-column label="约束" width="150">
            <template #default="scope">
              <div style="display: flex; flex-direction: column; gap: 5px">
                <el-tag v-if="scope.row.isPrimaryKey" type="danger" size="small">主键</el-tag>
                <el-tag v-if="scope.row.isForeignKey" type="warning" size="small">外键</el-tag>
                <el-tag v-if="scope.row.isUnique" type="info" size="small">唯一</el-tag>
                <el-tag v-if="!scope.row.isNullable" type="success" size="small">非空</el-tag>
              </div>
            </template>
          </el-table-column>
        </el-table>
  
        <!-- 操作按钮 -->
        <div style="margin-top: 20px; text-align: right">
          <el-button @click="handleBack" :disabled="actionLoading.get('back')">取消</el-button>
          <el-button type="primary" @click="createTask" :loading="actionLoading.get('createTask')" :disabled="actionLoading.get('createTask')">
            {{ isEditMode ? '更新任务' : '创建任务' }}
          </el-button>
        </div>
      </el-card>
  
      <!-- 保存模板对话框 -->
      <el-dialog v-model="showSaveTemplateDialog" title="保存为模板" width="500px">
        <el-form :model="templateForm" label-width="100px">
          <el-form-item label="模板名称" required>
            <el-input v-model="templateForm.name" placeholder="输入模板名称" />
          </el-form-item>
          <el-form-item label="描述">
            <el-input 
              v-model="templateForm.description" 
              type="textarea" 
              :rows="3"
              placeholder="输入模板描述（可选）"
            />
          </el-form-item>
        </el-form>
        <template #footer>
          <el-button @click="showSaveTemplateDialog = false" :disabled="actionLoading.get('saveTemplate')">取消</el-button>
          <el-button type="primary" @click="saveTemplate" :loading="actionLoading.get('saveTemplate')" :disabled="actionLoading.get('saveTemplate')">保存</el-button>
        </template>
      </el-dialog>
  
      <!-- 加载模板对话框 -->
      <el-dialog v-model="showLoadTemplateDialog" title="从模板加载" width="700px">
        <el-table :data="templates" style="width: 100%">
          <el-table-column prop="name" label="模板名称" />
          <el-table-column prop="table_name" label="表名" />
          <el-table-column prop="description" label="描述" />
          <el-table-column label="操作" width="220">
            <template #default="scope">
              <el-button size="small" type="info" @click="previewTemplate(scope.row)" :loading="actionLoading.get(`previewTemplate_${scope.row.id}`)" :disabled="actionLoading.get(`previewTemplate_${scope.row.id}`)">预览</el-button>
              <el-button size="small" type="primary" @click="loadTemplate(scope.row)" :loading="actionLoading.get(`loadTemplate_${scope.row.id}`)" :disabled="actionLoading.get(`loadTemplate_${scope.row.id}`)">加载</el-button>
              <el-button size="small" type="danger" @click="deleteTemplate(scope.row.id)" :loading="actionLoading.get(`deleteTemplate_${scope.row.id}`)" :disabled="actionLoading.get(`deleteTemplate_${scope.row.id}`)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-dialog>
  
      <!-- 预览模板对话框 -->
      <el-dialog v-model="showPreviewTemplateDialog" title="预览模板" width="1000px">
        <div v-if="previewTemplateData">
          <el-descriptions :column="2" border style="margin-bottom: 20px">
            <el-descriptions-item label="模板名称">{{ previewTemplateData.name }}</el-descriptions-item>
            <el-descriptions-item label="表名">{{ previewTemplateData.table_name }}</el-descriptions-item>
            <el-descriptions-item label="描述" :span="2">{{ previewTemplateData.description || '-' }}</el-descriptions-item>
          </el-descriptions>
  
          <div style="margin-bottom: 10px; color: #606266; font-size: 14px;">
            共 <strong style="color: #409EFF;">{{ previewTemplateFields.length }}</strong> 个字段
          </div>
  
          <el-table :data="previewTemplateFields" border style="width: 100%" max-height="500">
            <el-table-column prop="fieldName" label="字段名" width="150" fixed="left" />
            <el-table-column prop="fieldType" label="字段类型" width="120" />
            <el-table-column prop="ruleType" label="规则类型" width="150">
              <template #default="scope">
                <el-tag size="small">{{ getRuleTypeLabel(scope.row.ruleType, scope.row) }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="规则配置" min-width="300">
              <template #default="scope">
                <div v-if="scope.row.ruleType === 'random_string'" style="font-size: 12px; color: #606266;">
                  长度: {{ scope.row.config.minLength || '-' }} ~ {{ scope.row.config.maxLength || '-' }}, 
                  字符集: {{ getCharSetLabel(scope.row.config.charSet) }}
                </div>
                <div v-else-if="scope.row.ruleType === 'random_number'" style="font-size: 12px; color: #606266;">
                  范围: {{ scope.row.config.min || '-' }} ~ {{ scope.row.config.max || '-' }}, 
                  {{ scope.row.config.isInt ? '整数' : '小数' }}
                </div>
                <div v-else-if="scope.row.ruleType === 'random_date'" style="font-size: 12px; color: #606266;">
                  开始: {{ scope.row.config.startDate || '-' }}, 
                  结束: {{ scope.row.config.endDate || '-' }}
                </div>
                <div v-else-if="scope.row.ruleType === 'fixed'" style="font-size: 12px; color: #606266;">
                  固定值: {{ scope.row.config.value || '-' }}
                </div>
                <div v-else-if="scope.row.ruleType === 'increment'" style="font-size: 12px; color: #606266;">
                  起始: {{ scope.row.config.startValue || '-' }}, 
                  步长: {{ scope.row.config.step || '-' }}, 
                  {{ scope.row.config.cycle ? '循环' : '不循环' }}
                </div>
                <div v-else-if="scope.row.ruleType === 'list'" style="font-size: 12px; color: #606266;">
                  值列表: {{ (scope.row.config.values || []).join(', ') || '-' }}
                </div>
                <div v-else-if="scope.row.ruleType === 'regex'" style="font-size: 12px; color: #606266;">
                  正则: {{ scope.row.config.pattern || '-' }}
                </div>
                <div v-else-if="scope.row.ruleType === 'function'" style="font-size: 12px; color: #606266;">
                  函数: {{ scope.row.config.funcName || '-' }}
                </div>
                <div v-else-if="scope.row.ruleType === 'template'" style="font-size: 12px; color: #606266;">
                  模板: {{ scope.row.config.template || '-' }}
                </div>
                <div v-else-if="scope.row.ruleType === 'null'" style="font-size: 12px; color: #606266;">
                  生成 NULL 值
                </div>
                <div v-else-if="scope.row.ruleType === 'reference'" style="font-size: 12px; color: #606266;">
                  表达式: {{ scope.row.config.expression || '-' }}
                </div>
                <div v-else-if="scope.row.ruleType === 'geographic'" style="font-size: 12px; color: #606266;">
                  类型: {{ scope.row.config.type || '-' }}, 
                  国家: {{ scope.row.config.country || '-' }}
                </div>
                <div v-else-if="scope.row.ruleType === 'file'" style="font-size: 12px; color: #606266;">
                  文件路径: {{ scope.row.config.filePath || '-' }}, 
                  类型: {{ scope.row.config.fileType || '-' }}
                </div>
                <div v-else-if="scope.row.ruleType === 'binary'" style="font-size: 12px; color: #606266;">
                  模式: {{ scope.row.config.mode || '-' }}, 
                  格式: {{ scope.row.config.format || '-' }}
                </div>
                <div v-else-if="scope.row.ruleType === 'foreign'" style="font-size: 12px; color: #606266;">
                  外键: {{ scope.row.config.foreignDatabase || '-' }}/{{ scope.row.config.foreignTable || '-' }}/{{ scope.row.config.foreignField || '-' }}, 
                  {{ scope.row.config.generationMode === 'random' ? '随机' : scope.row.config.generationMode === 'non_repeating' ? '不重复' : '重复每个值' }}
                </div>
                <span v-else style="font-size: 12px; color: #909399;">-</span>
              </template>
            </el-table-column>
            <el-table-column label="约束" width="150">
              <template #default="scope">
                <div style="display: flex; flex-direction: column; gap: 5px">
                  <el-tag v-if="scope.row.isPrimaryKey" type="danger" size="small">主键</el-tag>
                  <el-tag v-if="scope.row.isForeignKey" type="warning" size="small">外键</el-tag>
                  <el-tag v-if="scope.row.isUnique && !scope.row.isPrimaryKey" type="success" size="small">唯一</el-tag>
                  <el-tag v-if="!scope.row.isNullable" type="success" size="small">非空</el-tag>
                  <span v-if="!scope.row.isPrimaryKey && !scope.row.isForeignKey && !scope.row.isUnique && scope.row.isNullable" style="color: #909399; font-size: 12px;">-</span>
                </div>
              </template>
            </el-table-column>
          </el-table>
        </div>
        <div v-else style="text-align: center; padding: 40px;">
          <el-empty description="暂无模板数据" />
        </div>
      </el-dialog>
  
      <!-- 批量设置对话框 -->
      <el-dialog v-model="showBatchDialog" title="批量设置规则" width="600px">
        <el-form label-width="120px">
          <el-form-item label="选择字段">
            <el-checkbox-group v-model="selectedFields">
              <el-checkbox 
                v-for="field in fieldRules" 
                :key="field.fieldName" 
                :label="field.fieldName"
              >
                {{ field.fieldName }} ({{ field.fieldType }})
              </el-checkbox>
            </el-checkbox-group>
          </el-form-item>
          <el-form-item label="规则类型">
            <el-select 
              v-model="batchRuleType" 
              placeholder="请先选择字段" 
              style="width: 100%"
              :disabled="selectedFields.length === 0"
            >
              <el-option 
                v-for="rule in getBatchAvailableRules()" 
                :key="rule.value"
                :label="rule.label" 
                :value="rule.value" 
              />
            </el-select>
            <div v-if="selectedFields.length > 0 && hasMixedFieldTypes()" style="margin-top: 8px; color: #E6A23C; font-size: 12px; display: flex; align-items: center;">
              <el-icon><WarningFilled /></el-icon>
              <span style="margin-left: 4px;">已选择不同类型的字段，仅显示通用规则类型</span>
            </div>
          </el-form-item>
          <el-form-item label="规则配置" v-if="batchRuleType">
            <div v-if="batchRuleType === 'random_string'" style="display: flex; gap: 10px; align-items: center">
              <el-input-number v-model="batchConfig.minLength" placeholder="最小长度" :min="1" :max="1000" style="width: 120px" />
              <el-input-number v-model="batchConfig.maxLength" placeholder="最大长度" :min="1" :max="1000" style="width: 120px" />
              <el-select 
                v-model="batchConfig.charSet" 
                multiple
                collapse-tags
                collapse-tags-tooltip
                placeholder="选择字符类型" 
                style="width: 200px"
              >
                <el-option label="字母" value="letters" />
                <el-option label="数字" value="numbers" />
                <el-option label="中文" value="chinese" />
                <el-option label="特殊字符" value="special" />
              </el-select>
            </div>
            <div v-else-if="batchRuleType === 'random_number'" style="display: flex; gap: 10px; align-items: center">
              <el-input-number v-model="batchConfig.min" placeholder="最小值" :precision="2" style="width: 120px" />
              <el-input-number v-model="batchConfig.max" placeholder="最大值" :precision="2" style="width: 120px" />
              <el-checkbox v-model="batchConfig.isInt">整数</el-checkbox>
            </div>
            <div v-else-if="batchRuleType === 'random_date'" style="display: flex; gap: 10px; flex-wrap: wrap">
              <el-date-picker v-model="batchConfig.startDate" type="datetime" placeholder="开始日期" style="width: 180px" />
              <el-date-picker v-model="batchConfig.endDate" type="datetime" placeholder="结束日期" style="width: 180px" />
              <el-input v-model="batchConfig.format" placeholder="格式（可选）" style="width: 150px" />
            </div>
            <el-input v-else-if="batchRuleType === 'fixed'" v-model="batchConfig.value" placeholder="固定值" />
            <div v-else-if="batchRuleType === 'increment'" style="display: flex; gap: 10px; align-items: center">
              <el-input-number v-model="batchConfig.startValue" placeholder="起始值" style="width: 120px" />
              <el-input-number v-model="batchConfig.step" placeholder="步长" :min="1" style="width: 120px" />
              <el-checkbox v-model="batchConfig.cycle">循环</el-checkbox>
            </div>
            <el-input 
              v-else-if="batchRuleType === 'list'" 
              v-model="batchConfig.valuesText" 
              type="textarea"
              :rows="6"
              placeholder="输入值列表，每行一个值"
            />
            <div v-else-if="batchRuleType === 'regex'" style="display: flex; flex-direction: column; gap: 8px">
              <el-select 
                v-model="batchConfig.presetPattern" 
                placeholder="选择常用正则表达式" 
                clearable
                @change="(val) => { if (val && val !== 'custom') { batchConfig.pattern = regexPresets[val] || '' } }"
              >
                <el-option label="IP地址" value="ip" />
                <el-option label="邮箱" value="email" />
                <el-option label="手机号（中国）" value="phone_cn" />
                <el-option label="身份证（中国）" value="idcard_cn" />
                <el-option label="URL" value="url" />
                <el-option label="日期" value="date" />
                <el-option label="时间" value="time" />
                <el-option label="邮编（中国）" value="postcode_cn" />
                <el-option label="车牌号（中国）" value="license_plate_cn" />
                <el-option label="自定义" value="custom" />
              </el-select>
              <el-input v-if="batchConfig.presetPattern === 'custom' || !batchConfig.presetPattern" v-model="batchConfig.pattern" placeholder="输入正则表达式" />
            </div>
            <el-select v-else-if="batchRuleType === 'function'" v-model="batchConfig.funcName" placeholder="选择函数" style="width: 100%">
              <el-option label="UUID" value="UUID" />
              <el-option label="NOW" value="NOW" />
              <el-option label="RANDOM" value="RANDOM" />
              <el-option label="TIMESTAMP" value="TIMESTAMP" />
            </el-select>
            <el-input v-else-if="batchRuleType === 'template'" v-model="batchConfig.template" placeholder="模板，支持 {name}, {date}, {number}, {uuid}" />
            <div v-else-if="batchRuleType === 'null'" style="display: flex; align-items: center; gap: 8px; padding: 8px; background: #f5f7fa; border-radius: 4px;">
              <span style="font-size: 12px; color: #606266;">生成 NULL 值</span>
            </div>
            <div v-else-if="batchRuleType === 'reference'" style="display: flex; flex-direction: column; gap: 8px">
              <el-input v-model="batchConfig.expression" placeholder="表达式，如 {first_name}_{last_name}" />
              <el-input v-model="batchConfig.fieldsText" placeholder="引用字段，用逗号分隔" />
            </div>
            <div v-else-if="batchRuleType === 'geographic'" style="display: flex; gap: 10px; flex-wrap: wrap">
              <el-select v-model="batchConfig.geoType" placeholder="类型" style="width: 150px">
                <el-option label="城市" value="city" />
                <el-option label="国家" value="country" />
                <el-option label="地址" value="address" />
                <el-option label="坐标" value="coordinates" />
              </el-select>
              <el-input v-model="batchConfig.country" placeholder="国家（可选）" style="width: 150px" />
            </div>
            <div v-else-if="batchRuleType === 'file'" style="display: flex; flex-direction: column; gap: 8px">
              <el-input v-model="batchConfig.filePath" placeholder="文件路径（绝对路径或相对路径）" />
              <el-select v-model="batchConfig.fileType" placeholder="文件类型" style="width: 150px">
                <el-option label="文本" value="txt" />
                <el-option label="CSV" value="csv" />
                <el-option label="JSON" value="json" />
              </el-select>
              <el-input-number v-model="batchConfig.columnIndex" placeholder="列索引" :min="0" style="width: 150px" />
              <el-checkbox v-model="batchConfig.loop">循环读取</el-checkbox>
            </div>
            <div v-else-if="batchRuleType === 'binary'" style="display: flex; flex-direction: column; gap: 8px">
              <el-select v-model="batchConfig.binaryMode" placeholder="模式" style="width: 150px">
                <el-option label="生成图片" value="generate" />
                <el-option label="从文件夹读取" value="folder" />
              </el-select>
              <div v-if="batchConfig.binaryMode === 'generate'" style="display: flex; gap: 10px; flex-wrap: wrap">
                <el-input-number v-model="batchConfig.width" placeholder="宽度" :min="1" style="width: 120px" />
                <el-input-number v-model="batchConfig.height" placeholder="高度" :min="1" style="width: 120px" />
                <el-select v-model="batchConfig.format" placeholder="格式" style="width: 120px">
                  <el-option label="PNG" value="png" />
                  <el-option label="JPEG" value="jpeg" />
                  <el-option label="GIF" value="gif" />
                </el-select>
              </div>
              <div v-else-if="batchConfig.binaryMode === 'folder'" style="display: flex; flex-direction: column; gap: 8px">
                <el-input v-model="batchConfig.folderPath" placeholder="文件夹路径" />
                <el-input v-model="batchConfig.extensionsText" placeholder="扩展名，用逗号分隔（如：jpg,png,gif）" />
                <el-checkbox v-model="batchConfig.loop">循环读取</el-checkbox>
              </div>
            </div>
            <div v-else-if="batchRuleType === 'foreign'" style="display: flex; flex-direction: column; gap: 8px">
              <el-input v-model="batchConfig.foreignTable" placeholder="外键关联表名" />
              <el-checkbox v-model="batchConfig.randomSelect">随机选择</el-checkbox>
            </div>
          </el-form-item>
        </el-form>
        <template #footer>
          <el-button @click="showBatchDialog = false" :disabled="actionLoading.get('batch')">取消</el-button>
          <el-button type="primary" @click="applyBatchSettings" :loading="actionLoading.get('batch')" :disabled="actionLoading.get('batch')">应用</el-button>
        </template>
      </el-dialog>
  
      <!-- 数据预览对话框 -->
      <el-dialog v-model="showPreviewDialog" title="数据预览" width="90%" :before-close="closePreview">
        <div v-loading="previewLoading">
          <div style="margin-bottom: 15px; display: flex; gap: 10px; align-items: center">
            <span>预览数量：</span>
            <el-input-number v-model="previewCount" :min="1" :max="50" :step="1" />
            <el-button type="primary" size="small" @click="doPreview" :loading="previewLoading" :disabled="previewLoading">刷新预览</el-button>
          </div>
          <el-table :data="previewData" border stripe max-height="500" style="width: 100%">
            <el-table-column 
              v-for="field in previewTableColumns" 
              :key="field.fieldName"
              :prop="field.fieldName"
              :label="field.fieldName"
              min-width="150"
              show-overflow-tooltip
            >
              <template #default="{ row }">
                <span>{{ row[field.fieldName] }}</span>
              </template>
            </el-table-column>
          </el-table>
          <div v-if="previewData.length === 0 && !previewLoading" style="text-align: center; padding: 40px; color: #909399">
            <el-empty description="暂无预览数据" />
          </div>
        </div>
      </el-dialog>
    </div>
  </template>
  
  <script setup>
  import { ref, onMounted, computed, watch } from 'vue'
  import { useRoute, useRouter } from 'vue-router'
  import { ElMessage } from 'element-plus'
  import { Search, WarningFilled, InfoFilled, Refresh } from '@element-plus/icons-vue'
  import api from '../api'
  import { useTemplateStore } from '../stores/template'
  
  const templateStore = useTemplateStore()
  
  const route = useRoute()
  const router = useRouter()
  
  const tableName = ref(route.query.table || '')
  const connectionId = ref(route.query.connection_id || '')
  const database = ref(route.query.database || '')
  const taskId = ref(route.query.task_id || '')
  const isEditMode = computed(() => !!taskId.value)
  const tableSchema = ref(null)
  const fieldRules = ref([])
  const fieldSearchText = ref('')
  
  // 外键配置相关数据（按字段名索引）
  const foreignDatabases = ref({}) // { fieldName: [db1, db2, ...] }
  const foreignTables = ref({}) // { fieldName: [table1, table2, ...] }
  const foreignFields = ref({}) // { fieldName: [{name: 'field1', type: 'varchar'}, ...] }
  const foreignDatabasesLoading = ref({}) // { fieldName: true/false }
  const foreignTablesLoading = ref({}) // { fieldName: true/false }
  const foreignFieldsLoading = ref({}) // { fieldName: true/false }
  
  // 基本信息配置折叠状态（空数组表示折叠）
  const showBasicConfig = ref([])
  
  // 表格最大高度（动态计算，根据基本信息配置是否展开）
  const tableMaxHeight = computed(() => {
    // 如果基本信息配置展开，表格高度小一些；如果折叠，表格高度大一些
    return showBasicConfig.value.includes('basic') ? 'calc(100vh - 500px)' : 'calc(100vh - 350px)'
  })
  const showPreviewDialog = ref(false)
  const previewData = ref([])
  const previewLoading = ref(false)
  const previewCount = ref(10)
  
  // 字段示例值缓存（fieldName -> exampleValue）
  const fieldExamples = ref(new Map())
  // 字段示例加载状态（fieldName -> loading）
  const fieldExampleLoading = ref(new Map())
  
  // 模板相关
  const showSaveTemplateDialog = ref(false)
  const showLoadTemplateDialog = ref(false)
  const showBatchDialog = ref(false)
  const showPreviewTemplateDialog = ref(false)
  const templateForm = ref({ name: '', description: '' })
  const templates = computed(() => templateStore.allTemplates || [])
  const previewTemplateData = ref(null)
  const previewTemplateFields = ref([])
  
  // 操作按钮的 loading 状态
  const actionLoading = ref(new Map())
  
  // 过滤后的字段规则
  const filteredFieldRules = computed(() => {
    if (!fieldSearchText.value) {
      return fieldRules.value
    }
    const searchText = fieldSearchText.value.toLowerCase()
    return fieldRules.value.filter(field => 
      field.fieldName.toLowerCase().includes(searchText) ||
      field.fieldType.toLowerCase().includes(searchText)
    )
  })
  
  // 预览表格列（按照字段规则顺序）
  const previewTableColumns = computed(() => {
    return fieldRules.value.map(field => ({
      fieldName: field.fieldName,
      fieldType: field.fieldType
    }))
  })
  
  const taskConfig = ref({
    name: '',
    totalRows: 1000,
    batchSize: 1000,
    threadCount: 1  // 线程数，1-10，主键将统一生成以避免冲突
  })
  
  // 获取缓存 key
  const getCacheKey = () => {
    return `task_config_${connectionId.value}_${database.value}_${tableName.value}`
  }
  
  // 保存配置到缓存
  const saveConfigToCache = () => {
    try {
      const cacheData = {
        fieldRules: fieldRules.value,
        taskConfig: taskConfig.value,
        timestamp: Date.now()
      }
      localStorage.setItem(getCacheKey(), JSON.stringify(cacheData))
    } catch (error) {
      // 保存配置到缓存失败（静默处理）
    }
  }
  
  // 从缓存恢复配置
  const loadConfigFromCache = () => {
    try {
      const cacheKey = getCacheKey()
      const cached = localStorage.getItem(cacheKey)
      if (cached) {
        const cacheData = JSON.parse(cached)
        // 检查缓存是否过期（24小时）
        const cacheAge = Date.now() - (cacheData.timestamp || 0)
        if (cacheAge < 24 * 60 * 60 * 1000) {
          return cacheData
        } else {
          // 缓存过期，清除
          localStorage.removeItem(cacheKey)
        }
      }
    } catch (error) {
      // 从缓存恢复配置失败（静默处理）
    }
    return null
  }
  
  // 清除缓存
  const clearCache = () => {
    try {
      localStorage.removeItem(getCacheKey())
    } catch (error) {
      // 清除缓存失败（静默处理）
    }
  }
  
  // 加载任务配置（编辑模式）
  const loadTaskConfig = async () => {
    if (!taskId.value) return
    
    try {
      const task = await api.getTask(taskId.value)
      if (!task) {
        ElMessage.error('任务不存在')
        router.push('/tasks')
        return
      }

      // 检查任务状态，只有未运行的任务可以编辑
      if (task.status === 'running' || task.status === 'paused') {
        ElMessage.warning('无法编辑正在运行或已暂停的任务')
        router.push('/tasks')
        return
      }

      // 填充任务基本信息
      if (task.name) {
        taskConfig.value.name = task.name
      }
      if (task.config) {
        if (task.config.total_rows) {
          taskConfig.value.totalRows = task.config.total_rows
        }
        if (task.config.batch_size) {
          taskConfig.value.batchSize = task.config.batch_size
        }
      }

      // 更新连接ID和数据库（如果任务中有）
      if (task.connection_id) {
        connectionId.value = task.connection_id
      }
      if (task.database) {
        database.value = task.database
      }
      if (task.table) {
        tableName.value = task.table
      }
    } catch (error) {
      ElMessage.error('加载任务配置失败: ' + (error.formattedMessage || error.message))
      router.push('/tasks')
    }
  }

  onMounted(async () => {
    if (!tableName.value && !taskId.value) {
      ElMessage.error('表名或任务ID不能为空')
      router.back()
      return
    }

    // 如果是编辑模式，先加载任务配置
    if (isEditMode.value) {
      await loadTaskConfig()
      // 确保在加载任务配置后再加载表结构
      if (tableName.value && database.value) {
        await loadTableSchema()
      } else {
        ElMessage.error('任务配置中缺少表名或数据库信息')
        router.push('/tasks')
        return
      }
    } else {
      await loadTableSchema()
    }
    
    await loadTemplates()
    
    // 如果是编辑模式，加载任务配置到字段规则
    if (isEditMode.value) {
      try {
        const task = await api.getTask(taskId.value)
        if (task && task.config && task.config.field_rules) {
          
          // 将任务配置的字段规则应用到字段规则列表
          const fieldMap = new Map()
          fieldRules.value.forEach(field => {
            fieldMap.set(field.fieldName, field)
          })

          // 应用任务配置的字段规则
          task.config.field_rules.forEach(rule => {
            const field = fieldMap.get(rule.field_name)
            if (field) {
              // 先设置规则类型
              field.ruleType = rule.rule_type
              // 然后获取默认配置作为基础
              const defaultConfig = getDefaultConfig(field)
              // 深度合并任务配置和默认配置，确保所有配置项都被恢复
              const mergedConfig = { ...defaultConfig, ...(rule.config || {}) }
              
              // 特殊处理：如果是序列类型且 max_value 为 0，使用生成数量作为默认值
              if (rule.rule_type === 'increment' && (!rule.config || !rule.config.max_value || rule.config.max_value === 0)) {
                mergedConfig.maxValue = taskConfig.value.totalRows || 1000
              }
              
              // 特殊处理：如果是 random_string 类型，根据 fixed_length 判断长度模式
              if (rule.rule_type === 'random_string' && rule.config) {
                const ruleConfig = rule.config
                // 根据配置判断长度模式（兼容旧配置）
                const lengthMode = (ruleConfig.fixed_length > 0) ? 'fixed' : 'random'
                mergedConfig.lengthMode = lengthMode
                
                // 如果是从旧配置加载，需要转换格式
                if (ruleConfig.fixed_length !== undefined) {
                  mergedConfig.fixedLength = ruleConfig.fixed_length
                }
                if (ruleConfig.min_length !== undefined) {
                  mergedConfig.minLength = ruleConfig.min_length
                }
                if (ruleConfig.max_length !== undefined) {
                  mergedConfig.maxLength = ruleConfig.max_length
                }
                if (ruleConfig.char_set !== undefined) {
                  mergedConfig.charSet = ruleConfig.char_set
                }
                if (ruleConfig.case !== undefined) {
                  mergedConfig.case = ruleConfig.case
                }
                if (ruleConfig.number_position !== undefined) {
                  mergedConfig.numberPosition = ruleConfig.number_position
                }
                if (ruleConfig.prefix !== undefined) {
                  mergedConfig.prefix = ruleConfig.prefix
                }
                if (ruleConfig.suffix !== undefined) {
                  mergedConfig.suffix = ruleConfig.suffix
                }
                if (ruleConfig.custom_chars !== undefined) {
                  mergedConfig.customChars = ruleConfig.custom_chars
                }
              }
              
              field.config = mergedConfig
              
              // 触发规则类型变化，确保配置正确应用
              // 注意：这里不调用 onRuleTypeChange，因为我们已经手动设置了配置
            } else {
            }
          })
          
          ElMessage.success('任务配置已恢复')
        } else {
        }
      } catch (error) {
        // 加载任务字段规则失败（已通过 ElMessage 提示用户）
        ElMessage.error('加载任务配置失败: ' + (error.formattedMessage || error.message))
      }
    }
    
    // 延迟生成示例值，等待字段规则初始化完成
    setTimeout(() => {
      generateAllFieldExamples()
    }, 500)
  })
  
  // 监听配置变化，自动保存到缓存（在 setup 阶段注册）
  // 监听 totalRows 变化，自动更新序列类型字段的 maxValue
  watch(() => taskConfig.value.totalRows, (newTotalRows) => {
    if (newTotalRows && newTotalRows > 0) {
      fieldRules.value.forEach(field => {
        if (field.ruleType === 'increment' && field.config) {
          // 如果 maxValue 为 0 或未设置，或者小于新的 totalRows，则更新
          if (!field.config.maxValue || field.config.maxValue === 0 || field.config.maxValue < newTotalRows) {
            field.config.maxValue = newTotalRows
          }
        }
      })
    }
  })
  
  watch([fieldRules, taskConfig], () => {
    // 只有在字段规则已初始化后才保存（避免初始化时保存空数据）
    if (fieldRules.value && fieldRules.value.length > 0) {
      saveConfigToCache()
    }
  }, { deep: true })
  
  // 加载外键数据库列表
  const loadForeignDatabases = async (fieldName) => {
    if (!connectionId.value) {
      ElMessage.warning('请先选择数据库连接')
      return
    }
    
    // 如果正在加载，直接返回
    if (foreignDatabasesLoading.value[fieldName]) {
      return
    }
    
    foreignDatabasesLoading.value[fieldName] = true
    try {
      const response = await api.getDatabases(connectionId.value)
      const databases = response.databases || []
      
      // 处理达梦数据库返回的对象数组格式 {name, username}
      // 其他数据库返回的是字符串数组
      const processedDatabases = databases.map(db => {
        if (typeof db === 'string') {
          return db
        } else if (db && typeof db === 'object' && db.name) {
          return db.name // 达梦数据库格式：使用 name 字段
        }
        return String(db)
      })
      
      foreignDatabases.value[fieldName] = processedDatabases
      
      // 如果当前字段没有选择数据库，默认使用当前数据库
      const field = fieldRules.value.find(f => f.fieldName === fieldName)
      if (field && field.config && !field.config.foreignDatabase && database.value) {
        field.config.foreignDatabase = database.value
      }
    } catch (error) {
      // 加载数据库列表失败（已通过 ElMessage 提示用户）
      ElMessage.error('加载数据库列表失败: ' + (error.formattedMessage || error.message))
    } finally {
      foreignDatabasesLoading.value[fieldName] = false
    }
  }
  
  // 加载外键表列表
  const loadForeignTables = async (fieldName, foreignDatabase) => {
    if (!connectionId.value || !foreignDatabase) {
      return
    }
    
    foreignTablesLoading.value[fieldName] = true
    try {
      const response = await api.getTables(foreignDatabase, connectionId.value)
      // 处理响应格式：可能是 response.tables 或 response.data.tables
      let tables = []
      if (response.tables) {
        tables = response.tables
      } else if (response.data && response.data.tables) {
        tables = response.data.tables
      } else if (Array.isArray(response)) {
        tables = response
      }
      
      // 处理可能的对象数组格式（虽然 GetTables 应该返回字符串数组，但为了兼容性还是处理一下）
      const processedTables = tables.map(table => {
        if (typeof table === 'string') {
          return table
        } else if (table && typeof table === 'object' && table.name) {
          return table.name
        }
        return String(table)
      })
      
      foreignTables.value[fieldName] = processedTables
    } catch (error) {
      // 加载表列表失败（已通过 ElMessage 提示用户）
      ElMessage.error('加载表列表失败: ' + (error.formattedMessage || error.message))
    } finally {
      foreignTablesLoading.value[fieldName] = false
    }
  }
  
  // 加载外键字段列表
  const loadForeignFields = async (fieldName, foreignDatabase, foreignTable) => {
    if (!connectionId.value || !foreignDatabase || !foreignTable) {
      return
    }
    
    foreignFieldsLoading.value[fieldName] = true
    try {
      const response = await api.getTableSchema(foreignDatabase, foreignTable, connectionId.value)
      const fields = (response.fields || []).map(f => ({
        name: f.name || f.Name,
        type: f.type || f.Type
      }))
      foreignFields.value[fieldName] = fields
    } catch (error) {
      // 加载字段列表失败（已通过 ElMessage 提示用户）
      ElMessage.error('加载字段列表失败: ' + (error.formattedMessage || error.message))
    } finally {
      foreignFieldsLoading.value[fieldName] = false
    }
  }
  
  // 处理外键数据库选择变化
  const onForeignDatabaseChange = (field) => {
    const foreignDatabase = field.config.foreignDatabase
    if (foreignDatabase) {
      loadForeignTables(field.fieldName, foreignDatabase)
      // 清空表和字段选择
      field.config.foreignTable = ''
      field.config.foreignField = ''
      foreignTables.value[field.fieldName] = []
      foreignFields.value[field.fieldName] = []
    }
  }
  
  // 处理外键表选择变化
  const onForeignTableChange = (field) => {
    const foreignDatabase = field.config.foreignDatabase
    const foreignTable = field.config.foreignTable
    if (foreignDatabase && foreignTable) {
      loadForeignFields(field.fieldName, foreignDatabase, foreignTable)
      // 清空字段选择
      field.config.foreignField = ''
      foreignFields.value[field.fieldName] = []
    }
  }
  
  // 处理外键生成模式变化
  const onForeignGenerationModeChange = (field) => {
    // 如果切换到重复模式，确保有默认的重复次数
    if (field.config.generationMode === 'repeat') {
      if (!field.config.repeatMin || field.config.repeatMin <= 0) {
        field.config.repeatMin = 1
      }
      if (!field.config.repeatMax || field.config.repeatMax < field.config.repeatMin) {
        field.config.repeatMax = field.config.repeatMin || 3
      }
    }
  }
  
  const loadTableSchema = async () => {
    // 如果是编辑模式但还没有表名，等待任务配置加载完成
    if (isEditMode.value && !tableName.value) {
      return
    }
    
    if (!tableName.value || !database.value) {
      return
    }
    
    try {
      const params = { database: database.value }
      if (connectionId.value) {
        params.connection_id = connectionId.value
      }
      const response = await api.getTableSchema(database.value, tableName.value, connectionId.value)
      tableSchema.value = response
      
      
      // 初始化字段规则
      if (!response.fields || response.fields.length === 0) {
        ElMessage.warning('表结构中没有字段数据')
        fieldRules.value = []
        return
      }
      
      // 编辑模式下不使用缓存，直接使用默认配置初始化，后续会应用任务配置
      // 非编辑模式下才尝试从缓存恢复配置
      if (!isEditMode.value) {
        const cachedData = loadConfigFromCache()
        if (cachedData && cachedData.fieldRules && cachedData.fieldRules.length > 0) {
          // 获取当前表结构的字段名
          const currentFieldNames = response.fields.map(f => f.name || f.Name).sort()
          const cachedFieldNames = cachedData.fieldRules.map(f => f.fieldName).sort()
          
          // 检查字段名是否完全匹配
          if (currentFieldNames.length === cachedFieldNames.length && 
              currentFieldNames.every((name, index) => name === cachedFieldNames[index])) {
            // 字段匹配，使用缓存的配置，但需要确保字段元数据是最新的
            const fieldMap = new Map()
            response.fields.forEach(field => {
              fieldMap.set(field.name || field.Name, field)
            })
            
            // 合并缓存的配置和最新的字段元数据
            fieldRules.value = cachedData.fieldRules.map(cachedRule => {
              const field = fieldMap.get(cachedRule.fieldName)
              if (!field) return cachedRule
              
              // 保留缓存的规则类型和配置，但更新字段元数据
              const restoredRule = {
                ...cachedRule,
                fieldType: field.type || field.Type,
                goType: (field.go_type || field.GoType || '').toLowerCase(),
                isPrimaryKey: field.is_primary_key || field.IsPrimaryKey || false,
                isForeignKey: field.is_foreign_key || field.IsForeignKey || false,
                isUnique: field.is_unique || field.IsUnique || false,
                isNullable: field.is_nullable !== false && field.IsNullable !== false,
                defaultValue: field.default_value || field.DefaultValue,
                maxLength: field.max_length || field.MaxLength || 0,
                precision: field.precision || field.Precision || 0,
                scale: field.scale || field.Scale || 0,
                // 更新外键信息（从接口返回的最新字段信息中获取）
                foreignTable: field.foreign_table || field.ForeignTable || field.foreignTable || cachedRule.foreignTable || '',
                foreignDatabase: field.foreign_database || field.ForeignDatabase || field.foreignDatabase || cachedRule.foreignDatabase || ''
              }
              
              // 兼容性处理：如果是 random_string 类型且没有 lengthMode，根据 fixedLength 自动设置
              if (restoredRule.ruleType === 'random_string' && restoredRule.config) {
                if (!restoredRule.config.lengthMode) {
                  restoredRule.config.lengthMode = (restoredRule.config.fixedLength > 0) ? 'fixed' : 'random'
                }
              }
              
              
              return restoredRule
            })
            
            // 恢复任务配置
            if (cachedData.taskConfig) {
              taskConfig.value = { ...taskConfig.value, ...cachedData.taskConfig }
            }
            
            ElMessage.info('已恢复上次的配置')
            return
          } else {
            // 字段不匹配，清除缓存
            clearCache()
          }
        }
      }
      
      // 没有缓存或缓存不匹配，使用默认配置初始化
      fieldRules.value = response.fields.map(field => {
        const defaultRuleType = getDefaultRuleType(field)
        // 创建一个包含 ruleType 的临时对象，以便 getDefaultConfig 能正确获取规则类型
        const fieldWithRuleType = {
          ...field,
          ruleType: defaultRuleType,
          go_type: field.go_type || field.GoType,
          GoType: field.go_type || field.GoType,
          fieldType: field.type || field.Type
        }
        const defaultConfig = getDefaultConfig(fieldWithRuleType)
        
        // 如果是序列类型，确保 maxValue 与生成数量一致
        if (defaultRuleType === 'increment') {
          const defaultMaxValue = taskConfig.value.totalRows || 1000
          defaultConfig.maxValue = defaultMaxValue
        }
        
        const fieldRule = {
          fieldName: field.name || field.Name,
          fieldType: field.type || field.Type,
          goType: (field.go_type || field.GoType || '').toLowerCase(),
          ruleType: defaultRuleType,
          config: defaultConfig,
          isPrimaryKey: field.is_primary_key || field.IsPrimaryKey || false,
          isForeignKey: field.is_foreign_key || field.IsForeignKey || false,
          isUnique: field.is_unique || field.IsUnique || false,
          isNullable: field.is_nullable !== false && field.IsNullable !== false,
          defaultValue: field.default_value || field.DefaultValue,
          maxLength: field.max_length || field.MaxLength || 0,
          precision: field.precision || field.Precision || 0,
          scale: field.scale || field.Scale || 0,
          // 保存外键信息（尝试多种可能的字段名格式）
          foreignTable: field.foreign_table || field.ForeignTable || field.foreignTable || '',
          foreignDatabase: field.foreign_database || field.ForeignDatabase || field.foreignDatabase || ''
        }
        
        
        return fieldRule
      })
      
    } catch (error) {
      // 加载表结构失败（已通过 ElMessage 提示用户）
      const errorMsg = error.formattedMessage || error.message || ''
      // 如果是连接失败的错误，提供更友好的提示
      if (errorMsg.includes('连接失败') || errorMsg.includes('连接未建立') || errorMsg.includes('请先连接')) {
        ElMessage.warning('数据库连接失败，系统已尝试自动重连。如果问题持续，请检查连接配置')
      } else {
        ElMessage.error('加载表结构失败: ' + errorMsg)
      }
    }
  }
  
  const getDefaultRuleType = (field) => {
    // 兼容多种属性命名方式：驼峰命名（isPrimaryKey）和下划线命名（is_primary_key）
    if (field.isPrimaryKey || field.is_primary_key || field.IsPrimaryKey) {
      // 主键默认使用序列（increment 类型）
      return 'increment'
    }
    if (field.isUnique || field.is_unique || field.IsUnique) {
      // 唯一约束默认使用 UUID（function 类型）
      return 'function'
    }
    if (field.isForeignKey || field.is_foreign_key || field.IsForeignKey) {
      return 'foreign'
    }
    
    const type = (field.go_type || field.GoType || '').toLowerCase()
    const fieldType = (field.type || field.Type || '').toLowerCase()
    if (type.includes('int')) {
      return 'random_number'
    } else if (type.includes('float') || type.includes('decimal')) {
      return 'random_number'
    } else if (type.includes('time') || type.includes('date')) {
      return 'random_date'
    } else if (fieldType.includes('bit')) {
      // BIT 类型默认使用固定值（0 或 1）
      return 'fixed'
    } else if (type === 'bool') {
      // 布尔类型（非 BIT）使用列表选择
      return 'list'
    } else if (fieldType.includes('varchar') || fieldType.includes('char') || (type === 'string' && !fieldType.includes('text'))) {
      // VARCHAR 类型默认使用正则表达式
      return 'regex'
    } else if (fieldType.includes('text') || fieldType.includes('clob')) {
      // TEXT 类型默认使用文本（random_string）
      return 'random_string'
    } else {
      return 'random_string'
    }
  }
  
  const getDefaultConfig = (field) => {
    const ruleType = field.ruleType || ''
    
    // 根据规则类型返回默认配置
    if (ruleType === 'reference') {
      return { expression: '', fields: [], fieldsText: '' }
    } else if (ruleType === 'geographic') {
      return { type: 'city', country: '' }
    } else if (ruleType === 'file') {
      return { filePath: '', fileType: 'txt', columnIndex: 0, loop: false }
    } else if (ruleType === 'binary') {
      return { mode: 'generate', width: 100, height: 100, format: 'png', folderPath: '', extensions: [], extensionsText: '', loop: false }
    } else if (ruleType === 'foreign') {
      return { 
        foreignDatabase: database.value || '', 
        foreignTable: '', 
        foreignField: field.fieldName || field.name || '',
        generationMode: 'random',
        repeatMin: 1,
        repeatMax: 3
      }
    } else if (ruleType === 'fixed') {
      // 固定值：根据字段类型返回合适的默认值
      const type = (field.go_type || field.GoType || '').toLowerCase()
      const fieldType = (field.fieldType || field.type || '').toLowerCase()
      if (fieldType.includes('bit')) {
        // BIT 类型：默认值为 0
        return { value: 0 }
      } else if (type === 'bool' || type.includes('boolean')) {
        // 布尔类型：默认值为 false
        return { value: false }
      } else {
        // 其他类型：默认值为空字符串
        return { value: '' }
      }
    } else if (ruleType === 'list') {
      // 列表选择：根据字段类型返回合适的默认值
      const type = (field.go_type || field.GoType || field.fieldType || '').toLowerCase()
      if (type === 'bool' || type.includes('boolean') || type.includes('bit')) {
        // 布尔类型：true, false
        return { values: [true, false], valuesText: 'true\nfalse', allowRepeat: true, weights: [] }
      } else if (type.includes('int') || type.includes('float') || type.includes('decimal') || type.includes('numeric')) {
        // 数字类型：1, 2, 3
        return { values: [1, 2, 3], valuesText: '1\n2\n3', allowRepeat: true, weights: [] }
      } else {
        // 字符串类型或其他：选项1, 选项2, 选项3
        return { values: ['选项1', '选项2', '选项3'], valuesText: '选项1\n选项2\n选项3', allowRepeat: true, weights: [] }
      }
    } else if (ruleType === 'regex') {
      // VARCHAR 类型默认使用 [A-Za-z0-9]{10}，但需要检查字段长度限制
      const fieldType = (field.fieldType || field.type || '').toLowerCase()
      const maxLength = field.maxLength || field.max_length || field.MaxLength || 0
      let patternLength = 10
      
      // 如果是 VARCHAR 类型，检查长度限制
      if (fieldType.includes('varchar') || fieldType.includes('char')) {
        if (maxLength > 0 && patternLength > maxLength) {
          patternLength = maxLength
        }
      }
      
      return { 
        pattern: `[A-Za-z0-9]{${patternLength}}`, 
        presetPattern: 'custom' 
      }
    } else if (ruleType === 'increment') {
      // 序列默认配置：开始1，递增1，最大值与生成数量一致
      const defaultMaxValue = taskConfig.value.totalRows || 1000
      return { 
        startValue: 1, 
        step: 1, 
        cycle: false, 
        maxValue: defaultMaxValue,
        format: '',
        precision: 0,
        scale: 0
      }
    } else if (ruleType === 'random_string') {
      // 随机文本规则：根据字段类型设置默认配置
      const fieldType = (field.fieldType || field.type || '').toLowerCase()
      const isTextType = fieldType.includes('text') || fieldType.includes('clob')
      
      if (isTextType) {
        // TEXT 类型：默认 100-10000 长度的字符串
        return { 
          minLength: 100, 
          maxLength: 10000, 
          charSet: ['letters', 'numbers', 'chinese', 'special'], // 默认全选
          fixedLength: 0,
          case: 'mixed',
          numberPosition: 'none',
          prefix: '',
          suffix: '',
          customChars: ''
        }
      } else {
        // 其他字符串类型：默认 5-20 长度的字符串
        return { 
          minLength: 5, 
          maxLength: 20, 
          charSet: ['letters', 'numbers', 'chinese', 'special'], // 默认全选
          fixedLength: 0,
          case: 'mixed',
          numberPosition: 'none',
          prefix: '',
          suffix: '',
          customChars: ''
        }
      }
    } else if (ruleType === 'function') {
      // 根据字段类型选择默认函数
      const type = (field.go_type || field.GoType || field.fieldType || '').toLowerCase()
      if (type.includes('time') || type.includes('date') || type.includes('timestamp')) {
        return { funcName: 'NOW', params: [] }
      } else {
        // UUID 默认配置：版本 v4，大小写混合，带-
        return { 
          funcName: 'UUID', 
          params: [],
          version: 'v4',
          case: 'mixed',
          withHyphen: true
        }
      }
    } else if (ruleType === 'template') {
      return { template: '' }
    } else if (ruleType === 'null') {
      return { probability: 1.0 } // 总是生成 NULL
    }
    
    // 根据字段类型返回默认配置
    const type = (field.go_type || field.GoType || '').toLowerCase()
    
    if (type.includes('int')) {
      // INT 类型：默认 0-1000 的整数，固定为整数，但要在字段范围内
      const range = getNumericFieldRange(field)
      const defaultMax = Math.min(1000, range.max)
      const defaultMin = Math.max(0, range.min)
      return { 
        min: defaultMin, 
        max: defaultMax, 
        isInt: true,  // INT 类型固定为整数，不能选择小数
        precision: 0,
        scale: 0,
        fixedLength: 0,
        distribution: 'uniform',
        mean: 0,
        stdDev: 0,
        lambda: 0
      }
    } else if (type.includes('float') || type.includes('decimal') || type.includes('numeric')) {
      // 浮点类型：默认 0-1000，可以选择整数或小数，可以设置小数位数，但要在字段范围内
      const range = getNumericFieldRange(field)
      const defaultMax = Math.min(1000, range.max)
      const defaultMin = Math.max(0, range.min)
      return { 
        min: defaultMin, 
        max: defaultMax, 
        isInt: true,  // 默认整数，但可以改为小数
        precision: 0,
        scale: 2,  // 默认2位小数（如果选择小数）
        fixedLength: 0,
        distribution: 'uniform',
        mean: 0,
        stdDev: 0,
        lambda: 0
      }
    } else if (type.includes('time') || type.includes('date')) {
      // 默认时间范围：2001-01-01 到今天
      const today = new Date()
      const todayStr = today.getFullYear() + '-' + 
        String(today.getMonth() + 1).padStart(2, '0') + '-' + 
        String(today.getDate()).padStart(2, '0')
      
      // 判断字段类型
      const isDateOnly = isDateOnlyType(field)
      const isTimestamp = isTimestampType(field)
      
      return { 
        startDate: '2001-01-01', 
        endDate: todayStr, 
        format: '',
        fullDay: isTimestamp ? true : undefined, // 日期时间类型默认一整天，DATE 类型不设置（undefined 表示不显示该选项）
        startTime: isTimestamp ? '09:00:00' : '', // 日期时间类型默认开始时间，DATE 类型不设置
        endTime: isTimestamp ? '18:00:00' : '', // 日期时间类型默认结束时间，DATE 类型不设置
        yearRange: [],
        yearList: [],
        monthRange: [],
        monthList: [],
        dayRange: [],
        dayList: [],
        hourRange: [],
        hourList: [],
        minuteRange: [],
        minuteList: [],
        secondRange: [],
        secondList: [],
        weekdayList: [],
        weekdayMode: 'all', // 'all' | 'weekdays' | 'custom'
        weekdayCustom: [1, 2, 3, 4, 5], // 默认选中周一到周五
        onlyWeekdays: false,
        onlyWeekends: false
      }
    } else if (type === 'bool') {
      return { values: [true, false], valuesText: 'true\nfalse', allowRepeat: true, weights: [] }
    } else {
      // 检查是否为 TEXT 类型
      const fieldType = (field.fieldType || field.type || '').toLowerCase()
      const isTextType = fieldType.includes('text') || fieldType.includes('clob')
      
      if (isTextType) {
        // TEXT 类型：默认 100-10000 长度的字符串
        return { 
          minLength: 100, 
          maxLength: 10000, 
          charSet: ['letters', 'numbers', 'chinese', 'special'], // 默认全选
          fixedLength: 0,
          case: 'mixed',
          numberPosition: 'none',
          prefix: '',
          suffix: '',
          customChars: ''
        }
      } else {
        // 其他字符串类型：默认 5-20 长度的字符串
        return { 
          minLength: 5, 
          maxLength: 20, 
          charSet: ['letters', 'numbers', 'chinese', 'special'], // 默认全选
          fixedLength: 0,
          case: 'mixed',
          numberPosition: 'none',
          prefix: '',
          suffix: '',
          customChars: ''
        }
      }
    }
  }
  
  const onRuleTypeChange = (field, options = {}) => {
    const { skipExample = false } = options
    // 切换规则类型时重置配置
    field.config = getDefaultConfig(field)
    
    // 如果是序列类型，确保 maxValue 与生成数量一致
    if (field.ruleType === 'increment') {
      const defaultMaxValue = taskConfig.value.totalRows || 1000
      field.config.maxValue = defaultMaxValue
    }
    
    // 如果是随机文本类型，确保 charSet 是数组且默认全选，lengthMode 默认为 'random'
    if (field.ruleType === 'random_string') {
      if (!field.config.lengthMode) {
        field.config.lengthMode = 'random'
      }
      if (!field.config.charSet || !Array.isArray(field.config.charSet) || field.config.charSet.length === 0) {
        field.config.charSet = ['letters', 'numbers', 'chinese', 'special']
      }
    }
    
    // 如果是日期时间类型，确保 fullDay 默认为 true
    if (field.ruleType === 'random_date') {
      const isTimestamp = isTimestampType(field)
      if (isTimestamp && field.config.fullDay === undefined) {
        field.config.fullDay = true
      }
    }
    
    // 如果是随机数字类型，确保 min 和 max 有默认值
    if (field.ruleType === 'random_number') {
      if (field.config.min === undefined || field.config.min === null) {
        const type = (field.goType || field.go_type || '').toLowerCase()
        const range = getNumericFieldRange(field)
        field.config.min = Math.max(0, range.min)
      }
      if (field.config.max === undefined || field.config.max === null) {
        const type = (field.goType || field.go_type || '').toLowerCase()
        const range = getNumericFieldRange(field)
        field.config.max = Math.min(1000, range.max)
      }
    }
    
    // 如果是外键类型，自动填充外键配置
    if (field.ruleType === 'foreign') {
      // 尝试多种可能的字段名格式来获取外键信息（在函数开始时就提取，确保在整个函数中都能使用）
      const fieldForeignTable = field.foreignTable || field.ForeignTable || field.foreign_table || field.ForeignTable || ''
      const fieldForeignDatabase = field.foreignDatabase || field.ForeignDatabase || field.foreign_database || field.ForeignDatabase || ''
      
      // 自动填充外键配置
      // 优先使用字段的外键数据库信息，如果没有则使用当前数据库
      if (!field.config.foreignDatabase) {
        if (fieldForeignDatabase) {
          field.config.foreignDatabase = fieldForeignDatabase
        } else if (database.value) {
          field.config.foreignDatabase = database.value
        }
      }
      
      // 如果字段有外键表信息，立即设置（不等待表列表加载）
      if (fieldForeignTable && !field.config.foreignTable) {
        field.config.foreignTable = fieldForeignTable
      }
      
      // 外键字段名通常是当前字段名本身
      if (!field.config.foreignField) {
        field.config.foreignField = field.fieldName
      }
      
      // 如果已选择数据库，自动加载表列表和字段列表（无论是否有外键信息）
      if (field.config.foreignDatabase) {
        // 先加载数据库列表（如果需要）
        loadForeignDatabases(field.fieldName).then(() => {
          // 加载表列表（无论是否已选择表）
          loadForeignTables(field.fieldName, field.config.foreignDatabase).then(() => {
            // 表列表加载完成后，如果还没有设置 foreignTable，再设置（确保表列表已存在）
            // 使用之前提取的外键信息（在函数开始时就提取了）
            if (fieldForeignTable && !field.config.foreignTable) {
              field.config.foreignTable = fieldForeignTable
            }
            
            // 如果已选择表，加载字段列表
            if (field.config.foreignTable) {
              loadForeignFields(field.fieldName, field.config.foreignDatabase, field.config.foreignTable)
            }
          })
        }).catch(() => {
          // 如果加载数据库列表失败，仍然尝试加载表列表
          loadForeignTables(field.fieldName, field.config.foreignDatabase).then(() => {
            // 表列表加载完成后，再设置外键表名（使用之前提取的外键信息）
            if (fieldForeignTable || field.isForeignKey) {
              const foreignTableName = fieldForeignTable
              if (foreignTableName) {
                field.config.foreignTable = foreignTableName
              }
              if (!field.config.foreignField) {
                field.config.foreignField = field.fieldName
              }
              if (field.config.foreignTable) {
                loadForeignFields(field.fieldName, field.config.foreignDatabase, field.config.foreignTable)
              }
            }
          })
        })
      } else {
        // 如果没有选择数据库，只加载数据库列表
        loadForeignDatabases(field.fieldName)
        
        // 即使没有数据库，也先设置外键表名（如果字段有外键信息，使用之前提取的值）
        if (fieldForeignTable || field.isForeignKey) {
          const foreignTableName = fieldForeignTable
          if (foreignTableName && !field.config.foreignTable) {
            field.config.foreignTable = foreignTableName
          }
          if (!field.config.foreignField) {
            field.config.foreignField = field.fieldName
          }
        }
      }
    }
    
    // 如果是 INT 类型且规则是 random_number，强制设置为整数
    if (isIntType(field) && field.ruleType === 'random_number') {
      field.config.isInt = true
    }
    
    // 确保数组字段已初始化（防止访问 undefined）
    if (field.config && field.ruleType === 'random_date') {
      if (!field.config.yearRange || !Array.isArray(field.config.yearRange)) field.config.yearRange = []
      if (!field.config.monthRange || !Array.isArray(field.config.monthRange)) field.config.monthRange = []
      if (!field.config.dayRange || !Array.isArray(field.config.dayRange)) field.config.dayRange = []
      if (!field.config.hourRange || !Array.isArray(field.config.hourRange)) field.config.hourRange = []
      if (!field.config.minuteRange || !Array.isArray(field.config.minuteRange)) field.config.minuteRange = []
      if (!field.config.secondRange || !Array.isArray(field.config.secondRange)) field.config.secondRange = []
      if (!field.config.yearList || !Array.isArray(field.config.yearList)) field.config.yearList = []
      if (!field.config.monthList || !Array.isArray(field.config.monthList)) field.config.monthList = []
      if (!field.config.dayList || !Array.isArray(field.config.dayList)) field.config.dayList = []
      if (!field.config.hourList || !Array.isArray(field.config.hourList)) field.config.hourList = []
      if (!field.config.minuteList || !Array.isArray(field.config.minuteList)) field.config.minuteList = []
      if (!field.config.secondList || !Array.isArray(field.config.secondList)) field.config.secondList = []
      if (!field.config.weekdayList || !Array.isArray(field.config.weekdayList)) field.config.weekdayList = []
      // 初始化星期模式
      if (!field.config.weekdayMode) field.config.weekdayMode = 'all'
      if (!field.config.weekdayCustom || !Array.isArray(field.config.weekdayCustom)) {
        field.config.weekdayCustom = [1, 2, 3, 4, 5]
      }
      // 根据模式设置 weekdayList
      if (field.config.weekdayMode === 'all') {
        field.config.weekdayList = []
      } else if (field.config.weekdayMode === 'weekdays') {
        field.config.weekdayList = [1, 2, 3, 4, 5]
      } else if (field.config.weekdayMode === 'custom') {
        field.config.weekdayList = field.config.weekdayCustom || []
      }
      // 确保范围数组至少有2个元素
      if (field.config.yearRange.length < 2) field.config.yearRange = [1900, 2100]
      if (field.config.monthRange.length < 2) field.config.monthRange = [1, 12]
      if (field.config.dayRange.length < 2) field.config.dayRange = [1, 31]
      if (field.config.hourRange.length < 2) field.config.hourRange = [0, 23]
      if (field.config.minuteRange.length < 2) field.config.minuteRange = [0, 59]
      if (field.config.secondRange.length < 2) field.config.secondRange = [0, 59]
      // 设置默认日期范围
      if (!field.config.startDate) {
        field.config.startDate = '2001-01-01'
      }
      if (!field.config.endDate) {
        const today = new Date()
        field.config.endDate = today.getFullYear() + '-' + 
          String(today.getMonth() + 1).padStart(2, '0') + '-' + 
          String(today.getDate()).padStart(2, '0')
      }
    }
    // 验证配置
    validateFieldRule(field)
    // 重新生成示例值（单个字段），除非明确跳过
    if (!skipExample) {
      generateFieldExample(field)
    }
  }
  
  // 验证字段规则配置是否符合字段边界要求
  const validateFieldRule = (field) => {
    if (!field || !field.ruleType) {
      return { valid: true, message: '' }
    }
  
    const goType = (field.goType || '').toLowerCase()
    const fieldType = (field.fieldType || '').toLowerCase()
    const maxLength = field.maxLength || 0
    const precision = field.precision || 0
    const scale = field.scale || 0
    const config = field.config || {}
  
    // 字符串类型（VARCHAR, CHAR, TEXT等）
    if (goType === 'string' || fieldType.includes('varchar') || fieldType.includes('char') || 
        fieldType.includes('text') || fieldType.includes('string')) {
      switch (field.ruleType) {
        case 'random_string':
          if (maxLength > 0) {
            if (config.minLength > maxLength) {
              return { valid: false, message: `最小长度(${config.minLength})不能超过字段最大长度(${maxLength})` }
            }
            if (config.maxLength > maxLength) {
              return { valid: false, message: `最大长度(${config.maxLength})不能超过字段最大长度(${maxLength})` }
            }
          }
          break
        
        case 'random_number':
          // 估算最大数字转换为字符串后的长度
          const maxNum = config.max || 1000
          const minNum = config.min || 0
          // 整数：最大数字的字符串长度
          // 浮点数：考虑小数点和精度
          let maxStrLength = 0
          if (config.isInt !== false) {
            maxStrLength = Math.max(String(Math.abs(Math.floor(maxNum))).length, String(Math.abs(Math.floor(minNum))).length)
            if (maxNum < 0 || minNum < 0) maxStrLength++ // 负号
          } else {
            // 浮点数：整数部分 + 小数点 + 小数部分（最多6位）
            const intPart = Math.max(String(Math.abs(Math.floor(maxNum))).length, String(Math.abs(Math.floor(minNum))).length)
            maxStrLength = intPart + 1 + 6 // 小数点 + 6位小数
            if (maxNum < 0 || minNum < 0) maxStrLength++ // 负号
          }
          if (maxLength > 0 && maxStrLength > maxLength) {
            return { valid: false, message: `生成的数字字符串长度(${maxStrLength})可能超过字段最大长度(${maxLength})，请减小数值范围` }
          }
          break
        
        case 'random_date':
          // 日期格式转换为字符串的长度
          // 默认格式：YYYY-MM-DD HH:mm:ss (19个字符)
          // 如果用户指定了format，使用format的长度；否则使用默认格式长度
          const dateFormat = config.format || 'YYYY-MM-DD HH:mm:ss'
          // 计算格式字符串的实际长度（将占位符替换为实际字符）
          // YYYY -> 4, MM -> 2, DD -> 2, HH -> 2, mm -> 2, ss -> 2
          // 简单估算：格式字符串长度（不考虑占位符）
          let estimatedLength = dateFormat.length
          // 如果格式字符串包含占位符，估算实际长度
          if (dateFormat.includes('YYYY')) {
            estimatedLength = estimatedLength - 4 + 4 // YYYY 占4个字符
          }
          if (dateFormat.includes('MM')) {
            estimatedLength = estimatedLength - 2 + 2
          }
          if (dateFormat.includes('DD')) {
            estimatedLength = estimatedLength - 2 + 2
          }
          if (dateFormat.includes('HH')) {
            estimatedLength = estimatedLength - 2 + 2
          }
          if (dateFormat.includes('mm')) {
            estimatedLength = estimatedLength - 2 + 2
          }
          if (dateFormat.includes('ss')) {
            estimatedLength = estimatedLength - 2 + 2
          }
          // 如果格式字符串中没有占位符，使用默认长度估算
          if (!dateFormat.includes('YYYY') && !dateFormat.includes('MM') && !dateFormat.includes('DD')) {
            estimatedLength = 19 // 默认日期时间格式长度
          }
          if (maxLength > 0 && estimatedLength > maxLength) {
            return { valid: false, message: `日期格式字符串长度(${estimatedLength})超过字段最大长度(${maxLength})，请使用更短的格式` }
          }
          break
        
        case 'fixed':
          const fixedValue = String(config.value || '')
          if (maxLength > 0 && fixedValue.length > maxLength) {
            return { valid: false, message: `固定值长度(${fixedValue.length})超过字段最大长度(${maxLength})` }
          }
          break
        
        case 'increment':
          // 估算递增数字的最大字符串长度
          const startValue = config.startValue || 1
          const step = config.step || 1
          // 简单估算：假设最多生成10000个值
          const maxIncrementValue = startValue + step * 10000
          const incrementStrLength = String(Math.abs(maxIncrementValue)).length
          if (maxIncrementValue < 0) incrementStrLength++ // 负号
          if (maxLength > 0 && incrementStrLength > maxLength) {
            return { valid: false, message: `递增数字字符串长度(${incrementStrLength})可能超过字段最大长度(${maxLength})，请调整起始值或步长` }
          }
          break
        
        case 'list':
          const values = config.values || []
          for (const val of values) {
            const valStr = String(val)
            if (maxLength > 0 && valStr.length > maxLength) {
              return { valid: false, message: `列表中的值"${valStr.substring(0, 20)}..."长度(${valStr.length})超过字段最大长度(${maxLength})` }
            }
          }
          break
        
        case 'template':
          // 模板生成的值长度难以精确估算，但可以检查模板本身
          const template = config.template || ''
          // 简单检查：如果模板很长，可能有问题
          if (maxLength > 0 && template.length > maxLength * 2) {
            return { valid: false, message: `模板长度可能导致生成的值超过字段最大长度(${maxLength})` }
          }
          break
      }
    }
  
    // 数值类型（INT, BIGINT, SMALLINT等）
    if (goType.includes('int') && !goType.includes('float') && !goType.includes('decimal')) {
      switch (field.ruleType) {
        case 'random_number':
          if (config.isInt === false) {
            return { valid: false, message: '整数类型字段不能使用浮点数规则' }
          }
          // 检查数值范围是否在字段类型支持的范围内
          const range = getNumericFieldRange(field)
          const minValue = config.min || 0
          const maxValue = config.max || 1000
          if (minValue < range.min) {
            return { valid: false, message: `最小值(${minValue})小于字段类型支持的最小值(${range.min})` }
          }
          if (maxValue > range.max) {
            return { valid: false, message: `最大值(${maxValue})大于字段类型支持的最大值(${range.max})` }
          }
          if (minValue > maxValue) {
            return { valid: false, message: `最小值(${minValue})不能大于最大值(${maxValue})` }
          }
          // 检查数值范围是否在整数范围内
          const maxInt = maxValue
          const minInt = minValue
          if (maxInt > Number.MAX_SAFE_INTEGER || minInt < Number.MIN_SAFE_INTEGER) {
            return { valid: false, message: '数值范围超出整数范围' }
          }
          break
        
        case 'increment':
          const startVal = config.startValue || 1
          const stepVal = config.step || 1
          if (!Number.isInteger(startVal) || !Number.isInteger(stepVal)) {
            return { valid: false, message: '整数类型字段的递增起始值和步长必须是整数' }
          }
          // 检查递增起始值是否在字段类型支持的范围内
          const incrementRange = getNumericFieldRange(field)
          if (startVal < incrementRange.min) {
            return { valid: false, message: `递增起始值(${startVal})小于字段类型支持的最小值(${incrementRange.min})` }
          }
          if (startVal > incrementRange.max) {
            return { valid: false, message: `递增起始值(${startVal})大于字段类型支持的最大值(${incrementRange.max})` }
          }
          // 估算最大可能值（假设生成10000个值）
          const estimatedMax = startVal + stepVal * 10000
          if (estimatedMax > incrementRange.max) {
            return { valid: false, message: `递增可能产生的最大值(${estimatedMax})超过字段类型支持的最大值(${incrementRange.max})，请调整起始值或步长` }
          }
          break
        
        case 'fixed':
          const fixedVal = config.value
          if (fixedVal !== null && fixedVal !== undefined && fixedVal !== '') {
            const numVal = Number(fixedVal)
            if (isNaN(numVal) || !Number.isInteger(numVal)) {
              return { valid: false, message: '整数类型字段的固定值必须是整数' }
            }
          }
          break
        
        case 'list':
          const listValues = config.values || []
          for (const val of listValues) {
            const numVal = Number(val)
            if (isNaN(numVal) || !Number.isInteger(numVal)) {
              return { valid: false, message: `列表中的值"${val}"不是有效的整数` }
            }
          }
          break
      }
    }
  
    // 浮点数类型（FLOAT, DOUBLE, DECIMAL, NUMERIC）
    if (goType.includes('float') || goType.includes('decimal') || goType.includes('numeric')) {
      switch (field.ruleType) {
        case 'random_number':
          // 检查数值范围是否在字段类型支持的范围内
          const floatRange = getNumericFieldRange(field)
          const floatMinValue = config.min || 0
          const floatMaxValue = config.max || 1000
          if (floatMinValue < floatRange.min) {
            return { valid: false, message: `最小值(${floatMinValue})小于字段类型支持的最小值(${floatRange.min})` }
          }
          if (floatMaxValue > floatRange.max) {
            return { valid: false, message: `最大值(${floatMaxValue})大于字段类型支持的最大值(${floatRange.max})` }
          }
          if (floatMinValue > floatMaxValue) {
            return { valid: false, message: `最小值(${floatMinValue})不能大于最大值(${floatMaxValue})` }
          }
          // 检查精度和小数位数
          if (precision > 0) {
            const maxNum = floatMaxValue
            const minNum = floatMinValue
            // 估算整数部分的位数
            const intPartDigits = Math.max(String(Math.abs(Math.floor(maxNum))).length, String(Math.abs(Math.floor(minNum))).length)
            if (intPartDigits + scale > precision) {
              return { valid: false, message: `数值范围可能导致精度超出字段定义(precision=${precision}, scale=${scale})` }
            }
          }
          break
        
        case 'fixed':
          const fixedVal = config.value
          if (fixedVal !== null && fixedVal !== undefined && fixedVal !== '') {
            const numVal = Number(fixedVal)
            if (isNaN(numVal)) {
              return { valid: false, message: '数值类型字段的固定值必须是有效数字' }
            }
            // 检查精度
            if (precision > 0) {
              const valStr = String(numVal)
              const parts = valStr.split('.')
              const intPart = parts[0].replace('-', '')
              const decPart = parts[1] || ''
              if (intPart.length + scale > precision) {
                return { valid: false, message: `固定值精度超出字段定义(precision=${precision}, scale=${scale})` }
              }
              if (decPart.length > scale) {
                return { valid: false, message: `固定值小数位数(${decPart.length})超过字段定义(${scale})` }
              }
            }
          }
          break
        
        case 'list':
          const listValues = config.values || []
          for (const val of listValues) {
            const numVal = Number(val)
            if (isNaN(numVal)) {
              return { valid: false, message: `列表中的值"${val}"不是有效数字` }
            }
            // 检查精度
            if (precision > 0) {
              const valStr = String(numVal)
              const parts = valStr.split('.')
              const intPart = parts[0].replace('-', '')
              const decPart = parts[1] || ''
              if (intPart.length + scale > precision) {
                return { valid: false, message: `列表中的值"${val}"精度超出字段定义` }
              }
              if (decPart.length > scale) {
                return { valid: false, message: `列表中的值"${val}"小数位数超过字段定义` }
              }
            }
          }
          break
      }
    }
  
    // 日期时间类型
    if (goType.includes('time') || goType.includes('date')) {
      switch (field.ruleType) {
        case 'random_number':
        case 'increment':
          return { valid: false, message: '日期时间类型字段不能使用数值规则' }
        
        case 'random_string':
          return { valid: false, message: '日期时间类型字段不能使用随机字符串规则' }
        
        case 'fixed':
          const fixedDateVal = config.value
          if (fixedDateVal !== null && fixedDateVal !== undefined && fixedDateVal !== '') {
            // 尝试解析为日期
            const dateVal = new Date(fixedDateVal)
            if (isNaN(dateVal.getTime())) {
              return { valid: false, message: '日期时间类型字段的固定值必须是有效日期' }
            }
          }
          break
        
        case 'list':
          const dateListValues = config.values || []
          for (const val of dateListValues) {
            const dateVal = new Date(val)
            if (isNaN(dateVal.getTime())) {
              return { valid: false, message: `列表中的值"${val}"不是有效日期` }
            }
          }
          break
      }
    }
  
    // 布尔类型
    if (goType === 'bool' || goType === 'boolean') {
      switch (field.ruleType) {
        case 'random_number':
        case 'increment':
        case 'random_date':
        case 'random_string':
          return { valid: false, message: '布尔类型字段只能使用列表选择或固定值规则' }
        
        case 'fixed':
          const fixedBoolVal = config.value
          if (fixedBoolVal !== null && fixedBoolVal !== undefined && fixedBoolVal !== '') {
            const boolVal = String(fixedBoolVal).toLowerCase()
            if (boolVal !== 'true' && boolVal !== 'false' && boolVal !== '1' && boolVal !== '0') {
              return { valid: false, message: '布尔类型字段的固定值必须是 true/false 或 1/0' }
            }
          }
          break
        
        case 'list':
          const boolListValues = config.values || []
          for (const val of boolListValues) {
            const boolVal = String(val).toLowerCase()
            if (boolVal !== 'true' && boolVal !== 'false' && boolVal !== '1' && boolVal !== '0') {
              return { valid: false, message: `列表中的值"${val}"不是有效的布尔值` }
            }
          }
          break
      }
    }
  
    // 二进制类型
    if (goType.includes('[]byte') || goType.includes('bytea') || 
        fieldType.includes('blob') || fieldType.includes('binary') || fieldType.includes('bytea')) {
      switch (field.ruleType) {
        case 'random_string':
        case 'random_number':
        case 'increment':
        case 'random_date':
          return { valid: false, message: '二进制类型字段只能使用二进制规则、固定值或空值' }
      }
    }
  
    return { valid: true, message: '' }
  }
  
  // 获取字段边界信息（用于tooltip显示）
  const getFieldBoundaryInfo = (field) => {
    if (!field) return ''
    
    const maxLength = field.maxLength || 0
    const precision = field.precision || 0
    const scale = field.scale || 0
    const goType = (field.goType || '').toLowerCase()
    const fieldType = (field.fieldType || '').toLowerCase()
    
    const info = []
    
    // 字符串类型显示最大长度
    if (goType === 'string' || fieldType.includes('varchar') || fieldType.includes('char') || 
        fieldType.includes('text') || fieldType.includes('string')) {
      if (maxLength > 0) {
        info.push(`字段最大长度: ${maxLength}`)
      }
    }
    
    // 数值类型显示精度和小数位
    if (goType.includes('int') || goType.includes('float') || goType.includes('decimal') || 
        goType.includes('numeric')) {
      if (precision > 0) {
        if (scale > 0) {
          info.push(`精度: ${precision}, 小数位: ${scale}`)
        } else {
          info.push(`精度: ${precision}`)
        }
      }
      // 如果是字符串类型但使用数值规则，也显示最大长度
      if (maxLength > 0 && (goType === 'string' || fieldType.includes('varchar') || fieldType.includes('char'))) {
        info.push(`字段最大长度: ${maxLength}`)
      }
    }
    
    // 日期时间类型显示最大长度（如果适用）
    if ((goType.includes('time') || goType.includes('date')) && maxLength > 0) {
      info.push(`字段最大长度: ${maxLength}`)
    }
    
    return info.length > 0 ? info.join('\n') : ''
  }
  
  // 验证所有字段规则
  const validateAllFieldRules = () => {
    const errors = []
    for (const field of fieldRules.value) {
      const result = validateFieldRule(field)
      if (!result.valid) {
        errors.push({
          field: field.fieldName,
          message: result.message
        })
      }
    }
    return errors
  }
  
  const parseListValues = (field) => {
    if (field.config.valuesText) {
      // 按行分割，过滤空行
      field.config.values = field.config.valuesText.split('\n')
        .map(v => v.trim())
        .filter(v => v.length > 0)
    } else {
      field.config.values = []
    }
  }
  
  const parseReferenceFields = (field) => {
    if (field.config.fieldsText) {
      field.config.fields = field.config.fieldsText.split(',').map(v => v.trim()).filter(v => v)
    } else {
      field.config.fields = []
    }
  }
  
  const parseBinaryExtensions = (field) => {
    if (field.config.extensionsText) {
      field.config.extensions = field.config.extensionsText.split(',').map(v => v.trim()).filter(v => v)
    } else {
      field.config.extensions = []
    }
  }
  
  // 判断是否为 INT 类型字段
  const isIntType = (field) => {
    const goType = (field.goType || '').toLowerCase()
    const fieldType = (field.fieldType || '').toLowerCase()
    // INT 类型：int, bigint, smallint, tinyint 等
    return goType.includes('int') && !goType.includes('float') && !goType.includes('decimal') && !goType.includes('numeric')
  }
  
  // 判断是否为浮点类型字段
  const isFloatType = (field) => {
    const goType = (field.goType || '').toLowerCase()
    const fieldType = (field.fieldType || '').toLowerCase()
    // 浮点类型：float, double, decimal, numeric
    return goType.includes('float') || goType.includes('decimal') || goType.includes('numeric') || 
           fieldType.includes('float') || fieldType.includes('double') || fieldType.includes('decimal') || fieldType.includes('numeric')
  }
  
  // 判断是否为纯日期类型字段（DATE，不包含时间部分）
  const isDateOnlyType = (field) => {
    const goType = (field.goType || field.go_type || '').toLowerCase()
    const fieldType = (field.fieldType || field.field_type || field.type || '').toLowerCase()
    // 纯日期类型：date（不包含 timestamp, datetime, time）
    return fieldType === 'date' || (fieldType.includes('date') && !fieldType.includes('timestamp') && !fieldType.includes('datetime') && !fieldType.includes('time'))
  }
  
  // 判断是否为日期时间类型字段（TIMESTAMP, DATETIME等，包含时间部分）
  const isTimestampType = (field) => {
    const goType = (field.goType || field.go_type || '').toLowerCase()
    const fieldType = (field.fieldType || field.field_type || field.type || '').toLowerCase()
    // 日期时间类型：timestamp, datetime, timestamptz
    return fieldType.includes('timestamp') || fieldType.includes('datetime') || 
           goType.includes('time') && !goType.includes('date') && !fieldType.includes('date') && !fieldType.includes('time')
  }
  
  // 格式化日期示例值显示
  const formatDateExample = (value, field) => {
    if (!value) return String(value)
    
    const valueStr = String(value)
    // 如果是 ISO 8601 格式（包含 T 和 Z）
    if (valueStr.includes('T') || valueStr.includes('Z')) {
      try {
        const date = new Date(valueStr)
        if (!isNaN(date.getTime())) {
          // 判断是否为纯日期类型
          if (isDateOnlyType(field)) {
            // DATE 类型：只显示日期部分 YYYY-MM-DD
            const year = date.getFullYear()
            const month = String(date.getMonth() + 1).padStart(2, '0')
            const day = String(date.getDate()).padStart(2, '0')
            return `${year}-${month}-${day}`
          } else {
            // 日期时间类型：显示日期和时间 YYYY-MM-DD HH:mm:ss
            const year = date.getFullYear()
            const month = String(date.getMonth() + 1).padStart(2, '0')
            const day = String(date.getDate()).padStart(2, '0')
            const hours = String(date.getHours()).padStart(2, '0')
            const minutes = String(date.getMinutes()).padStart(2, '0')
            const seconds = String(date.getSeconds()).padStart(2, '0')
            return `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`
          }
        }
      } catch (e) {
        // 解析失败，返回原值
      }
    }
    return valueStr
  }
  
  // 获取数值类型字段的范围限制
  const getNumericFieldRange = (field) => {
    const fieldType = (field.fieldType || '').toLowerCase()
    const goType = (field.goType || '').toLowerCase()
    const precision = field.precision || field.Precision || 0
    const scale = field.scale || field.Scale || 0
    
    // 整数类型
    if (goType.includes('int') && !goType.includes('float') && !goType.includes('decimal')) {
      if (fieldType.includes('tinyint')) {
        // TINYINT: -128 到 127 (有符号) 或 0 到 255 (无符号)
        // 默认假设有符号，但实际应该根据数据库配置判断，这里保守处理
        return { min: -128, max: 127 }
      } else if (fieldType.includes('smallint')) {
        // SMALLINT: -32,768 到 32,767 (有符号) 或 0 到 65,535 (无符号)
        return { min: -32768, max: 32767 }
      } else if (fieldType.includes('mediumint')) {
        // MEDIUMINT: -8,388,608 到 8,388,607 (有符号) 或 0 到 16,777,215 (无符号)
        return { min: -8388608, max: 8388607 }
      } else if (fieldType.includes('bigint')) {
        // BIGINT: -9,223,372,036,854,775,808 到 9,223,372,036,854,775,807
        // JavaScript 的 Number.MAX_SAFE_INTEGER 是 2^53 - 1，所以使用这个作为上限
        return { min: Number.MIN_SAFE_INTEGER, max: Number.MAX_SAFE_INTEGER }
      } else {
        // INT: -2,147,483,648 到 2,147,483,647 (有符号) 或 0 到 4,294,967,295 (无符号)
        return { min: -2147483648, max: 2147483647 }
      }
    }
    
    // 浮点类型
    if (goType.includes('float') || goType.includes('double') || goType.includes('decimal') || goType.includes('numeric')) {
      if (fieldType.includes('float')) {
        // FLOAT: 约 -3.4E+38 到 3.4E+38
        return { min: -3.4e38, max: 3.4e38 }
      } else if (fieldType.includes('double')) {
        // DOUBLE: 约 -1.7E+308 到 1.7E+308
        return { min: -1.7e308, max: 1.7e308 }
      } else if (fieldType.includes('decimal') || fieldType.includes('numeric')) {
        // DECIMAL/NUMERIC: 根据 precision 和 scale 确定
        // precision 是总位数，scale 是小数位数
        // 例如 DECIMAL(10,2) 可以存储 -99999999.99 到 99999999.99
        if (precision > 0 && scale >= 0 && scale <= precision) {
          const maxIntegerDigits = precision - scale
          // 计算最大值：例如 precision=10, scale=2，则整数部分最多8位，最大值是 99999999.99
          // 整数部分最大值：10^8 - 1 = 99999999
          // 小数部分最大值：0.99 (对于 scale=2)
          const maxIntegerPart = Math.pow(10, maxIntegerDigits) - 1
          const maxDecimalPart = scale > 0 ? (Math.pow(10, scale) - 1) / Math.pow(10, scale) : 0
          const maxValue = maxIntegerPart + maxDecimalPart
          const minValue = -maxValue
          return { min: minValue, max: maxValue }
        }
        // 如果没有 precision，使用默认范围（但限制在安全范围内）
        return { min: -999999999999.99, max: 999999999999.99 }
      }
    }
    
    // 默认范围（如果无法确定类型）
    return { min: Number.MIN_SAFE_INTEGER, max: Number.MAX_SAFE_INTEGER }
  }
  
  // 根据字段类型获取可用的规则选项（只保留常用规则）
  const getAvailableRules = (field) => {
    // 常用规则列表
    const commonRules = [
      { label: '随机文本', value: 'random_string' },
      { label: '随机数字', value: 'random_number' },
      { label: '随机日期', value: 'random_date' },
      { label: '固定值', value: 'fixed' },
      { label: '列表选择', value: 'list' },
      { label: '正则表达式', value: 'regex' }
    ]
    
    // 所有规则（用于特殊字段类型）
    const allRules = [
      ...commonRules,
      { label: '模板', value: 'template' },
      { label: '空值', value: 'null' },
      { label: '引用字段', value: 'reference' },
      { label: '地理数据', value: 'geographic' },
      { label: '从文件读取', value: 'file' },
      { label: '二进制/图片', value: 'binary' }
    ]
  
    const goType = (field.goType || '').toLowerCase()
    const fieldType = (field.fieldType || '').toLowerCase()
    const isNullable = field.isNullable !== false // 默认为可空
  
    // 如果是主键或唯一约束，只支持 UUID 和序列
    // 兼容多种属性命名方式：驼峰命名（isPrimaryKey）和下划线命名（is_primary_key）
    const isPrimaryKey = field.isPrimaryKey || field.is_primary_key || field.IsPrimaryKey || false
    const isUnique = field.isUnique || field.is_unique || field.IsUnique || false
    const isForeignKey = field.isForeignKey || field.is_foreign_key || field.IsForeignKey || false
    
    if (isPrimaryKey || isUnique) {
      return [
        { label: 'UUID', value: 'function' },
        { label: '序列', value: 'increment' }
      ]
    }
  
    // 如果是外键，只显示外键相关规则
    if (isForeignKey) {
      const rules = [
        { label: '外键引用', value: 'foreign' },
        { label: '固定值', value: 'fixed' }
      ]
      // 如果可空，添加 NULL 规则
      if (isNullable) {
        rules.push({ label: '空值', value: 'null' })
      }
      return rules
    }
  
    // 根据 Go 类型过滤规则（只返回常用规则）
    if (goType.includes('int') || goType.includes('float') || goType.includes('decimal') || goType.includes('numeric')) {
      // 数字类型
      const rules = [
        { label: '随机数字', value: 'random_number' },
        { label: '固定值', value: 'fixed' },
        { label: '序列', value: 'increment' },
        { label: '列表选择', value: 'list' }
      ]
      // 如果可空，添加 NULL 规则
      if (isNullable) {
        rules.push({ label: '空值', value: 'null' })
      }
      return rules
    } else if (goType.includes('time') || goType.includes('date')) {
      // 日期时间类型：根据字段类型区分显示
      const isDateOnly = isDateOnlyType(field)
      const isTimestamp = isTimestampType(field)
      const rules = []
      
      if (isDateOnly) {
        // DATE 类型：显示"随机日期"
        rules.push({ label: '随机日期', value: 'random_date' })
      } else if (isTimestamp) {
        // TIMESTAMP/DATETIME 类型：显示"日期时间"
        rules.push({ label: '日期时间', value: 'random_date' })
      } else {
        // 其他日期时间类型：默认显示"随机日期"
        rules.push({ label: '随机日期', value: 'random_date' })
      }
      
      rules.push({ label: '固定值', value: 'fixed' })
      // 如果可空，添加 NULL 规则
      if (isNullable) {
        rules.push({ label: '空值', value: 'null' })
      }
      return rules
    } else if (goType === 'bool' || goType === 'boolean') {
      // 布尔类型
      const rules = [
        { label: '列表选择', value: 'list' },
        { label: '固定值', value: 'fixed' }
      ]
      // 如果可空，添加 NULL 规则
      if (isNullable) {
        rules.push({ label: '空值', value: 'null' })
      }
      return rules
    } else if (goType.includes('[]byte') || goType.includes('bytea') || 
               fieldType.includes('blob') || fieldType.includes('binary') || fieldType.includes('bytea')) {
      // 二进制类型（隐藏二进制规则，只保留固定值）
      const rules = [
        { label: '固定值', value: 'fixed' }
      ]
      // 如果可空，添加 NULL 规则
      if (isNullable) {
        rules.push({ label: '空值', value: 'null' })
      }
      return rules
    } else {
      // 字符串类型（varchar, char, text 等）
      // 字符串类型可以容纳数值、日期等类型（转换为字符串），所以支持更多规则
      // 只返回常用规则
      const rules = [
        { label: '随机文本', value: 'random_string' },
        { label: '随机数字', value: 'random_number' }, // 可以转换为字符串
        { label: '随机日期', value: 'random_date' }, // 可以转换为字符串
        { label: '固定值', value: 'fixed' },
        { label: '序列', value: 'increment' }, // 可以转换为字符串
        { label: '列表选择', value: 'list' },
        { label: '正则表达式', value: 'regex' }
      ]
      // 如果可空，添加 NULL 规则
      if (isNullable) {
        rules.push({ label: '空值', value: 'null' })
      }
      return rules
    }
  }
  
  const createTask = async () => {
    if (actionLoading.value.get('createTask')) return
    actionLoading.value.set('createTask', true)
    try {
      // 使用buildTableConfig构建配置
      const config = buildTableConfig()
      if (!config) {
        ElMessage.error('配置无效，无法创建任务')
        return
      }

      const taskName = taskConfig.value.name || `${tableName.value}_${Date.now()}`
      
      if (isEditMode.value) {
        // 编辑模式：更新任务
        await api.updateTask(taskId.value, taskName, connectionId.value, config)
        ElMessage.success('任务更新成功')
      } else {
        // 创建模式：创建新任务
        await api.createTask(taskName, connectionId.value, config)
        ElMessage.success('任务创建成功')
      }
      
      // 清除缓存
      clearCache()
      router.push('/tasks')
    } catch (error) {
      const action = isEditMode.value ? '更新' : '创建'
      ElMessage.error(`任务${action}失败: ` + (error.formattedMessage || error.message))
    } finally {
      actionLoading.value.set('createTask', false)
    }
  }
  
  // 构建表配置（用于预览和创建任务）
  const buildTableConfig = () => {
    try {
      // 构建字段规则
      const rules = fieldRules.value.map(field => {
        const rule = {
          field_name: field.fieldName,
          field_type: field.fieldType,
          rule_type: field.ruleType,
          config: {},
          is_primary_key: field.isPrimaryKey,
          is_foreign_key: field.isForeignKey,
          is_unique: field.isUnique,
          is_nullable: field.isNullable,
          default_value: field.defaultValue,
          max_length: field.maxLength || 0,
          precision: field.precision || 0,
          scale: field.scale || 0
        }
  
        // 根据规则类型构建配置（复用createTask中的逻辑）
        switch (field.ruleType) {
          case 'random_string':
            // 处理字符集：如果是数组，转换为数组；如果是字符串，转换为数组（兼容旧格式）
            let charSet = field.config.charSet || ['letters', 'numbers', 'chinese', 'special']
            if (typeof charSet === 'string') {
              // 兼容旧格式：如果是 'all'，转换为全部字符类型
              if (charSet === 'all') {
                charSet = ['letters', 'numbers', 'chinese', 'special']
              } else {
                charSet = [charSet]
              }
            }
            // 确保是数组
            if (!Array.isArray(charSet)) {
              charSet = ['letters', 'numbers', 'chinese', 'special']
            }
            // 根据长度模式决定使用固定长度还是随机长度
            const lengthMode = field.config.lengthMode || (field.config.fixedLength > 0 ? 'fixed' : 'random')
            const config = {
              char_set: charSet,
              case: field.config.case || 'mixed',
              number_position: field.config.numberPosition || 'none',
              prefix: field.config.prefix || '',
              suffix: field.config.suffix || '',
              custom_chars: field.config.customChars || ''
            }
            
            if (lengthMode === 'fixed' && field.config.fixedLength > 0) {
              // 固定长度模式
              config.fixed_length = field.config.fixedLength
              config.min_length = 0
              config.max_length = 0
            } else {
              // 随机长度模式
              config.fixed_length = 0
              config.min_length = field.config.minLength || 10
              config.max_length = field.config.maxLength || 50
            }
            
            rule.config = config
            break
          case 'random_number':
            rule.config = {
              min: field.config.min || 0,
              max: field.config.max || 1000,
              is_int: field.config.isInt !== undefined ? field.config.isInt : true
            }
            break
          case 'random_date':
            // 将日期格式转换为 ISO 8601 格式（DATE 类型不包含时间部分）
            let startDate = field.config.startDate || ''
            let endDate = field.config.endDate || ''
            
            // 判断字段类型
            const isDateOnly = isDateOnlyType(field)
            const isTimestamp = isTimestampType(field)
            
            // DATE 类型：只使用日期部分，不添加时间
            // TIMESTAMP/DATETIME 类型：转换为 ISO 8601 格式
            if (!isDateOnly && startDate && startDate.match(/^\d{4}-\d{2}-\d{2}$/)) {
              if (isTimestamp && !field.config.fullDay && field.config.startTime) {
                // 日期时间类型且不是一整天，使用配置的开始时间
                startDate = startDate + 'T' + field.config.startTime + 'Z'
              } else {
                // 日期时间类型或一整天，使用默认时间
                startDate = startDate + 'T00:00:00Z'
              }
            }
            if (!isDateOnly && endDate && endDate.match(/^\d{4}-\d{2}-\d{2}$/)) {
              if (isTimestamp && !field.config.fullDay && field.config.endTime) {
                // 日期时间类型且不是一整天，使用配置的结束时间
                endDate = endDate + 'T' + field.config.endTime + 'Z'
              } else {
                // 日期时间类型或一整天，使用默认时间
                endDate = endDate + 'T23:59:59Z'
              }
            }
            
            // 根据 weekdayMode 设置 weekday_list
            let weekdayList = []
            if (field.config.weekdayMode === 'all') {
              weekdayList = []
            } else if (field.config.weekdayMode === 'weekdays') {
              weekdayList = [1, 2, 3, 4, 5]
            } else if (field.config.weekdayMode === 'custom') {
              weekdayList = field.config.weekdayCustom || []
            } else {
              // 兼容旧配置：直接使用 weekdayList
              weekdayList = field.config.weekdayList || []
            }
            
            // 根据时间配置设置 hourRange（只有日期时间类型才需要）
            let hourRange = []
            if (isTimestamp && !field.config.fullDay && field.config.startTime && field.config.endTime) {
              const startHour = parseInt(field.config.startTime.split(':')[0])
              const endHour = parseInt(field.config.endTime.split(':')[0])
              if (startHour >= 0 && startHour <= 23 && endHour >= 0 && endHour <= 23) {
                hourRange = [startHour, endHour]
              }
            }
            
            rule.config = {
              start_date: startDate,
              end_date: endDate,
              format: field.config.format || '',
              hour_range: hourRange.length > 0 ? hourRange : undefined,
              weekday_list: weekdayList,
              only_weekdays: field.config.onlyWeekdays || false,
              only_weekends: field.config.onlyWeekends || false
            }
            break
          case 'fixed':
            rule.config = { value: field.config.value || '' }
            break
          case 'increment':
            // 如果 maxValue 为 0 或未设置，使用生成数量作为默认值
            const incrementMaxValue = (field.config.maxValue && field.config.maxValue > 0) 
              ? field.config.maxValue 
              : (taskConfig.value.totalRows || 1000)
            rule.config = {
              start_value: field.config.startValue || 1,
              step: field.config.step || 1,
              cycle: field.config.cycle || false,
              max_value: incrementMaxValue,
              format: field.config.format || '',
              precision: field.config.precision || 0,
              scale: field.config.scale || 0
            }
            break
          case 'list':
            rule.config = {
              values: field.config.values || [],
              allow_repeat: field.config.allowRepeat !== undefined ? field.config.allowRepeat : true,
              weights: field.config.weights || []
            }
            break
          case 'regex':
            rule.config = { pattern: field.config.pattern || '' }
            break
          case 'function':
            const funcName = field.config.funcName || 'UUID'
            console.log('[构建配置] 函数规则:', {
              fieldName: field.fieldName,
              funcName: funcName,
              config: field.config
            })
            rule.config = {
              func_name: funcName,
              params: field.config.params || []
            }
            // 如果是 UUID，添加配置选项
            if (funcName === 'UUID') {
              rule.config.version = field.config.version || 'v4'
              rule.config.case = field.config.case || 'mixed'
              rule.config.with_hyphen = field.config.withHyphen !== undefined ? field.config.withHyphen : true
            }
            break
          case 'template':
            rule.config = { template: field.config.template || '' }
            break
          case 'null':
            rule.config = {
              probability: 1.0 // 总是生成 NULL
            }
            break
          case 'reference':
            rule.config = {
              expression: field.config.expression || '',
              fields: field.config.fields || []
            }
            break
          case 'geographic':
            rule.config = {
              type: field.config.type || 'city',
              country: field.config.country || ''
            }
            break
          case 'file':
            rule.config = {
              file_path: field.config.filePath || '',
              file_type: field.config.fileType || 'txt',
              column_index: field.config.columnIndex || 0,
              loop: field.config.loop || false
            }
            break
          case 'binary':
            rule.config = {
              mode: field.config.mode || 'generate',
              width: field.config.width || 100,
              height: field.config.height || 100,
              format: field.config.format || 'png',
              folder_path: field.config.folderPath || '',
              extensions: field.config.extensions || [],
              loop: field.config.loop || false
            }
            break
          case 'foreign':
            // 清理外键配置，移除可能的换行符和空白字符
            const foreignDatabase = (field.config.foreignDatabase || database.value || '').trim().replace(/\n/g, '').replace(/\t/g, '')
            const foreignTable = (field.config.foreignTable || field.foreignTable || '').trim().replace(/\n/g, '').replace(/\t/g, '')
            const foreignField = (field.config.foreignField || field.fieldName || '').trim().replace(/\n/g, '').replace(/\t/g, '')
            const generationMode = field.config.generationMode || (field.config.randomSelect !== false ? 'random' : 'non_repeating')
            
            rule.config = {
              foreign_database: foreignDatabase,
              foreign_table: foreignTable,
              foreign_field: foreignField,
              generation_mode: generationMode,
              repeat_min: field.config.repeatMin || 1,
              repeat_max: field.config.repeatMax || 3
            }
            
            // 向后兼容：如果配置中没有外键表，尝试使用字段信息中的外键表
            if (!rule.config.foreign_table && field.isForeignKey && field.foreignTable) {
              rule.config.foreign_table = field.foreignTable.trim().replace(/\n/g, '').replace(/\t/g, '')
            }
            break
        }
  
        // 如果是外键规则，确保 foreign_table 字段被设置到规则对象上
        if (field.ruleType === 'foreign' || field.isForeignKey) {
          const foreignTable = (rule.config.foreign_table || field.foreignTable || '').trim().replace(/\n/g, '').replace(/\t/g, '')
          rule.foreign_table = foreignTable
        }
  
        return rule
      })
  
      // 构建表配置
      return {
        table_name: tableName.value,
        database: database.value,
        total_rows: taskConfig.value.totalRows,
        batch_size: taskConfig.value.batchSize,
        thread_count: taskConfig.value.threadCount || 1,
        field_rules: rules,
        use_transaction: true,
        on_error: 'skip',
        retry_times: 3
      }
    } catch (error) {
      console.error('构建配置失败:', error)
      return null
    }
  }
  
  // 预览数据
  const handlePreviewData = async () => {
    if (actionLoading.value.get('preview')) return
    actionLoading.value.set('preview', true)
    showPreviewDialog.value = true
    try {
      await doPreview()
    } finally {
      actionLoading.value.set('preview', false)
    }
  }
  
  const doPreview = async () => {
    if (previewLoading.value) return
    previewLoading.value = true
    try {
      // 构建配置
      const config = buildTableConfig()
      if (!config) {
        ElMessage.error('配置无效，无法预览')
        return
      }
  
      const response = await api.previewData(connectionId.value, config, previewCount.value)
      previewData.value = response.data || []
      ElMessage.success(`成功生成 ${response.count || 0} 条预览数据`)
    } catch (error) {
      ElMessage.error('预览失败: ' + (error.formattedMessage || error.message))
      previewData.value = []
    } finally {
      previewLoading.value = false
    }
  }
  
  const closePreview = () => {
    showPreviewDialog.value = false
    previewData.value = []
  }
  
  // 构建单个字段的配置（用于生成示例值，只包含当前字段）
  const buildSingleFieldConfig = (field) => {
    try {
      // 对于 reference 和 template 规则，可能需要其他字段的值，使用完整配置
      const needsOtherFields = field.ruleType === 'reference' || field.ruleType === 'template'
      
      if (needsOtherFields) {
        // 需要其他字段，使用完整配置
        return buildTableConfig()
      }
      
      // 只包含当前字段的简化配置
      const rule = {
        field_name: field.fieldName,
        field_type: field.fieldType,
        rule_type: field.ruleType,
        config: {},
        is_primary_key: field.isPrimaryKey,
        is_foreign_key: field.isForeignKey,
        is_unique: field.isUnique,
        is_nullable: field.isNullable,
        default_value: field.defaultValue,
        max_length: field.maxLength || 0,
        precision: field.precision || 0,
        scale: field.scale || 0
      }
      
      // 根据规则类型构建配置（复用 buildTableConfig 中的逻辑）
      switch (field.ruleType) {
        case 'random_string':
          // 处理字符集：如果是数组，转换为数组；如果是字符串，转换为数组（兼容旧格式）
          let charSet2 = field.config.charSet || ['letters', 'numbers', 'chinese', 'special']
          if (typeof charSet2 === 'string') {
            // 兼容旧格式：如果是 'all'，转换为全部字符类型
            if (charSet2 === 'all') {
              charSet2 = ['letters', 'numbers', 'chinese', 'special']
            } else {
              charSet2 = [charSet2]
            }
          }
          // 确保是数组
          if (!Array.isArray(charSet2)) {
            charSet2 = ['letters', 'numbers', 'chinese', 'special']
          }
          rule.config = {
            min_length: field.config.minLength || 10,
            max_length: field.config.maxLength || 50,
            char_set: charSet2
          }
          break
        case 'random_number':
          rule.config = {
            min: field.config.min || 0,
            max: field.config.max || 1000,
            is_int: field.config.isInt !== false
          }
          break
        case 'random_date':
          let startDate = field.config.startDate || ''
          let endDate = field.config.endDate || ''
          const isDateOnly = isDateOnlyType(field)
          const isTimestamp = isTimestampType(field)
          
          // DATE 类型：只使用日期部分，不添加时间
          // TIMESTAMP/DATETIME 类型：转换为 ISO 8601 格式
          if (!isDateOnly) {
            if (startDate && startDate.match(/^\d{4}-\d{2}-\d{2}$/)) {
              startDate = startDate + 'T00:00:00Z'
            }
            if (endDate && endDate.match(/^\d{4}-\d{2}-\d{2}$/)) {
              endDate = endDate + 'T23:59:59Z'
            }
          }
          
          let hourRange = []
          // 只有日期时间类型才需要时间配置
          if (isTimestamp && !field.config.fullDay) {
            const startTime = field.config.startTime || '09:00:00'
            const endTime = field.config.endTime || '18:00:00'
            const startHour = parseInt(startTime.split(':')[0]) || 9
            const endHour = parseInt(endTime.split(':')[0]) || 18
            if (startHour !== endHour) {
              hourRange = [startHour, endHour]
            }
          }
          let weekdayList = []
          if (field.config.weekdayMode === 'all') {
            weekdayList = []
          } else if (field.config.weekdayMode === 'weekdays') {
            weekdayList = [1, 2, 3, 4, 5]
          } else if (field.config.weekdayMode === 'custom') {
            weekdayList = field.config.weekdayCustom || []
          } else {
            weekdayList = field.config.weekdayList || []
          }
          rule.config = {
            start_date: startDate,
            end_date: endDate,
            format: field.config.format || '',
            hour_range: hourRange.length > 0 ? hourRange : undefined,
            weekday_list: weekdayList,
            only_weekdays: field.config.onlyWeekdays || false,
            only_weekends: field.config.onlyWeekends || false
          }
          break
        case 'fixed':
          rule.config = { value: field.config.value || '' }
          break
        case 'increment':
          const incrementMaxValue = (field.config.maxValue && field.config.maxValue > 0) 
            ? field.config.maxValue 
            : (taskConfig.value.totalRows || 1000)
          rule.config = {
            start_value: field.config.startValue || 1,
            step: field.config.step || 1,
            cycle: field.config.cycle || false,
            max_value: incrementMaxValue,
            format: field.config.format || '',
            precision: field.config.precision || 0,
            scale: field.config.scale || 0
          }
          break
        case 'list':
          rule.config = {
            values: field.config.values || [],
            allow_repeat: field.config.allowRepeat !== undefined ? field.config.allowRepeat : true,
            weights: field.config.weights || []
          }
          break
        case 'regex':
          rule.config = { pattern: field.config.pattern || '' }
          break
        case 'function':
          const funcName = field.config.funcName || 'UUID'
          rule.config = {
            func_name: funcName,
            params: field.config.params || []
          }
          if (funcName === 'UUID') {
            rule.config.case = field.config.case || 'mixed'
            rule.config.with_hyphen = field.config.withHyphen !== undefined ? field.config.withHyphen : true
          }
          break
        case 'foreign':
          const foreignDatabase = (field.config.foreignDatabase || database.value || '').trim().replace(/\n/g, '').replace(/\t/g, '')
          const foreignTable = (field.config.foreignTable || field.foreignTable || '').trim().replace(/\n/g, '').replace(/\t/g, '')
          const foreignField = (field.config.foreignField || field.fieldName || '').trim().replace(/\n/g, '').replace(/\t/g, '')
          const generationMode = field.config.generationMode || (field.config.randomSelect !== false ? 'random' : 'non_repeating')
          
          rule.config = {
            foreign_database: foreignDatabase,
            foreign_table: foreignTable,
            foreign_field: foreignField,
            generation_mode: generationMode,
            repeat_min: field.config.repeatMin || 1,
            repeat_max: field.config.repeatMax || 3
          }
          
          if (!rule.config.foreign_table && field.isForeignKey && field.foreignTable) {
            rule.config.foreign_table = field.foreignTable.trim().replace(/\n/g, '').replace(/\t/g, '')
          }
          if (field.ruleType === 'foreign' || field.isForeignKey) {
            rule.foreign_table = rule.config.foreign_table || field.foreignTable || ''
          }
          break
        default:
          rule.config = {}
      }
      
      // 对于序列规则，生成5条数据以显示序列效果；其他规则只生成1条
      const isIncrement = rule.rule_type === 'increment'
      
      return {
        table_name: tableName.value,
        database: database.value,
        total_rows: isIncrement ? 5 : 1, // 序列规则生成5条，其他规则生成1条
        batch_size: isIncrement ? 5 : 1,
        field_rules: [rule],
        use_transaction: false,
        on_error: 'skip',
        retry_times: 0
      }
    } catch (error) {
      console.error('构建单字段配置失败:', error)
      return null
    }
  }
  
  // 生成单个字段的示例值
  const fieldExampleTimers = new Map() // 每个字段的防抖定时器
  const generateFieldExample = async (field) => {
    if (!field || !field.fieldName) return
    
    // 如果该字段正在生成，跳过
    if (fieldExampleLoading.value.get(field.fieldName)) {
      return
    }
    
    // 清除之前的定时器（防抖）
    const existingTimer = fieldExampleTimers.get(field.fieldName)
    if (existingTimer) {
      clearTimeout(existingTimer)
    }
    
    // 延迟执行，避免短时间内多次调用
    return new Promise((resolve) => {
      const timer = setTimeout(async () => {
        fieldExampleTimers.delete(field.fieldName)
        
        // 设置加载状态
        fieldExampleLoading.value.set(field.fieldName, true)
        
        try {
      // 构建配置（只包含当前字段，除非是 reference/template 规则）
      const config = buildSingleFieldConfig(field)
      if (!config) {
        fieldExamples.value.set(field.fieldName, '配置无效')
        return
      }
  
      // 对于序列规则，生成5条数据以显示序列效果
      const previewCount = field.ruleType === 'increment' ? 5 : 1
      
      // 调用预览接口
      const response = await api.previewData(connectionId.value, config, previewCount)
      if (response.data && response.data.length > 0) {
        // 对于序列规则，格式化为 "1, 2, 3, 4, 5, ..." 的格式
        if (field.ruleType === 'increment' && response.data.length > 0) {
          const values = response.data.map(row => {
            const value = row[field.fieldName]
            if (value !== null && value !== undefined) {
              return String(value)
            }
            return ''
          }).filter(v => v !== '')
          
          if (values.length > 0) {
            fieldExamples.value.set(field.fieldName, values.join(', ') + ', ...')
          } else {
            fieldExamples.value.set(field.fieldName, '生成失败')
          }
        } else {
          // 其他规则类型，只显示第一条数据
          const exampleValue = response.data[0][field.fieldName]
          if (exampleValue !== null && exampleValue !== undefined) {
            // 格式化显示
            let displayValue = String(exampleValue)
            
            // 如果是日期类型，格式化日期显示
            const fieldType = (field.fieldType || field.type || '').toLowerCase()
            const goType = (field.goType || field.GoType || '').toLowerCase()
            if (fieldType.includes('date') || fieldType.includes('time') || goType.includes('time') || goType.includes('date')) {
              displayValue = formatDateExample(exampleValue, field)
            } else {
              displayValue = String(exampleValue)
            }
            
            // 如果太长，截断
            if (displayValue.length > 50) {
              displayValue = displayValue.substring(0, 50) + '...'
            }
            fieldExamples.value.set(field.fieldName, displayValue)
          } else {
            fieldExamples.value.set(field.fieldName, '(空)')
          }
        }
      } else {
        fieldExamples.value.set(field.fieldName, '生成失败')
      }
        } catch (error) {
          console.error('生成字段示例失败:', error)
          fieldExamples.value.set(field.fieldName, '生成失败')
        } finally {
          fieldExampleLoading.value.set(field.fieldName, false)
          resolve()
        }
      }, 200) // 200ms 防抖延迟
      
      fieldExampleTimers.set(field.fieldName, timer)
    })
  }
  
  // 批量生成所有字段的示例值（一次请求）
  let isGeneratingAllExamples = false // 防止重复调用
  let generateAllExamplesTimer = null // 防抖定时器
  const generateAllFieldExamples = async () => {
    // 如果正在生成，直接返回
    if (isGeneratingAllExamples) {
      return
    }
    
    // 清除之前的定时器（防抖）
    if (generateAllExamplesTimer) {
      clearTimeout(generateAllExamplesTimer)
      generateAllExamplesTimer = null
    }
    
    // 延迟执行，避免短时间内多次调用
    return new Promise((resolve) => {
      generateAllExamplesTimer = setTimeout(async () => {
        isGeneratingAllExamples = true
        generateAllExamplesTimer = null
        try {
          // 构建完整配置（包含所有字段）
          const config = buildTableConfig()
          if (!config) {
            console.warn('配置无效，无法生成示例值')
            return
          }
          
          // 检查是否有需要生成示例值的字段
          const hasValidFields = fieldRules.value.some(field => field.ruleType && field.ruleType !== 'null')
          if (!hasValidFields) {
            return
          }
          
          // 分离序列规则字段和其他字段
          const incrementFields = [] // 序列规则字段（需要生成5条数据）
          const otherFields = [] // 其他字段（只需要1条数据）
          
          fieldRules.value.forEach(field => {
            if (field.ruleType && field.ruleType !== 'null') {
              if (field.ruleType === 'increment') {
                incrementFields.push(field)
              } else {
                otherFields.push(field)
              }
              // 设置所有字段的加载状态
              fieldExampleLoading.value.set(field.fieldName, true)
            }
          })
          
          // 顺序处理：先批量请求其他字段，再单独请求序列规则字段（避免后端去重器冲突）
          
          // 1. 先批量请求其他字段（生成1条数据）
          if (otherFields.length > 0) {
            try {
              // 构建只包含非序列规则字段的配置
              const otherFieldsConfig = {
                ...config,
                field_rules: config.field_rules.filter(rule => {
                  const field = fieldRules.value.find(f => f.fieldName === rule.field_name)
                  return field && field.ruleType !== 'increment'
                })
              }
              
              const response = await api.previewData(connectionId.value, otherFieldsConfig, 1)
              if (response.data && response.data.length > 0) {
                const firstRow = response.data[0]
                
                // 处理非序列规则字段的示例值
                otherFields.forEach(field => {
                  const exampleValue = firstRow[field.fieldName]
                  
                  if (exampleValue !== null && exampleValue !== undefined) {
                    // 格式化显示
                    let displayValue = String(exampleValue)
                    
                    // 如果是日期类型，格式化日期显示
                    const fieldType = (field.fieldType || field.type || '').toLowerCase()
                    const goType = (field.goType || field.GoType || '').toLowerCase()
                    if (fieldType.includes('date') || fieldType.includes('time') || goType.includes('time') || goType.includes('date')) {
                      displayValue = formatDateExample(exampleValue, field)
                    } else {
                      displayValue = String(exampleValue)
                    }
                    
                    // 如果太长，截断
                    if (displayValue.length > 50) {
                      displayValue = displayValue.substring(0, 50) + '...'
                    }
                    fieldExamples.value.set(field.fieldName, displayValue)
                  } else {
                    fieldExamples.value.set(field.fieldName, '(空)')
                  }
                  
                  fieldExampleLoading.value.set(field.fieldName, false)
                })
              }
            } catch (err) {
              console.error('批量生成非序列字段示例失败:', err)
              // 批量请求失败，逐个请求非序列字段
              for (const field of otherFields) {
                try {
                  await generateFieldExample(field)
                } catch (e) {
                  fieldExamples.value.set(field.fieldName, '生成失败')
                  fieldExampleLoading.value.set(field.fieldName, false)
                }
              }
            }
          }
          
          // 2. 再单独请求序列规则字段（生成5条数据，顺序执行避免去重器冲突）
          if (incrementFields.length > 0) {
            for (const field of incrementFields) {
              try {
                // 构建配置（只包含当前字段）
                const fieldConfig = buildSingleFieldConfig(field)
                if (!fieldConfig) {
                  fieldExamples.value.set(field.fieldName, '配置无效')
                  fieldExampleLoading.value.set(field.fieldName, false)
                  continue
                }
                
                // 序列规则生成5条数据
                const response = await api.previewData(connectionId.value, fieldConfig, 5)
                if (response.data && response.data.length > 0) {
                  const values = response.data.map(row => {
                    const value = row[field.fieldName]
                    if (value !== null && value !== undefined) {
                      return String(value)
                    }
                    return ''
                  }).filter(v => v !== '')
                  
                  if (values.length > 0) {
                    fieldExamples.value.set(field.fieldName, values.join(', ') + ', ...')
                  } else {
                    fieldExamples.value.set(field.fieldName, '生成失败')
                  }
                } else {
                  fieldExamples.value.set(field.fieldName, '生成失败')
                }
              } catch (err) {
                console.error(`生成序列字段 ${field.fieldName} 示例失败:`, err)
                fieldExamples.value.set(field.fieldName, '生成失败')
              } finally {
                fieldExampleLoading.value.set(field.fieldName, false)
              }
            }
          }
        } catch (error) {
          console.error('批量生成示例值失败:', error)
          // 如果批量请求失败，回退到逐个请求
          for (const field of fieldRules.value) {
            if (field.ruleType && field.ruleType !== 'null') {
              try {
                await generateFieldExample(field)
              } catch (err) {
                console.error(`生成字段 ${field.fieldName} 示例失败:`, err)
                fieldExamples.value.set(field.fieldName, '生成失败')
                fieldExampleLoading.value.set(field.fieldName, false)
              }
            }
          }
        } finally {
          isGeneratingAllExamples = false
          resolve()
        }
      }, 300) // 300ms 防抖延迟
    })
  }
  
  const loadTemplates = async () => {
    try {
      await templateStore.loadTemplates(tableName.value)
    } catch (error) {
      console.error('加载模板列表失败:', error)
    }
  }
  
  const saveTemplate = async () => {
    if (!templateForm.value.name) {
      ElMessage.warning('请输入模板名称')
      return
    }
    if (actionLoading.value.get('saveTemplate')) return
    
    actionLoading.value.set('saveTemplate', true)
    try {
      // 构建配置
      const rules = fieldRules.value.map(field => {
        const rule = {
          field_name: field.fieldName,
          field_type: field.fieldType,
          rule_type: field.ruleType,
          config: {},
          is_primary_key: field.isPrimaryKey,
          is_foreign_key: field.isForeignKey,
          is_unique: field.isUnique,
          is_nullable: field.isNullable,
          default_value: field.defaultValue,
          max_length: field.maxLength || 0,
          precision: field.precision || 0,
          scale: field.scale || 0
        }
  
        // 根据规则类型构建配置（复用 createTask 中的逻辑）
        switch (field.ruleType) {
          case 'random_string':
            // 处理字符集：如果是数组，转换为数组；如果是字符串，转换为数组（兼容旧格式）
            let charSet = field.config.charSet || ['letters', 'numbers', 'chinese', 'special']
            if (typeof charSet === 'string') {
              // 兼容旧格式：如果是 'all'，转换为全部字符类型
              if (charSet === 'all') {
                charSet = ['letters', 'numbers', 'chinese', 'special']
              } else {
                charSet = [charSet]
              }
            }
            // 确保是数组
            if (!Array.isArray(charSet)) {
              charSet = ['letters', 'numbers', 'chinese', 'special']
            }
            // 根据长度模式决定使用固定长度还是随机长度
            const lengthMode = field.config.lengthMode || (field.config.fixedLength > 0 ? 'fixed' : 'random')
            const config = {
              char_set: charSet,
              case: field.config.case || 'mixed',
              number_position: field.config.numberPosition || 'none',
              prefix: field.config.prefix || '',
              suffix: field.config.suffix || '',
              custom_chars: field.config.customChars || ''
            }
            
            if (lengthMode === 'fixed' && field.config.fixedLength > 0) {
              // 固定长度模式
              config.fixed_length = field.config.fixedLength
              config.min_length = 0
              config.max_length = 0
            } else {
              // 随机长度模式
              config.fixed_length = 0
              config.min_length = field.config.minLength || 10
              config.max_length = field.config.maxLength || 50
            }
            
            rule.config = config
            break
          case 'random_number':
            rule.config = {
              min: field.config.min || 0,
              max: field.config.max || 1000,
              is_int: field.config.isInt || false
            }
            break
          case 'random_date':
            // 将日期格式转换为 ISO 8601 格式（如果只是日期，添加时间部分）
            let startDate4 = field.config.startDate || ''
            let endDate4 = field.config.endDate || ''
            
            // 如果日期格式是 YYYY-MM-DD，转换为 ISO 8601
            if (startDate4 && startDate4.match(/^\d{4}-\d{2}-\d{2}$/)) {
              startDate4 = startDate4 + 'T00:00:00Z'
            }
            if (endDate4 && endDate4.match(/^\d{4}-\d{2}-\d{2}$/)) {
              endDate4 = endDate4 + 'T23:59:59Z'
            }
            
            // 根据 weekdayMode 设置 weekday_list
            let weekdayList4 = []
            if (field.config.weekdayMode === 'all') {
              weekdayList4 = []
            } else if (field.config.weekdayMode === 'weekdays') {
              weekdayList4 = [1, 2, 3, 4, 5]
            } else if (field.config.weekdayMode === 'custom') {
              weekdayList4 = field.config.weekdayCustom || []
            } else {
              // 兼容旧配置：直接使用 weekdayList
              weekdayList4 = field.config.weekdayList || []
            }
            
            rule.config = {
              start_date: startDate4,
              end_date: endDate4,
              format: field.config.format || '',
              weekday_list: weekdayList4,
              only_weekdays: field.config.onlyWeekdays || false,
              only_weekends: field.config.onlyWeekends || false
            }
            break
          case 'fixed':
            rule.config = { value: field.config.value || '' }
            break
          case 'increment':
            // 如果 maxValue 为 0 或未设置，使用生成数量作为默认值
            const templateIncrementMaxValue = (field.config.maxValue && field.config.maxValue > 0) 
              ? field.config.maxValue 
              : (taskConfig.value.totalRows || 1000)
            rule.config = {
              start_value: field.config.startValue || 1,
              step: field.config.step || 1,
              cycle: field.config.cycle || false,
              max_value: templateIncrementMaxValue,
              format: field.config.format || '',
              precision: field.config.precision || 0,
              scale: field.config.scale || 0
            }
            break
          case 'list':
            rule.config = {
              values: field.config.values || [],
              allow_repeat: field.config.allowRepeat !== undefined ? field.config.allowRepeat : true,
              weights: field.config.weights || []
            }
            break
          case 'regex':
            rule.config = { pattern: field.config.pattern || '' }
            break
          case 'function':
            const funcName = field.config.funcName || 'UUID'
            console.log('[构建配置] 函数规则:', {
              fieldName: field.fieldName,
              funcName: funcName,
              config: field.config
            })
            rule.config = {
              func_name: funcName,
              params: field.config.params || []
            }
            // 如果是 UUID，添加配置选项
            if (funcName === 'UUID') {
              rule.config.version = field.config.version || 'v4'
              rule.config.case = field.config.case || 'mixed'
              rule.config.with_hyphen = field.config.withHyphen !== undefined ? field.config.withHyphen : true
            }
            break
          case 'template':
            rule.config = { template: field.config.template || '' }
            break
          case 'null':
            rule.config = {
              probability: 1.0 // 总是生成 NULL
            }
            break
          case 'reference':
            rule.config = {
              expression: field.config.expression || '',
              fields: field.config.fields || []
            }
            break
          case 'geographic':
            rule.config = {
              type: field.config.type || 'city',
              country: field.config.country || ''
            }
            break
          case 'file':
            rule.config = {
              file_path: field.config.filePath || '',
              file_type: field.config.fileType || 'txt',
              column_index: field.config.columnIndex || 0,
              loop: field.config.loop || false
            }
            break
          case 'binary':
            rule.config = {
              mode: field.config.mode || 'generate',
              width: field.config.width || 100,
              height: field.config.height || 100,
              format: field.config.format || 'png',
              folder_path: field.config.folderPath || '',
              extensions: field.config.extensions || [],
              loop: field.config.loop || false
            }
            break
        }
  
        return rule
      })
  
      const config = {
        table_name: tableName.value,
        database: database.value,
        total_rows: taskConfig.value.totalRows,
        batch_size: taskConfig.value.batchSize,
        field_rules: rules,
        use_transaction: true,
        on_error: 'skip',
        retry_times: 3
      }
  
      await api.saveTemplate({
        name: templateForm.value.name,
        description: templateForm.value.description,
        table_name: tableName.value,
        config
      })
  
      ElMessage.success('模板保存成功')
      showSaveTemplateDialog.value = false
      templateForm.value = { name: '', description: '' }
      await loadTemplates()
    } catch (error) {
      ElMessage.error('保存模板失败: ' + (error.formattedMessage || error.message))
    } finally {
      actionLoading.value.set('saveTemplate', false)
    }
  }
  
  const loadTemplate = async (template) => {
    const key = `loadTemplate_${template.id}`
    if (actionLoading.value.get(key)) return
    actionLoading.value.set(key, true)
    try {
      // 加载模板配置
      const config = template.config
  
      // 恢复基本信息
      taskConfig.value.totalRows = config.total_rows || 1000
      taskConfig.value.batchSize = config.batch_size || 1000
      taskConfig.value.threadCount = 1  // 固定为1，单个任务不拆分多线程
  
      // 恢复字段规则
      fieldRules.value = config.field_rules.map(rule => {
        const field = tableSchema.value.fields.find(f => 
          (f.name || f.Name) === rule.field_name
        ) || {}
  
        const fieldRule = {
          fieldName: rule.field_name,
          fieldType: rule.field_type,
          goType: (field.go_type || field.GoType || '').toLowerCase(),
          ruleType: rule.rule_type,
          config: {},
          isPrimaryKey: rule.is_primary_key || false,
          isForeignKey: rule.is_foreign_key || false,
          isUnique: rule.is_unique || false,
          isNullable: rule.is_nullable !== false,
          defaultValue: rule.default_value,
          maxLength: field.max_length || field.MaxLength || rule.max_length || 0,
          precision: field.precision || field.Precision || rule.precision || 0,
          scale: field.scale || field.Scale || rule.scale || 0
        }
  
        // 恢复配置
        const ruleConfig = rule.config || {}
        switch (rule.rule_type) {
          case 'random_string':
            // 处理字符集：如果是数组，直接使用；如果是字符串，转换为数组（兼容旧格式）
            let charSet = ruleConfig.char_set || ['letters', 'numbers', 'chinese', 'special']
            if (typeof charSet === 'string') {
              // 兼容旧格式：如果是 'all'，转换为全部字符类型
              if (charSet === 'all') {
                charSet = ['letters', 'numbers', 'chinese', 'special']
              } else {
                charSet = [charSet]
              }
            }
            // 确保是数组
            if (!Array.isArray(charSet)) {
              charSet = ['letters', 'numbers', 'chinese', 'special']
            }
            // 根据配置判断长度模式（兼容旧配置）
            const lengthMode = (ruleConfig.fixed_length > 0) ? 'fixed' : 'random'
            fieldRule.config = {
              lengthMode: lengthMode,
              minLength: ruleConfig.min_length || 10,
              maxLength: ruleConfig.max_length || 50,
              charSet: charSet,
              fixedLength: ruleConfig.fixed_length || 0,
              case: ruleConfig.case || 'mixed',
              numberPosition: ruleConfig.number_position || 'none',
              prefix: ruleConfig.prefix || '',
              suffix: ruleConfig.suffix || '',
              customChars: ruleConfig.custom_chars || ''
            }
            break
          case 'random_number':
            fieldRule.config = {
              min: ruleConfig.min || 0,
              max: ruleConfig.max || 1000,
              isInt: ruleConfig.is_int || false
            }
            break
          case 'random_date':
            // 从 ISO 8601 格式转换为 YYYY-MM-DD 格式（如果包含时间部分）
            let startDate8 = ruleConfig.start_date || ''
            let endDate8 = ruleConfig.end_date || ''
            
            // 如果包含时间部分，提取日期和时间
            let startTime8 = ''
            let endTime8 = ''
            if (startDate8 && startDate8.includes('T')) {
              const parts = startDate8.split('T')
              startDate8 = parts[0]
              if (parts[1]) {
                startTime8 = parts[1].replace('Z', '').substring(0, 8) // 提取 HH:mm:ss
              }
            }
            if (endDate8 && endDate8.includes('T')) {
              const parts = endDate8.split('T')
              endDate8 = parts[0]
              if (parts[1]) {
                endTime8 = parts[1].replace('Z', '').substring(0, 8) // 提取 HH:mm:ss
              }
            }
            
            // 判断是否为日期时间类型
            const fieldType8 = (field.field_type || field.type || '').toLowerCase()
            const isTimestamp8 = fieldType8.includes('timestamp') || fieldType8.includes('datetime')
            
            // 根据 hour_range 判断是否为一整天
            let fullDay8 = true
            if (isTimestamp8 && ruleConfig.hour_range && Array.isArray(ruleConfig.hour_range) && ruleConfig.hour_range.length === 2) {
              fullDay8 = false
              // 如果没有从日期中提取到时间，使用 hour_range 推断
              if (!startTime8) {
                startTime8 = String(ruleConfig.hour_range[0]).padStart(2, '0') + ':00:00'
              }
              if (!endTime8) {
                endTime8 = String(ruleConfig.hour_range[1]).padStart(2, '0') + ':00:00'
              }
            } else if (isTimestamp8 && (startTime8 || endTime8)) {
              fullDay8 = false
            }
            
            // 根据 weekday_list 设置 weekdayMode
            let weekdayMode8 = 'all'
            let weekdayCustom8 = []
            if (ruleConfig.weekday_list && ruleConfig.weekday_list.length > 0) {
              const weekdayList8 = ruleConfig.weekday_list
              if (weekdayList8.length === 5 && weekdayList8.every(w => [1, 2, 3, 4, 5].includes(w))) {
                weekdayMode8 = 'weekdays'
              } else {
                weekdayMode8 = 'custom'
                weekdayCustom8 = weekdayList8
              }
            }
            
            fieldRule.config = {
              startDate: startDate8,
              endDate: endDate8,
              format: ruleConfig.format || '',
              fullDay: fullDay8,
              startTime: startTime8 || '09:00:00',
              endTime: endTime8 || '18:00:00',
              weekdayMode: weekdayMode8,
              weekdayCustom: weekdayCustom8,
              weekdayList: ruleConfig.weekday_list || [],
              onlyWeekdays: ruleConfig.only_weekdays || false,
              onlyWeekends: ruleConfig.only_weekends || false
            }
            break
          case 'fixed':
            fieldRule.config = { value: ruleConfig.value || '' }
            break
          case 'increment':
            // 如果 max_value 为 0 或未设置，使用生成数量作为默认值
            const defaultMaxValue = ruleConfig.max_value && ruleConfig.max_value > 0 
              ? ruleConfig.max_value 
              : (taskConfig.value.totalRows || 1000)
            fieldRule.config = {
              startValue: ruleConfig.start_value || 1,
              step: ruleConfig.step || 1,
              cycle: ruleConfig.cycle || false,
              maxValue: defaultMaxValue,
              format: ruleConfig.format || '',
              precision: ruleConfig.precision || 0,
              scale: ruleConfig.scale || 0
            }
            break
          case 'list':
            fieldRule.config = {
              values: ruleConfig.values || [],
              valuesText: (ruleConfig.values || []).join('\n'),
              allowRepeat: ruleConfig.allow_repeat !== undefined ? ruleConfig.allow_repeat : true,
              weights: ruleConfig.weights || [],
              weightsText: (ruleConfig.weights || []).join(',')
            }
            break
          case 'regex':
            fieldRule.config = { 
              pattern: ruleConfig.pattern || '',
              presetPattern: ''
            }
            // 检查是否匹配预设
            if (ruleConfig.pattern) {
              for (const [key, value] of Object.entries(regexPresets)) {
                if (value === ruleConfig.pattern) {
                  fieldRule.config.presetPattern = key
                  break
                }
              }
              if (!fieldRule.config.presetPattern) {
                fieldRule.config.presetPattern = 'custom'
              }
            }
            break
          case 'function':
            fieldRule.config = {
              funcName: ruleConfig.func_name || 'UUID',
              params: ruleConfig.params || []
            }
            // 如果是 UUID，恢复配置选项
            if (fieldRule.config.funcName === 'UUID') {
              fieldRule.config.version = ruleConfig.version || 'v4'
              fieldRule.config.case = ruleConfig.case || 'mixed'
              fieldRule.config.withHyphen = ruleConfig.with_hyphen !== undefined ? ruleConfig.with_hyphen : true
            }
            break
          case 'template':
            fieldRule.config = { template: ruleConfig.template || '' }
            break
          case 'null':
            fieldRule.config = {
              probability: 100 // 总是生成 NULL（前端显示用，实际后端使用 1.0）
            }
            break
          case 'reference':
            fieldRule.config = {
              expression: ruleConfig.expression || '',
              fields: ruleConfig.fields || [],
              fieldsText: (ruleConfig.fields || []).join(',')
            }
            break
          case 'geographic':
            fieldRule.config = {
              type: ruleConfig.type || 'city',
              country: ruleConfig.country || ''
            }
            break
          case 'file':
            fieldRule.config = {
              filePath: ruleConfig.file_path || '',
              fileType: ruleConfig.file_type || 'txt',
              columnIndex: ruleConfig.column_index || 0,
              loop: ruleConfig.loop || false
            }
            break
          case 'binary':
            fieldRule.config = {
              mode: ruleConfig.mode || 'generate',
              width: ruleConfig.width || 100,
              height: ruleConfig.height || 100,
              format: ruleConfig.format || 'png',
              folderPath: ruleConfig.folder_path || '',
              extensions: ruleConfig.extensions || [],
              extensionsText: (ruleConfig.extensions || []).join(','),
              loop: ruleConfig.loop || false
            }
            break
          default:
            fieldRule.config = getDefaultConfig(field)
        }
  
        return fieldRule
      })
  
      ElMessage.success('模板加载成功')
      showLoadTemplateDialog.value = false
    } catch (error) {
      ElMessage.error('加载模板失败: ' + (error.formattedMessage || error.message))
    } finally {
      actionLoading.value.set(key, false)
    }
  }
  
  const deleteTemplate = async (templateId) => {
    try {
      await api.deleteTemplate(templateId)
      ElMessage.success('模板已删除')
      await loadTemplates()
    } catch (error) {
      ElMessage.error('删除模板失败: ' + (error.formattedMessage || error.message))
    }
  }
  
  // 预览模板
  const previewTemplate = (template) => {
    const key = `previewTemplate_${template.id}`
    if (actionLoading.value.get(key)) return
    actionLoading.value.set(key, true)
    try {
      previewTemplateData.value = template
      const config = template.config || {}
      const fieldRules = config.field_rules || []
      
      // 转换字段规则为预览格式
      previewTemplateFields.value = fieldRules.map(rule => {
        const ruleConfig = rule.config || {}
        const fieldRule = {
          fieldName: rule.field_name || '',
          fieldType: rule.field_type || '',
          ruleType: rule.rule_type || '',
          config: {},
          isPrimaryKey: rule.is_primary_key || false,
          isForeignKey: rule.is_foreign_key || false,
          isUnique: rule.is_unique || false,
          isNullable: rule.is_nullable !== false
        }
  
        // 根据规则类型转换配置
        switch (rule.rule_type) {
          case 'random_string':
            // 处理字符集：如果是数组，直接使用；如果是字符串，转换为数组（兼容旧格式）
            let charSet3 = ruleConfig.char_set || ['letters', 'numbers', 'chinese', 'special']
            if (typeof charSet3 === 'string') {
              // 兼容旧格式：如果是 'all'，转换为全部字符类型
              if (charSet3 === 'all') {
                charSet3 = ['letters', 'numbers', 'chinese', 'special']
              } else {
                charSet3 = [charSet3]
              }
            }
            // 确保是数组
            if (!Array.isArray(charSet3)) {
              charSet3 = ['letters', 'numbers', 'chinese', 'special']
            }
            fieldRule.config = {
              minLength: ruleConfig.min_length || 10,
              maxLength: ruleConfig.max_length || 50,
              charSet: charSet3,
              fixedLength: ruleConfig.fixed_length || 0,
              case: ruleConfig.case || 'mixed',
              numberPosition: ruleConfig.number_position || 'none',
              prefix: ruleConfig.prefix || '',
              suffix: ruleConfig.suffix || '',
              customChars: ruleConfig.custom_chars || ''
            }
            break
          case 'random_number':
            fieldRule.config = {
              min: ruleConfig.min,
              max: ruleConfig.max,
              isInt: ruleConfig.is_int !== false
            }
            break
          case 'random_date':
            fieldRule.config = {
              startDate: ruleConfig.start_date || '',
              endDate: ruleConfig.end_date || '',
              format: ruleConfig.format || ''
            }
            break
          case 'fixed':
            fieldRule.config = { value: ruleConfig.value || '' }
            break
          case 'increment':
            fieldRule.config = {
              startValue: ruleConfig.start_value || 1,
              step: ruleConfig.step || 1,
              cycle: ruleConfig.cycle || false,
              maxValue: ruleConfig.max_value || 0,
              format: ruleConfig.format || '',
              precision: ruleConfig.precision || 0,
              scale: ruleConfig.scale || 0
            }
            break
          case 'list':
            fieldRule.config = {
              values: ruleConfig.values || [],
              valuesText: (ruleConfig.values || []).join('\n'),
              allowRepeat: ruleConfig.allow_repeat !== undefined ? ruleConfig.allow_repeat : true,
              weights: ruleConfig.weights || [],
              weightsText: (ruleConfig.weights || []).join(',')
            }
            break
          case 'regex':
            fieldRule.config = { 
              pattern: ruleConfig.pattern || '',
              presetPattern: ''
            }
            // 检查是否匹配预设
            if (ruleConfig.pattern) {
              for (const [key, value] of Object.entries(regexPresets)) {
                if (value === ruleConfig.pattern) {
                  fieldRule.config.presetPattern = key
                  break
                }
              }
              if (!fieldRule.config.presetPattern) {
                fieldRule.config.presetPattern = 'custom'
              }
            }
            break
          case 'function':
            fieldRule.config = {
              funcName: ruleConfig.func_name || 'UUID',
              params: ruleConfig.params || []
            }
            // 如果是 UUID，恢复配置选项
            if (fieldRule.config.funcName === 'UUID') {
              fieldRule.config.version = ruleConfig.version || 'v4'
              fieldRule.config.case = ruleConfig.case || 'mixed'
              fieldRule.config.withHyphen = ruleConfig.with_hyphen !== undefined ? ruleConfig.with_hyphen : true
            }
            break
          case 'template':
            fieldRule.config = { template: ruleConfig.template || '' }
            break
          case 'null':
            fieldRule.config = {
              probability: 100 // 总是生成 NULL（前端显示用，实际后端使用 1.0）
            }
            break
          case 'reference':
            fieldRule.config = {
              expression: ruleConfig.expression || '',
              fields: ruleConfig.fields || []
            }
            break
          case 'geographic':
            fieldRule.config = {
              type: ruleConfig.type || 'city',
              country: ruleConfig.country || ''
            }
            break
          case 'file':
            fieldRule.config = {
              filePath: ruleConfig.file_path || '',
              fileType: ruleConfig.file_type || 'txt',
              columnIndex: ruleConfig.column_index || 0,
              loop: ruleConfig.loop || false
            }
            break
          case 'binary':
            fieldRule.config = {
              mode: ruleConfig.mode || 'generate',
              width: ruleConfig.width || 100,
              height: ruleConfig.height || 100,
              format: ruleConfig.format || 'png',
              folderPath: ruleConfig.folder_path || '',
              extensions: ruleConfig.extensions || [],
              loop: ruleConfig.loop || false
            }
            break
          case 'foreign':
            // 向后兼容：如果配置中有 random_select，转换为 generation_mode
            let generationMode = ruleConfig.generation_mode || 'random'
            if (!ruleConfig.generation_mode && ruleConfig.random_select !== undefined) {
              generationMode = ruleConfig.random_select !== false ? 'random' : 'non_repeating'
            }
            
            fieldRule.config = {
              foreignDatabase: ruleConfig.foreign_database || database.value || '',
              foreignTable: ruleConfig.foreign_table || '',
              foreignField: ruleConfig.foreign_field || fieldRule.fieldName || '',
              generationMode: generationMode,
              repeatMin: ruleConfig.repeat_min || 1,
              repeatMax: ruleConfig.repeat_max || 3,
              // 向后兼容
              randomSelect: generationMode === 'random'
            }
            
            // 如果配置了外键，自动加载相关数据
            if (fieldRule.config.foreignDatabase) {
              loadForeignDatabases(fieldRule.fieldName)
              if (fieldRule.config.foreignTable) {
                loadForeignTables(fieldRule.fieldName, fieldRule.config.foreignDatabase)
                if (fieldRule.config.foreignField) {
                  loadForeignFields(fieldRule.fieldName, fieldRule.config.foreignDatabase, fieldRule.config.foreignTable)
                }
              }
            }
            break
          default:
            fieldRule.config = {}
        }
  
        return fieldRule
      })
  
      showPreviewTemplateDialog.value = true
    } catch (error) {
      console.error('预览模板失败:', error)
      ElMessage.error('预览模板失败: ' + (error.formattedMessage || error.message))
    } finally {
      actionLoading.value.set(key, false)
    }
  }
  
  // 获取规则类型标签
  const getRuleTypeLabel = (ruleType, field = null) => {
    const labels = {
      'random_string': '随机文本',
      'random_number': '随机数字',
      'random_date': '随机日期', // 默认标签，会根据字段类型动态调整
      'fixed': '固定值',
      'increment': '递增',
      'list': '列表选择',
      'regex': '正则表达式',
      'function': '函数',
      'template': '模板',
      'null': '空值',
      'reference': '引用字段',
      'geographic': '地理数据',
      'file': '从文件读取',
      'binary': '二进制/图片',
      'foreign': '外键引用'
    }
    
    // 如果是 random_date 规则，根据字段类型返回不同的标签
    if (ruleType === 'random_date' && field) {
      if (isTimestampType(field)) {
        return '日期时间'
      } else if (isDateOnlyType(field)) {
        return '随机日期'
      }
    }
    
    return labels[ruleType] || ruleType
  }
  
  // 获取字符集标签
  const getCharSetLabel = (charSet) => {
    const labels = {
      'letters': '字母',
      'numbers': '数字',
      'chinese': '中文',
      'special': '特殊字符',
      'all': '全部'
    }
    // 如果是数组，转换为标签文本
    if (Array.isArray(charSet)) {
      if (charSet.length === 0) {
        return '未选择'
      }
      if (charSet.length === 4) {
        return '全部'
      }
      return charSet.map(c => labels[c] || c).join('、')
    }
    // 兼容旧格式（字符串）
    return labels[charSet] || charSet
  }
  
  const selectedFields = ref([])
  const batchRuleType = ref('')
  const batchConfig = ref({
    // random_string
    minLength: 10,
    maxLength: 50,
    charSet: ['letters', 'numbers', 'chinese', 'special'], // 默认全选
    // random_number
    min: 0,
    max: 1000,
    isInt: true,
    // random_date
    startDate: '',
    endDate: '',
    format: '',
    // fixed
    value: '',
    // increment
    startValue: 1,
    step: 1,
    cycle: false,
    // list
    valuesText: '',
    // regex
    presetPattern: '',
    pattern: '',
    // function
    funcName: 'UUID',
    // template
    template: '',
    // null
    probability: 100, // 空值规则：总是生成 NULL
    // reference
    expression: '',
    fieldsText: '',
    // geographic
    geoType: 'city',
    country: '',
    // file
    filePath: '',
    fileType: 'txt',
    columnIndex: 0,
    loop: false,
    // binary
    binaryMode: 'generate',
    width: 100,
    height: 100,
    format: 'png',
    folderPath: '',
    extensionsText: '',
    loop: false,
    // foreign
    foreignTable: '',
    randomSelect: true
  })
  
  // 常用正则表达式映射
  const regexPresets = {
    ip: '^((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)){3})$',
    email: '^[A-Za-z0-9]+([-._][A-Za-z0-9]+)*@[A-Za-z0-9]+(-[A-Za-z0-9]+)*(\\.[A-Za-z]{2,6}|[A-Za-z]{2,4}\\.[A-Za-z]{2,3})$',
    phone_cn: '^1[3-9]\\d{9}$',
    idcard_cn: '^[1-9]\\d{5}(18|19|20)\\d{2}(0[1-9]|1[0-2])(0[1-9]|[12]\\d|3[01])\\d{3}[0-9Xx]$',
    url: '^https?:\\/\\/(www\\.)?[-a-zA-Z0-9@:%._\\+~#=]{1,256}\\.[a-zA-Z0-9()]{1,6}\\b([-a-zA-Z0-9()@:%_\\+.~#?&//=]*)$',
    date: '^\\d{4}-(0[1-9]|1[0-2])-(0[1-9]|[12]\\d|3[01])$',
    time: '^([01]\\d|2[0-3]):([0-5]\\d):([0-5]\\d)$',
    postcode_cn: '^[1-9]\\d{5}$',
    license_plate_cn: '^[京津沪渝冀豫云辽黑湘皖鲁新苏浙赣鄂桂甘晋蒙陕吉闽贵粤青藏川宁琼使领][A-Z][A-HJ-NP-Z0-9]{4,5}[A-HJ-NP-Z0-9挂学警港澳]$'
  }
  
  // 处理正则表达式预设选择
  const handleRegexPresetChange = (field) => {
    if (field.config.presetPattern && field.config.presetPattern !== 'custom') {
      field.config.pattern = regexPresets[field.config.presetPattern] || ''
    } else if (field.config.presetPattern === 'custom') {
      field.config.pattern = ''
    }
  }
  
  // 获取批量设置可用的规则类型（根据选中的字段类型，只返回常用规则）
  const getBatchAvailableRules = () => {
    if (selectedFields.value.length === 0) {
      // 如果没有选中字段，返回常用规则
      return [
        { label: '随机文本', value: 'random_string' },
        { label: '随机数字', value: 'random_number' },
        { label: '随机日期', value: 'random_date' },
        { label: '固定值', value: 'fixed' },
        { label: '序列', value: 'increment' },
        { label: '列表选择', value: 'list' },
        { label: '正则表达式', value: 'regex' },
        { label: '外键引用', value: 'foreign' }
      ]
    }
  
    // 获取所有选中字段
    const selectedFieldObjects = selectedFields.value.map(fieldName => 
      fieldRules.value.find(f => f.fieldName === fieldName)
    ).filter(f => f !== undefined)
  
    if (selectedFieldObjects.length === 0) {
      return []
    }
  
    // 检查所有字段是否都可空（所有字段都可空时才显示 NULL 规则）
    const allNullable = selectedFieldObjects.every(f => f.isNullable !== false)
  
    // 检查是否有外键字段
    const hasForeignKey = selectedFieldObjects.some(f => f.isForeignKey)
    if (hasForeignKey) {
      // 如果选中的字段中有外键，只显示外键相关规则
      const rules = [
        { label: '外键引用', value: 'foreign' },
        { label: '固定值', value: 'fixed' }
      ]
      // 如果所有字段都可空，添加 NULL 规则
      if (allNullable) {
        rules.push({ label: '空值', value: 'null' })
      }
      return rules
    }
  
    // 收集所有字段类型
    const fieldTypes = new Set()
    selectedFieldObjects.forEach(field => {
      const goType = (field.goType || '').toLowerCase()
      const fieldType = (field.fieldType || '').toLowerCase()
      
      if (goType.includes('int') || goType.includes('float') || goType.includes('decimal') || goType.includes('numeric')) {
        fieldTypes.add('number')
      } else if (goType.includes('time') || goType.includes('date')) {
        fieldTypes.add('datetime')
      } else if (goType === 'bool' || goType === 'boolean') {
        fieldTypes.add('bool')
      } else if (goType.includes('[]byte') || goType.includes('bytea') || 
                 fieldType.includes('blob') || fieldType.includes('binary') || fieldType.includes('bytea')) {
        fieldTypes.add('binary')
      } else {
        fieldTypes.add('string')
      }
    })
  
    // 如果只有一种类型，返回该类型的规则
    if (fieldTypes.size === 1) {
      const type = Array.from(fieldTypes)[0]
      let rules = []
      switch (type) {
        case 'number':
          rules = [
            { label: '随机数字', value: 'random_number' },
            { label: '固定值', value: 'fixed' },
            { label: '序列', value: 'increment' },
            { label: '列表选择', value: 'list' }
          ]
          break
        case 'datetime':
          rules = [
            { label: '随机日期', value: 'random_date' },
            { label: '固定值', value: 'fixed' }
          ]
          break
        case 'bool':
          rules = [
            { label: '列表选择', value: 'list' },
            { label: '固定值', value: 'fixed' }
          ]
          break
        case 'binary':
          rules = [
            { label: '固定值', value: 'fixed' }
          ]
          break
        case 'string':
        default:
          // 字符串类型可以容纳数值、日期等类型（转换为字符串），所以支持更多规则
          // 只返回常用规则（隐藏 random_string）
          rules = [
            { label: '随机数字', value: 'random_number' }, // 可以转换为字符串
            { label: '随机日期', value: 'random_date' }, // 可以转换为字符串
            { label: '固定值', value: 'fixed' },
            { label: '序列', value: 'increment' }, // 可以转换为字符串
            { label: '列表选择', value: 'list' },
            { label: '正则表达式', value: 'regex' }
          ]
          break
      }
      // 如果所有字段都可空，添加 NULL 规则
      if (allNullable) {
        rules.push({ label: '空值', value: 'null' })
      }
      return rules
    }
  
    // 如果混合类型，返回所有字段都支持的通用规则
    // 注意：这种情况应该被 hasMixedFieldTypes() 检测到并阻止应用
    // 只返回常用规则
    const rules = [
      { label: '固定值', value: 'fixed' }
    ]
    // 如果所有字段都可空，添加 NULL 规则
    if (allNullable) {
      rules.push({ label: '空值', value: 'null' })
    }
    return rules
  }
  
  // 处理正则表达式输入
  const handleRegexPatternInput = (field) => {
    // 如果输入的正则表达式匹配某个预设，自动选择该预设
    const pattern = field.config.pattern || ''
    if (pattern) {
      for (const [key, value] of Object.entries(regexPresets)) {
        if (value === pattern) {
          field.config.presetPattern = key
          return
        }
      }
      // 如果不匹配任何预设，设置为自定义
      if (field.config.presetPattern && field.config.presetPattern !== 'custom') {
        field.config.presetPattern = 'custom'
      }
    }
  }
  
  // 检查是否有混合字段类型
  const hasMixedFieldTypes = () => {
    if (selectedFields.value.length === 0) {
      return false
    }
  
    const selectedFieldObjects = selectedFields.value.map(fieldName => 
      fieldRules.value.find(f => f.fieldName === fieldName)
    ).filter(f => f !== undefined)
  
    if (selectedFieldObjects.length === 0) {
      return false
    }
  
    // 检查是否有外键字段
    const hasForeignKey = selectedFieldObjects.some(f => f.isForeignKey)
    if (hasForeignKey && selectedFieldObjects.length > 1) {
      // 如果选中的字段中有外键且还有其他字段，类型不一致
      return true
    }
  
    // 收集所有字段类型
    const fieldTypes = new Set()
    selectedFieldObjects.forEach(field => {
      const goType = (field.goType || '').toLowerCase()
      const fieldType = (field.fieldType || '').toLowerCase()
      
      if (goType.includes('int') || goType.includes('float') || goType.includes('decimal') || goType.includes('numeric')) {
        fieldTypes.add('number')
      } else if (goType.includes('time') || goType.includes('date')) {
        fieldTypes.add('datetime')
      } else if (goType === 'bool' || goType === 'boolean') {
        fieldTypes.add('bool')
      } else if (goType.includes('[]byte') || goType.includes('bytea') || 
                 fieldType.includes('blob') || fieldType.includes('binary') || fieldType.includes('bytea')) {
        fieldTypes.add('binary')
      } else {
        fieldTypes.add('string')
      }
    })
  
    // 如果类型超过一种，说明类型不一致
    return fieldTypes.size > 1
  }
  
  const applyBatchSettings = () => {
    if (actionLoading.value.get('batch')) return
    if (selectedFields.value.length === 0) {
      ElMessage.warning('请先选择要设置的字段')
      return
    }
    if (!batchRuleType.value) {
      ElMessage.warning('请选择规则类型')
      return
    }
    
    // 检查类型一致性
    if (hasMixedFieldTypes()) {
      ElMessage.warning('选中的字段类型不一致，请选择相同类型的字段进行批量设置')
      return
    }
    
    actionLoading.value.set('batch', true)
    try {
      selectedFields.value.forEach(fieldName => {
        const field = fieldRules.value.find(f => f.fieldName === fieldName)
        if (field) {
          field.ruleType = batchRuleType.value
          onRuleTypeChange(field)
          
          // 应用配置
          const config = batchConfig.value
          switch (batchRuleType.value) {
            case 'random_string':
              field.config.minLength = config.minLength
              field.config.maxLength = config.maxLength
              // 确保 charSet 是数组
              field.config.charSet = Array.isArray(config.charSet) ? config.charSet : ['letters', 'numbers', 'chinese', 'special']
              break
            case 'random_number':
              field.config.min = config.min
              field.config.max = config.max
              field.config.isInt = config.isInt !== false
              break
            case 'random_date':
              field.config.startDate = config.startDate || ''
              field.config.endDate = config.endDate || ''
              field.config.format = config.format || ''
              break
            case 'fixed':
              field.config.value = config.value || ''
              break
            case 'increment':
              field.config.startValue = config.startValue || 1
              field.config.step = config.step || 1
              field.config.cycle = config.cycle || false
              field.config.maxValue = config.maxValue || 0
              field.config.format = config.format || ''
              field.config.precision = config.precision || 0
              field.config.scale = config.scale || 0
              break
            case 'list':
              if (config.valuesText) {
                // 使用换行符分割，兼容逗号分隔（向后兼容）
                if (config.valuesText.includes('\n')) {
                  field.config.values = config.valuesText.split('\n').map(v => v.trim()).filter(v => v.length > 0)
                } else {
                  // 兼容旧格式：逗号分隔
                  field.config.values = config.valuesText.split(',').map(v => v.trim()).filter(v => v.length > 0)
                }
                // 统一转换为换行符分隔的格式
                field.config.valuesText = field.config.values.join('\n')
              } else if (config.values && Array.isArray(config.values)) {
                // 如果只有 values 数组，转换为 valuesText
                field.config.values = config.values
                field.config.valuesText = config.values.join('\n')
              }
              field.config.allowRepeat = config.allowRepeat !== undefined ? config.allowRepeat : true
              if (config.weightsText) {
                field.config.weights = config.weightsText.split(',').map(v => {
                  const num = parseFloat(v.trim())
                  return isNaN(num) ? 1.0 : num
                }).filter(v => v > 0)
                field.config.weightsText = config.weightsText
              } else {
                field.config.weights = []
                field.config.weightsText = ''
              }
              break
            case 'regex':
              if (config.presetPattern && config.presetPattern !== 'custom') {
                field.config.presetPattern = config.presetPattern
                field.config.pattern = regexPresets[config.presetPattern] || ''
              } else {
                field.config.presetPattern = 'custom'
                field.config.pattern = config.pattern || ''
              }
              break
            case 'function':
              field.config.funcName = config.funcName || 'UUID'
              break
            case 'template':
              field.config.template = config.template || ''
              break
            case 'null':
              field.config.probability = 100 // 空值规则：总是生成 NULL（前端显示用）
              break
            case 'reference':
              field.config.expression = config.expression || ''
              if (config.fieldsText) {
                field.config.fields = config.fieldsText.split(',').map(v => v.trim()).filter(v => v)
                field.config.fieldsText = config.fieldsText
              }
              break
            case 'geographic':
              field.config.type = config.geoType || 'city'
              field.config.country = config.country || ''
              break
            case 'file':
              field.config.filePath = config.filePath || ''
              field.config.fileType = config.fileType || 'txt'
              field.config.columnIndex = config.columnIndex || 0
              field.config.loop = config.loop || false
              break
            case 'binary':
              field.config.mode = config.binaryMode || 'generate'
              if (config.binaryMode === 'generate') {
                field.config.width = config.width || 100
                field.config.height = config.height || 100
                field.config.format = config.format || 'png'
              } else {
                field.config.folderPath = config.folderPath || ''
                if (config.extensionsText) {
                  field.config.extensions = config.extensionsText.split(',').map(v => v.trim()).filter(v => v)
                  field.config.extensionsText = config.extensionsText
                }
                field.config.loop = config.loop || false
              }
              break
            case 'foreign':
              field.config.foreignDatabase = config.foreignDatabase || database.value || ''
              field.config.foreignTable = config.foreignTable || ''
              field.config.foreignField = config.foreignField || field.fieldName || ''
              field.config.generationMode = config.generationMode || 'random'
              field.config.repeatMin = config.repeatMin || 1
              field.config.repeatMax = config.repeatMax || 3
              // 向后兼容
              field.config.randomSelect = field.config.generationMode === 'random'
              break
          }
        }
      })
  
      ElMessage.success(`已批量设置 ${selectedFields.value.length} 个字段`)
      showBatchDialog.value = false
      selectedFields.value = []
      batchRuleType.value = ''
      // 重置配置
      batchConfig.value = {
        minLength: 10,
        maxLength: 50,
        charSet: ['letters', 'numbers', 'chinese', 'special'], // 默认全选
        min: 0,
        max: 1000,
        isInt: true,
        startDate: '',
        endDate: '',
        format: '',
        value: '',
        startValue: 1,
        step: 1,
        cycle: false,
        valuesText: '',
        presetPattern: '',
        pattern: '',
        funcName: 'UUID',
        template: '',
        probability: 100, // 空值规则：总是生成 NULL
        expression: '',
        fieldsText: '',
        geoType: 'city',
        country: '',
        filePath: '',
        fileType: 'txt',
        columnIndex: 0,
        loop: false,
        binaryMode: 'generate',
        width: 100,
        height: 100,
        format: 'png',
        folderPath: '',
        extensionsText: '',
        foreignTable: '',
        randomSelect: true
      }
    } finally {
      actionLoading.value.set('batch', false)
    }
  }
  
  const resetAllRules = () => {
    fieldRules.value.forEach(field => {
      // 重置规则类型
      const defaultRuleType = getDefaultRuleType(field)
      // 检查默认规则类型是否在可用规则列表中
      const availableRules = getAvailableRules(field)
      const ruleExists = availableRules.find(r => r.value === defaultRuleType)
      
      if (ruleExists) {
        field.ruleType = defaultRuleType
      } else {
        // 如果默认规则类型不在可用规则列表中，使用第一个可用规则
        field.ruleType = availableRules.length > 0 ? availableRules[0].value : defaultRuleType
      }
      
      // 使用 onRuleTypeChange 来确保配置正确应用（包括序列规则的 maxValue、随机文本规则的 charSet 等）
      // 传入 skipExample: true 跳过单个字段的示例值生成，统一在最后批量生成
      onRuleTypeChange(field, { skipExample: true })
    })
    // 重新生成所有字段的示例值（延迟更长时间，确保外键表列表加载完成）
    // 使用批量请求，一次获取所有字段的示例值
    setTimeout(() => {
      generateAllFieldExamples()
    }, 500)
    ElMessage.success('已重置所有字段规则')
  }
  
  // 返回上一页，保持左侧树的展开状态
  const handleBack = () => {
    // 如果是编辑模式，返回到任务列表
    if (isEditMode.value) {
      router.push('/tasks')
      return
    }
    
    // 使用 push 而不是 back，这样可以保持路由状态
    // 如果是从数据库选择页面来的，返回到数据库选择页面
    if (route.query.database && route.query.connection_id) {
      router.push({
        name: 'DatabaseSelect',
        query: {
          connection_id: route.query.connection_id,
          database: route.query.database
        }
      })
    } else {
      router.push('/database-select')
    }
  }
  </script>
  
  <style scoped>
  .el-table {
    margin-top: 20px;
  }
  /* 移除序列规则高级配置底部的白条 */
  :deep(.increment-collapse .el-collapse-item__wrap) {
    border-bottom: none !important;
  }
  :deep(.increment-collapse-item .el-collapse-item__content) {
    padding-bottom: 0 !important;
  }
  </style>
  
  
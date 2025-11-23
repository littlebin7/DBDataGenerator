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
  
        <!-- 基本信息配置 -->
        <el-form :model="taskConfig" label-width="150px" style="margin-bottom: 30px">
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
                  :max="20"
                  style="width: 100%"
                />
              </el-form-item>
            </el-col>
          </el-row>
        </el-form>
  
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
  
        <el-table :data="filteredFieldRules" border style="width: 100%" max-height="600">
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
                <!-- 基础配置 -->
                <div style="display: flex; gap: 10px; flex-wrap: wrap; align-items: center">
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
                  <el-select v-model="scope.row.config.charSet" size="small" style="width: 120px">
                    <el-option label="字母" value="letters" />
                    <el-option label="数字" value="numbers" />
                    <el-option label="中文" value="chinese" />
                    <el-option label="特殊字符" value="special" />
                    <el-option label="全部" value="all" />
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
                      <!-- 固定长度 -->
                      <div style="display: flex; gap: 10px; align-items: center">
                        <span style="width: 80px; font-size: 12px;">固定长度:</span>
                        <el-input-number 
                          v-model="scope.row.config.fixedLength" 
                          :min="1"
                          :max="scope.row.maxLength || 1000"
                          placeholder="固定长度（可选）"
                          size="small"
                          style="width: 120px"
                          @change="() => validateFieldRule(scope.row)"
                        />
                      </div>
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
                  <el-input-number 
                    v-model="scope.row.config.min" 
                    :precision="2"
                    placeholder="最小值"
                    size="small"
                    style="width: 120px"
                    @change="() => validateFieldRule(scope.row)"
                  />
                  <el-input-number 
                    v-model="scope.row.config.max" 
                    :precision="2"
                    placeholder="最大值"
                    size="small"
                    style="width: 120px"
                    @change="() => validateFieldRule(scope.row)"
                  />
                  <el-checkbox v-model="scope.row.config.isInt" size="small" @change="() => validateFieldRule(scope.row)">整数</el-checkbox>
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
                      <!-- 步长 -->
                      <div style="display: flex; gap: 10px; align-items: center">
                        <span style="width: 80px; font-size: 12px;">步长:</span>
                        <el-input-number 
                          v-model="scope.row.config.step" 
                          :precision="2"
                          placeholder="步长（可选）"
                          size="small"
                          style="width: 120px"
                          @change="() => validateFieldRule(scope.row)"
                        />
                      </div>
                      <!-- 精度和小数位数 -->
                      <div style="display: flex; gap: 10px; align-items: center">
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
                          :min="1"
                          :max="scope.row.maxLength || 1000"
                          placeholder="固定长度（可选）"
                          size="small"
                          style="width: 120px"
                          @change="() => validateFieldRule(scope.row)"
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
  
              <!-- 随机日期配置 -->
              <div v-if="scope.row.ruleType === 'random_date'" style="display: flex; flex-direction: column; gap: 10px">
                <!-- 基础配置 -->
                <div style="display: flex; gap: 10px; flex-wrap: wrap; align-items: center">
                  <el-date-picker
                    v-model="scope.row.config.startDate"
                    type="datetime"
                    placeholder="开始日期"
                    size="small"
                    style="width: 180px"
                    value-format="YYYY-MM-DDTHH:mm:ssZ"
                    @change="() => validateFieldRule(scope.row)"
                  />
                  <el-date-picker
                    v-model="scope.row.config.endDate"
                    type="datetime"
                    placeholder="结束日期"
                    size="small"
                    style="width: 180px"
                    value-format="YYYY-MM-DDTHH:mm:ssZ"
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
                <!-- 粒度配置（可折叠） -->
                <el-collapse v-model="scope.row.showGranularity" style="border: none;">
                  <el-collapse-item :name="scope.row.fieldName + '_granularity'" style="border: none;">
                    <template #title>
                      <span style="font-size: 12px; color: #909399;">高级配置（粒度控制）</span>
                    </template>
                    <div style="display: flex; flex-direction: column; gap: 10px; padding: 10px; background: #f5f7fa; border-radius: 4px;">
                      <!-- 年 -->
                      <div style="display: flex; gap: 10px; align-items: center">
                        <span style="width: 80px; font-size: 12px;">年:</span>
                        <el-input-number 
                          v-model="scope.row.config.yearRange[0]" 
                          :min="1900"
                          :max="2100"
                          placeholder="最小年"
                          size="small"
                          style="width: 100px"
                          @change="() => validateFieldRule(scope.row)"
                        />
                        <span style="font-size: 12px;">~</span>
                        <el-input-number 
                          v-model="scope.row.config.yearRange[1]" 
                          :min="1900"
                          :max="2100"
                          placeholder="最大年"
                          size="small"
                          style="width: 100px"
                          @change="() => validateFieldRule(scope.row)"
                        />
                        <el-button size="small" text @click="scope.row.config.yearRange = null">清除</el-button>
                      </div>
                      <!-- 月 -->
                      <div style="display: flex; gap: 10px; align-items: center">
                        <span style="width: 80px; font-size: 12px;">月:</span>
                        <el-input-number 
                          v-model="scope.row.config.monthRange[0]" 
                          :min="1"
                          :max="12"
                          placeholder="最小月"
                          size="small"
                          style="width: 100px"
                          @change="() => validateFieldRule(scope.row)"
                        />
                        <span style="font-size: 12px;">~</span>
                        <el-input-number 
                          v-model="scope.row.config.monthRange[1]" 
                          :min="1"
                          :max="12"
                          placeholder="最大月"
                          size="small"
                          style="width: 100px"
                          @change="() => validateFieldRule(scope.row)"
                        />
                        <el-button size="small" text @click="scope.row.config.monthRange = null">清除</el-button>
                      </div>
                      <!-- 日 -->
                      <div style="display: flex; gap: 10px; align-items: center">
                        <span style="width: 80px; font-size: 12px;">日:</span>
                        <el-input-number 
                          v-model="scope.row.config.dayRange[0]" 
                          :min="1"
                          :max="31"
                          placeholder="最小日"
                          size="small"
                          style="width: 100px"
                          @change="() => validateFieldRule(scope.row)"
                        />
                        <span style="font-size: 12px;">~</span>
                        <el-input-number 
                          v-model="scope.row.config.dayRange[1]" 
                          :min="1"
                          :max="31"
                          placeholder="最大日"
                          size="small"
                          style="width: 100px"
                          @change="() => validateFieldRule(scope.row)"
                        />
                        <el-button size="small" text @click="scope.row.config.dayRange = null">清除</el-button>
                      </div>
                      <!-- 小时 -->
                      <div style="display: flex; gap: 10px; align-items: center">
                        <span style="width: 80px; font-size: 12px;">小时:</span>
                        <el-input-number 
                          v-model="scope.row.config.hourRange[0]" 
                          :min="0"
                          :max="23"
                          placeholder="最小小时"
                          size="small"
                          style="width: 100px"
                          @change="() => validateFieldRule(scope.row)"
                        />
                        <span style="font-size: 12px;">~</span>
                        <el-input-number 
                          v-model="scope.row.config.hourRange[1]" 
                          :min="0"
                          :max="23"
                          placeholder="最大小时"
                          size="small"
                          style="width: 100px"
                          @change="() => validateFieldRule(scope.row)"
                        />
                        <el-button size="small" text @click="scope.row.config.hourRange = null">清除</el-button>
                      </div>
                      <!-- 分钟 -->
                      <div style="display: flex; gap: 10px; align-items: center">
                        <span style="width: 80px; font-size: 12px;">分钟:</span>
                        <el-input-number 
                          v-model="scope.row.config.minuteRange[0]" 
                          :min="0"
                          :max="59"
                          placeholder="最小分钟"
                          size="small"
                          style="width: 100px"
                          @change="() => validateFieldRule(scope.row)"
                        />
                        <span style="font-size: 12px;">~</span>
                        <el-input-number 
                          v-model="scope.row.config.minuteRange[1]" 
                          :min="0"
                          :max="59"
                          placeholder="最大分钟"
                          size="small"
                          style="width: 100px"
                          @change="() => validateFieldRule(scope.row)"
                        />
                        <el-button size="small" text @click="scope.row.config.minuteRange = null">清除</el-button>
                      </div>
                      <!-- 秒 -->
                      <div style="display: flex; gap: 10px; align-items: center">
                        <span style="width: 80px; font-size: 12px;">秒:</span>
                        <el-input-number 
                          v-model="scope.row.config.secondRange[0]" 
                          :min="0"
                          :max="59"
                          placeholder="最小秒"
                          size="small"
                          style="width: 100px"
                          @change="() => validateFieldRule(scope.row)"
                        />
                        <span style="font-size: 12px;">~</span>
                        <el-input-number 
                          v-model="scope.row.config.secondRange[1]" 
                          :min="0"
                          :max="59"
                          placeholder="最大秒"
                          size="small"
                          style="width: 100px"
                          @change="() => validateFieldRule(scope.row)"
                        />
                        <el-button size="small" text @click="scope.row.config.secondRange = null">清除</el-button>
                      </div>
                      <!-- 星期 -->
                      <div style="display: flex; gap: 10px; align-items: center">
                        <span style="width: 80px; font-size: 12px;">星期:</span>
                        <el-select 
                          v-model="scope.row.config.weekdayList" 
                          multiple
                          placeholder="选择星期（0=周日, 1=周一...）"
                          size="small"
                          style="width: 200px"
                          @change="() => validateFieldRule(scope.row)"
                        >
                          <el-option label="周日" :value="0" />
                          <el-option label="周一" :value="1" />
                          <el-option label="周二" :value="2" />
                          <el-option label="周三" :value="3" />
                          <el-option label="周四" :value="4" />
                          <el-option label="周五" :value="5" />
                          <el-option label="周六" :value="6" />
                        </el-select>
                        <el-button size="small" text @click="scope.row.config.weekdayList = []">清除</el-button>
                      </div>
                      <!-- 工作日/周末 -->
                      <div style="display: flex; gap: 10px; align-items: center">
                        <el-checkbox v-model="scope.row.config.onlyWeekdays" size="small" @change="() => { if(scope.row.config.onlyWeekdays) scope.row.config.onlyWeekends = false; validateFieldRule(scope.row) }">仅工作日（周一到周五）</el-checkbox>
                        <el-checkbox v-model="scope.row.config.onlyWeekends" size="small" @change="() => { if(scope.row.config.onlyWeekends) scope.row.config.onlyWeekdays = false; validateFieldRule(scope.row) }">仅周末（周六和周日）</el-checkbox>
                      </div>
                    </div>
                  </el-collapse-item>
                </el-collapse>
              </div>
  
              <!-- 固定值配置 -->
              <div v-if="scope.row.ruleType === 'fixed'" style="display: flex; gap: 8px; align-items: center">
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
                  <el-input-number 
                    v-model="scope.row.config.startValue" 
                    placeholder="起始值"
                    size="small"
                    style="width: 120px"
                    @change="() => validateFieldRule(scope.row)"
                  />
                  <el-input-number 
                    v-model="scope.row.config.step" 
                    placeholder="步长"
                    size="small"
                    style="width: 120px"
                    @change="() => validateFieldRule(scope.row)"
                  />
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
                <!-- 粒度配置（可折叠） -->
                <el-collapse v-model="scope.row.showGranularity" style="border: none;">
                  <el-collapse-item :name="scope.row.fieldName + '_granularity'" style="border: none;">
                    <template #title>
                      <span style="font-size: 12px; color: #909399;">高级配置（粒度控制）</span>
                    </template>
                    <div style="display: flex; flex-direction: column; gap: 10px; padding: 10px; background: #f5f7fa; border-radius: 4px;">
                      <!-- 最大值（循环时使用） -->
                      <div style="display: flex; gap: 10px; align-items: center">
                        <span style="width: 80px; font-size: 12px;">最大值:</span>
                        <el-input-number 
                          v-model="scope.row.config.maxValue" 
                          placeholder="最大值（循环时使用，0表示不限制）"
                          size="small"
                          style="width: 200px"
                          @change="() => validateFieldRule(scope.row)"
                        />
                      </div>
                      <!-- 格式（用于字符串类型） -->
                      <div v-if="scope.row.fieldType.toLowerCase().includes('varchar') || scope.row.fieldType.toLowerCase().includes('char')" style="display: flex; gap: 10px; align-items: center">
                        <span style="width: 80px; font-size: 12px;">格式:</span>
                        <el-input 
                          v-model="scope.row.config.format" 
                          placeholder="格式（如 USER_{:05d} 表示 USER_00001）"
                          size="small"
                          style="flex: 1"
                          @blur="() => validateFieldRule(scope.row)"
                        />
                      </div>
                      <!-- 精度和小数位数（用于数字类型） -->
                      <div v-if="scope.row.goType && (scope.row.goType.includes('int') || scope.row.goType.includes('float') || scope.row.goType.includes('decimal'))" style="display: flex; gap: 10px; align-items: center">
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
                          placeholder="小数位数"
                          size="small"
                          style="width: 120px"
                          @change="() => validateFieldRule(scope.row)"
                        />
                      </div>
                    </div>
                  </el-collapse-item>
                </el-collapse>
              </div>
  
              <!-- 列表配置 -->
              <div v-if="scope.row.ruleType === 'list'" style="display: flex; flex-direction: column; gap: 10px">
                <!-- 基础配置 -->
                <div style="display: flex; gap: 8px; align-items: center">
                  <el-input 
                    v-model="scope.row.config.valuesText" 
                    placeholder="输入值列表，用逗号分隔"
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
                <el-select 
                  v-model="scope.row.config.presetPattern" 
                  placeholder="选择常用正则表达式"
                  size="small"
                  clearable
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
                <el-input 
                  v-model="scope.row.config.pattern" 
                  placeholder="输入正则表达式（选择预设会自动填充）"
                  size="small"
                  @input="handleRegexPatternInput(scope.row)"
                />
              </div>
  
              <!-- 函数配置 -->
              <div v-if="scope.row.ruleType === 'function'" style="display: flex; gap: 10px">
                <el-select 
                  v-model="scope.row.config.funcName" 
                  size="small" 
                  style="width: 150px"
                  @change="(val) => { console.log('[函数选择]', scope.row.fieldName, '选择了:', val, '当前config:', scope.row.config) }"
                >
                  <el-option label="NOW()" value="NOW" />
                  <el-option label="TODAY()" value="TODAY" />
                  <el-option label="UUID()" value="UUID" />
                  <el-option label="RAND()" value="RAND" />
                  <el-option label="RAND_INT(min, max)" value="RAND_INT" />
                  <el-option label="CONCAT(...)" value="CONCAT" />
                </el-select>
              </div>
  
              <!-- 模板配置 -->
              <el-input 
                v-if="scope.row.ruleType === 'template'"
                v-model="scope.row.config.template" 
                placeholder="模板，支持 {name}, {date}, {number}, {uuid}"
                size="small"
              />
  
              <!-- 空值配置 -->
              <el-slider 
                v-if="scope.row.ruleType === 'null'"
                v-model="scope.row.config.probability" 
                :min="0" 
                :max="100"
                :step="1"
                show-input
                size="small"
              />
  
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
                <div style="display: flex; gap: 8px; align-items: center">
                  <el-input 
                    v-model="scope.row.config.foreignTable" 
                    placeholder="外键关联表名"
                    size="small"
                    style="flex: 1"
                  />
                </div>
                <el-checkbox v-model="scope.row.config.randomSelect" size="small">随机选择</el-checkbox>
              </div>
              
              <!-- 示例值显示（所有规则类型通用） -->
              <div v-if="scope.row.ruleType && scope.row.ruleType !== 'null'" style="margin-top: 8px; padding-top: 8px; border-top: 1px solid #e4e7ed;">
                <div style="display: flex; align-items: center; gap: 8px; padding: 6px 8px; background: #f5f7fa; border-radius: 4px; font-size: 12px;">
                  <span style="color: #909399; font-weight: 500;">示例值:</span>
                  <span style="color: #606266; flex: 1; overflow: hidden; text-overflow: ellipsis; white-space: nowrap;">
                    {{ fieldExamples.get(scope.row.fieldName) || '-' }}
                  </span>
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
          <el-button type="primary" @click="createTask" :loading="actionLoading.get('createTask')" :disabled="actionLoading.get('createTask')">创建任务</el-button>
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
                <el-tag size="small">{{ getRuleTypeLabel(scope.row.ruleType) }}</el-tag>
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
                  空值概率: {{ scope.row.config.probability || 0 }}%
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
                  外键表: {{ scope.row.config.foreignTable || '-' }}, 
                  {{ scope.row.config.randomSelect ? '随机选择' : '顺序选择' }}
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
              <el-select v-model="batchConfig.charSet" placeholder="字符集" style="width: 150px">
                <el-option label="字母" value="letters" />
                <el-option label="数字" value="numbers" />
                <el-option label="中文" value="chinese" />
                <el-option label="特殊字符" value="special" />
                <el-option label="全部" value="all" />
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
            <el-input v-else-if="batchRuleType === 'list'" v-model="batchConfig.valuesText" placeholder="输入值列表，用逗号分隔" />
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
            <el-slider v-else-if="batchRuleType === 'null'" v-model="batchConfig.probability" :min="0" :max="100" :step="1" show-input />
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
            />
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
  const tableSchema = ref(null)
  const fieldRules = ref([])
  const fieldSearchText = ref('')
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
    batchSize: 500,
    threadCount: 4
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
      console.error('保存配置到缓存失败:', error)
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
      console.error('从缓存恢复配置失败:', error)
    }
    return null
  }
  
  // 清除缓存
  const clearCache = () => {
    try {
      localStorage.removeItem(getCacheKey())
    } catch (error) {
      console.error('清除缓存失败:', error)
    }
  }
  
  onMounted(async () => {
    if (!tableName.value) {
      ElMessage.error('表名不能为空')
      router.back()
      return
    }
  
    await loadTableSchema()
    await loadTemplates()
    
    // 延迟生成示例值，等待字段规则初始化完成
    setTimeout(() => {
      generateAllFieldExamples()
    }, 500)
  })
  
  // 监听配置变化，自动保存到缓存（在 setup 阶段注册）
  watch([fieldRules, taskConfig], () => {
    // 只有在字段规则已初始化后才保存（避免初始化时保存空数据）
    if (fieldRules.value && fieldRules.value.length > 0) {
      saveConfigToCache()
    }
  }, { deep: true })
  
  const loadTableSchema = async () => {
    try {
      const params = { database: database.value }
      if (connectionId.value) {
        params.connection_id = connectionId.value
      }
      const response = await api.getTableSchema(database.value, tableName.value, connectionId.value)
      tableSchema.value = response
      
      // 调试：打印响应数据
      console.log('表结构响应数据:', response)
      console.log('字段列表:', response.fields)
      
      // 初始化字段规则
      if (!response.fields || response.fields.length === 0) {
        ElMessage.warning('表结构中没有字段数据')
        fieldRules.value = []
        return
      }
      
      // 先尝试从缓存恢复配置
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
            return {
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
              scale: field.scale || field.Scale || 0
            }
          })
          
          // 恢复任务配置
          if (cachedData.taskConfig) {
            taskConfig.value = { ...taskConfig.value, ...cachedData.taskConfig }
          }
          
          ElMessage.info('已恢复上次的配置')
          console.log('从缓存恢复的字段规则:', fieldRules.value)
          return
        } else {
          // 字段不匹配，清除缓存
          clearCache()
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
        return {
          fieldName: field.name || field.Name,
          fieldType: field.type || field.Type,
          goType: (field.go_type || field.GoType || '').toLowerCase(),
          ruleType: defaultRuleType,
          config: getDefaultConfig(fieldWithRuleType),
          isPrimaryKey: field.is_primary_key || field.IsPrimaryKey || false,
          isForeignKey: field.is_foreign_key || field.IsForeignKey || false,
          isUnique: field.is_unique || field.IsUnique || false,
          isNullable: field.is_nullable !== false && field.IsNullable !== false,
          defaultValue: field.default_value || field.DefaultValue,
          maxLength: field.max_length || field.MaxLength || 0,
          precision: field.precision || field.Precision || 0,
          scale: field.scale || field.Scale || 0
        }
      })
      
      console.log('初始化后的字段规则:', fieldRules.value)
    } catch (error) {
      console.error('加载表结构失败:', error)
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
    if (field.is_primary_key || field.IsPrimaryKey) {
      return 'function' // 默认使用 UUID
    }
    if (field.is_foreign_key || field.IsForeignKey) {
      return 'foreign'
    }
    
    const type = (field.go_type || field.GoType || '').toLowerCase()
    if (type.includes('int')) {
      return 'random_number'
    } else if (type.includes('float') || type.includes('decimal')) {
      return 'random_number'
    } else if (type.includes('time') || type.includes('date')) {
      return 'random_date'
    } else if (type === 'bool') {
      return 'list'
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
      return { foreignTable: '', randomSelect: true }
    } else if (ruleType === 'regex') {
      return { pattern: '', presetPattern: '' }
    } else if (ruleType === 'function') {
      // 根据字段类型选择默认函数
      const type = (field.go_type || field.GoType || field.fieldType || '').toLowerCase()
      if (type.includes('time') || type.includes('date') || type.includes('timestamp')) {
        return { funcName: 'NOW', params: [] }
      } else {
        return { funcName: 'UUID', params: [] }
      }
    } else if (ruleType === 'template') {
      return { template: '' }
    } else if (ruleType === 'null') {
      return { probability: 0 }
    }
    
    // 根据字段类型返回默认配置
    const type = (field.go_type || field.GoType || '').toLowerCase()
    
    if (type.includes('int')) {
      return { 
        min: 1, 
        max: 1000000, 
        isInt: true,
        step: 0,
        precision: 0,
        scale: 0,
        fixedLength: 0,
        distribution: 'uniform',
        mean: 0,
        stdDev: 0,
        lambda: 0
      }
    } else if (type.includes('float') || type.includes('decimal')) {
      return { 
        min: 0, 
        max: 10000, 
        isInt: false,
        step: 0,
        precision: 0,
        scale: 0,
        fixedLength: 0,
        distribution: 'uniform',
        mean: 0,
        stdDev: 0,
        lambda: 0
      }
    } else if (type.includes('time') || type.includes('date')) {
      return { 
        startDate: '', 
        endDate: '', 
        format: '',
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
        onlyWeekdays: false,
        onlyWeekends: false
      }
    } else if (type === 'bool') {
      return { values: [true, false], valuesText: 'true,false' }
    } else {
      return { 
        minLength: 5, 
        maxLength: 20, 
        charSet: 'all',
        fixedLength: 0,
        case: 'mixed',
        numberPosition: 'none',
        prefix: '',
        suffix: '',
        customChars: ''
      }
    }
  }
  
  const onRuleTypeChange = (field) => {
    // 切换规则类型时重置配置
    field.config = getDefaultConfig(field)
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
      // 确保范围数组至少有2个元素
      if (field.config.yearRange.length < 2) field.config.yearRange = [1900, 2100]
      if (field.config.monthRange.length < 2) field.config.monthRange = [1, 12]
      if (field.config.dayRange.length < 2) field.config.dayRange = [1, 31]
      if (field.config.hourRange.length < 2) field.config.hourRange = [0, 23]
      if (field.config.minuteRange.length < 2) field.config.minuteRange = [0, 59]
      if (field.config.secondRange.length < 2) field.config.secondRange = [0, 59]
    }
    // 验证配置
    validateFieldRule(field)
    // 重新生成示例值
    generateFieldExample(field)
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
          // 检查数值范围是否在整数范围内
          const maxInt = config.max || 1000
          const minInt = config.min || 0
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
          // 检查精度和小数位数
          if (precision > 0) {
            const maxNum = config.max || 1000
            const minNum = config.min || 0
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
      field.config.values = field.config.valuesText.split(',').map(v => v.trim())
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
  
  // 根据字段类型获取可用的规则选项
  const getAvailableRules = (field) => {
    const allRules = [
      { label: '随机值', value: 'random_string' },
      { label: '随机数字', value: 'random_number' },
      { label: '随机日期', value: 'random_date' },
      { label: '固定值', value: 'fixed' },
      { label: '递增', value: 'increment' },
      { label: '列表选择', value: 'list' },
      { label: '正则表达式', value: 'regex' },
      { label: '函数', value: 'function' },
      { label: '模板', value: 'template' },
      { label: '空值', value: 'null' },
      { label: '引用字段', value: 'reference' },
      { label: '地理数据', value: 'geographic' },
      { label: '从文件读取', value: 'file' },
      { label: '二进制/图片', value: 'binary' }
    ]
  
    const goType = (field.goType || '').toLowerCase()
    const fieldType = (field.fieldType || '').toLowerCase()
  
    // 如果是外键，只显示外键相关规则
    if (field.isForeignKey) {
      return [
        { label: '外键引用', value: 'foreign' },
        { label: '固定值', value: 'fixed' },
        { label: '空值', value: 'null' }
      ]
    }
  
    // 根据 Go 类型过滤规则
    if (goType.includes('int') || goType.includes('float') || goType.includes('decimal') || goType.includes('numeric')) {
      // 数字类型
      return [
        { label: '随机数字', value: 'random_number' },
        { label: '固定值', value: 'fixed' },
        { label: '递增', value: 'increment' },
        { label: '列表选择', value: 'list' },
        { label: '函数', value: 'function' },
        { label: '引用字段', value: 'reference' },
        { label: '空值', value: 'null' }
      ]
    } else if (goType.includes('time') || goType.includes('date')) {
      // 日期时间类型
      return [
        { label: '随机日期', value: 'random_date' },
        { label: '固定值', value: 'fixed' },
        { label: '函数', value: 'function' },
        { label: '引用字段', value: 'reference' },
        { label: '空值', value: 'null' }
      ]
    } else if (goType === 'bool' || goType === 'boolean') {
      // 布尔类型
      return [
        { label: '列表选择', value: 'list' },
        { label: '固定值', value: 'fixed' },
        { label: '空值', value: 'null' }
      ]
    } else if (goType.includes('[]byte') || goType.includes('bytea') || 
               fieldType.includes('blob') || fieldType.includes('binary') || fieldType.includes('bytea')) {
      // 二进制类型
      return [
        { label: '二进制/图片', value: 'binary' },
        { label: '固定值', value: 'fixed' },
        { label: '空值', value: 'null' }
      ]
    } else {
      // 字符串类型（varchar, char, text 等）
      // 字符串类型可以容纳数值、日期等类型（转换为字符串），所以支持更多规则
      return [
        { label: '随机值', value: 'random_string' },
        { label: '随机数字', value: 'random_number' }, // 可以转换为字符串
        { label: '随机日期', value: 'random_date' }, // 可以转换为字符串
        { label: '固定值', value: 'fixed' },
        { label: '递增', value: 'increment' }, // 可以转换为字符串
        { label: '列表选择', value: 'list' },
        { label: '正则表达式', value: 'regex' },
        { label: '函数', value: 'function' },
        { label: '模板', value: 'template' },
        { label: '引用字段', value: 'reference' },
        { label: '地理数据', value: 'geographic' },
        { label: '从文件读取', value: 'file' },
        { label: '空值', value: 'null' }
      ]
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
      
      await api.createTask(taskName, connectionId.value, config)
      ElMessage.success('任务创建成功')
      // 清除缓存
      clearCache()
      router.push('/tasks')
    } catch (error) {
      ElMessage.error('创建任务失败: ' + (error.formattedMessage || error.message))
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
            rule.config = {
              min_length: field.config.minLength || 10,
              max_length: field.config.maxLength || 50,
              char_set: field.config.charSet || 'all'
            }
            break
          case 'random_number':
            rule.config = {
              min: field.config.min || 0,
              max: field.config.max || 1000,
              is_int: field.config.isInt !== undefined ? field.config.isInt : true
            }
            break
          case 'random_date':
            rule.config = {
              start_date: field.config.startDate || '',
              end_date: field.config.endDate || '',
              format: field.config.format || ''
            }
            break
          case 'fixed':
            rule.config = { value: field.config.value || '' }
            break
          case 'increment':
            rule.config = {
              start_value: field.config.startValue || 1,
              step: field.config.step || 1,
              cycle: field.config.cycle || false,
              max_value: field.config.maxValue || 0,
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
            break
          case 'template':
            rule.config = { template: field.config.template || '' }
            break
          case 'null':
            rule.config = {
              probability: (field.config.probability || 0) / 100
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
  
      // 构建表配置
      return {
        table_name: tableName.value,
        database: database.value,
        total_rows: taskConfig.value.totalRows,
        batch_size: taskConfig.value.batchSize,
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
  
      // 调试：打印函数规则配置
      const functionRules = config.field_rules.filter(r => r.rule_type === 'function')
      if (functionRules.length > 0) {
        console.log('[预览数据] 函数规则配置:', functionRules)
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
  
  // 生成单个字段的示例值
  const generateFieldExample = async (field) => {
    if (!field || !field.fieldName) return
    
    // 设置加载状态
    fieldExampleLoading.value.set(field.fieldName, true)
    
    try {
      // 构建配置（只包含当前字段）
      const config = buildTableConfig()
      if (!config) {
        fieldExamples.value.set(field.fieldName, '配置无效')
        return
      }
  
      // 调用预览接口，只生成1条数据
      const response = await api.previewData(connectionId.value, config, 1)
      if (response.data && response.data.length > 0) {
        const exampleValue = response.data[0][field.fieldName]
        if (exampleValue !== null && exampleValue !== undefined) {
          // 格式化显示
          let displayValue = String(exampleValue)
          // 如果太长，截断
          if (displayValue.length > 50) {
            displayValue = displayValue.substring(0, 50) + '...'
          }
          fieldExamples.value.set(field.fieldName, displayValue)
        } else {
          fieldExamples.value.set(field.fieldName, '(空)')
        }
      } else {
        fieldExamples.value.set(field.fieldName, '生成失败')
      }
    } catch (error) {
      console.error('生成字段示例失败:', error)
      fieldExamples.value.set(field.fieldName, '生成失败')
    } finally {
      fieldExampleLoading.value.set(field.fieldName, false)
    }
  }
  
  // 初始化时自动生成所有字段的示例值
  const generateAllFieldExamples = async () => {
    for (const field of fieldRules.value) {
      if (field.ruleType && field.ruleType !== 'null') {
        await generateFieldExample(field)
      }
    }
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
            rule.config = {
              min_length: field.config.minLength || 10,
              max_length: field.config.maxLength || 50,
              char_set: field.config.charSet || 'all'
            }
            break
          case 'random_number':
            rule.config = {
              min: field.config.min || 0,
              max: field.config.max || 1000,
              is_int: field.config.isInt || false
            }
            break
          case 'random_date':
            rule.config = {
              start_date: field.config.startDate || '',
              end_date: field.config.endDate || '',
              format: field.config.format || ''
            }
            break
          case 'fixed':
            rule.config = { value: field.config.value || '' }
            break
          case 'increment':
            rule.config = {
              start_value: field.config.startValue || 1,
              step: field.config.step || 1,
              cycle: field.config.cycle || false,
              max_value: field.config.maxValue || 0,
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
            break
          case 'template':
            rule.config = { template: field.config.template || '' }
            break
          case 'null':
            rule.config = {
              probability: (field.config.probability || 0) / 100
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
      taskConfig.value.batchSize = config.batch_size || 500
      taskConfig.value.threadCount = 4
  
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
            fieldRule.config = {
              minLength: ruleConfig.min_length || 10,
              maxLength: ruleConfig.max_length || 50,
              charSet: ruleConfig.char_set || 'all'
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
              valuesText: (ruleConfig.values || []).join(','),
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
            break
          case 'template':
            fieldRule.config = { template: ruleConfig.template || '' }
            break
          case 'null':
            fieldRule.config = {
              probability: (ruleConfig.probability || 0) * 100
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
            fieldRule.config = {
              minLength: ruleConfig.min_length,
              maxLength: ruleConfig.max_length,
              charSet: ruleConfig.char_set || 'all'
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
              valuesText: (ruleConfig.values || []).join(','),
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
            break
          case 'template':
            fieldRule.config = { template: ruleConfig.template || '' }
            break
          case 'null':
            fieldRule.config = {
              probability: (ruleConfig.probability || 0) * 100
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
            fieldRule.config = {
              foreignTable: ruleConfig.foreign_table || '',
              randomSelect: ruleConfig.random_select !== false
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
  const getRuleTypeLabel = (ruleType) => {
    const labels = {
      'random_string': '随机值',
      'random_number': '随机数字',
      'random_date': '随机日期',
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
    return labels[charSet] || charSet
  }
  
  const selectedFields = ref([])
  const batchRuleType = ref('')
  const batchConfig = ref({
    // random_string
    minLength: 10,
    maxLength: 50,
    charSet: 'all',
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
    probability: 0,
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
  
  // 获取批量设置可用的规则类型（根据选中的字段类型）
  const getBatchAvailableRules = () => {
    if (selectedFields.value.length === 0) {
      // 如果没有选中字段，返回所有规则
      return [
        { label: '随机值', value: 'random_string' },
        { label: '随机数字', value: 'random_number' },
        { label: '随机日期', value: 'random_date' },
        { label: '固定值', value: 'fixed' },
        { label: '递增', value: 'increment' },
        { label: '列表选择', value: 'list' },
        { label: '正则表达式', value: 'regex' },
        { label: '函数', value: 'function' },
        { label: '模板', value: 'template' },
        { label: '空值', value: 'null' },
        { label: '引用字段', value: 'reference' },
        { label: '地理数据', value: 'geographic' },
        { label: '从文件读取', value: 'file' },
        { label: '二进制/图片', value: 'binary' },
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
  
    // 检查是否有外键字段
    const hasForeignKey = selectedFieldObjects.some(f => f.isForeignKey)
    if (hasForeignKey) {
      // 如果选中的字段中有外键，只显示外键相关规则
      return [
        { label: '外键引用', value: 'foreign' },
        { label: '固定值', value: 'fixed' },
        { label: '空值', value: 'null' }
      ]
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
      switch (type) {
        case 'number':
          return [
            { label: '随机数字', value: 'random_number' },
            { label: '固定值', value: 'fixed' },
            { label: '递增', value: 'increment' },
            { label: '列表选择', value: 'list' },
            { label: '函数', value: 'function' },
            { label: '引用字段', value: 'reference' },
            { label: '空值', value: 'null' }
          ]
        case 'datetime':
          return [
            { label: '随机日期', value: 'random_date' },
            { label: '固定值', value: 'fixed' },
            { label: '函数', value: 'function' },
            { label: '引用字段', value: 'reference' },
            { label: '空值', value: 'null' }
          ]
        case 'bool':
          return [
            { label: '列表选择', value: 'list' },
            { label: '固定值', value: 'fixed' },
            { label: '空值', value: 'null' }
          ]
        case 'binary':
          return [
            { label: '二进制/图片', value: 'binary' },
            { label: '固定值', value: 'fixed' },
            { label: '空值', value: 'null' }
          ]
        case 'string':
        default:
          // 字符串类型可以容纳数值、日期等类型（转换为字符串），所以支持更多规则
          return [
            { label: '随机值', value: 'random_string' },
            { label: '随机数字', value: 'random_number' }, // 可以转换为字符串
            { label: '随机日期', value: 'random_date' }, // 可以转换为字符串
            { label: '固定值', value: 'fixed' },
            { label: '递增', value: 'increment' }, // 可以转换为字符串
            { label: '列表选择', value: 'list' },
            { label: '正则表达式', value: 'regex' },
            { label: '函数', value: 'function' },
            { label: '模板', value: 'template' },
            { label: '引用字段', value: 'reference' },
            { label: '地理数据', value: 'geographic' },
            { label: '从文件读取', value: 'file' },
            { label: '空值', value: 'null' }
          ]
      }
    }
  
    // 如果混合类型，返回所有字段都支持的通用规则
    // 注意：这种情况应该被 hasMixedFieldTypes() 检测到并阻止应用
    return [
      { label: '固定值', value: 'fixed' },
      { label: '函数', value: 'function' },
      { label: '引用字段', value: 'reference' },
      { label: '空值', value: 'null' }
    ]
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
              field.config.charSet = config.charSet || 'all'
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
                field.config.values = config.valuesText.split(',').map(v => v.trim()).filter(v => v)
                field.config.valuesText = config.valuesText
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
              field.config.probability = (config.probability || 0) / 100
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
              field.config.foreignTable = config.foreignTable || ''
              field.config.randomSelect = config.randomSelect !== false
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
        charSet: 'all',
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
        probability: 0,
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
      field.ruleType = getDefaultRuleType(field)
      field.config = getDefaultConfig(field)
    })
    ElMessage.success('已重置所有字段规则')
  }
  
  // 返回上一页，保持左侧树的展开状态
  const handleBack = () => {
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
  </style>
  
  
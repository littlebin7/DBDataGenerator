# 数据生成支持的类型说明

## 一、支持的数据库字段类型

系统会自动识别数据库字段类型，并映射到 Go 类型，然后根据类型推荐合适的生成规则。

### PostgreSQL 支持的类型

| 数据库类型 | Go 类型映射 | 说明 |
|-----------|------------|------|
| `int`, `int2`, `int4`, `int8`, `smallint`, `bigint`, `integer` | `int64` | 整数类型 |
| `decimal`, `numeric` | `float64` | 精确小数 |
| `float`, `float4`, `float8`, `double precision` | `float64` | 浮点数 |
| `boolean`, `bool` | `bool` | 布尔值 |
| `date`, `time`, `timestamp`, `timestamptz` | `time.Time` | 日期时间 |
| `json`, `jsonb` | `string` | JSON 类型（作为字符串处理） |
| `bytea` | `[]byte` | 二进制数据 |
| `uuid` | `string` | UUID 类型 |
| `varchar`, `char`, `text` | `string` | 字符串类型 |
| 其他 | `string` | 默认作为字符串处理 |

### MySQL/MariaDB 支持的类型

| 数据库类型 | Go 类型映射 | 说明 |
|-----------|------------|------|
| `tinyint`, `smallint`, `mediumint`, `int`, `bigint` | `int64` | 整数类型 |
| `decimal`, `numeric` | `float64` | 精确小数 |
| `float`, `double` | `float64` | 浮点数 |
| `bool`, `boolean`, `tinyint(1)` | `bool` | 布尔值 |
| `date`, `time`, `datetime`, `timestamp` | `time.Time` | 日期时间 |
| `json` | `string` | JSON 类型 |
| `blob`, `binary`, `varbinary` | `[]byte` | 二进制数据 |
| `enum` | `string` | 枚举类型（会解析枚举值） |
| `varchar`, `char`, `text` | `string` | 字符串类型 |
| 其他 | `string` | 默认作为字符串处理 |

### 达梦数据库支持的类型

| 数据库类型 | Go 类型映射 | 说明 |
|-----------|------------|------|
| `INT`, `BIGINT`, `SMALLINT` | `int64` | 整数类型 |
| `DECIMAL`, `NUMERIC` | `float64` | 精确小数 |
| `FLOAT`, `DOUBLE`, `REAL` | `float64` | 浮点数 |
| `BOOLEAN`, `BIT` | `bool` | 布尔值 |
| `DATE`, `TIME`, `TIMESTAMP` | `time.Time` | 日期时间 |
| `CHAR`, `VARCHAR`, `TEXT` | `string` | 字符串类型 |
| `BLOB`, `BINARY` | `[]byte` | 二进制数据 |
| 其他 | `string` | 默认作为字符串处理 |

---

## 二、支持的生成规则类型

系统支持 **14 种**数据生成规则，每种规则适用于不同的场景。

### 1. 随机字符串 (random_string)

**适用场景**：字符串类型字段（VARCHAR, CHAR, TEXT 等）

**配置参数**：
```json
{
  "min_length": 5,        // 最小长度
  "max_length": 20,       // 最大长度
  "char_set": "all",      // 字符集：all/letters/numbers/chinese/special
  "custom_chars": "",     // 自定义字符集（可选）
  "prefix": "",           // 前缀（可选）
  "suffix": ""            // 后缀（可选）
}
```

**字符集选项**：
- `all` - 字母+数字+特殊字符
- `letters` - 仅字母（大小写）
- `numbers` - 仅数字
- `chinese` - 中文字符
- `special` - 特殊字符
- `custom_chars` - 自定义字符集

**示例**：
- 用户名：`{"min_length": 6, "max_length": 12, "char_set": "letters"}`
- 手机号：`{"min_length": 11, "max_length": 11, "char_set": "numbers", "prefix": "1"}`
- 邮箱：`{"min_length": 8, "max_length": 20, "char_set": "all", "suffix": "@example.com"}`

### 2. 随机数字 (random_number)

**适用场景**：数字类型字段（INT, BIGINT, DECIMAL, FLOAT 等）

**配置参数**：
```json
{
  "min": 0,              // 最小值
  "max": 1000,           // 最大值
  "is_int": true         // 是否整数（false 为浮点数）
}
```

**示例**：
- 年龄：`{"min": 18, "max": 80, "is_int": true}`
- 价格：`{"min": 0.01, "max": 9999.99, "is_int": false}`
- 评分：`{"min": 1, "max": 5, "is_int": true}`

### 3. 随机日期 (random_date)

**适用场景**：日期时间类型字段（DATE, DATETIME, TIMESTAMP 等）

**配置参数**：
```json
{
  "start_date": "2020-01-01",           // 开始日期（ISO 8601 格式）
  "end_date": "2024-12-31",             // 结束日期
  "format": "2006-01-02 15:04:05"       // 输出格式（Go 时间格式）
}
```

**日期格式说明**（Go 时间格式）：
- `2006-01-02` - 日期
- `15:04:05` - 时间
- `2006-01-02 15:04:05` - 日期时间
- `2006-01-02T15:04:05Z` - ISO 8601

**示例**：
- 生日：`{"start_date": "1950-01-01", "end_date": "2005-12-31", "format": "2006-01-02"}`
- 创建时间：`{"start_date": "2020-01-01", "end_date": "2024-12-31", "format": "2006-01-02 15:04:05"}`

### 4. 固定值 (fixed)

**适用场景**：需要固定值的字段

**配置参数**：
```json
{
  "value": "固定值"  // 可以是字符串、数字、布尔值等
}
```

**示例**：
- 状态：`{"value": "active"}`
- 类型：`{"value": "user"}`
- 版本号：`{"value": "1.0"}`

### 5. 递增 (increment)

**适用场景**：需要连续递增的字段（如序号、ID 等）

**配置参数**：
```json
{
  "start_value": 1,     // 起始值
  "step": 1,            // 步长（可为负数实现递减）
  "cycle": false,       // 是否循环（达到最大值后重新开始）
  "max_value": 0        // 最大值（循环时使用，0 表示不限制）
}
```

**示例**：
- 序号：`{"start_value": 1, "step": 1}`
- 倒计时：`{"start_value": 100, "step": -1}`
- 循环编号：`{"start_value": 1, "step": 1, "cycle": true, "max_value": 10}`

### 6. 列表选择 (list)

**适用场景**：从预定义列表中选择值（如状态、类型、枚举等）

**配置参数**：
```json
{
  "values": ["选项1", "选项2", "选项3"],  // 值列表
  "weights": [1, 2, 1],                   // 权重（可选，控制选择概率）
  "allow_repeat": true                    // 是否允许重复（唯一约束时自动设为 false）
}
```

**示例**：
- 状态：`{"values": ["active", "inactive", "pending"], "allow_repeat": true}`
- 性别：`{"values": ["男", "女"], "allow_repeat": true}`
- 优先级（带权重）：`{"values": ["low", "medium", "high"], "weights": [3, 2, 1]}`

### 7. 正则表达式 (regex)

**适用场景**：需要符合特定格式的字符串（如手机号、邮箱、身份证等）

**配置参数**：
```json
{
  "pattern": "^1[3-9]\\d{9}$"  // 正则表达式模式
}
```

**常用正则示例**：
- 手机号：`"^1[3-9]\\d{9}$"`
- 邮箱：`"^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"`
- 身份证：`"^\\d{17}[\\dXx]$"`
- 车牌号：`"^[京津沪渝冀豫云辽黑湘皖鲁新苏浙赣鄂桂甘晋蒙陕吉闽贵粤青藏川宁琼使领][A-Z][A-Z0-9]{5}$"`

**注意**：正则表达式需要转义（`\\d` 表示数字）

### 8. 函数 (function)

**适用场景**：使用内置函数生成值

**配置参数**：
```json
{
  "func_name": "UUID",    // 函数名
  "params": []            // 函数参数
}
```

**支持的内置函数**：

| 函数名 | 说明 | 参数 | 返回值 | 适用场景 |
|--------|------|------|--------|----------|
| `UUID` | 生成 UUID | 无 | 字符串 | 主键、唯一标识 |
| `NOW` | 当前时间 | 无 | time.Time | 创建时间、更新时间 |
| `RAND` | 随机数（0-1） | 无 | float64 | 随机值 |
| `RAND_INT` | 随机整数 | `[min, max]` | int64 | 随机整数 |

**示例**：
- UUID 主键：`{"func_name": "UUID", "params": []}`
- 当前时间：`{"func_name": "NOW", "params": []}`
- 随机整数：`{"func_name": "RAND_INT", "params": [1, 100]}`

### 9. 模板 (template)

**适用场景**：使用模板字符串生成复合值

**配置参数**：
```json
{
  "template": "用户_{INDEX}_{RAND(1000,9999)}"  // 模板字符串
}
```

**支持的占位符**：
- `{INDEX}` - 当前行索引（从 0 开始）
- `{RAND(min,max)}` - 随机数
- `{UUID}` - UUID
- `{NOW}` - 当前时间
- `{DATE}` - 当前日期

**示例**：
- 用户名：`{"template": "user_{INDEX}"}`
- 订单号：`{"template": "ORD{INDEX:06d}_{RAND(1000,9999)}"}`
- 编号：`{"template": "SN-{DATE}-{RAND(10000,99999)}"}`

### 10. 空值 (null)

**适用场景**：需要一定概率生成 NULL 值的字段

**配置参数**：
```json
{
  "probability": 0.1  // NULL 概率（0.0-1.0，0.1 表示 10% 概率）
}
```

**注意**：
- 如果字段有 `NOT NULL` 约束，此规则会被忽略
- 概率范围：0.0（从不为空）到 1.0（总是为空）

**示例**：
- 可选字段：`{"probability": 0.2}` （20% 概率为空）
- 备注字段：`{"probability": 0.5}` （50% 概率为空）

### 11. 引用其他字段 (reference)

**适用场景**：需要基于同一行中其他字段的值生成数据

**配置参数**：
```json
{
  "expression": "{first_name}_{last_name}",  // 表达式，支持 {field_name} 占位符
  "fields": ["first_name", "last_name"]       // 引用的字段列表（可选，用于验证）
}
```

**功能说明**：
- 支持引用同一行中已生成的其他字段值
- 支持简单的数学表达式：`{price} * {quantity}`
- 支持字符串拼接：`{first_name}_{last_name}`
- 支持基本运算符：`+`, `-`, `*`, `/`

**注意**：
- 引用的字段必须在当前字段之前生成
- 如果引用的字段不存在，占位符不会被替换
- 数学表达式仅支持简单的二元运算

**示例**：
- 全名：`{"expression": "{first_name} {last_name}"}`
- 总价：`{"expression": "{price} * {quantity}"}`
- 邮箱：`{"expression": "{username}@example.com"}`
- 地址：`{"expression": "{street}, {city}, {country}"}`

### 12. 地理数据 (geographic)

**适用场景**：需要生成地理位置相关数据的字段

**配置参数**：
```json
{
  "type": "city",      // 类型：city/country/address/coordinates/latitude/longitude/postal_code
  "country": "CN"      // 国家代码（可选，如 CN, US）
}
```

**类型说明**：
- `city` - 城市名称
- `country` - 国家名称
- `address` - 完整地址（街道+城市）
- `coordinates` - 经纬度坐标（格式：lat,lon）
- `latitude` - 纬度（-90 到 90）
- `longitude` - 经度（-180 到 180）
- `postal_code` - 邮政编码

**支持的国家**：
- `CN` / `CHN` / `中国` - 中国城市和邮编
- `US` / `USA` / `美国` - 美国城市和邮编
- 未指定时随机选择

**示例**：
- 城市：`{"type": "city", "country": "CN"}`
- 国家：`{"type": "country"}`
- 地址：`{"type": "address", "country": "US"}`
- 坐标：`{"type": "coordinates"}`
- 邮编：`{"type": "postal_code", "country": "CN"}`

### 13. 从文件读取 (file)

**适用场景**：需要从外部文件读取数据的字段

**配置参数**：
```json
{
  "file_path": "/path/to/data.csv",  // 文件路径
  "file_type": "csv",                 // 文件类型：csv/txt/json
  "column_index": 0,                  // 列索引（CSV/JSON 数组使用，从 0 开始）
  "loop": true                        // 是否循环读取（文件读完后从头开始）
}
```

**文件类型说明**：
- `csv` - CSV 文件，使用 `column_index` 指定列
- `txt` - 文本文件，每行一个值
- `json` - JSON 文件，每行一个 JSON 对象或数组，使用 `column_index` 指定数组索引

**注意**：
- 文件路径可以是绝对路径或相对路径
- 文件会被缓存，避免重复读取
- 如果 `loop` 为 `false` 且索引超出文件行数，会报错
- CSV 文件使用标准 CSV 格式（支持逗号分隔）

**示例**：
- CSV 第一列：`{"file_path": "names.csv", "file_type": "csv", "column_index": 0, "loop": true}`
- 文本文件：`{"file_path": "words.txt", "file_type": "txt", "loop": true}`
- JSON 数组：`{"file_path": "data.json", "file_type": "json", "column_index": 1, "loop": false}`

### 14. 二进制/图片 (binary)

**适用场景**：需要生成或读取二进制数据的字段（BLOB, BYTEA, BINARY 等）

**配置参数**：
```json
{
  "mode": "generate",           // 模式：generate（生成图片）/folder（从文件夹读取）
  "width": 100,                 // 图片宽度（生成模式，默认 100）
  "height": 100,                // 图片高度（生成模式，默认 100）
  "format": "png",              // 图片格式（生成模式：png/jpeg/gif，默认 png）
  "folder_path": "/path/to/images",  // 文件夹路径（文件夹模式）
  "extensions": ["jpg", "png"], // 文件扩展名过滤（文件夹模式，如 ["jpg", "png"]）
  "loop": true                  // 是否循环（文件夹模式）
}
```

**模式说明**：

**生成图片模式** (`generate`)：
- 自动生成随机颜色的图片
- 支持 PNG、JPEG、GIF 格式
- 可自定义图片尺寸
- 适合测试场景，快速生成大量图片数据

**从文件夹读取模式** (`folder`)：
- 从指定文件夹读取真实图片文件
- 支持扩展名过滤（jpg, jpeg, png, gif, bmp, webp）
- 支持循环读取
- 文件夹会被缓存，避免重复扫描

**注意**：
- 生成模式：图片为随机颜色，中心有简单图案
- 文件夹模式：路径可以是绝对路径或相对路径
- 文件夹会被递归扫描（包括子目录）
- 如果 `loop` 为 `false` 且索引超出文件数量，会报错

**示例**：
- 生成 200x200 PNG 图片：`{"mode": "generate", "width": 200, "height": 200, "format": "png"}`
- 从文件夹读取 JPG/PNG：`{"mode": "folder", "folder_path": "/images", "extensions": ["jpg", "png"], "loop": true}`
- 生成 JPEG 图片：`{"mode": "generate", "width": 150, "height": 150, "format": "jpeg"}`

---

## 三、字段类型与规则推荐

系统会根据字段的 Go 类型自动推荐合适的规则，并在前端界面中**智能过滤**，只显示适用的规则选项。

### 规则过滤逻辑

| Go 类型 | 可用规则 | 说明 |
|---------|---------|------|
| `int64`, `float64` | `random_number`, `fixed`, `increment`, `list`, `function`, `reference`, `null` | 数字类型 |
| `time.Time` | `random_date`, `fixed`, `function`, `reference`, `null` | 日期时间类型 |
| `bool` | `list`, `fixed`, `null` | 布尔类型 |
| `[]byte` | `binary`, `fixed`, `null` | 二进制类型 |
| `string` | `random_string`, `fixed`, `list`, `regex`, `function`, `template`, `reference`, `geographic`, `file`, `null` | 字符串类型 |
| 外键字段 | `foreign`, `fixed`, `null` | 外键类型 |

### 智能过滤功能

**功能说明**：
- 前端界面会根据字段类型自动过滤规则选项
- 只显示与该字段类型兼容的规则
- 避免选择不兼容的规则，减少配置错误
- 提升用户体验，提高配置效率

**示例**：
- 数字字段：只显示数字相关规则（随机数字、递增、列表等）
- 日期字段：只显示日期相关规则（随机日期、函数等）
- 二进制字段：只显示二进制相关规则（二进制/图片、固定值等）
- 字符串字段：显示所有字符串相关规则（随机值、模板、地理数据等）

### 特殊字段处理

#### 主键字段
- **自增主键**：自动使用 `increment` 规则
- **UUID 主键**：自动使用 `function` (UUID) 规则
- **字符串主键**：自动使用 `function` (UUID) 规则

#### 外键字段
- 自动从关联表查询现有数据
- 随机选择一条记录的值
- 如果关联表为空，使用默认值或报错

#### 唯一约束字段
- 自动维护已生成值的集合
- 确保生成的值唯一
- 如果冲突，重新生成（最多重试 10 次）

#### 非空约束字段
- 确保生成的值不为 NULL
- 如果规则可能生成空值，会被忽略

#### 默认值字段
- 如果字段有默认值且未配置规则，使用默认值

#### ENUM 类型字段（MySQL）
- 自动解析枚举值
- 推荐使用 `list` 规则，值列表为枚举值

---

## 四、规则组合使用

可以组合多个规则实现复杂的数据生成：

### 示例 1：带前缀的递增编号
```json
{
  "rule_type": "template",
  "config": {
    "template": "USER_{INDEX:06d}"
  }
}
```
生成：`USER_000001`, `USER_000002`, `USER_000003` ...

### 示例 1.1：引用其他字段生成全名
```json
{
  "first_name": {"rule_type": "random_string", "config": {"min_length": 3, "max_length": 10}},
  "last_name": {"rule_type": "random_string", "config": {"min_length": 3, "max_length": 10}},
  "full_name": {"rule_type": "reference", "config": {"expression": "{first_name} {last_name}"}}
}
```
生成：`John Smith`, `Jane Doe` ...

### 示例 1.2：计算总价
```json
{
  "price": {"rule_type": "random_number", "config": {"min": 10, "max": 100, "is_int": false}},
  "quantity": {"rule_type": "random_number", "config": {"min": 1, "max": 10, "is_int": true}},
  "total": {"rule_type": "reference", "config": {"expression": "{price} * {quantity}"}}
}
```
生成：`price: 25.5, quantity: 3, total: 76.5` ...

### 示例 1.3：生成地理位置数据
```json
{
  "city": {"rule_type": "geographic", "config": {"type": "city", "country": "CN"}},
  "country": {"rule_type": "geographic", "config": {"type": "country"}},
  "address": {"rule_type": "geographic", "config": {"type": "address", "country": "US"}},
  "coordinates": {"rule_type": "geographic", "config": {"type": "coordinates"}}
}
```
生成：`city: 北京, country: 中国, address: 123 Main Street, New York, coordinates: 39.904200,116.407396` ...

### 示例 1.4：从文件读取数据
```json
{
  "name": {"rule_type": "file", "config": {"file_path": "names.csv", "file_type": "csv", "column_index": 0, "loop": true}},
  "description": {"rule_type": "file", "config": {"file_path": "descriptions.txt", "file_type": "txt", "loop": true}}
}
```

### 示例 2：带权重的状态选择
```json
{
  "rule_type": "list",
  "config": {
    "values": ["pending", "processing", "completed", "failed"],
    "weights": [10, 5, 3, 1]  // pending 出现概率最高
  }
}
```

### 示例 3：随机日期 + 空值概率
```json
{
  "rule_type": "random_date",
  "config": {
    "start_date": "2020-01-01",
    "end_date": "2024-12-31",
    "format": "2006-01-02 15:04:05"
  }
}
```
配合空值规则可以实现：80% 有值，20% 为空

---

## 五、约束处理优先级

1. **主键约束** - 最高优先级，自动处理
2. **外键约束** - 从关联表查询
3. **唯一约束** - 维护唯一值集合
4. **非空约束** - 确保不为 NULL
5. **默认值** - 无规则时使用默认值

---

## 六、性能优化建议

1. **批量插入**：使用合适的批次大小（默认 500）
2. **多线程**：根据数据库性能调整线程数（1-20）
3. **事务控制**：大量数据生成时建议使用事务
4. **唯一约束**：唯一字段过多会影响性能（需要维护唯一值集合）

---

## 七、常见使用场景

### 用户表
```json
{
  "id": {"rule_type": "function", "config": {"func_name": "UUID"}},
  "username": {"rule_type": "random_string", "config": {"min_length": 6, "max_length": 12, "char_set": "letters"}},
  "email": {"rule_type": "template", "config": {"template": "user_{INDEX}@example.com"}},
  "age": {"rule_type": "random_number", "config": {"min": 18, "max": 80, "is_int": true}},
  "created_at": {"rule_type": "function", "config": {"func_name": "NOW"}}
}
```

### 订单表
```json
{
  "order_no": {"rule_type": "template", "config": {"template": "ORD{INDEX:08d}"}},
  "user_id": {"rule_type": "foreign", "config": {"foreign_table": "users", "foreign_field": "id"}},
  "amount": {"rule_type": "random_number", "config": {"min": 0.01, "max": 9999.99, "is_int": false}},
  "status": {"rule_type": "list", "config": {"values": ["pending", "paid", "shipped", "completed"]}},
  "created_at": {"rule_type": "random_date", "config": {"start_date": "2024-01-01", "end_date": "2024-12-31"}}
}
```

### 产品表
```json
{
  "name": {"rule_type": "random_string", "config": {"min_length": 5, "max_length": 50, "char_set": "all", "prefix": "产品_"}},
  "price": {"rule_type": "random_number", "config": {"min": 10, "max": 1000, "is_int": false}},
  "stock": {"rule_type": "random_number", "config": {"min": 0, "max": 1000, "is_int": true}},
  "description": {"rule_type": "null", "config": {"probability": 0.3}}  // 30% 概率为空
}
```

---

## 八、限制和注意事项

1. **正则表达式**：复杂正则可能影响性能
2. **唯一约束**：大量唯一字段会降低生成速度
3. **外键约束**：关联表必须有数据
4. **模板占位符**：目前支持的占位符有限
5. **二进制数据**：BLOB 类型使用字符串生成后转为字节
6. **引用字段**：引用的字段必须在当前字段之前生成，否则无法正确引用
7. **表达式计算**：仅支持简单的二元数学运算（+, -, *, /），复杂表达式需要扩展
8. **文件读取**：文件路径必须存在且可读，大文件可能影响性能
9. **地理数据**：内置的城市和国家列表有限，可根据需要扩展

---

**文档版本**：v1.0  
**最后更新**：2024年


package generator

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// RuleGenerator 规则生成器接口
type RuleGenerator interface {
	Generate(rule *FieldRule, index int64) (interface{}, error)
}

// 各种规则生成器实现

type RandomStringGenerator struct{}

func (g *RandomStringGenerator) Generate(rule *FieldRule, index int64) (interface{}, error) {
	// 先检查原始配置中的 char_set，处理向后兼容
	var charSetTypes []string
	if ruleConfig, ok := rule.Config.(map[string]interface{}); ok {
		if charSetRaw, exists := ruleConfig["char_set"]; exists {
			switch v := charSetRaw.(type) {
			case []interface{}:
				// 新格式：数组
				charSetTypes = make([]string, 0, len(v))
				for _, item := range v {
					if str, ok := item.(string); ok {
						charSetTypes = append(charSetTypes, str)
					}
				}
			case []string:
				// 新格式：字符串数组
				charSetTypes = v
			case string:
				// 旧格式：字符串，转换为数组
				if v == "all" {
					charSetTypes = []string{"letters", "numbers", "chinese", "special"}
				} else {
					charSetTypes = []string{v}
				}
				// 更新原始配置为数组格式，以便后续解析
				ruleConfig["char_set"] = charSetTypes
			}
		}
	}

	// 如果仍然为空，使用默认值
	if len(charSetTypes) == 0 {
		charSetTypes = []string{"letters", "numbers", "chinese", "special"} // 默认全选
		// 更新原始配置
		if ruleConfig, ok := rule.Config.(map[string]interface{}); ok {
			ruleConfig["char_set"] = charSetTypes
		}
	}

	var config StringRandomConfig
	if err := unmarshalConfig(rule.Config, &config); err != nil {
		return nil, err
	}

	// 确保 config.CharSet 有值（如果解析后为空，使用我们处理过的值）
	if len(config.CharSet) == 0 {
		config.CharSet = charSetTypes
	} else {
		charSetTypes = config.CharSet
	}

	// 如果设置了固定长度，使用固定长度
	var length int
	if config.FixedLength > 0 {
		length = config.FixedLength
	} else {
		minLen := config.MinLength
		if minLen == 0 {
			minLen = 10
		}
		maxLen := config.MaxLength
		if maxLen == 0 {
			maxLen = 50
		}
		if maxLen < minLen {
			maxLen = minLen
		}
		length = minLen + rand.Intn(maxLen-minLen+1)
	}

	// 选择字符集（支持多选）
	var charSet string

	// 定义各字符类型的字符集
	charSetMap := map[string]string{
		"letters": "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ",
		"numbers": "0123456789",
		"chinese": "的一是在不了有和人这中大为上个国我以要他时来用们生到作地于出就分对成会可主发年动同工也能下过子说产种面而方后多定行学法所民得经十三之进着等部度家电力里如水化高自二理起小物现实加量都两体制机当使点从业本去把性好应开它合还因由其些然前外天政四日那社义事平形相全表间样与关各重新线内数正心反你明看原又么利比或但质气第向道命此变条只没结解问意建月公无系军很情者最立代想已通并提直题党程展五果料象员革位入常文总次品式活设及管特件长求老头基资边流路级少图山统接知较将组见计别她手角期根论运农指几九区强放决西被干做必战先回则任取据处队南给色光门即保治北造百规热领七海口东导器压志世金增争济阶油思术极交受联什认六共权收证改清己美再采转更单风切打白教速花带安场身车例真务具万每目至达走积示议声报斗完类八离华名确才科张信马节话米整空元况今集温传土许步群广石记需段研界拉林律叫且究观越织装影算低持音众书布复容儿须际商非验连断深难近矿千周委素技备半办青省列习响约支般史感劳便团往酸历市克何除消构府称太准精值号率族维划选标写存候毛亲快效斯院查江型眼王按格养易置派层片始却专状育厂京识适属圆包火住调满县局照参红细引听该铁价严龙飞",
		"special": "!@#$%^&*()_+-=[]{}|;:,.<>?",
	}

	// 使用 config.CharSet（已经处理过向后兼容）
	charSetTypes = config.CharSet

	// 合并多个字符集
	charSetBuilder := strings.Builder{}
	charSetMapUsed := make(map[string]bool) // 用于去重

	for _, charType := range charSetTypes {
		// 处理向后兼容：如果是 "all"，转换为所有类型
		if charType == "all" {
			for k, v := range charSetMap {
				if !charSetMapUsed[k] {
					charSetBuilder.WriteString(v)
					charSetMapUsed[k] = true
				}
			}
		} else if charset, exists := charSetMap[charType]; exists {
			if !charSetMapUsed[charType] {
				charSetBuilder.WriteString(charset)
				charSetMapUsed[charType] = true
			}
		}
	}

	charSet = charSetBuilder.String()

	// 如果合并后的字符集为空，使用默认值
	if charSet == "" {
		if config.CustomChars != "" {
			charSet = config.CustomChars
		} else {
			// 默认：字母+数字
			charSet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
		}
	}

	// 处理数字位置
	numberPosition := config.NumberPosition
	if numberPosition == "" {
		numberPosition = "none"
	}

	// 生成随机字符串
	// 将字符集转换为 rune 切片，以正确处理多字节字符（如中文）
	charSetRunes := []rune(charSet)
	numberChars := "0123456789"
	numberCharsRunes := []rune(numberChars)

	result := make([]rune, length)

	for i := range result {
		if numberPosition == "start" && i == 0 {
			// 开头必须是数字
			result[i] = numberCharsRunes[rand.Intn(len(numberCharsRunes))]
		} else if numberPosition == "end" && i == length-1 {
			// 结尾必须是数字
			result[i] = numberCharsRunes[rand.Intn(len(numberCharsRunes))]
		} else if numberPosition == "random" && rand.Float64() < 0.3 && len(numberCharsRunes) > 0 {
			// 30%概率是数字
			result[i] = numberCharsRunes[rand.Intn(len(numberCharsRunes))]
		} else {
			// 使用完整字符集（按 rune 索引，正确处理多字节字符）
			if len(charSetRunes) > 0 {
				result[i] = charSetRunes[rand.Intn(len(charSetRunes))]
			} else {
				// 如果字符集为空，使用默认字符
				result[i] = 'a'
			}
		}
	}

	str := string(result)

	// 应用大小写
	caseMode := config.Case
	if caseMode == "" {
		caseMode = "mixed"
	}
	switch caseMode {
	case "lower":
		str = strings.ToLower(str)
	case "upper":
		str = strings.ToUpper(str)
		// "mixed" 保持原样
	}

	// 应用前缀和后缀
	if config.Prefix != "" {
		str = config.Prefix + str
	}
	if config.Suffix != "" {
		str = str + config.Suffix
	}

	return str, nil
}

type RandomNumberGenerator struct{}

func (g *RandomNumberGenerator) Generate(rule *FieldRule, index int64) (interface{}, error) {
	var config NumberRandomConfig
	if err := unmarshalConfig(rule.Config, &config); err != nil {
		return nil, err
	}

	min := config.Min
	max := config.Max
	if max < min {
		max = min + 1000
	}

	var value float64

	// 根据分布类型生成数值
	distribution := config.Distribution
	if distribution == "" {
		distribution = "uniform" // 默认均匀分布
	}

	switch distribution {
	case "normal":
		// 正态分布
		mean := config.Mean
		if mean == 0 {
			mean = (min + max) / 2 // 默认均值为范围中点
		}
		stdDev := config.StdDev
		if stdDev == 0 {
			stdDev = (max - min) / 6 // 默认标准差为范围的1/6
		}
		// 使用Box-Muller变换生成正态分布随机数
		u1 := rand.Float64()
		u2 := rand.Float64()
		z := math.Sqrt(-2*math.Log(u1)) * math.Cos(2*math.Pi*u2)
		value = mean + z*stdDev
		// 限制在[min, max]范围内
		if value < min {
			value = min
		}
		if value > max {
			value = max
		}
	case "exponential":
		// 指数分布
		lambda := config.Lambda
		if lambda <= 0 {
			lambda = 1.0 / ((max - min) / 2) // 默认lambda
		}
		u := rand.Float64()
		value = min - (1.0/lambda)*math.Log(1-u)
		// 限制在[min, max]范围内
		if value > max {
			value = max
		}
		if value < min {
			value = min
		}
	default:
		// 均匀分布（默认）
		if config.Step > 0 {
			// 使用步长
			steps := int64((max - min) / config.Step)
			if steps > 0 {
				stepIndex := rand.Int63n(steps + 1)
				value = min + float64(stepIndex)*config.Step
			} else {
				value = min + rand.Float64()*(max-min)
			}
		} else {
			if config.IsInt {
				value = float64(int64(min) + rand.Int63n(int64(max-min+1)))
			} else {
				value = min + rand.Float64()*(max-min)
			}
		}
	}

	// 应用精度和小数位数
	if config.Precision > 0 || config.Scale > 0 {
		scale := config.Scale
		if scale == 0 && config.Precision > 0 {
			// 如果没有指定小数位数，根据精度估算
			scale = config.Precision / 3
		}
		// 格式化数值
		format := fmt.Sprintf("%%.%df", scale)
		valueStr := fmt.Sprintf(format, value)
		parsedValue, err := strconv.ParseFloat(valueStr, 64)
		if err == nil {
			value = parsedValue
		}
	}

	// 如果是整数类型，转换为整数
	if config.IsInt {
		value = float64(int64(value))
	}

	return value, nil
}

type RandomDateGenerator struct{}

func (g *RandomDateGenerator) Generate(rule *FieldRule, index int64) (interface{}, error) {
	var config DateRandomConfig
	if err := unmarshalConfig(rule.Config, &config); err != nil {
		return nil, err
	}

	startDate := time.Now().AddDate(-1, 0, 0) // 默认一年前
	endDate := time.Now()                     // 默认现在

	if config.StartDate != "" {
		if t, err := time.Parse(time.RFC3339, config.StartDate); err == nil {
			startDate = t
		}
	}
	if config.EndDate != "" {
		if t, err := time.Parse(time.RFC3339, config.EndDate); err == nil {
			endDate = t
		}
	}

	// 应用粒度配置
	var candidateTime time.Time
	maxAttempts := 100 // 最多尝试100次找到符合条件的日期
	for attempt := 0; attempt < maxAttempts; attempt++ {
		delta := endDate.Sub(startDate)
		sec := rand.Int63n(int64(delta.Seconds()))
		candidateTime = startDate.Add(time.Duration(sec) * time.Second)

		// 检查年范围
		if len(config.YearRange) == 2 {
			year := candidateTime.Year()
			if year < config.YearRange[0] || year > config.YearRange[1] {
				continue
			}
		} else if len(config.YearList) > 0 {
			year := candidateTime.Year()
			found := false
			for _, y := range config.YearList {
				if year == y {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		// 检查月范围
		if len(config.MonthRange) == 2 {
			month := int(candidateTime.Month())
			if month < config.MonthRange[0] || month > config.MonthRange[1] {
				continue
			}
		} else if len(config.MonthList) > 0 {
			month := int(candidateTime.Month())
			found := false
			for _, m := range config.MonthList {
				if month == m {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		// 检查日范围
		if len(config.DayRange) == 2 {
			day := candidateTime.Day()
			if day < config.DayRange[0] || day > config.DayRange[1] {
				continue
			}
		} else if len(config.DayList) > 0 {
			day := candidateTime.Day()
			found := false
			for _, d := range config.DayList {
				if day == d {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		// 检查小时范围
		if len(config.HourRange) == 2 {
			hour := candidateTime.Hour()
			if hour < config.HourRange[0] || hour > config.HourRange[1] {
				continue
			}
		} else if len(config.HourList) > 0 {
			hour := candidateTime.Hour()
			found := false
			for _, h := range config.HourList {
				if hour == h {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		// 检查分钟范围
		if len(config.MinuteRange) == 2 {
			minute := candidateTime.Minute()
			if minute < config.MinuteRange[0] || minute > config.MinuteRange[1] {
				continue
			}
		} else if len(config.MinuteList) > 0 {
			minute := candidateTime.Minute()
			found := false
			for _, m := range config.MinuteList {
				if minute == m {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		// 检查秒范围
		if len(config.SecondRange) == 2 {
			second := candidateTime.Second()
			if second < config.SecondRange[0] || second > config.SecondRange[1] {
				continue
			}
		} else if len(config.SecondList) > 0 {
			second := candidateTime.Second()
			found := false
			for _, s := range config.SecondList {
				if second == s {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		// 检查星期
		if len(config.WeekdayList) > 0 {
			weekday := int(candidateTime.Weekday())
			found := false
			for _, w := range config.WeekdayList {
				if weekday == w {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		// 检查仅工作日
		if config.OnlyWeekdays {
			weekday := candidateTime.Weekday()
			if weekday == time.Saturday || weekday == time.Sunday {
				continue
			}
		}

		// 检查仅周末
		if config.OnlyWeekends {
			weekday := candidateTime.Weekday()
			if weekday != time.Saturday && weekday != time.Sunday {
				continue
			}
		}

		// 所有条件都满足
		if config.Format != "" {
			return candidateTime.Format(config.Format), nil
		}
		return candidateTime, nil
	}

	// 如果尝试多次都找不到符合条件的日期，返回随机日期（不应用粒度限制）
	delta := endDate.Sub(startDate)
	sec := rand.Int63n(int64(delta.Seconds()))
	randomTime := startDate.Add(time.Duration(sec) * time.Second)

	if config.Format != "" {
		return randomTime.Format(config.Format), nil
	}
	return randomTime, nil
}

type FixedGenerator struct{}

func (g *FixedGenerator) Generate(rule *FieldRule, index int64) (interface{}, error) {
	var config FixedConfig
	if err := unmarshalConfig(rule.Config, &config); err != nil {
		return nil, err
	}
	return config.Value, nil
}

type IncrementGenerator struct {
	counters map[string]int64
	mu       sync.RWMutex // 保护 counters 的并发访问
}

func NewIncrementGenerator() *IncrementGenerator {
	return &IncrementGenerator{
		counters: make(map[string]int64),
	}
}

func (g *IncrementGenerator) Generate(rule *FieldRule, index int64) (interface{}, error) {
	var config IncrementConfig
	if err := unmarshalConfig(rule.Config, &config); err != nil {
		return nil, err
	}

	key := rule.FieldName

	// 使用互斥锁保护 map 的并发访问
	g.mu.Lock()
	defer g.mu.Unlock()

	if _, exists := g.counters[key]; !exists {
		g.counters[key] = config.StartValue
	}

	value := g.counters[key]
	g.counters[key] += config.Step

	if config.Cycle && config.MaxValue > 0 && g.counters[key] > config.MaxValue {
		g.counters[key] = config.StartValue
	}

	return value, nil
}

type ListGenerator struct{}

func (g *ListGenerator) Generate(rule *FieldRule, index int64) (interface{}, error) {
	var config ListConfig
	if err := unmarshalConfig(rule.Config, &config); err != nil {
		return nil, err
	}

	if len(config.Values) == 0 {
		return nil, fmt.Errorf("列表为空")
	}

	// 如果有权重，使用加权随机
	if len(config.Weights) == len(config.Values) {
		totalWeight := 0.0
		for _, w := range config.Weights {
			totalWeight += w
		}
		r := rand.Float64() * totalWeight
		sum := 0.0
		for i, w := range config.Weights {
			sum += w
			if r <= sum {
				return config.Values[i], nil
			}
		}
	}

	// 普通随机选择
	return config.Values[rand.Intn(len(config.Values))], nil
}

type RegexGenerator struct{}

func (g *RegexGenerator) Generate(rule *FieldRule, index int64) (interface{}, error) {
	var config RegexConfig
	if err := unmarshalConfig(rule.Config, &config); err != nil {
		return nil, err
	}

	// 检查常见模式（特殊处理，高效准确）
	if result := g.generateByCommonPattern(config.Pattern); result != "" {
		return result, nil
	}

	// 编译正则表达式
	re, err := regexp.Compile(config.Pattern)
	if err != nil {
		return nil, fmt.Errorf("无效的正则表达式: %w", err)
	}

	// 尝试智能生成：根据正则表达式结构生成
	if candidate := g.generateByPatternStructure(config.Pattern, re); candidate != "" {
		if re.MatchString(candidate) {
			return candidate, nil
		}
	}

	// 尝试根据模式特征生成
	if candidate := g.generateByPattern(config.Pattern); candidate != "" {
		if re.MatchString(candidate) {
			return candidate, nil
		}
	}

	// 使用改进的生成策略（随机生成+验证）
	maxAttempts := 3000
	for i := 0; i < maxAttempts; i++ {
		candidate := g.generateCandidateString(config.Pattern, i)
		if re.MatchString(candidate) {
			return candidate, nil
		}
	}

	return nil, fmt.Errorf("无法生成匹配正则表达式的值（尝试 %d 次后失败）", maxAttempts)
}

// generateByCommonPattern 根据常见模式生成（特殊处理）
func (g *RegexGenerator) generateByCommonPattern(pattern string) string {
	// IP地址
	if g.isIPAddressPattern(pattern) {
		return g.generateIPAddress()
	}

	// 邮箱地址
	if g.isEmailPattern(pattern) {
		return g.generateEmailAddress()
	}

	// 手机号（中国）
	if g.isPhoneCNPattern(pattern) {
		return g.generatePhoneCN()
	}

	// 身份证号（中国）
	if g.isIDCardCNPattern(pattern) {
		return g.generateIDCardCN()
	}

	// URL地址
	if g.isURLPattern(pattern) {
		return g.generateURL()
	}

	// 日期（YYYY-MM-DD）
	if g.isDatePattern(pattern) {
		return g.generateDate()
	}

	// 时间（HH:MM:SS）
	if g.isTimePattern(pattern) {
		return g.generateTime()
	}

	// 邮政编码（中国）
	if g.isPostcodeCNPattern(pattern) {
		return g.generatePostcodeCN()
	}

	return ""
}

// isIPAddressPattern 检查是否是IP地址的正则表达式
func (g *RegexGenerator) isIPAddressPattern(pattern string) bool {
	// 检查是否包含IP地址的典型模式
	// 匹配类似 ^((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)){3})$ 的模式
	ipIndicators := []string{
		"25[0-5]",
		"2[0-4][0-9]",
		"[01]?[0-9][0-9]?",
		"\\d{1,3}",
	}

	// 检查是否包含IP地址的典型数字范围模式
	hasIPRange := false
	for _, indicator := range ipIndicators {
		if strings.Contains(pattern, indicator) {
			hasIPRange = true
			break
		}
	}

	// 检查是否包含点号分隔符（IP地址格式）
	hasDots := strings.Contains(pattern, "\\.") || strings.Contains(pattern, ".")

	// 检查是否包含4个部分的模式（IP地址有4个八位组）
	hasFourParts := strings.Contains(pattern, "{3}") || strings.Count(pattern, "\\.") >= 3 || strings.Count(pattern, ".") >= 3

	return hasIPRange && hasDots && (hasFourParts || strings.Contains(pattern, "25[0-5]"))
}

// generateIPAddress 生成随机IP地址
func (g *RegexGenerator) generateIPAddress() string {
	octet1 := rand.Intn(256)
	octet2 := rand.Intn(256)
	octet3 := rand.Intn(256)
	octet4 := rand.Intn(256)
	return fmt.Sprintf("%d.%d.%d.%d", octet1, octet2, octet3, octet4)
}

// isEmailPattern 检查是否是邮箱地址的正则表达式
func (g *RegexGenerator) isEmailPattern(pattern string) bool {
	// 检查是否包含邮箱地址的典型特征
	// 1. 包含 @ 符号
	// 2. 包含域名部分（通常有 . 和字母）
	// 3. 包含用户名部分（通常有字母、数字、点、下划线、连字符）

	hasAtSymbol := strings.Contains(pattern, "@")
	hasDomainPattern := strings.Contains(pattern, "\\.[A-Za-z]") ||
		strings.Contains(pattern, "\\.[a-z]") ||
		strings.Contains(pattern, "\\.[A-Z]") ||
		(strings.Contains(pattern, ".") &&
			(strings.Contains(pattern, "[A-Za-z]") || strings.Contains(pattern, "[a-z]")))

	hasUsernamePattern := strings.Contains(pattern, "[A-Za-z0-9]") ||
		strings.Contains(pattern, "[a-z0-9]") ||
		strings.Contains(pattern, "[-._]")

	return hasAtSymbol && hasDomainPattern && hasUsernamePattern
}

// generateEmailAddress 生成随机邮箱地址
func (g *RegexGenerator) generateEmailAddress() string {
	// 生成用户名部分（5-15个字符）
	usernameLength := 5 + rand.Intn(11)
	username := g.generateEmailUsername(usernameLength)

	// 生成域名部分
	domains := []string{
		"com", "net", "org", "edu", "gov", "cn", "io", "co", "uk", "de",
		"fr", "jp", "au", "ca", "info", "biz", "tech", "online", "site", "xyz",
	}

	// 主域名
	mainDomain := domains[rand.Intn(len(domains))]

	// 有时添加二级域名
	if rand.Float32() < 0.3 {
		subdomains := []string{"mail", "email", "web", "www", "app", "api", "blog", "news"}
		subdomain := subdomains[rand.Intn(len(subdomains))]
		return fmt.Sprintf("%s@%s.%s", username, subdomain, mainDomain)
	}

	// 有时使用多级域名（如 .com.cn）
	if rand.Float32() < 0.2 && mainDomain == "cn" {
		return fmt.Sprintf("%s@example.%s", username, mainDomain)
	}

	return fmt.Sprintf("%s@example.%s", username, mainDomain)
}

// generateEmailUsername 生成邮箱用户名部分
func (g *RegexGenerator) generateEmailUsername(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	separators := []string{"", ".", "_", "-"}

	username := make([]byte, 0, length)

	for i := 0; i < length; i++ {
		// 随机决定是否插入分隔符
		if i > 0 && i < length-1 && rand.Float32() < 0.15 {
			sep := separators[rand.Intn(len(separators))]
			if sep != "" {
				username = append(username, []byte(sep)...)
				i += len(sep) - 1
				continue
			}
		}

		// 添加随机字符
		char := charset[rand.Intn(len(charset))]
		username = append(username, char)
	}

	return string(username)
}

// generateByPattern 根据正则表达式模式尝试生成候选字符串
func (g *RegexGenerator) generateByPattern(pattern string) string {
	// 处理数字范围模式 [0-9], \d
	if matched, _ := regexp.MatchString(`\d+`, pattern); matched {
		// 尝试生成数字
		return fmt.Sprintf("%d", rand.Intn(10000))
	}

	// 处理字母模式 [a-z], [A-Z]
	if strings.Contains(pattern, "[a-z]") || strings.Contains(pattern, "[A-Z]") {
		return generateRandomString(10)
	}

	return ""
}

// generateCandidateString 根据尝试次数生成不同长度的候选字符串
func (g *RegexGenerator) generateCandidateString(pattern string, attempt int) string {
	// 分析模式特征
	hasDigits := strings.Contains(pattern, "\\d") || strings.Contains(pattern, "[0-9]")
	hasLetters := strings.Contains(pattern, "[A-Za-z]") || strings.Contains(pattern, "[a-z]") || strings.Contains(pattern, "[A-Z]")
	hasSpecialChars := strings.Contains(pattern, "[-._]") || strings.Contains(pattern, "[@]")

	// 根据尝试次数调整长度
	baseLength := 5
	length := baseLength + (attempt % 30)
	if length > 50 {
		length = 50
	}

	// 根据模式特征生成字符串
	if hasDigits && hasLetters {
		// 混合数字和字母
		return g.generateMixedString(length)
	} else if hasDigits {
		// 纯数字
		return g.generateNumericString(length)
	} else if hasLetters {
		// 纯字母
		return generateRandomString(length)
	} else if hasSpecialChars {
		// 包含特殊字符
		return g.generateStringWithSpecialChars(pattern, length)
	}

	// 默认生成随机字符串
	return generateRandomString(length)
}

// generateNumericString 生成纯数字字符串
func (g *RegexGenerator) generateNumericString(length int) string {
	const charset = "0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// generateStringWithSpecialChars 生成包含特殊字符的字符串
func (g *RegexGenerator) generateStringWithSpecialChars(pattern string, length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	specialChars := []string{".", "-", "_", "@"}

	result := make([]byte, 0, length)

	for i := 0; i < length; i++ {
		// 随机决定是否插入特殊字符
		if i > 0 && i < length-1 && rand.Float32() < 0.2 {
			special := specialChars[rand.Intn(len(specialChars))]
			result = append(result, []byte(special)...)
			i += len(special) - 1
			continue
		}

		// 添加随机字符
		char := charset[rand.Intn(len(charset))]
		result = append(result, char)
	}

	return string(result)
}

// generateMixedString 生成包含数字和字母的混合字符串
func (g *RegexGenerator) generateMixedString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// 常见模式检测和生成函数

// isPhoneCNPattern 检查是否是中国手机号的正则表达式
func (g *RegexGenerator) isPhoneCNPattern(pattern string) bool {
	return strings.Contains(pattern, "1[3-9]") && strings.Contains(pattern, "\\d{9}") ||
		strings.Contains(pattern, "1[3-9]\\d{9}")
}

// generatePhoneCN 生成中国手机号
func (g *RegexGenerator) generatePhoneCN() string {
	prefixes := []string{"130", "131", "132", "133", "134", "135", "136", "137", "138", "139",
		"150", "151", "152", "153", "155", "156", "157", "158", "159",
		"180", "181", "182", "183", "184", "185", "186", "187", "188", "189",
		"191", "193", "195", "196", "197", "198", "199"}
	prefix := prefixes[rand.Intn(len(prefixes))]
	suffix := fmt.Sprintf("%08d", rand.Intn(100000000))
	return prefix + suffix
}

// isIDCardCNPattern 检查是否是中国身份证号的正则表达式
func (g *RegexGenerator) isIDCardCNPattern(pattern string) bool {
	return strings.Contains(pattern, "\\d{5}") && strings.Contains(pattern, "\\d{2}") &&
		strings.Contains(pattern, "\\d{3}") && (strings.Contains(pattern, "[0-9Xx]") || strings.Contains(pattern, "\\d"))
}

// generateIDCardCN 生成中国身份证号
func (g *RegexGenerator) generateIDCardCN() string {
	// 地区码（6位）
	areaCode := fmt.Sprintf("%06d", 110000+rand.Intn(900000))

	// 出生日期（8位）
	year := 1950 + rand.Intn(70)
	month := 1 + rand.Intn(12)
	daysInMonth := 28
	if month == 2 {
		daysInMonth = 28
	} else if month == 4 || month == 6 || month == 9 || month == 11 {
		daysInMonth = 30
	} else {
		daysInMonth = 31
	}
	day := 1 + rand.Intn(daysInMonth)
	birthDate := fmt.Sprintf("%04d%02d%02d", year, month, day)

	// 顺序码（3位）
	sequence := fmt.Sprintf("%03d", rand.Intn(1000))

	// 校验码（1位，可以是0-9或X）
	checkCodes := []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "X"}
	checkCode := checkCodes[rand.Intn(len(checkCodes))]

	return areaCode + birthDate + sequence + checkCode
}

// isURLPattern 检查是否是URL的正则表达式
func (g *RegexGenerator) isURLPattern(pattern string) bool {
	return strings.Contains(pattern, "http") && (strings.Contains(pattern, "://") || strings.Contains(pattern, "https?"))
}

// generateURL 生成URL地址
func (g *RegexGenerator) generateURL() string {
	schemes := []string{"http", "https"}
	scheme := schemes[rand.Intn(len(schemes))]

	domains := []string{"example.com", "test.org", "demo.net", "sample.io", "website.com"}
	domain := domains[rand.Intn(len(domains))]

	paths := []string{"", "/index.html", "/page", "/api/v1", "/users", "/products"}
	path := paths[rand.Intn(len(paths))]

	if path != "" && rand.Float32() < 0.3 {
		path += fmt.Sprintf("/%d", rand.Intn(1000))
	}

	return fmt.Sprintf("%s://%s%s", scheme, domain, path)
}

// isDatePattern 检查是否是日期（YYYY-MM-DD）的正则表达式
func (g *RegexGenerator) isDatePattern(pattern string) bool {
	return strings.Contains(pattern, "\\d{4}") && strings.Contains(pattern, "-") &&
		(strings.Contains(pattern, "0[1-9]|1[0-2]") || strings.Contains(pattern, "\\d{2}"))
}

// generateDate 生成日期（YYYY-MM-DD）
func (g *RegexGenerator) generateDate() string {
	year := 2000 + rand.Intn(25)
	month := 1 + rand.Intn(12)
	daysInMonth := 28
	if month == 2 {
		daysInMonth = 28
	} else if month == 4 || month == 6 || month == 9 || month == 11 {
		daysInMonth = 30
	} else if month == 1 || month == 3 || month == 5 || month == 7 || month == 8 || month == 10 || month == 12 {
		daysInMonth = 31
	}
	day := 1 + rand.Intn(daysInMonth)
	return fmt.Sprintf("%04d-%02d-%02d", year, month, day)
}

// isTimePattern 检查是否是时间（HH:MM:SS）的正则表达式
func (g *RegexGenerator) isTimePattern(pattern string) bool {
	return strings.Contains(pattern, ":") && (strings.Contains(pattern, "[01]\\d|2[0-3]") || strings.Contains(pattern, "\\d{2}"))
}

// generateTime 生成时间（HH:MM:SS）
func (g *RegexGenerator) generateTime() string {
	hour := rand.Intn(24)
	minute := rand.Intn(60)
	second := rand.Intn(60)
	return fmt.Sprintf("%02d:%02d:%02d", hour, minute, second)
}

// isPostcodeCNPattern 检查是否是中国邮政编码的正则表达式
func (g *RegexGenerator) isPostcodeCNPattern(pattern string) bool {
	return strings.Contains(pattern, "[1-9]") && strings.Contains(pattern, "\\d{5}")
}

// generatePostcodeCN 生成中国邮政编码
func (g *RegexGenerator) generatePostcodeCN() string {
	// 中国邮政编码：1-9开头，6位数字
	firstDigit := 1 + rand.Intn(9)
	rest := fmt.Sprintf("%05d", rand.Intn(100000))
	return fmt.Sprintf("%d%s", firstDigit, rest)
}

// generateByPatternStructure 根据正则表达式结构智能生成
func (g *RegexGenerator) generateByPatternStructure(pattern string, re *regexp.Regexp) string {
	// 尝试解析正则表达式的结构并生成
	// 这是一个简化的实现，可以进一步优化

	// 提取字符类
	if strings.Contains(pattern, "[0-9]") || strings.Contains(pattern, "\\d") {
		// 包含数字
		if strings.Contains(pattern, "[A-Za-z]") || strings.Contains(pattern, "[a-z]") {
			// 数字+字母
			return g.generateMixedString(10)
		}
		// 纯数字
		return g.generateNumericString(10)
	}

	// 提取长度限制
	if strings.Contains(pattern, "{") {
		// 尝试提取 {n} 或 {n,m} 格式
		// 这里简化处理，使用固定长度
		return generateRandomString(10)
	}

	return ""
}

type FunctionGenerator struct {
	funcs map[string]func([]interface{}) (interface{}, error)
}

func NewFunctionGenerator() *FunctionGenerator {
	gen := &FunctionGenerator{
		funcs: make(map[string]func([]interface{}) (interface{}, error)),
	}
	gen.registerBuiltinFunctions()
	return gen
}

func (g *FunctionGenerator) Generate(rule *FieldRule, index int64) (interface{}, error) {
	var config FunctionConfig

	// 优先从 map 中直接获取（更可靠）
	if configMap, ok := rule.Config.(map[string]interface{}); ok {
		// 尝试获取 func_name（下划线格式）
		if funcName, ok := configMap["func_name"].(string); ok && funcName != "" {
			config.FuncName = funcName
		} else if funcName, ok := configMap["funcName"].(string); ok && funcName != "" {
			// 尝试获取 funcName（驼峰格式）
			config.FuncName = funcName
		}

		// 获取参数
		if params, ok := configMap["params"].([]interface{}); ok {
			config.Params = params
		}

		// 获取 UUID 配置选项
		if v, exists := configMap["version"]; exists {
			if versionStr, ok := v.(string); ok {
				config.Version = versionStr
			}
		}
		if c, exists := configMap["case"]; exists {
			if caseStr, ok := c.(string); ok {
				config.Case = caseStr
			}
		}
		if h, exists := configMap["with_hyphen"]; exists {
			if withHyphenBool, ok := h.(bool); ok {
				config.WithHyphen = withHyphenBool
			}
		}
	}

	// 如果从 map 中获取失败，尝试使用 JSON 序列化/反序列化
	if config.FuncName == "" {
		if err := unmarshalConfig(rule.Config, &config); err != nil {
			return nil, fmt.Errorf("解析函数配置失败: %w", err)
		}
	}

	if config.FuncName == "" {
		return nil, fmt.Errorf("函数名不能为空，配置: %+v", rule.Config)
	}

	fn, exists := g.funcs[config.FuncName]
	if !exists {
		return nil, fmt.Errorf("未知函数: %s，可用函数: NOW, TODAY, UUID, RAND, RAND_INT, CONCAT", config.FuncName)
	}

	// 对于 UUID 函数，传递配置选项
	if config.FuncName == "UUID" {
		version := config.Version
		if version == "" {
			version = "v4" // 默认使用 v4
		}
		uuidParams := []interface{}{
			map[string]interface{}{
				"version":     version,
				"case":        config.Case,
				"with_hyphen": config.WithHyphen,
			},
		}
		result, err := fn(uuidParams)
		if err != nil {
			return nil, fmt.Errorf("执行函数 %s 失败: %w", config.FuncName, err)
		}
		return result, nil
	}

	result, err := fn(config.Params)
	if err != nil {
		return nil, fmt.Errorf("执行函数 %s 失败: %w", config.FuncName, err)
	}

	return result, nil
}

func (g *FunctionGenerator) registerBuiltinFunctions() {
	g.funcs["NOW"] = func(params []interface{}) (interface{}, error) {
		return time.Now(), nil
	}
	g.funcs["TODAY"] = func(params []interface{}) (interface{}, error) {
		now := time.Now()
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()), nil
	}
	g.funcs["UUID"] = func(params []interface{}) (interface{}, error) {
		// 从配置中获取 UUID 选项（如果存在）
		var caseOption string = "mixed"
		var withHyphen bool = true
		var version string = "v4" // 默认使用 v4

		// 尝试从 params 中获取配置（如果通过 FunctionConfig 传递）
		if len(params) > 0 {
			if configMap, ok := params[0].(map[string]interface{}); ok {
				if c, exists := configMap["case"]; exists {
					if caseStr, ok := c.(string); ok {
						caseOption = caseStr
					}
				}
				if h, exists := configMap["with_hyphen"]; exists {
					if withHyphenBool, ok := h.(bool); ok {
						withHyphen = withHyphenBool
					}
				}
				if v, exists := configMap["version"]; exists {
					if versionStr, ok := v.(string); ok {
						version = versionStr
					}
				}
			}
		}

		var uuidStr string

		// 根据版本生成 UUID
		switch strings.ToLower(version) {
		case "v1":
			// V1: 基于时间戳和 MAC 地址
			uuidVal, err := uuid.NewUUID()
			if err != nil {
				return nil, fmt.Errorf("生成 UUID v1 失败: %w", err)
			}
			uuidStr = uuidVal.String()
		case "v3":
			// V3: 基于命名空间和名称的 MD5 哈希
			// 需要命名空间 UUID 和名称，这里使用默认命名空间和随机名称
			namespace := uuid.NameSpaceDNS
			name := fmt.Sprintf("default-%d", time.Now().UnixNano())
			uuidStr = uuid.NewMD5(namespace, []byte(name)).String()
		case "v4":
			// V4: 随机生成（默认）
			uuidStr = uuid.New().String()
		case "v5":
			// V5: 基于命名空间和名称的 SHA-1 哈希
			// 需要命名空间 UUID 和名称，这里使用默认命名空间和随机名称
			namespace := uuid.NameSpaceDNS
			name := fmt.Sprintf("default-%d", time.Now().UnixNano())
			uuidStr = uuid.NewSHA1(namespace, []byte(name)).String()
		default:
			// 默认使用 v4
			uuidStr = uuid.New().String()
		}

		// 处理大小写
		switch caseOption {
		case "lower":
			uuidStr = strings.ToLower(uuidStr)
		case "upper":
			uuidStr = strings.ToUpper(uuidStr)
		}

		// 处理连字符
		if !withHyphen {
			uuidStr = strings.ReplaceAll(uuidStr, "-", "")
		}

		return uuidStr, nil
	}
	g.funcs["RAND"] = func(params []interface{}) (interface{}, error) {
		return rand.Float64(), nil
	}
	g.funcs["RAND_INT"] = func(params []interface{}) (interface{}, error) {
		min := 0
		max := 100
		if len(params) >= 1 {
			if v, ok := params[0].(float64); ok {
				min = int(v)
			}
		}
		if len(params) >= 2 {
			if v, ok := params[1].(float64); ok {
				max = int(v)
			}
		}
		return min + rand.Intn(max-min+1), nil
	}
	g.funcs["CONCAT"] = func(params []interface{}) (interface{}, error) {
		var parts []string
		for _, p := range params {
			parts = append(parts, fmt.Sprintf("%v", p))
		}
		return strings.Join(parts, ""), nil
	}
}

type NullGenerator struct{}

func (g *NullGenerator) Generate(rule *FieldRule, index int64) (interface{}, error) {
	// 空值规则总是返回 NULL
	return nil, nil
}

type TemplateGenerator struct{}

func (g *TemplateGenerator) Generate(rule *FieldRule, index int64) (interface{}, error) {
	var config TemplateConfig
	if err := unmarshalConfig(rule.Config, &config); err != nil {
		return nil, err
	}

	result := config.Template
	result = strings.ReplaceAll(result, "{name}", generateRandomString(10))
	result = strings.ReplaceAll(result, "{date}", time.Now().Format("2006-01-02"))
	result = strings.ReplaceAll(result, "{number}", fmt.Sprintf("%d", rand.Intn(10000)))
	result = strings.ReplaceAll(result, "{uuid}", uuid.New().String())

	return result, nil
}

// GenerateWithRow 使用当前行数据生成值（用于引用其他字段）
func (g *TemplateGenerator) GenerateWithRow(rule *FieldRule, index int64, rowData map[string]interface{}) (interface{}, error) {
	var config TemplateConfig
	if err := unmarshalConfig(rule.Config, &config); err != nil {
		return nil, err
	}

	result := config.Template
	// 替换字段引用 {field_name}
	for fieldName, value := range rowData {
		placeholder := fmt.Sprintf("{%s}", fieldName)
		result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", value))
	}
	// 替换内置占位符
	result = strings.ReplaceAll(result, "{name}", generateRandomString(10))
	result = strings.ReplaceAll(result, "{date}", time.Now().Format("2006-01-02"))
	result = strings.ReplaceAll(result, "{number}", fmt.Sprintf("%d", rand.Intn(10000)))
	result = strings.ReplaceAll(result, "{uuid}", uuid.New().String())

	return result, nil
}

// ReferenceGenerator 引用其他字段生成器
type ReferenceGenerator struct{}

func (g *ReferenceGenerator) Generate(rule *FieldRule, index int64) (interface{}, error) {
	// 这个方法不应该被直接调用，应该使用 GenerateWithRow
	return nil, fmt.Errorf("引用字段规则需要行数据，请使用 GenerateWithRow")
}

// GenerateWithRow 使用当前行数据生成值
func (g *ReferenceGenerator) GenerateWithRow(rule *FieldRule, index int64, rowData map[string]interface{}) (interface{}, error) {
	var config ReferenceConfig
	if err := unmarshalConfig(rule.Config, &config); err != nil {
		return nil, err
	}

	if config.Expression == "" {
		return nil, fmt.Errorf("表达式不能为空")
	}

	result := config.Expression
	// 替换字段引用 {field_name}
	for fieldName, value := range rowData {
		placeholder := fmt.Sprintf("{%s}", fieldName)
		result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", value))
	}

	// 简单的数学表达式计算（支持 +, -, *, /）
	// 注意：这是一个简化实现，复杂的表达式可能需要使用表达式解析库
	result = evaluateSimpleExpression(result)

	return result, nil
}

// GeographicGenerator 地理数据生成器
type GeographicGenerator struct{}

func (g *GeographicGenerator) Generate(rule *FieldRule, index int64) (interface{}, error) {
	var config GeographicConfig
	if err := unmarshalConfig(rule.Config, &config); err != nil {
		return nil, err
	}

	switch config.Type {
	case "city":
		return generateCity(config.Country), nil
	case "country":
		return generateCountry(), nil
	case "address":
		return generateAddress(config.Country), nil
	case "coordinates":
		return generateCoordinates(), nil
	case "latitude":
		return generateLatitude(), nil
	case "longitude":
		return generateLongitude(), nil
	case "postal_code":
		return generatePostalCode(config.Country), nil
	default:
		return generateCity(config.Country), nil
	}
}

// FileGenerator 从文件读取生成器
type FileGenerator struct {
	fileCache map[string][]string
	mu        sync.Mutex
}

func NewFileGenerator() *FileGenerator {
	return &FileGenerator{
		fileCache: make(map[string][]string),
	}
}

func (g *FileGenerator) Generate(rule *FieldRule, index int64) (interface{}, error) {
	var config FileConfig
	if err := unmarshalConfig(rule.Config, &config); err != nil {
		return nil, err
	}

	// 从缓存或文件读取数据
	lines, err := g.loadFileData(config)
	if err != nil {
		return nil, err
	}

	if len(lines) == 0 {
		return nil, fmt.Errorf("文件 %s 没有数据", config.FilePath)
	}

	// 根据索引选择行
	lineIndex := int(index) % len(lines)
	if config.Loop {
		lineIndex = int(index) % len(lines)
	} else if int(index) >= len(lines) {
		return nil, fmt.Errorf("索引超出文件行数")
	}

	line := lines[lineIndex]

	// 根据文件类型解析
	switch config.FileType {
	case "csv":
		return parseCSVLine(line, config.ColumnIndex)
	case "txt":
		return line, nil
	case "json":
		return parseJSONLine(line, config.ColumnIndex)
	default:
		return line, nil
	}
}

func (g *FileGenerator) loadFileData(config FileConfig) ([]string, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	// 检查缓存
	if lines, exists := g.fileCache[config.FilePath]; exists {
		return lines, nil
	}

	// 读取文件
	data, err := os.ReadFile(config.FilePath)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败: %w", err)
	}

	// 按行分割
	lines := strings.Split(string(data), "\n")
	// 过滤空行
	var filteredLines []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			filteredLines = append(filteredLines, line)
		}
	}

	// 缓存
	g.fileCache[config.FilePath] = filteredLines

	return filteredLines, nil
}

// 辅助函数

func unmarshalConfig(config interface{}, target interface{}) error {
	data, err := json.Marshal(config)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}

func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// 地理数据生成函数

var chineseCities = []string{"北京", "上海", "广州", "深圳", "杭州", "南京", "成都", "武汉", "西安", "重庆", "天津", "苏州", "长沙", "郑州", "青岛", "大连", "宁波", "厦门", "无锡", "福州"}
var usCities = []string{"New York", "Los Angeles", "Chicago", "Houston", "Phoenix", "Philadelphia", "San Antonio", "San Diego", "Dallas", "San Jose"}
var countries = []string{"中国", "美国", "日本", "德国", "英国", "法国", "意大利", "加拿大", "澳大利亚", "韩国", "印度", "巴西", "俄罗斯", "西班牙", "墨西哥"}

func generateCity(country string) string {
	switch country {
	case "CN", "CHN", "中国":
		return chineseCities[rand.Intn(len(chineseCities))]
	case "US", "USA", "美国":
		return usCities[rand.Intn(len(usCities))]
	default:
		// 随机选择
		allCities := append(chineseCities, usCities...)
		return allCities[rand.Intn(len(allCities))]
	}
}

func generateCountry() string {
	return countries[rand.Intn(len(countries))]
}

func generateAddress(country string) string {
	city := generateCity(country)
	streetNum := rand.Intn(999) + 1
	streetNames := []string{"Main Street", "Park Avenue", "First Avenue", "Second Street", "Oak Street", "Elm Street", "Maple Drive", "Cedar Lane"}
	streetName := streetNames[rand.Intn(len(streetNames))]
	return fmt.Sprintf("%d %s, %s", streetNum, streetName, city)
}

func generateCoordinates() string {
	lat := generateLatitude()
	lon := generateLongitude()
	return fmt.Sprintf("%.6f,%.6f", lat, lon)
}

func generateLatitude() float64 {
	// 中国纬度范围大约 18°N - 54°N
	// 全球纬度范围 -90° 到 90°
	return -90 + rand.Float64()*180
}

func generateLongitude() float64 {
	// 中国经度范围大约 73°E - 135°E
	// 全球经度范围 -180° 到 180°
	return -180 + rand.Float64()*360
}

func generatePostalCode(country string) string {
	switch country {
	case "CN", "CHN", "中国":
		// 中国邮编 6 位数字
		return fmt.Sprintf("%06d", rand.Intn(1000000))
	case "US", "USA", "美国":
		// 美国邮编 5 位数字，可选 4 位扩展
		return fmt.Sprintf("%05d", rand.Intn(100000))
	default:
		// 默认 6 位数字
		return fmt.Sprintf("%06d", rand.Intn(1000000))
	}
}

// 文件解析函数

func parseCSVLine(line string, columnIndex int) (interface{}, error) {
	reader := csv.NewReader(strings.NewReader(line))
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("解析 CSV 行失败: %w", err)
	}

	if len(records) == 0 || len(records[0]) == 0 {
		return "", nil
	}

	if columnIndex < 0 || columnIndex >= len(records[0]) {
		return records[0][0], nil // 返回第一列
	}

	return records[0][columnIndex], nil
}

func parseJSONLine(line string, columnIndex int) (interface{}, error) {
	var data interface{}
	if err := json.Unmarshal([]byte(line), &data); err != nil {
		return nil, fmt.Errorf("解析 JSON 行失败: %w", err)
	}

	// 如果是数组，返回指定索引的元素
	if arr, ok := data.([]interface{}); ok {
		if columnIndex >= 0 && columnIndex < len(arr) {
			return arr[columnIndex], nil
		}
		if len(arr) > 0 {
			return arr[0], nil
		}
	}

	// 如果是对象，返回整个对象（columnIndex 可以用于指定字段，但这里简化处理）
	return data, nil
}

// 简单表达式计算函数（支持基本的 +, -, *, /）
func evaluateSimpleExpression(expr string) string {
	// 这是一个非常简化的实现，只处理基本的数学运算
	// 更复杂的表达式可以使用专门的表达式解析库

	// 尝试解析数学表达式，如 "10 + 20", "5 * 3" 等
	// 使用正则表达式匹配数字和运算符
	re := regexp.MustCompile(`(\d+\.?\d*)\s*([+\-*/])\s*(\d+\.?\d*)`)
	matches := re.FindStringSubmatch(expr)

	if len(matches) == 4 {
		val1, err1 := strconv.ParseFloat(matches[1], 64)
		val2, err2 := strconv.ParseFloat(matches[3], 64)
		op := matches[2]

		if err1 == nil && err2 == nil {
			var result float64
			switch op {
			case "+":
				result = val1 + val2
			case "-":
				result = val1 - val2
			case "*":
				result = val1 * val2
			case "/":
				if val2 != 0 {
					result = val1 / val2
				} else {
					return expr // 除零，返回原表达式
				}
			default:
				return expr
			}

			// 如果是整数，返回整数格式
			if result == float64(int64(result)) {
				return fmt.Sprintf("%d", int64(result))
			}
			return fmt.Sprintf("%.2f", result)
		}
	}

	// 如果无法解析为数学表达式，返回原字符串（可能包含字段引用）
	return expr
}

// BinaryGenerator 二进制/图片生成器
type BinaryGenerator struct {
	folderCache map[string][]string // 文件夹路径 -> 文件列表
	mu          sync.Mutex
}

func NewBinaryGenerator() *BinaryGenerator {
	return &BinaryGenerator{
		folderCache: make(map[string][]string),
	}
}

func (g *BinaryGenerator) Generate(rule *FieldRule, index int64) (interface{}, error) {
	var config BinaryConfig
	if err := unmarshalConfig(rule.Config, &config); err != nil {
		return nil, err
	}

	switch config.Mode {
	case "generate":
		return g.generateImage(config, index)
	case "folder":
		return g.readImageFromFolder(config, index)
	default:
		return nil, fmt.Errorf("未知的二进制模式: %s", config.Mode)
	}
}

// generateImage 生成图片
func (g *BinaryGenerator) generateImage(config BinaryConfig, index int64) ([]byte, error) {
	width := config.Width
	if width <= 0 {
		width = 100
	}
	height := config.Height
	if height <= 0 {
		height = 100
	}
	format := config.Format
	if format == "" {
		format = "png"
	}

	// 创建图片
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// 生成随机颜色作为背景
	r := uint8(rand.Intn(256))
	gr := uint8(rand.Intn(256))
	b := uint8(rand.Intn(256))
	bgColor := color.RGBA{r, gr, b, 255}

	// 填充背景
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, bgColor)
		}
	}

	// 在图片中心绘制一个简单的图案（可选）
	centerX := width / 2
	centerY := height / 2
	patternColor := color.RGBA{255 - r, 255 - gr, 255 - b, 255}
	for y := centerY - 10; y < centerY+10; y++ {
		for x := centerX - 10; x < centerX+10; x++ {
			if x >= 0 && x < width && y >= 0 && y < height {
				img.Set(x, y, patternColor)
			}
		}
	}

	// 编码图片
	var buf bytes.Buffer
	switch format {
	case "png":
		if err := png.Encode(&buf, img); err != nil {
			return nil, fmt.Errorf("PNG 编码失败: %w", err)
		}
	case "jpeg", "jpg":
		if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90}); err != nil {
			return nil, fmt.Errorf("JPEG 编码失败: %w", err)
		}
	case "gif":
		// GIF 需要 palette
		palette := color.Palette{color.White, color.Black, bgColor, patternColor}
		palettedImg := image.NewPaletted(image.Rect(0, 0, width, height), palette)
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				palettedImg.Set(x, y, img.At(x, y))
			}
		}
		if err := gif.Encode(&buf, palettedImg, &gif.Options{NumColors: 256}); err != nil {
			return nil, fmt.Errorf("GIF 编码失败: %w", err)
		}
	default:
		return nil, fmt.Errorf("不支持的图片格式: %s", format)
	}

	return buf.Bytes(), nil
}

// readImageFromFolder 从文件夹读取图片
func (g *BinaryGenerator) readImageFromFolder(config BinaryConfig, index int64) ([]byte, error) {
	if config.FolderPath == "" {
		return nil, fmt.Errorf("文件夹路径不能为空")
	}

	// 获取文件列表
	files, err := g.getImageFiles(config.FolderPath, config.Extensions)
	if err != nil {
		return nil, err
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("文件夹 %s 中没有找到图片文件", config.FolderPath)
	}

	// 根据索引选择文件
	fileIndex := int(index) % len(files)
	if config.Loop {
		fileIndex = int(index) % len(files)
	} else if int(index) >= len(files) {
		return nil, fmt.Errorf("索引超出文件数量")
	}

	filePath := files[fileIndex]

	// 读取文件
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("读取图片文件失败: %w", err)
	}

	return data, nil
}

// getImageFiles 获取文件夹中的图片文件列表
func (g *BinaryGenerator) getImageFiles(folderPath string, extensions []string) ([]string, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	// 检查缓存
	if files, exists := g.folderCache[folderPath]; exists {
		return files, nil
	}

	// 默认支持的图片扩展名
	defaultExtensions := []string{"jpg", "jpeg", "png", "gif", "bmp", "webp"}
	if len(extensions) > 0 {
		defaultExtensions = extensions
	}

	// 转换为小写并添加点号
	extMap := make(map[string]bool)
	for _, ext := range defaultExtensions {
		ext = strings.ToLower(strings.TrimPrefix(ext, "."))
		extMap["."+ext] = true
	}

	var imageFiles []string

	// 遍历文件夹
	err := filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 只处理文件，不处理目录
		if info.IsDir() {
			return nil
		}

		// 检查扩展名
		ext := strings.ToLower(filepath.Ext(path))
		if extMap[ext] {
			imageFiles = append(imageFiles, path)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("遍历文件夹失败: %w", err)
	}

	// 缓存结果
	g.folderCache[folderPath] = imageFiles

	return imageFiles, nil
}

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
	var config StringRandomConfig
	if err := unmarshalConfig(rule.Config, &config); err != nil {
		return nil, err
	}

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

	length := minLen + rand.Intn(maxLen-minLen+1)

	// 选择字符集
	var charSet string
	switch config.CharSet {
	case "letters":
		charSet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	case "numbers":
		charSet = "0123456789"
	case "chinese":
		// 简单的中文字符范围
		charSet = "的一是在不了有和人这中大为上个国我以要他时来用们生到作地于出就分对成会可主发年动同工也能下过子说产种面而方后多定行学法所民得经十三之进着等部度家电力里如水化高自二理起小物现实加量都两体制机当使点从业本去把性好应开它合还因由其些然前外天政四日那社义事平形相全表间样与关各重新线内数正心反你明看原又么利比或但质气第向道命此变条只没结解问意建月公无系军很情者最立代想已通并提直题党程展五果料象员革位入常文总次品式活设及管特件长求老头基资边流路级少图山统接知较将组见计别她手角期根论运农指几九区强放决西被干做必战先回则任取据处队南给色光门即保治北造百规热领七海口东导器压志世金增争济阶油思术极交受联什认六共权收证改清己美再采转更单风切打白教速花带安场身车例真务具万每目至达走积示议声报斗完类八离华名确才科张信马节话米整空元况今集温传土许步群广石记需段研界拉林律叫且究观越织装影算低持音众书布复容儿须际商非验连断深难近矿千周委素技备半办青省列习响约支般史感劳便团往酸历市克何除消构府称太准精值号率族维划选标写存候毛亲快效斯院查江型眼王按格养易置派层片始却专状育厂京识适属圆包火住调满县局照参红细引听该铁价严龙飞"
	case "special":
		charSet = "!@#$%^&*()_+-=[]{}|;:,.<>?"
	case "all":
		charSet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()_+-=[]{}|;:,.<>?"
	default:
		if config.CustomChars != "" {
			charSet = config.CustomChars
		} else {
			charSet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
		}
	}

	// 生成随机字符串
	result := make([]byte, length)
	for i := range result {
		result[i] = charSet[rand.Intn(len(charSet))]
	}

	str := string(result)
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
	if config.IsInt {
		value = float64(int64(min) + rand.Int63n(int64(max-min+1)))
	} else {
		value = min + rand.Float64()*(max-min)
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

	// 简单的正则生成实现（实际应该使用专门的库如 github.com/dlclark/regexp2）
	// 这里提供一个基础实现
	re, err := regexp.Compile(config.Pattern)
	if err != nil {
		return nil, fmt.Errorf("无效的正则表达式: %w", err)
	}

	// 简单实现：生成随机字符串并验证
	for i := 0; i < 100; i++ {
		candidate := generateRandomString(20)
		if re.MatchString(candidate) {
			return candidate, nil
		}
	}

	return nil, fmt.Errorf("无法生成匹配正则表达式的值")
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
	if err := unmarshalConfig(rule.Config, &config); err != nil {
		return nil, err
	}

	fn, exists := g.funcs[config.FuncName]
	if !exists {
		return nil, fmt.Errorf("未知函数: %s", config.FuncName)
	}

	return fn(config.Params)
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
		return uuid.New().String(), nil
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
	var config NullConfig
	if err := unmarshalConfig(rule.Config, &config); err != nil {
		return nil, err
	}

	if rand.Float64() < config.Probability {
		return nil, nil
	}
	// 如果概率不满足，返回默认值或空字符串
	return "", nil
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

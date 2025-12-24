package gtsharding

import (
	"fmt"
	"time"

	gscron "github.com/DontBeProud/go-treasure/gt-cron"
)

// TimeLevel 基于时间分割的规则等级
type TimeLevel uint

const (
	minTimeLevel       TimeLevel = TimeLevelDisabled
	TimeLevelDisabled  TimeLevel = 0  // 不使用基于时间分割的规则
	TimeLevelHour      TimeLevel = 1  // 小时
	TimeLevelHalfDay   TimeLevel = 2  // 半天(每12个小时分割一次)
	TimeLevelDay       TimeLevel = 3  // 天
	TimeLevelWeek      TimeLevel = 4  // 周(每周一为分割日)
	TimeLevelHalfMonth TimeLevel = 5  // 半月(每个月的1号和16号位分割月)
	TimeLevelMonth     TimeLevel = 6  // 月
	TimeLevelTwoMonth  TimeLevel = 7  // 2月
	TimeLevelQuarter   TimeLevel = 8  // 季度
	TimeLevelHalfYear  TimeLevel = 9  // 半年
	TimeLevelYear      TimeLevel = 10 // 年
	maxTimeLevel       TimeLevel = TimeLevelYear
)

// RuleWithTimeConfig 基于时间的分割规则配置
type RuleWithTimeConfig struct {
	Level              TimeLevel // 等级
	SplitCharacter     string    // 分割字符
	TimeSplitCharacter string    // 各时间成员之间的分割字符
	EarliestValidTime  time.Time // 最早的有效时间(分库分表的业务场景一般需要设置一个最早有效时间, 否则在查询等场景中可能出现表未创建的情况)
	PrefixMode         bool      // 时间信息是否置于命名前缀(默认位false，即采用raw - date的格式)
}

// GenerateFnModifyRule 生成时间相关的分割修饰规则
func (cfg *RuleWithTimeConfig) GenerateFnModifyRule() (FnStringModifyRule, error) {
	return generateFnModifyRuleWithTime(cfg.Level, cfg.TimeSplitCharacter)
}

func (cfg *RuleWithTimeConfig) check() error {
	if l := cfg.Level; l < minTimeLevel || l > maxTimeLevel {
		return fmt.Errorf("RuleWithTimeConfig: unknown time level: %d", l)
	}

	if cfg.Level != TimeLevelDisabled && cfg.EarliestValidTime.IsZero() {
		return fmt.Errorf("RuleWithTimeConfig: 请设置合理的EarliestValidTime, 否则在查询等场景中可能出现表未创建的情况")
	}

	return nil
}

func (r *rootRule) generateShardingRuleWithTime(cfg *RuleWithTimeConfig) (*shardingRuleWithTime, error) {
	errFmt := "rootRule.generateShardingRuleWithTime: %s"
	if cfg == nil {
		return nil, fmt.Errorf(errFmt, "cfg == nil")
	}

	if err := cfg.check(); err != nil {
		return nil, fmt.Errorf(errFmt, err)
	}

	var prefixRules, suffixRules []FnStringModifyRule
	modifyRule, err := cfg.GenerateFnModifyRule()
	if err != nil {
		return nil, fmt.Errorf(errFmt, err)
	}
	if modifyRule != nil {
		var splitRule FnStringModifyRule
		if cfg.SplitCharacter != "" {
			splitRule = func(param *StringModifyParam) string { return cfg.SplitCharacter }
		}
		if cfg.PrefixMode {
			// date-raw
			prefixRules = []FnStringModifyRule{modifyRule, splitRule}
		} else {
			// raw-date
			suffixRules = []FnStringModifyRule{splitRule, modifyRule}
		}
	}

	return &shardingRuleWithTime{
		RootRule:     r,
		cfg:          *cfg,
		shardingRule: r.generateNamingModificationRule(prefixRules, suffixRules),
	}, nil
}

// 基于时间的分割规则
type shardingRuleWithTime struct {
	RootRule
	shardingRule *shardingRule
	cfg          RuleWithTimeConfig
}

// Modify 修饰
func (r *shardingRuleWithTime) Modify(baseString string, t time.Time) string {
	return r.shardingRule.Modify(baseString, NewStringModifyParam().setTime(t))
}

// BatchModify 批量修饰
func (r *shardingRuleWithTime) BatchModify(baseString string, tList []time.Time) []string {
	result := make([]string, len(tList))
	for index, modifier := range r.BatchGenerateStringModifiers(tList) {
		result[index] = modifier.ModifyString(baseString)
	}
	return result
}

// BatchModifyByTimeRange 【返回去重结果】基于传入的时间范围批量修饰(起始时间会被自动修正为不早于EarliestValidTime的值)
func (r *shardingRuleWithTime) BatchModifyByTimeRange(baseString string, start *time.Time, end *time.Time) []string {
	nodes := r.ExpandValidTimeNodeList(start, end)
	if r.cfg.Level == TimeLevelDisabled && len(nodes) > 1 {
		nodes = nodes[:1]
	}
	return r.BatchModify(baseString, nodes)
}

// GenerateStringModifier 生成字符串修饰器
func (r *shardingRuleWithTime) GenerateStringModifier(t time.Time) StringModifier {
	return r.shardingRule.GenerateStringModifier(NewStringModifyParam().setTime(t))
}

// BatchGenerateStringModifiers 批量生成字符串修饰器
func (r *shardingRuleWithTime) BatchGenerateStringModifiers(tList []time.Time) []StringModifier {
	modifiers := make([]StringModifier, len(tList))
	for index, t := range tList {
		modifiers[index] = r.GenerateStringModifier(t)
	}
	return modifiers
}

// BatchGenerateStringModifiersByTimeRange 【返回去重结果】基于传入的时间范围批量生成字符串修饰器(起始时间会被自动修正为不早于EarliestValidTime的值)
func (r *shardingRuleWithTime) BatchGenerateStringModifiersByTimeRange(start *time.Time, end *time.Time) []StringModifier {
	nodes := r.ExpandValidTimeNodeList(start, end)
	if r.cfg.Level == TimeLevelDisabled && len(nodes) > 1 {
		nodes = nodes[:1]
	}
	return r.BatchGenerateStringModifiers(nodes)
}

const (
	defaultModifyParamKeyTime = "__time__"
)

func (p *StringModifyParam) setTime(t time.Time) *StringModifyParam {
	return p.Set(defaultModifyParamKeyTime, t)
}

func (p *StringModifyParam) getTime() *time.Time {
	_t := p.Get(defaultModifyParamKeyTime)
	if _t == nil {
		return nil
	}
	t := _t.(time.Time)
	return &t
}

// 生成时间相关的分割修饰规则
func generateFnModifyRuleWithTime(level TimeLevel, splitCharacter string) (FnStringModifyRule, error) {
	if level == TimeLevelDisabled {
		return nil, nil
	}

	var timeRule = func(t *time.Time) string {
		return t.Format("2006")
	}
	var indexRule func(t *time.Time) string
	switch level {
	case TimeLevelHour:
		timeRule = func(t *time.Time) string {
			return t.Format(fmt.Sprintf("2006%s01%s02%s15%s04", splitCharacter, splitCharacter, splitCharacter, splitCharacter))
		}
	case TimeLevelHalfDay:
		timeRule = func(t *time.Time) string {
			return t.Format(fmt.Sprintf("2006%s01%s02%s15", splitCharacter, splitCharacter, splitCharacter))
		}
		indexRule = func(t *time.Time) string {
			return map[bool]string{
				true:  "12",
				false: "00",
			}[t.Hour() > 11]
		}
	case TimeLevelDay:
		timeRule = func(t *time.Time) string {
			return t.Format(fmt.Sprintf("2006%s01%s02", splitCharacter, splitCharacter))
		}
	case TimeLevelWeek:
		timeRule = func(t *time.Time) string {
			// 所在周的周一的日期
			return t.AddDate(0, 0, 1-int(t.Weekday())).Format(fmt.Sprintf("2006%s01%s02", splitCharacter, splitCharacter))
		}
	case TimeLevelHalfMonth:
		timeRule = func(t *time.Time) string {
			return t.Format(fmt.Sprintf("2006%s01", splitCharacter))
		}
		indexRule = func(t *time.Time) string {
			return fmt.Sprintf("%02d", 1+15*((int(t.Day())-1)/15))
		}
	case TimeLevelMonth:
		timeRule = func(t *time.Time) string {
			return t.Format(fmt.Sprintf("2006%s01", splitCharacter))
		}
	case TimeLevelTwoMonth:
		indexRule = func(t *time.Time) string {
			return fmt.Sprintf("%02d", 1+2*((int(t.Month())-1)/2))
		}
	case TimeLevelQuarter:
		indexRule = func(t *time.Time) string {
			return fmt.Sprintf("%02d", 1+3*((int(t.Month())-1)/3))
		}
	case TimeLevelHalfYear:
		indexRule = func(t *time.Time) string {
			return fmt.Sprintf("%02d", 1+6*((int(t.Month())-1)/6))
		}
	case TimeLevelYear:
	default:
		return nil, fmt.Errorf("unknown time level: %d", level)
	}

	rule := func(param *StringModifyParam) string {
		ret := ""
		if param == nil {
			return ret
		}

		t := param.getTime()
		if t == nil {
			return ret
		}

		ret = timeRule(t)
		if indexRule != nil {
			if index := indexRule(t); index != "" {
				ret = fmt.Sprintf("%s%s%s", ret, splitCharacter, index)
			}
		}
		return ret
	}
	return rule, nil
}

// GetTimeNode 获取给定时间节点对应的时间节点
func (r *shardingRuleWithTime) GetTimeNode(t time.Time) time.Time {
	switch r.cfg.Level {
	case TimeLevelDisabled:
		break
	case TimeLevelHour:
		t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, t.Location())
	case TimeLevelHalfDay:
		t = time.Date(t.Year(), t.Month(), t.Day(), map[bool]int{
			true:  12,
			false: 0,
		}[t.Hour() >= 12], 0, 0, 0, t.Location())
	case TimeLevelDay:
		t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	case TimeLevelWeek:
		t = t.AddDate(0, 0, 1-int(t.Weekday()))
	case TimeLevelHalfMonth:
		t = time.Date(t.Year(), t.Month(), 1+15*((int(t.Day())-1)/15), 0, 0, 0, 0, t.Location())
	case TimeLevelMonth:
		t = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
	case TimeLevelTwoMonth:
		t = time.Date(t.Year(), time.Month(1+2*((int(t.Month())-1)/2)), 1, 0, 0, 0, 0, t.Location())
	case TimeLevelQuarter:
		t = time.Date(t.Year(), time.Month(1+3*((int(t.Month())-1)/3)), 1, 0, 0, 0, 0, t.Location())
	case TimeLevelHalfYear:
		t = time.Date(t.Year(), time.Month(1+6*((int(t.Month())-1)/6)), 1, 0, 0, 0, 0, t.Location())
	case TimeLevelYear:
		t = time.Date(t.Year(), 1, 1, 0, 0, 0, 0, t.Location())
	default:
		break
	}
	return t
}

// NextTimeNode 推算给定时间节点的后N个时间节点
// round 周期累加轮数
func (r *shardingRuleWithTime) NextTimeNode(raw time.Time, round uint) time.Time {
	t := raw
	for i := uint(0); i < round; i++ {
		t = r.nextTimeNode(t)
	}
	return t
}

// PreTimeNode 推算给定时间节点的前N个时间节点
// round 周期累加轮数
func (r *shardingRuleWithTime) PreTimeNode(raw time.Time, round uint) time.Time {
	t := raw
	for i := uint(0); i < round; i++ {
		t = r.preTimeNode(t)
	}
	// 不能早于最早有效时间
	if t.Before(r.cfg.EarliestValidTime) {
		t = r.GetTimeNode(r.cfg.EarliestValidTime)
	}
	return t
}

// preTimeNode 推算给定时间节点的上一个时间节点
func (r *shardingRuleWithTime) preTimeNode(raw time.Time) time.Time {
	t := raw
	switch r.cfg.Level {
	case TimeLevelDisabled:
		break
	case TimeLevelHour:
		t = t.Add(-1 * time.Hour)
	case TimeLevelHalfDay:
		t = t.Add(-12 * time.Hour)
	case TimeLevelDay:
		t = t.AddDate(0, 0, -1)
	case TimeLevelWeek:
		t = t.AddDate(0, 0, -7)
	case TimeLevelHalfMonth:
		if t.Day() > 15 {
			t = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location()) // 本月1号
		} else {
			t = time.Date(t.Year(), t.Month()-1, 16, 0, 0, 0, 0, t.Location()) // 上月16号
		}
	case TimeLevelMonth:
		t = t.AddDate(0, -1, 0)
	case TimeLevelTwoMonth:
		t = t.AddDate(0, -2, 0)
	case TimeLevelQuarter:
		if t.Month() <= 3 {
			t = time.Date(t.Year()-1, 10, 1, 0, 0, 0, 0, t.Location()) // 上年10月
		} else {
			t = time.Date(t.Year(), ((t.Month()-4)/3)*3+1, 1, 0, 0, 0, 0, t.Location()) // 上一季度
		}
	case TimeLevelHalfYear:
		if t.Month() > 6 {
			t = time.Date(t.Year(), 1, 1, 0, 0, 0, 0, t.Location()) // 今年1月
		} else {
			t = time.Date(t.Year()-1, 7, 1, 0, 0, 0, 0, t.Location()) // 上年7月
		}
	case TimeLevelYear:
		t = time.Date(t.Year()-1, 1, 1, 0, 0, 0, 0, t.Location())
	default:
		break
	}
	return r.GetTimeNode(t)
}

// NextTimeNode 推算给定时间节点的下一个时间节点
func (r *shardingRuleWithTime) nextTimeNode(raw time.Time) time.Time {
	t := raw
	switch r.cfg.Level {
	case TimeLevelDisabled:
		break
	case TimeLevelHour:
		t = t.Add(1 * time.Hour)
	case TimeLevelHalfDay:
		t = t.Add(12 * time.Hour)
	case TimeLevelDay:
		t = t.AddDate(0, 0, 1)
	case TimeLevelWeek:
		t = t.AddDate(0, 0, 7)
	case TimeLevelHalfMonth:
		if t.Day() > 15 {
			t = time.Date(t.Year(), t.Month()+1, 1, 0, 0, 0, 0, t.Location()) // 次月1号
		} else {
			t = time.Date(t.Year(), t.Month(), 16, 0, 0, 0, 0, t.Location()) // 当月16号
		}
	case TimeLevelMonth:
		t = t.AddDate(0, 1, 0)
	case TimeLevelTwoMonth:
		t = t.AddDate(0, 2, 0)
	case TimeLevelQuarter:
		if t.Month() > 9 {
			t = time.Date(t.Year()+1, 1, 1, 0, 0, 0, 0, t.Location()) // 次年1月
		} else {
			t = time.Date(t.Year(), ((t.Month()-1)/3)*3+4, 1, 0, 0, 0, 0, t.Location()) // 下一季度
		}
	case TimeLevelHalfYear:
		if t.Month() > 6 {
			t = time.Date(t.Year()+1, 1, 1, 0, 0, 0, 0, t.Location()) // 次年1月
		} else {
			t = time.Date(t.Year(), 7, 1, 0, 0, 0, 0, t.Location()) // 本年7月
		}
	case TimeLevelYear:
		t = time.Date(t.Year()+1, 1, 1, 0, 0, 0, 0, t.Location())
	default:
		break
	}
	return r.GetTimeNode(t)
}

// ExpandValidTimeNodeList 【返回去重结果】根据传入的时间区间，展开生成有效的时间节点列表，用于后续生成修饰器列表(起始时间会被自动修正为不早于EarliestValidTime的值)
func (r *shardingRuleWithTime) ExpandValidTimeNodeList(start *time.Time, end *time.Time) []time.Time {
	_start := r.cfg.EarliestValidTime
	if start != nil && start.After(_start) {
		_start = *start
	}

	_end := time.Now()
	if end != nil {
		_end = *end
	}

	tList := make([]time.Time, 0)
	for t := _start; !t.After(_end); {
		tList = append(tList, t)
		t = r.nextTimeNode(t)
	}

	return tList
}

// GetDefaultCronExpressionGenerator 根据时间分割规则等级，生成对应的默认计划任务表达式
// 秒和分钟默认设为0
// 日期在TimeLevelMonth及以上的等级默认设为对应时间周期内的第一天
func (r *shardingRuleWithTime) GetDefaultCronExpressionGenerator() gscron.ExpressionGenerator {
	generator := gscron.NewExpressionGenerator().SetSecondConstValue(0).SetMinuteConstValue(0).SetHourConstValue(0)
	switch r.cfg.Level {
	case TimeLevelHour:
		// 每小时
		var interval uint32 = 1
		generator.SetHourRange(*gscron.NewTimeRange(), &interval)
	case TimeLevelHalfDay:
		// 每半天
		var interval uint32 = 12
		generator.SetHourRange(*gscron.NewTimeRange(), &interval)
	case TimeLevelDay:
		// 每天
		var interval uint32 = 1
		generator.SetDayRange(*gscron.NewTimeRange(), &interval)
	case TimeLevelWeek:
		// 每周一
		generator.SetWeekConstValue(1)
	case TimeLevelHalfMonth:
		// 每月1/16号
		generator.SetMonthsConstValueList([]uint32{1, 16})
	case TimeLevelMonth:
		// 每个月第一天
		var interval uint32 = 1
		generator.SetDayConstValue(1).SetMonthRange(*gscron.NewTimeRange(), &interval)
	case TimeLevelTwoMonth:
		// 1/3/5/7/9/11月第1天
		var interval uint32 = 2
		generator.SetDayConstValue(1).SetMonthRange(*gscron.NewTimeRange().SetStart(1), &interval)
	case TimeLevelQuarter:
		// 1/4/7/10月第1天
		var interval uint32 = 3
		generator.SetDayConstValue(1).SetMonthRange(*gscron.NewTimeRange().SetStart(1), &interval)
	case TimeLevelHalfYear:
		// 6/12月第1天
		var interval uint32 = 6
		generator.SetDayConstValue(1).SetMonthRange(*gscron.NewTimeRange().SetStart(1), &interval)
	case TimeLevelYear:
		// 每年第1天
		generator.SetDayConstValue(1).SetMonthConstValue(1)
	default:
		return nil
	}
	return generator
}

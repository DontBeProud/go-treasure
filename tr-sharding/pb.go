package trsharding

import (
	trconfpb "github.com/DontBeProud/go-treasure/pb/tr-conf-pb"
	"time"
)

// NewRootRuleWithPb 基于pb创建根命名规则
func NewRootRuleWithPb(cfg *trconfpb.ShardingRootConfig) RootRule {
	if cfg == nil {
		return &rootRule{}
	}
	return &rootRule{
		constPrefix: cfg.GetConstPrefix(),
		constSuffix: cfg.GetConstSuffix(),
	}
}

// GenerateSubShardingRuleWithTimeWithPb 生成基于时间的子分割规则(通过pb)
func (r *rootRule) GenerateSubShardingRuleWithTimeWithPb(cfg *trconfpb.ShardingWithTimeConfig) (SubRuleWithTime, error) {
	return r.GenerateSubShardingRuleWithTime(ParsesShardingWithTimeConfigPb(cfg))
}

// GenerateSubShardingRuleWithGroupWithPb 生成基于分组的子分割规则(通过pb)
func (r *rootRule) GenerateSubShardingRuleWithGroupWithPb(cfg *trconfpb.ShardingWithGroupConfig) (SubRuleWithGroup, error) {
	return r.GenerateSubShardingRuleWithGroup(ParseRuleWithGroupConfigWithPb(cfg))
}

func ParseRuleWithGroupConfigWithPb(cfg *trconfpb.ShardingWithGroupConfig) *RuleWithGroupConfig {
	if cfg == nil {
		return nil
	}

	return &RuleWithGroupConfig{
		GroupSize:             cfg.GetGroupSize(),
		SplitCharacter:        cfg.GetSplitCharacter(),
		PrefixMode:            cfg.GetPrefixMode(),
		IndexIncreaseFromZero: cfg.GetIndexIncreaseFromZero(),
	}
}

// ParsesShardingWithTimeConfigPb 解析ShardingWithTimeConfigPb
func ParsesShardingWithTimeConfigPb(cfg *trconfpb.ShardingWithTimeConfig) *RuleWithTimeConfig {
	if cfg == nil {
		return nil
	}

	res := &RuleWithTimeConfig{
		Level:              TimeLevel(cfg.GetTimeLevel().Number()),
		SplitCharacter:     cfg.GetSplitCharacter(),
		TimeSplitCharacter: cfg.GetTimeSplitCharacter(),
		PrefixMode:         cfg.GetPrefixMode(),
	}

	if t := cfg.GetEarliestValidTime(); t != 0 {
		res.EarliestValidTime = time.Unix(int64(cfg.GetEarliestValidTime()), 0)
	}

	return res
}

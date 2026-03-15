package gtredis

import (
	"github.com/redis/go-redis/v9"
)

var (
	// ErrNil is returned when the reply is nil but a non-nil value was expected.
	ErrNil = redis.Nil
	// ErrClosed is returned when an operation is performed on a closed client.
	ErrClosed = redis.ErrClosed
	// TxFailedErr transaction redis failed.
	TxFailedErr = redis.TxFailedErr
)

type (
	Cmder     = redis.Cmder
	Pipeliner = redis.Pipeliner
	Scripter  = redis.Scripter
)

// command.go
type (
	Cmd                        = redis.Cmd
	SliceCmd                   = redis.SliceCmd
	StatusCmd                  = redis.StatusCmd
	IntCmd                     = redis.IntCmd
	IntSliceCmd                = redis.IntSliceCmd
	DurationCmd                = redis.DurationCmd
	TimeCmd                    = redis.TimeCmd
	BoolCmd                    = redis.BoolCmd
	StringCmd                  = redis.StringCmd
	FloatCmd                   = redis.FloatCmd
	FloatSliceCmd              = redis.FloatSliceCmd
	StringSliceCmd             = redis.StringSliceCmd
	KeyValue                   = redis.KeyValue
	KeyValueSliceCmd           = redis.KeyValueSliceCmd
	BoolSliceCmd               = redis.BoolSliceCmd
	MapStringStringCmd         = redis.MapStringStringSliceCmd
	MapStringIntCmd            = redis.MapStringIntCmd
	StringStructMapCmd         = redis.StringStructMapCmd
	XMessage                   = redis.XMessage
	XMessageSliceCmd           = redis.XMessageSliceCmd
	XStream                    = redis.XStream
	XStreamSliceCmd            = redis.XStreamSliceCmd
	XPending                   = redis.XPending
	XPendingCmd                = redis.XPendingCmd
	XPendingExt                = redis.XPendingExt
	XPendingExtCmd             = redis.XPendingExtCmd
	XAutoClaimCmd              = redis.XAutoClaimCmd
	XAutoClaimJustIDCmd        = redis.XAutoClaimJustIDCmd
	XInfoConsumersCmd          = redis.XInfoConsumersCmd
	XInfoConsumer              = redis.XInfoConsumer
	XInfoGroupsCmd             = redis.XInfoGroupsCmd
	XInfoGroup                 = redis.XInfoGroup
	XInfoStreamCmd             = redis.XInfoStreamCmd
	XInfoStream                = redis.XInfoStream
	XInfoStreamFullCmd         = redis.XInfoStreamFullCmd
	XInfoStreamFull            = redis.XInfoStreamFull
	XInfoStreamGroup           = redis.XInfoStreamGroup
	XInfoStreamGroupPending    = redis.XInfoStreamGroupPending
	XInfoStreamConsumer        = redis.XInfoStreamConsumer
	XInfoStreamConsumerPending = redis.XInfoStreamConsumerPending
	ZSliceCmd                  = redis.ZSliceCmd
	ZWithKeyCmd                = redis.ZWithKeyCmd
	ScanCmd                    = redis.ScanCmd
	ClusterNode                = redis.ClusterNode
	ClusterSlot                = redis.ClusterSlot
	ClusterSlotsCmd            = redis.ClusterSlotsCmd
	GeoLocation                = redis.GeoLocation
	GeoRadiusQuery             = redis.GeoRadiusQuery
	GeoLocationCmd             = redis.GeoLocationCmd
	GeoSearchQuery             = redis.GeoSearchQuery
	GeoSearchLocationQuery     = redis.GeoSearchLocationQuery
	GeoSearchStoreQuery        = redis.GeoSearchStoreQuery
	GeoSearchLocationCmd       = redis.GeoSearchLocationCmd
	GeoPos                     = redis.GeoPos
	GeoPosCmd                  = redis.GeoPosCmd
	CommandInfo                = redis.CommandInfo
	CommandsInfoCmd            = redis.CommandsInfoCmd
	SlowLog                    = redis.SlowLog
	SlowLogCmd                 = redis.SlowLogCmd
	MapStringInterfaceCmd      = redis.MapStringInterfaceCmd
	MapStringStringSliceCmd    = redis.MapStringStringSliceCmd
	KeyValuesCmd               = redis.KeyValuesCmd
	ZSliceWithKeyCmd           = redis.ZSliceWithKeyCmd
	Function                   = redis.Function
	Library                    = redis.Library
	FunctionListCmd            = redis.FunctionListCmd
	FunctionStats              = redis.FunctionStats
	RunningScript              = redis.RunningScript
	Engine                     = redis.Engine
	FunctionStatsCmd           = redis.FunctionStatsCmd
	LCSQuery                   = redis.LCSQuery
	LCSMatch                   = redis.LCSMatch
	LCSMatchedPosition         = redis.LCSMatchedPosition
	LCSPosition                = redis.LCSPosition
	LCSCmd                     = redis.LCSCmd
	KeyFlags                   = redis.KeyFlags
	KeyFlagsCmd                = redis.KeyFlagsCmd
	ClusterLink                = redis.ClusterLink
	ClusterLinksCmd            = redis.ClusterLinksCmd
	SlotRange                  = redis.SlotRange
	Node                       = redis.Node
	ClusterShard               = redis.ClusterShard
	ClusterShardsCmd           = redis.ClusterShardsCmd
	RankScore                  = redis.RankScore
	RankWithScoreCmd           = redis.RankWithScoreCmd
	ClientFlags                = redis.ClientFlags
	ClientInfo                 = redis.ClientInfo
	ClientInfoCmd              = redis.ClientInfoCmd
	ACLLogEntry                = redis.ACLLogEntry
	ACLLogCmd                  = redis.ACLLogCmd
)

// commands.go
type (
	FilterBy           = redis.FilterBy
	Sort               = redis.Sort
	SetArgs            = redis.SetArgs
	BitCount           = redis.BitCount
	LPosArgs           = redis.LPosArgs
	XAddArgs           = redis.XAddArgs
	XReadArgs          = redis.XReadArgs
	XReadGroupArgs     = redis.XReadGroupArgs
	XPendingExtArgs    = redis.XPendingExtArgs
	XClaimArgs         = redis.XClaimArgs
	Z                  = redis.Z
	ZWithKey           = redis.ZWithKey
	ZStore             = redis.ZStore
	ZAddArgs           = redis.ZAddArgs
	ZRangeArgs         = redis.ZRangeArgs
	ZRangeBy           = redis.ZRangeBy
	FunctionListQuery  = redis.FunctionListQuery
	ModuleLoadexConfig = redis.ModuleLoadexConfig
)

// pubsub.go
type (
	PubSub       = redis.PubSub
	Subscription = redis.Subscription
	Message      = redis.Message
	Pong         = redis.Pong
)

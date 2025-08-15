package provider

import (
	"confremote-pilot/codec"
	"confremote-pilot/mediator"
)

type option struct {
	coordinator *mediator.Coordinator
	configType  codec.CfgFileType
	properties  map[string]interface{} // 配置参数
	sources     []*Source              // 每个source对应一个具体的远端配置文件
	customKey   string
}
type Option func(*option)

func WithMediator(coordinator *mediator.Coordinator) Option {
	return func(o *option) {
		o.coordinator = coordinator
	}
}
func WithProperties(properties map[string]interface{}) Option {
	return func(o *option) {
		o.properties = properties
	}
}

func WithSources(sources []*Source) Option {
	return func(o *option) {
		o.sources = sources
	}
}
func WithConfigType(tp codec.CfgFileType) Option {
	return func(o *option) {
		o.configType = tp
	}
}
func WithCustomKey(key string) Option {
	return func(o *option) {
		o.customKey = key
	}
}

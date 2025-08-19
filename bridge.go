package confremote_pilot

import (
	"context"
	"github.com/fufuzion/confremote-pilot/codec"
	"github.com/fufuzion/confremote-pilot/mediator"
	"github.com/fufuzion/confremote-pilot/provider"
	"github.com/spf13/viper"
	"sync"
	"sync/atomic"
)

var bridge *Bridge
var once sync.Once

type Bridge struct {
	ctx         context.Context
	vp          atomic.Value
	mu          *sync.RWMutex
	pvm         map[string]provider.Provider
	coordinator *mediator.Coordinator
	hook        func(key string, msg map[string]any)
}

func Instance(ctx context.Context) *Bridge {
	once.Do(func() {
		bridge = &Bridge{
			ctx: ctx,
			vp:  atomic.Value{},
			mu:  &sync.RWMutex{},
			pvm: make(map[string]provider.Provider),
		}
		bridge.vp.Store(viper.New())
		bridge.coordinator = mediator.NewCoordinator(bridge)

	})
	return bridge
}

func (b *Bridge) Config() *viper.Viper {
	return b.vp.Load().(*viper.Viper)
}
func (b *Bridge) Get(key string) any {
	return b.vp.Load().(*viper.Viper).Get(key)
}
func (b *Bridge) All() map[string]any {
	return b.vp.Load().(*viper.Viper).AllSettings()
}

func (b *Bridge) SetHook(hook func(key string, msg map[string]any)) {
	b.hook = hook
}

func (b *Bridge) Update(key string, msg map[string]any) {
	mergeConfig := viper.New()
	b.mu.RLock()
	defer b.mu.RUnlock()
	for _, pv := range b.pvm {
		subConfig, err := pv.Load()
		if err != nil {
			return
		}
		err = mergeConfig.MergeConfigMap(subConfig)
		if err != nil {
			return
		}
	}
	b.vp.Store(mergeConfig)
	if b.hook != nil {
		b.hook(key, msg)
	}
}

/*
	properties 为初始化client参数，内容跟随provider变化
	provider = 'nacos'时：
		properties := map[string]interface{}{
			constant.KEY_CLIENT_CONFIG:  constant.ClientConfig{},
			constant.KEY_SERVER_CONFIGS: []constant.ServerConfig{},
		}
	provider = 'etcd' | 'consul' | 'firestore'时：
		properties := map[string]interface{}{
			"endpoint": string,
			"path": string,
		}
	provider = 'zookeeper'时：
		properties := map[string]interface{}{
			"endpoint": string, // 多个节点用逗号分隔
			"path": string,
			“timeout”： time.Duration/number/string, // number类型时单位为Millisecond，默认5s
		}
*/

type Config struct {
	Provider   provider.CfgProviderType `json:"provider"`    // Provider CfgProviderType, e.g., "nacos".
	Properties map[string]interface{}   `json:"properties"`  // client and server init param.
	Sources    []*provider.Source       `json:"sources"`     // nacos支持同一个实例下支持加载多个source，provider='nacos'时必传
	ConfigType codec.CfgFileType        `json:"config_type"` // 配置文件的格式类型，目前支持"yaml"和"json"
}

func (b *Bridge) RegisterSource(key string, cfg *Config) error {
	pv, err := provider.NewProvider(
		b.ctx,
		cfg.Provider,
		provider.WithMediator(b.coordinator),
		provider.WithProperties(cfg.Properties),
		provider.WithSources(cfg.Sources),
		provider.WithConfigType(cfg.ConfigType),
		provider.WithCustomKey(key),
	)
	if err != nil {
		return err
	}

	newConfig := viper.New()
	setting, err := pv.Load()
	if err != nil {
		return err
	}
	if err = newConfig.MergeConfigMap(setting); err != nil {
		return err
	}

	b.mu.Lock()
	defer b.mu.Unlock()
	b.pvm[key] = pv

	current := b.vp.Load().(*viper.Viper).AllSettings()
	if err = newConfig.MergeConfigMap(current); err != nil {
		return err
	}
	b.vp.Store(newConfig)
	return nil
}

func (b *Bridge) RegisterSourceBatch(sources map[string]*Config) error {
	newConfig := viper.New()
	for key, cfg := range sources {
		pv, err := provider.NewProvider(
			b.ctx,
			cfg.Provider,
			provider.WithMediator(b.coordinator),
			provider.WithProperties(cfg.Properties),
			provider.WithSources(cfg.Sources),
			provider.WithConfigType(cfg.ConfigType),
			provider.WithCustomKey(key),
		)
		if err != nil {
			return err
		}
		setting, err := pv.Load()
		if err != nil {
			return err
		}
		if err = newConfig.MergeConfigMap(setting); err != nil {
			return err
		}
		b.mu.Lock()
		b.pvm[key] = pv
		b.mu.Unlock()
	}

	b.mu.Lock()
	defer b.mu.Unlock()
	current := b.vp.Load().(*viper.Viper).AllSettings()
	if err := newConfig.MergeConfigMap(current); err != nil {
		return err
	}
	b.vp.Store(newConfig)
	return nil
}

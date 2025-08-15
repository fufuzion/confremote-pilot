package provider

import (
	"context"
	"errors"
	"github.com/spf13/viper"
	"github.com/thoas/go-funk"
	"time"
)

type viperBaseProvider struct {
	ctx context.Context
	tp  CfgProviderType
	vp  *viper.Viper
	o   *option
}

func newViperBaseProvider(ctx context.Context, tp CfgProviderType, o *option) (Provider, error) {
	if funk.IsEmpty(o.properties) {
		return nil, errors.New("properties is required")
	}
	endpoint, ok := o.properties["endpoint"].(string)
	if !ok {
		return nil, errors.New("endpoint is required")
	}
	path, ok := o.properties["path"].(string)
	if !ok {
		return nil, errors.New("path is required")
	}
	vp := viper.New()
	vp.SetConfigType(o.configType.ToString())
	if err := vp.AddRemoteProvider(tp.ToString(), endpoint, path); err != nil {
		return nil, err
	}
	if err := vp.ReadRemoteConfig(); err != nil {
		return nil, err
	}
	provider := &viperBaseProvider{
		ctx: ctx,
		tp:  tp,
		vp:  vp,
		o:   o,
	}
	provider.listen()
	return provider, nil
}
func (p *viperBaseProvider) Name() string {
	return p.tp.ToString()
}

func (p *viperBaseProvider) Load() (map[string]interface{}, error) {
	return p.vp.AllSettings(), nil
}

func (p *viperBaseProvider) listen() {
	go func() {
		backoff := time.Second
		for {
			select {
			case <-p.ctx.Done():
				return
			default:
				if err := p.vp.WatchRemoteConfig(); err != nil {
					time.Sleep(backoff)
					if backoff < 30*time.Second {
						backoff *= 2
					}
					continue
				}
				backoff = time.Second
				settings := func() map[string]interface{} {
					return p.vp.AllSettings()
				}()
				p.notify(settings)
			}
		}
	}()
}
func (p *viperBaseProvider) notify(change map[string]interface{}) {
	if p.o.coordinator == nil {
		return
	}
	p.o.coordinator.Notify(p.o.customKey, change)
}

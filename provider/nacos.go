package provider

import (
	"errors"
	"github.com/fufuzion/confremote-pilot/codec"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/thoas/go-funk"
	"sync"
)

// https://nacos.io/docs/latest/manual/user/go-sdk/usage/
// nacosProvider 包装nacos-group/nacos-sdk-go/v2
type nacosProvider struct {
	tp     CfgProviderType
	mu     *sync.RWMutex
	data   map[string]map[string]interface{}
	client config_client.IConfigClient
	codec  codec.Codec
	o      *option
}

func checkParam(o *option) error {
	if o.properties == nil {
		return errors.New("properties is required")
	}
	if len(o.sources) == 0 {
		return errors.New("sources is required")
	}
	return nil
}

func newNacosProvider(o *option) (Provider, error) {
	if err := checkParam(o); err != nil {
		return nil, err
	}
	client, err := clients.CreateConfigClient(o.properties)
	if err != nil {
		return nil, err
	}
	p := &nacosProvider{
		tp:     CfgProviderNacos,
		mu:     &sync.RWMutex{},
		codec:  codec.NewCodec(o.configType),
		data:   make(map[string]map[string]interface{}),
		client: client,
		o:      o,
	}

	for _, source := range o.sources {
		subSetting, err := p.readRemote(source.DataId, source.Group)
		if err != nil {
			return nil, err
		}
		p.mu.Lock()
		p.data[p.dataKey(source.DataId, source.Group)] = subSetting
		p.mu.Unlock()

		err = p.listen(source.DataId, source.Group)
		if err != nil {
			return nil, err
		}
	}
	return p, nil
}

func (p *nacosProvider) Name() string {
	return p.tp.ToString()
}

func (p *nacosProvider) Load() (map[string]interface{}, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	ret := make(map[string]interface{})
	for _, vm := range p.data {
		for k, v := range vm {
			ret[k] = v
		}
	}
	return ret, nil
}

func (p *nacosProvider) readRemote(dataId string, group string) (map[string]interface{}, error) {
	content, err := p.client.GetConfig(vo.ConfigParam{
		DataId: dataId,
		Group:  group,
	})
	if err != nil {
		return nil, err
	}
	setting := make(map[string]interface{})
	err = p.codec.Decode([]byte(content), &setting)
	if err != nil {
		return nil, err
	}
	return setting, nil
}
func (p *nacosProvider) listen(dataId string, group string) error {
	return p.client.ListenConfig(vo.ConfigParam{
		DataId: dataId,
		Group:  group,
		OnChange: func(namespace, group, dataId, data string) {
			p.onChange(group, dataId, data)
		},
	})
}

func (p *nacosProvider) onChange(group, dataId string, data string) {
	setting := make(map[string]interface{})
	err := p.codec.Decode([]byte(data), &setting)
	if err != nil {
		return
	}
	if funk.IsEmpty(setting) {
		return
	}
	go p.notify(setting)
	p.mu.Lock()
	defer p.mu.Unlock()
	p.data[p.dataKey(dataId, group)] = setting
}

func (p *nacosProvider) notify(change map[string]interface{}) {
	if p.o.coordinator == nil {
		return
	}
	p.o.coordinator.Notify(p.o.customKey, change)
}
func (p *nacosProvider) dataKey(dataId string, group string) string {
	return group + ":" + dataId
}

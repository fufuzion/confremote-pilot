package confremote_pilot

import (
	"confremote-pilot/codec"
	"confremote-pilot/provider"
	"encoding/json"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"testing"
	"time"
)

func TestBridge_RegisterNacosSource(t *testing.T) {
	var (
		clientConfig = constant.ClientConfig{
			NamespaceId: "xxxx",
			AccessKey:   "xxxx",
			SecretKey:   "xxxx",
			ContextPath: "/nacos",
		}
		serverConfig = constant.ServerConfig{
			Scheme:      provider.SchemeTypeHttp,
			ContextPath: "/nacos",
			IpAddr:      "xxx.xxx.xxx.xxx",
			Port:        8848,
		}
	)
	properties := map[string]interface{}{
		constant.KEY_CLIENT_CONFIG:  clientConfig,
		constant.KEY_SERVER_CONFIGS: []constant.ServerConfig{serverConfig},
	}
	sources := make([]*provider.Source, 0)
	sources = append(sources, &provider.Source{
		DataId: "xxxx",
		Group:  "xxxx",
	})
	sources = append(sources, &provider.Source{
		DataId: "xxxx",
		Group:  "xxxx",
	})
	sources = append(sources, &provider.Source{
		DataId: "xxxx",
		Group:  "xxxx",
	})

	err := Instance().RegisterSource("xxxx", &Config{
		Provider:   provider.CfgProviderNacos,
		Properties: properties,
		Sources:    sources,
		ConfigType: codec.CfgFileTypeYaml,
	})
	if err != nil {
		t.Error(err)
		return
	}
	Instance().SetHook(func(key string, msg map[string]any) {
		b, _ := json.Marshal(msg)
		t.Logf("onchange: key %s msg = %s", key, string(b))
	})
	for i := 0; i < 5; i++ {
		b, _ := json.Marshal(Instance().All())
		t.Log(string(b))
		time.Sleep(10 * time.Second)
	}
}

func TestBridge_RegisterNacosSourceBatch(t *testing.T) {
	properties1 := map[string]interface{}{
		constant.KEY_CLIENT_CONFIG: constant.ClientConfig{
			NamespaceId: "xxxx",
			AccessKey:   "xxxx",
			SecretKey:   "xxxx",
			ContextPath: "/nacos",
		},
		constant.KEY_SERVER_CONFIGS: []constant.ServerConfig{
			{
				Scheme:      provider.SchemeTypeHttp,
				ContextPath: "/nacos",
				IpAddr:      "xxx.xxx.xxx.xxx",
				Port:        8848,
			},
		},
	}
	sources1 := make([]*provider.Source, 0)
	sources1 = append(sources1, &provider.Source{
		DataId: "xxxx",
		Group:  "xxxx",
	})
	properties2 := map[string]interface{}{
		constant.KEY_CLIENT_CONFIG: constant.ClientConfig{
			NamespaceId: "xxxx",
			AccessKey:   "xxxx",
			SecretKey:   "xxxx",
			ContextPath: "/nacos",
			Endpoint:    "xxxx", // 阿里云由acm托管
		},
	}
	sources2 := make([]*provider.Source, 0)
	sources2 = append(sources2, &provider.Source{
		DataId: "xxxx",
		Group:  "xxxx",
	})
	sourceM := map[string]*Config{
		"tx-1": {
			Provider:   provider.CfgProviderNacos,
			Properties: properties1,
			Sources:    sources1,
			ConfigType: codec.CfgFileTypeYaml,
		},
		"ali-1": {
			Provider:   provider.CfgProviderNacos,
			Properties: properties2,
			Sources:    sources2,
			ConfigType: codec.CfgFileTypeYaml,
		},
	}
	err := Instance().RegisterSourceBatch(sourceM)
	if err != nil {
		t.Error(err)
		return
	}
	Instance().SetHook(func(key string, msg map[string]any) {
		b, _ := json.Marshal(msg)
		t.Logf("onchange: key %s msg = %s", key, string(b))
	})
	for i := 0; i < 3; i++ {
		b, _ := json.Marshal(Instance().All())
		t.Log(string(b))
		time.Sleep(10 * time.Second)
	}
}

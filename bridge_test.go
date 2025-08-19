package confremote_pilot

import (
	"context"
	"encoding/json"
	"github.com/fufuzion/confremote-pilot/codec"
	"github.com/fufuzion/confremote-pilot/provider"
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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
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

	err := Instance(ctx).RegisterSource("xxxx", &Config{
		Provider:   provider.CfgProviderNacos,
		Properties: properties,
		Sources:    sources,
		ConfigType: codec.CfgFileTypeYaml,
	})
	if err != nil {
		t.Error(err)
		return
	}
	Instance(ctx).SetHook(func(key string, msg map[string]any) {
		b, _ := json.Marshal(msg)
		t.Logf("onchange: key %s msg = %s", key, string(b))
	})
	tk := time.NewTicker(time.Second * 50)
	defer tk.Stop()
	tk2 := time.NewTicker(time.Second * 10)
	defer tk2.Stop()
	for {
		select {
		case <-tk.C:
			t.Log("Test Done.")
			return
		case <-tk2.C:
			b, _ := json.Marshal(Instance(ctx).All())
			t.Log(string(b))
		}
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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err := Instance(ctx).RegisterSourceBatch(sourceM)
	if err != nil {
		t.Error(err)
		return
	}
	Instance(ctx).SetHook(func(key string, msg map[string]any) {
		b, _ := json.Marshal(msg)
		t.Logf("onchange: key %s msg = %s", key, string(b))
	})
	tk := time.NewTicker(time.Second * 50)
	defer tk.Stop()
	tk2 := time.NewTicker(time.Second * 10)
	defer tk2.Stop()
	for {
		select {
		case <-tk.C:
			t.Log("Test Done.")
			return
		case <-tk2.C:
			b, _ := json.Marshal(Instance(ctx).All())
			t.Log(string(b))
		}
	}
}

func TestBridge_RegisterZookeeperSource(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	properties := map[string]interface{}{
		"endpoint": "xxxx,xxxx",
		"path":     "xxxx",
		"timeout":  2000,
	}
	err := Instance(ctx).RegisterSource("zk_config", &Config{
		Provider:   provider.CfgProviderZookeeper,
		Properties: properties,
		ConfigType: codec.CfgFileTypeYaml,
	})
	if err != nil {
		t.Error(err)
		return
	}
	Instance(ctx).SetHook(func(key string, msg map[string]any) {
		b, _ := json.Marshal(msg)
		t.Logf("onchange: key %s msg = %s", key, string(b))
	})
	tk := time.NewTicker(time.Second * 50)
	defer tk.Stop()
	tk2 := time.NewTicker(time.Second * 10)
	defer tk2.Stop()
	for {
		select {
		case <-tk.C:
			t.Log("Test Done.")
			return
		case <-tk2.C:
			b, _ := json.Marshal(Instance(ctx).All())
			t.Log(string(b))
		}
	}
}

package provider

import (
	"context"
	"errors"
	"fmt"
	"github.com/fufuzion/confremote-pilot/codec"
	"github.com/go-zookeeper/zk"
	"github.com/thoas/go-funk"
	"golang.org/x/exp/maps"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type zookeeperProvider struct {
	ctx          context.Context
	tp           CfgProviderType
	conn         *zk.Conn
	stateCh      <-chan zk.Event
	o            *option
	codec        codec.Codec
	mu           *sync.RWMutex
	data         map[string]interface{}
	watchPath    string
	listenCancel context.CancelFunc
	reconnecting uint32
}

func newZookeeperProvider(ctx context.Context, o *option) (Provider, error) {
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
	var err error
	timeout := 5 * time.Second
	if timeoutVal, ok := o.properties["timeout"]; ok {
		timeout, err = parseTimeout(timeoutVal)
		if err != nil {
			return nil, err
		}
	}
	servers := strings.Split(endpoint, ",")
	conn, eventCh, err := zk.Connect(servers, timeout)
	if err != nil {
		return nil, fmt.Errorf("connect zookeeper failed: %w", err)
	}
	provider := &zookeeperProvider{
		ctx:       ctx,
		tp:        CfgProviderZookeeper,
		conn:      conn,
		stateCh:   eventCh,
		codec:     codec.NewCodec(o.configType),
		o:         o,
		data:      make(map[string]interface{}),
		mu:        &sync.RWMutex{},
		watchPath: path,
	}
	setting, err := provider.readRemote(path)
	if err != nil {
		return nil, err
	}
	provider.mu.Lock()
	provider.data = setting
	provider.mu.Unlock()

	provider.listen(path)
	provider.watchConnection()

	return provider, nil
}

func parseTimeout(val interface{}) (time.Duration, error) {
	switch v := val.(type) {
	case time.Duration:
		return v, nil
	case int:
		return time.Duration(v) * time.Millisecond, nil
	case int32:
		return time.Duration(v) * time.Millisecond, nil
	case int64:
		return time.Duration(v) * time.Millisecond, nil
	case float32:
		return time.Duration(v * float32(time.Millisecond)), nil
	case float64:
		return time.Duration(v * float64(time.Millisecond)), nil
	case string:
		return time.ParseDuration(v)
	default:
		return 0, fmt.Errorf("invalid timeout type: %T", val)
	}
}
func (p *zookeeperProvider) Name() string {
	return p.tp.ToString()
}
func (p *zookeeperProvider) Load() (map[string]interface{}, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	ret := make(map[string]interface{})
	maps.Copy(ret, p.data)
	return ret, nil
}
func (p *zookeeperProvider) readRemote(path string) (map[string]interface{}, error) {
	content, _, err := p.conn.Get(path)
	switch {
	case errors.Is(err, zk.ErrNoNode): // 支持空节点启动后再写入数据
		return make(map[string]interface{}), nil
	case err != nil:
		return nil, err
	}
	setting := make(map[string]interface{})
	err = p.codec.Decode(content, &setting)
	if err != nil {
		return nil, err
	}
	return setting, nil
}
func (p *zookeeperProvider) onChange(path string) {
	settings, err := p.readRemote(path)
	if err == nil {
		p.mu.Lock()
		p.data = settings
		p.mu.Unlock()
		p.notify(settings)
	}
}
func (p *zookeeperProvider) listen(path string) {
	if p.listenCancel != nil {
		p.listenCancel()
	}
	ctx, cancel := context.WithCancel(p.ctx)
	p.listenCancel = cancel
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				_, _, ch, err := p.conn.GetW(path)
				if err != nil {
					fmt.Println("zookeeperProvider.listen.GetW failed: ", err)
					time.Sleep(2 * time.Second)
					continue
				}
				select {
				case <-ctx.Done():
					return
				case ev, ok := <-ch:
					if !ok {
						time.Sleep(2 * time.Second)
						continue
					}
					switch ev.Type {
					case zk.EventNodeDataChanged, zk.EventNodeCreated, zk.EventNodeDeleted:
						p.onChange(path)
					}
				}
			}
		}
	}()
}

func (p *zookeeperProvider) notify(change map[string]interface{}) {
	if p.o.coordinator == nil {
		return
	}
	p.o.coordinator.Notify(p.o.customKey, change)
}
func (p *zookeeperProvider) watchConnection() {
	go func() {
		for {
			select {
			case <-p.ctx.Done():
				return
			default:
				stateCh := p.stateCh
				select {
				case <-p.ctx.Done():
					return
				case ev, ok := <-stateCh:
					if !ok {
						time.Sleep(time.Second)
						continue
					}
					if ev.State == zk.StateExpired || ev.State == zk.StateDisconnected || ev.State == zk.StateUnknown {
						time.Sleep(time.Second)
						p.reconnect()
					}
				}
			}
		}
	}()
}

func (p *zookeeperProvider) reconnect() {
	if !atomic.CompareAndSwapUint32(&p.reconnecting, 0, 1) {
		return
	}
	defer atomic.StoreUint32(&p.reconnecting, 0)

	endpoint, _ := p.o.properties["endpoint"].(string)
	servers := strings.Split(endpoint, ",")
	timeout := 5 * time.Second
	if timeoutVal, ok := p.o.properties["timeout"]; ok {
		if timeoutInt, ok := timeoutVal.(int); ok {
			timeout = time.Duration(timeoutInt) * time.Second
		}
	}
	conn, eventCh, err := zk.Connect(servers, timeout)
	if err != nil {
		return
	}
	oldConn := p.conn
	defer oldConn.Close()
	p.conn = conn
	p.stateCh = eventCh

	if settings, err := p.readRemote(p.watchPath); err == nil {
		p.mu.Lock()
		p.data = settings
		p.mu.Unlock()
	}

	p.listen(p.watchPath)
}

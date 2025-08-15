package provider

type Source struct {
	DataId string
	Group  string
}

type CfgProviderType string

func (p CfgProviderType) ToString() string { return string(p) }

const (
	CfgProviderNacos CfgProviderType = "nacos"
)

const (
	SchemeTypeHttp = "http"
)

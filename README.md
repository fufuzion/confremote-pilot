# Config Bridge

`Config Bridge` 是一个支持多配置源动态合并更新的Go库

## ✨ 特性

- 支持 **多Provider**（目前仅封装了nacos）
- 支持跨云平台/跨组件/跨配置文件类型，多命名空间、多组配置源组合使用
- 自动合并多源配置，感知配置增减
- 支持自定义 Hook 回调监听配置变化
- 本地使用 `viper` 管理统一配置视图

## 📦 安装

```bash
go get github.com/yourusername/config-bridge




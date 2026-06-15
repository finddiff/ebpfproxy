# ebpfproxy

<img src="https://github.com/daeuniverse/dae/blob/main/logo.png" border="0" width="25%">

**_ebpfproxy_** 是 [dae](https://github.com/daeuniverse/dae) 的增强 fork，在原有高性能透明代理基础上增加了强大的 **Web 管理面板**。

为了尽可能提升流量分流的性能，ebpfproxy 使用 eBPF 在 Linux 内核中实现透明代理和流量分流。因此，ebpfproxy 可以让直连流量绕过代理程序的转发，实现真正的直连通过。得益于这一机制，直连流量几乎没有性能损失，也几乎不消耗额外的系统资源。

作为 [v2rayA](https://github.com/v2rayA/v2rayA) 的后继者，dae 放弃了 v2ray-core，以更自由地满足用户需求。

## 功能特性

### 核心代理功能（来自 dae）

- [x] 实现 `Real Direct` 流量分流（需开启 ipforward），达到[高性能](https://docs.google.com/spreadsheets/d/1UaWU6nNho7edBNjNqC8dfGXLlW0-cm84MM7sH6Gp7UE/edit?usp=sharing)
- [x] 支持按本地进程名分流流量
- [x] 支持按局域网 MAC 地址分流流量
- [x] 支持反向匹配规则进行流量分流
- [x] 支持按策略自动切换节点。即自动测试 TCP/UDP/IPv4/IPv6 独立延迟，然后根据用户自定义策略为对应流量选择最优节点
- [x] 支持高级 DNS 解析流程
- [x] 支持 shadowsocks、trojan(-go) 和 socks5 的 full-cone NAT
- [x] 支持多种主流代理协议，详见 [proxy-protocols.md](./docs/en/proxy-protocols.md)

### Web 管理面板（ebpfproxy 新增）

- [x] **实时系统概览** - 通过 WebSocket 每 2 秒实时刷新 CPU、内存、系统负载、网络吞吐、连接数、UDP 会话
- [x] **连接监控** - 查看所有活跃 TCP/UDP 连接，包含协议、源/目标地址、传输速率、持续时间、域名解析。支持搜索、排序、分页以及关闭单个或全部连接
- [x] **代理组管理** - 查看代理组及每个节点的延迟、存活状态、协议检测。支持 `select` 策略组的节点手动切换
- [x] **路由规则查看** - 展示所有编译后的路由规则，包含匹配类型、出站目标、规则原文
- [x] **DNS 监控** - 查看 DNS 上游服务器和请求/响应路由规则
- [x] **可视化配置编辑器** - 基于表单的配置编辑器，支持可视化表单与 dae 配置语法之间的双向转换。支持编辑 Global、DNS、Subscription、Node、Group、Routing 各节
- [x] **实时日志流** - 所有守护进程日志通过 Logrus Hook 捕获并通过 WebSocket 推送到 Web UI，服务端环形缓冲区保留最近 2000 条日志
- [x] **硬件传感器** - 监控 CPU 温度/频率、GPU（NVIDIA 通过 nvidia-smi）、风扇转速、电压、磁盘 I/O、网卡统计，通过 hwmon/thermal zones/lm-sensors 采集
- [x] **DHCP 租约查看** - 解析 ISC DHCP 和 dnsmasq 租约文件，查看局域网设备信息
- [x] **Token 认证** - 通过可配置的 Token 保护 API 访问安全
- [x] **配置热重载** - 在 Web UI 中保存配置后触发在线热重载，无需重启服务
- [x] **Docker 支持** - 提供多阶段构建 Dockerfile 和 docker-compose.yml 便于容器化部署

## 快速开始

### 命令行使用

请参考 [快速入门指南](./docs/en/README.md) 立即开始使用。

### 启用 Web UI

在配置文件中添加以下内容：

```ini
global {
    webui_port: 8080
    webui_token: "your-secret-token"
}
```

或通过命令行参数指定：

```bash
dae run -c /etc/dae/config.dae --webui-port 8080
```

然后在浏览器中打开 `http://<your-host>:8080`。如果设置了 Token，需要在登录页面输入。

### Docker

```bash
docker-compose up -d
```

## 注意事项

1. 如果你在公网环境的同一台机器（如 VPS）上同时部署了 dae 和 shadowsocks 服务端（或任何 UDP 服务），请不要忘记为你的 UDP 服务端口添加 `l4proto(udp) && sport(你的服务端口) -> must_direct` 规则。因为 UDP 状态难以维护，所有出站 UDP 数据包都可能被代理（取决于你的路由配置），包括发往客户端的流量。这不是我们期望的行为。`must_direct` 会让来自该端口的所有流量（包括 DNS 流量）直连。
1. 如果中国大陆用户首次访问某些国内网站时发现首屏加载时间很长，请检查是否在 DNS 路由中使用海外 DNS 处理了某些国内域名。这有时很难发现。例如 `ocsp.digicert.cn` 意外地被包含在 `geosite:geolocation-!cn` 中，会导致某些 TLS 握手耗时很长。在 DNS 路由中使用此类域名集合时需格外注意。

## Web UI 页面

| 页面 | 说明 |
|------|------|
| **概览** | 系统资源仪表盘，展示 CPU、内存、负载、流量速率、连接数 |
| **连接** | 实时 TCP/UDP 连接列表，支持搜索、排序和关闭连接 |
| **代理** | 代理组/节点状态，延迟展示和手动选择 |
| **规则** | 编译后的路由规则展示 |
| **DNS** | DNS 上游和路由配置查看 |
| **配置** | 可视化配置编辑器，支持表单填写和原始文本预览 |
| **日志** | 实时日志流，支持按级别过滤 |
| **传感器** | 硬件传感器数据（CPU、GPU、风扇、电压、磁盘、网卡） |
| **DHCP** | DHCP 租约列表，展示局域网设备信息 |

## 工作原理

详见 [How it works](./docs/en/how-it-works.md)。

## TODO

- [ ] 自动检测 DNS 上游与来源环路（上游是否也是我们的客户端），提示用户添加 sip 规则
- [ ] MACv2 扩展提取
- [ ] 日志输出到用户空间
- [ ] 面向协议的节点特性检测（或过滤），如 full-cone（特别是 VMess 和 VLESS）
- [ ] 添加快速入门指南
- [ ] ...

## 许可证

[AGPL-3.0 (C) daeuniverse](https://github.com/daeuniverse/dae/blob/main/LICENSE)

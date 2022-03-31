# Getting Started

A command line interactive interface for [clash](https://github.com/Dreamacro/clash).

```bash
$ go get -u -v github.com/Dreamacro/clash-ctl
```

The binary is built under $GOPATH/bin

```bash
$ clash-ctl
```

## Usage

```bash
>>> server add
server name: clash
server address: 127.0.0.1
server port: 9090
server secret: 
API is HTTPS?[y/N]: 
write server success
```

The configuration file is stored in `$HOME/.config/clash/ctl.toml`

```toml
selected = "clash"

[servers]

  [servers.clash]
    host = "127.0.0.1"
    https = false
    port = "9090"
    secret = ""
```

Use `ping` command to check if the server is working.

```bash
>>> ping
clash                [success]
```

Use `proxy ls` to list all proxies groups.

```bash
>>> proxy ls
╭───────┬─────────────────────┬──────────┬────────────────────────────────╮
│ INDEX │ NAME                │ TYPE     │ NOW                            │
├───────┼─────────────────────┼──────────┼────────────────────────────────┤
│ 4fdd  │ 🚥 其他流量         │ Selector │ 🌐 国外流量                    │
│ abd6  │ 🎬 国际流媒体       │ Selector │ 🌐 国外流量                    │
│ 235f  │ 🏠 大陆流量         │ Selector │ ➡️ 直接连接                    │
│ db64  │ GLOBAL              │ Selector │ DIRECT                         │
│ e342  │ 🌐 国际网站         │ Selector │ 🌐 国外流量                    │
│ fc52  │ 🌐 国外流量         │ Selector │ 🇭🇰 香港 11 HKBN家宽（0.1倍率） │
│ dd1e  │ 🎬 大陆流媒体国际版 │ Selector │ ➡️ 直接连接                    │
│ c5b1  │ ➡️ 直接连接         │ Selector │ DIRECT                         │
│ 0ef6  │ 🎮 Steam            │ Selector │ ➡️ 直接连接                    │
│ 91d8  │ 🎬 大陆流媒体       │ Selector │ 🏠 大陆流量                    │
│ a1b3  │ 🏠 大陆网站         │ Selector │ 🏠 大陆流量                    │
╰───────┴─────────────────────┴──────────┴────────────────────────────────╯
>>> proxy ls fc52
╭───────┬────────────────────────────────────────┬──────────────┬───────╮
│ INDEX │ NAME                                   │ TYPE         │ DELAY │
├───────┼────────────────────────────────────────┼──────────────┼───────┤
│ 2fba  │ ♻️ 故障切换                            │ Fallback     │     0 │
│ c5b1  │ ➡️ 直接连接                            │ Selector     │     0 │
│ 9a6e  │ 🇯🇵 日本 05 NTT IPv6                    │ Vmess        │     0 │
│ 623e  │ 🇯🇵 日本 06 M247 IPv6                   │ Vmess        │  1106 │
│ 9ac2  │ 🇯🇵 日本 07 IIJ                         │ Vmess        │   418 │
│ 8c6c  │ 🇯🇵 日本 09 IIJ                         │ Vmess        │   421 │
│ 52b4  │ 🇯🇵 日本 11 IIJ IPv6                    │ Vmess        │   438 │
│ c737  │ 🇯🇵 日本 13 IIJ IPv6                    │ Vmess        │   406 │
│ 364d  │ 🇯🇵 日本 15 IIJ                         │ Vmess        │   412 │
│ 6218  │ 🇯🇵 日本 17 IIJ                         │ Vmess        │   440 │
│ 83ac  │ 🇯🇵 日本 19 IIJ                         │ Vmess        │   415 │
│ e5f7  │ 🇯🇵 日本 21 NTT                         │ Vmess        │   514 │
│ 7e60  │ 🇭🇰 香港 01 电信/深港专线（2.5倍率）    │ Vmess        │   100 │
│ bfa4  │ 🇭🇰 香港 02 电信/深港专线（2.5倍率）    │ Vmess        │     0 │
│ 87dd  │ 🇭🇰 香港 05 移动/深港专线（2倍率）      │ Vmess        │    88 │
│ b381  │ 🇭🇰 香港 06 移动/深港专线（2倍率）      │ Vmess        │    86 │
│ fc94  │ 🇭🇰 香港 09 HKT家宽（0.1倍率）          │ Vmess        │     0 │
│ cd34  │ 🇭🇰 香港 10 HKT家宽（0.1倍率）          │ Vmess        │     0 │
│ e4ec  │ 🇭🇰 香港 11 HKBN家宽（0.1倍率）         │ Vmess        │  1620 │
│ 0a03  │ 🇭🇰 香港 12 CN2 IPv6                    │ Vmess        │   211 │
╰───────┴────────────────────────────────────────┴──────────────┴───────╯
```

And you can use `proxy set group proxyName` to configure the proxy for the group

```bash
>>> proxy set fc52 0a03
Set proxy 🇭🇰 香港 12 CN2 IPv6 to group 🌐 国外流量
```

Of course `proxy delay` will test all the proxies in the group.

### What is index

Index is the firt four hex characters of [SHA1](https://en.wikipedia.org/wiki/SHA-1) of the name.
Why? Because usually the name of proxy group/node is Chinese characters, even emoji, which is hard to input in terminal.

Firt four hex is pretty enough to prevent duplication I think.

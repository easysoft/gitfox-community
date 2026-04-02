# `gitfox start` 核心服务注册指引

`gitfox start` 是 GitFox 的一键后台守护进程注册及启动指令。该设计旨在屏蔽不同操作系统运维之间的割裂感，让您可以跨平台“单行代码”将服务挂载进入机器的底层常驻管理机制（开机自启、崩溃自动重启）。

---

## 🏃 适用场景与边界
此指令仅适用于**生产部署或后台挂载运行**场景。当您通过 `gitfox init` 完成了交互式系统向导，并成功生成 `prod.env` / `prod.yaml` 后，即可调用此指令脱离前台运行。

## ⚠️ 前置要求与安全守则
1. **工作目录**：**必须**在与 `prod.env` 同级的目录中执行此指令，因为服务将被强绑定在该目录之上。
2. **系统级提权**：
   - **Linux**：必须赋予 ROOT 权限（例如使用 `sudo ./gitfox start`）。
   - **Windows**：必须使用**“管理员身份运行”**调出 Terminal，并且系统环境变量（`PATH`）中需提前准备好 `nssm.exe` 工具。
   - **macOS**：以当前普通用户态运行即可，直接挂载进 User LaunchAgents。
3. **环境前置探针**：服务器必须已经装载好了 `git` 及 `docker` 核心包，命令运行的第一秒会严格对这些包的版本与 PATH 抓取能力做检查。

---

## 💡 使用方式

不需要任何额外的标志参数，只需直接在存放应用的目录下敲击：

```bash
./gitfox start
```

### 幕后执行流
整个命令执行如丝般顺滑，但它底层一共接管了这几件事：
1. **底层强依赖检测**：扫描并汇报 `git` 与 `docker` 所处的版本信息。
2. **环境加载**：加载 `prod.env` 并嗅探所需通讯的 HTTP API 端口。
3. **分发注册表**：
   - 在 **Linux**, 直接向 `/etc/systemd/system/` 注入标准的 `gitfox.service`。
   - 在 **Windows**, 先执行清理策略防御冲突，随后唤醒 `nssm install` 等三连底座动作。
   - 在 **macOS**, 生成 `~/Library/LaunchAgents/com.gitfox.service.plist` XML 配置表并基于 `launchctl load` 拉起。 
4. **短路熔断与存活确认**：命令下发后不是立刻回弹，系统将在 15 秒内疯狂轮询刚刚嗅探的 API `/public/health` 通道，**确保系统真正从代码级活过来**后，才对屏幕爆出漂亮的拉起成功提示。

---

## 🔍 运维检查：如何确认挂载服务情况？

即便终端里回显了绿色的 "服务已成功启动"，后续您依然可以用对应系统的原生运维方式来查看及重置这支服务：

### 🐧 Linux (使用 Systemd)
最主流的运维操作体验：

```bash
# 查看实时的服务组件状态与内存占有（应显示 active (running)）：
sudo systemctl status gitfox

# 滚动排查由于 Panic 或运行时打出的后台标准错误日志：
sudo journalctl -u gitfox -f -n 100

# 临时停止或重启
sudo systemctl stop gitfox
sudo systemctl restart gitfox
```

### 🪟 Windows (使用 NSSM Native Service)
您可以打开系统最常见的 **"服务记录表"**：
1. 按 `Win + R` 并输入 `services.msc` 回车。
2. 滑动在列表中找到名为 **`gitfox`** 的服务项。
3. 观察它的“状态”列是否显示为 **“正在运行”**。在此处也可以方便地使用鼠标随时点击重启或停止。

*或者也可以通过纯终端状态确认:*
```powershell
nssm status gitfox
```

### 🍎 macOS (使用 LaunchDaemon)
通过原生的终端调度器 `launchctl` 去确认 PID 的存在：

```bash
# 确认进程编号是否存在
launchctl list | grep com.gitfox.service

# 当您不需要开启自启时去卸载服务
launchctl unload ~/Library/LaunchAgents/com.gitfox.service.plist
```

# 应用定时管家 (App Scheduler)

定时启动和停止懒猫微服上的应用，支持自定义调度计划和推送通知。

## ✨ 新功能：倒计时显示

每个定时任务卡片上现在会显示下次执行的倒计时，实时更新：

- ⏱️ **实时倒计时**：显示距离下次执行还有多长时间
- 🔄 **自动刷新**：每秒自动更新倒计时
- 🌐 **多语言支持**：中文、英文、日文
- 🎨 **视觉反馈**：即将执行的任务会有特殊动画效果

### 倒计时显示示例

```
📅 定时任务卡片
━━━━━━━━━━━━━━━━━━━━━━
任务名称：每日启动应用
应用：MyApp
⏱️ 下次执行：2小时35分12秒
⏰ 执行时间：09:30
▶️  操作：启动应用
📆 重复：一 二 三 四 五
━━━━━━━━━━━━━━━━━━━━━━
```

## ⚙️ 配置说明

### 时区设置

应用现在支持通过环境变量 `TZ` 来配置时区。在安装应用时，您可以在设置向导中指定时区。

**默认时区**：`Asia/Shanghai`（中国标准时间）

**常用时区示例**：
- `Asia/Shanghai` - 中国标准时间 (UTC+8)
- `Asia/Tokyo` - 日本标准时间 (UTC+9)
- `America/New_York` - 美国东部时间 (UTC-5/-4)
- `Europe/London` - 英国时间 (UTC+0/+1)
- `America/Los_Angeles` - 美国西部时间 (UTC-8/-7)

### 为什么需要时区设置？

时区设置确保：
1. ✅ 定时任务在正确的本地时间执行
2. ✅ 倒计时显示准确的剩余时间
3. ✅ 日志时间戳显示正确的本地时间
4. ✅ 多地区用户可以使用各自的时区

### 配置文件

应用包含以下配置文件：

1. **manifest.yml** - 应用清单文件
   - 定义应用的基本信息
   - 配置环境变量（包括 TZ）
   - 定义路由和服务

2. **lzc-deploy-params.yml** - 部署参数文件
   - 定义安装向导中的配置项
   - 支持多语言（中文、英文、日文）
   - 包含时区选择器

3. **lzc-build.yml** - 构建配置文件
   - 定义构建脚本
   - 指定内容目录和图标

## 🚀 安装

在懒猫应用商店安装时，系统会提示您配置时区：

```
┌─────────────────────────────────────┐
│  应用定时管家 - 安装向导              │
├─────────────────────────────────────┤
│  时区设置                            │
│  ┌───────────────────────────────┐  │
│  │ Asia/Shanghai              ▼ │  │
│  └───────────────────────────────┘  │
│                                     │
│  常用时区：                          │
│  • Asia/Shanghai (中国)             │
│  • Asia/Tokyo (日本)                │
│  • America/New_York (美国东部)       │
│  • Europe/London (英国)             │
└─────────────────────────────────────┘
```

## 📖 使用说明

1. **创建定时任务**：点击"新建任务"按钮
2. **配置任务**：
   - 任务名称：给任务起一个容易识别的名字
   - 选择应用：从已安装的应用中选择
   - 操作：选择"启动应用"或"停止应用"
   - 执行时间：设置小时和分钟
   - 重复日期：选择一周中的哪几天执行
3. **查看倒计时**：任务卡片会实时显示下次执行的倒计时
4. **管理任务**：可以编辑、删除或临时禁用任务

## 🔔 推送通知

支持通过 Server酱 发送推送通知：

1. 在设置中配置 Server酱 SendKey
2. 选择通知时机：成功时、失败时
3. 发送测试通知验证配置

## 🌍 多语言支持

界面支持三种语言：
- 🇨🇳 中文（简体）
- 🇬🇧 English
- 🇯🇵 日本語

## 📝 技术说明

### 前端技术栈
- 纯 JavaScript（无框架依赖）
- CSS3（支持深色/浅色主题）
- RemixIcon 图标库
- 多语言国际化支持

### 后端技术栈
- Go 语言
- Echo Web 框架
- Ent ORM
- SQLite 数据库
- gRPC（与懒猫 API 通信）

### 倒计时实现原理

```javascript
// 计算下次执行时间
function getNextExecutionTime(schedule) {
  const now = new Date();
  const currentDay = now.getDay();

  // 在未来7天内查找下一次执行
  for (let dayOffset = 0; dayOffset < 7; dayOffset++) {
    const checkDay = (currentDay + dayOffset) % 7;

    if (schedule.weekDays.includes(checkDay)) {
      const nextExecution = new Date(now);
      nextExecution.setDate(now.getDate() + dayOffset);
      nextExecution.setHours(schedule.hour, schedule.minute, 0, 0);

      // 如果是今天，检查时间是否已过
      if (dayOffset === 0 && nextExecution <= now) {
        continue;
      }

      return nextExecution;
    }
  }

  return null;
}

// 每秒更新倒计时
setInterval(updateCountdowns, 1000);
```

## 📦 开发

### 构建应用

```bash
# 执行构建脚本
./build.sh

# 或手动构建
lzc-cli project build
```

### 本地开发

```bash
# 启动开发环境
go run cmd/apps-scheduler/main.go
```

## 📄 许可证

MIT License

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

---

★ Insight ─────────────────────────────────────
**懒猫应用配置最佳实践**
1. **lzc-deploy-params.yml**：定义用户可配置的参数，使用小写 ID + 英文描述
2. **manifest.yml**：使用 `{{.U.参数名}}` 引用用户配置的参数
3. **多语言支持**：在 locales 中为每个参数提供翻译
4. **默认值**：为非必需参数提供合理的默认值（如 Asia/Shanghai）
5. **验证规则**：使用 regex 确保用户输入的格式正确
─────────────────────────────────────────────────
# apps-scheduler

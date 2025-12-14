# Tools CLI

一个基于 Cobra 框架开发的命令行工具，用于快速创建和管理 Go 和 .NET 项目。

## 功能特性

- 🚀 快速创建 Go 和 .NET 项目脚手架
- 📦 自动配置项目结构和依赖
- 🔧 内置项目初始化和运行命令
- 🌐 支持 SOCKS5 代理配置

## 安装

# 将 tools.exe 添加到 PATH 环境变量

## 命令列表

### 项目管理

#### 创建新项目

```bash
# 创建 Go 项目
tools project create go <项目名>

# 创建 .NET 项目（DDD 分层架构）
tools project create dotnet <项目名>
```

**Go 项目结构：**

```
项目名/
├── cmd/
│   └── main.go
├── internal/
├── pkg/
├── configs/
├── api/
└── go.mod
```

**\.NET 项目结构（DDD）：**

```
项目名/
├── src/
│   ├── 项目名.Domain/        # 领域层
│   ├── 项目名.Application/   # 应用层
│   ├── 项目名.Infrastructure/ # 基础设施层
│   ├── 项目名.Api/           # API层
│   └── 项目名.Shared/        # 共享层
└── 项目名.sln
```

#### 初始化现有项目

```bash
tools project init <项目名>
```

- 检测项目类型（Go 或 .NET）
- 生成 `toolsconfig.json` 配置文件

### 运行项目

#### 运行 Go 项目

```bash
tools run
```

#### 运行 .NET 项目

```bash
# 运行默认项目
tools run

# 指定项目运行（支持 Tab 补全）
tools run -t Api
tools run --hostType Api
```

### 构建项目

```bash
tools build
```

- 自动检测项目类型
- Go: 执行 `go build`
- .NET: 执行 `dotnet build`

### 代理配置

#### 查看代理配置

```bash
tools proxy
```

#### 设置 SOCKS5 代理

```bash
tools proxy set socks5://127.0.0.1:1080
```

#### 使用代理运行命令

```bash
# 示例：通过代理拉取 Git 代码
tools proxy run git pull

# 示例：通过代理安装 Go 依赖
tools proxy run go get github.com/spf13/cobra
```

## 配置文件

项目根目录的 `toolsconfig.json`：

```json
{
  "ProjectName": "MyProject",
  "BuildPath": "",
  "DotNetHosts": {
    "Api": "path/to/project.csproj"
  },
  "Proxy": {
    "Socks5": "socks5://127.0.0.1:1080"
  }
}
```

## 依赖包

- [github.com/spf13/cobra](https://github.com/spf13/cobra) - CLI 框架

## License

本项目采用 GNU General Public License v3.0 许可证，详见 [LICENSE](LICENSE) 文件。

# Tools CLI

一个基于 Cobra 开发的命令行工具，用于创建/初始化/运行/发布 Go 与 .NET 项目，并提供代理命令执行、HTTP 请求、Shell 补全等能力。

## 功能特性

- 快速创建 Go / .NET 项目脚手架
- 初始化现有项目并生成 `toolsconfig.json`
- 一键运行（Go / .NET）
- 多平台发布（Go: `linux/windows/darwin`；.NET: `linux/win/osx`）
- 构建 Docker 镜像（Go / .NET）
- 代理环境下运行任意命令（读取环境变量 `socks5`）
- HTTP 请求命令（支持自定义方法、Header、Body、输出文件）
- Shell 自动补全（PowerShell/Bash/Zsh）

## 安装

### 从源码编译

```bash
go build -o tools.exe
```

将可执行文件加入 PATH 后使用 `tools` 即可。

### 自动补全

- Windows PowerShell：执行 `tools create` / `tools init` 会自动安装补全（幂等）。
- 也可以手动生成：

```powershell
tools completion powershell > $HOME\tools-completion.ps1
. $HOME\tools-completion.ps1
```

## 命令参考

### 创建项目

```bash
tools create <go|dotnet> <项目名>
```

### 初始化项目

```bash
tools init <项目名>
```

会检测当前目录项目类型并生成/更新 `toolsconfig.json`。

### 运行项目

```bash
# 自动检测项目类型
tools run

# .NET 项目指定启动项（Key 来自 toolsconfig.json 的 DotNetHosts）
tools run -t <启动项名称>
tools run --hosttype <启动项名称>
```

### 发布项目（多平台）

```bash
# Go：按 toolsconfig.json 的 TargetPlatforms 交叉编译，返回当前系统对应可执行文件路径
tools publish

# .NET：按 TargetPlatforms 发布；需要指定启动项
tools publish -t <启动项名称>
tools publish --hosttype <启动项名称>
```

发布输出目录由 `PublishPath` 控制，工具会按 OS 创建子目录：

```text
<PublishPath>/
├── linux/
├── windows/
└── darwin/      (Go)

<DotNetHost.PublishPath>/
├── linux/
├── win/
└── osx/         (.NET)
```

如果.net 发布没有指定目标平台就会使用原始的发布命令方式

### 构建 Docker 镜像

```bash
tools build-image

# .NET 项目指定启动项
tools build-image -t <启动项名称>
```

### 代理管理

工具读取环境变量 `socks5`（例如 `socks5://127.0.0.1:1080`）。

```powershell
# PowerShell 设置环境变量
$env:socks5 = "socks5://127.0.0.1:1080"

# 输出用于设置 http_proxy / https_proxy 的 PowerShell 命令（复制执行）
tools proxy set

# 在代理环境下运行任意命令
tools proxy run git pull
tools proxy run go mod download
```

### HTTP 请求

```bash
tools http-req <url> \
  -X GET \
  -H "Accept: application/json" \
  -d '{"page":1,"limit":10}' \
  -o out.json
```

参数：

- `-X, --method`：HTTP 方法（默认 GET）
- `-H, --header`：请求头（可重复）
- `-d, --params`：请求参数/Body
- `-o, --output`：输出文件路径

### 生成补全脚本

```bash
tools completion [powershell|bash|zsh]
```

## 配置文件（toolsconfig.json）

### Go 项目示例

```json
{
  "ProjectName": "my-go-app",
  "Config": {
    "PublishPath": "./bin",
    "TargetPlatforms": [
      "linux/amd64",
      "windows/amd64",
      "darwin/amd64",
      "darwin/arm64"
    ]
  }
}
```

### .NET 项目示例

```json
{
  "ProjectName": "MyDotNetApp",
  "Config": {
    "DotNetHosts": {
      "MyDotNetApp.Api": {
        "CsProjectPath": "D:\\Projects\\MyApp\\src\\MyDotNetApp.Api\\MyDotNetApp.Api.csproj",
        "DockerfilePath": "D:\\Projects\\MyApp\\src\\MyDotNetApp.Api\\Dockerfile",
        "PublishPath": "D:\\Projects\\MyApp\\publish\\MyDotNetApp.Api",
        "TargetPlatform": "win-x64"
      }
    },
    "TargetPlatforms": ["linux-x64", "win-x64", "osx-x64", "osx-arm64"]
  }
}
```

### git 工作区整理

tools git refresh
因为目前自己电脑 windows 安装的最近版本的 git 对文件的修改和移动不明感，总是需要手动的使用 git st git rm 命令才行，
所以吧这打包成一个方便自己使用

## License

MIT License，详见 [LICENSE](LICENSE)。

### 方式 2：从源码编译

```bash
git clone <repository-url>
cd go-cli
go build -o tools.exe
```

# .NET 项目需要指定启动项（支持 Tab 补全）

tools publish -t <启动项名称>
tools publish --hosttype <启动项名称>
**PowerShell：**

````powershell
# 生成补全脚本
 - Go: `linux/amd64`, `windows/amd64`, `darwin/amd64`, `darwin/arm64`
 - .NET: `linux-x64`, `win-x64`, `osx-x64`, `osx-arm64`
# 在 PowerShell 配置文件中添加
notepad $PROFILE
# 添加以下内容：
. $HOME\tools-completion.ps1

# 重新加载配置
 <PublishPath>/
 ├── linux/
 │   └── <name>
 ├── windows/
 │   └── <name>.exe
 └── darwin/
     └── <name>
```bash
tools completion bash > ~/.tools-completion.bash
echo "source ~/.tools-completion.bash" >> ~/.bashrc
source ~/.bashrc
````

# 先设置环境变量 socks5（PowerShell 示例）

$env:socks5 = "socks5://127.0.0.1:1080"

# 查看并复制工具输出的设置命令

tools proxy set

## 📚 命令参考

### 项目管理

#### 创建新项目

```bash
# 创建 Go 项目
    "TargetPlatforms": [
      "linux/amd64",
      "windows/amd64",
      "darwin/amd64",
      "darwin/arm64"
    ]
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
│   ├── 项目名.Api/           # API 层
│   └── 项目名.Shared/        # 共享层
└── 项目名.sln
```

#### 初始化现有项目

```bash
tools project init
```

自动检测项目类型并生成配置文件 `toolsconfig.json`

### 运行项目

```bash
# 运行 Go 项目（自动检测）
tools run

# 运行 .NET 项目并指定启动项（支持 Tab 补全）
tools run -t <启动项名称>
tools run --hosttype <启动项名称>
```

### 发布项目（交叉编译）

```bash
# 自动检测项目类型并发布到配置的目标平台
tools publish

# .NET 项目指定启动项
tools publish -t Api
```

**支持的平台：**

- Go: `linux/amd64`, `windows/amd64`
- .NET: `linux-x64`, `win-x64`

**输出结构：**

```
bin/
├── linux/
│   └── myapp
└── windows/
    └── myapp.exe
```

### 构建 Docker 镜像

```bash
# 自动检测项目类型并构建 Docker 镜像
tools build-image

# .NET 项目指定启动项
tools build-image -t Api
```

### 代理管理

#### 查看代理配置

```bash
tools proxy
```

#### 设置 SOCKS5 代理

```bash
tools proxy set socks5://127.0.0.1:1080
```

#### 通过代理运行命令

```bash
# 通过代理拉取代码
tools proxy run git pull

# 通过代理安装依赖
tools proxy run go get github.com/spf13/cobra
tools proxy run dotnet restore
```

### HTTP 请求工具

```bash
# GET 请求
tools http-req https://api.example.com/data

# POST 请求
tools http-req https://api.example.com/data \
  -X POST \
  -H "Content-Type: application/json" \
  -d '{"name":"test"}'

# 添加多个请求头
tools http-req https://api.example.com \
  -H "Authorization: Bearer token123" \
  -H "User-Agent: MyTool/1.0"

# 保存响应到文件
tools http-req https://example.com -o output.html
```

**参数说明：**

- `-X, --method`: HTTP 方法（GET/POST/PUT/DELETE 等）
- `-H, --header`: 请求头（可多次使用）
- `-d, --params`: 请求参数/Body
- `-o, --output`: 保存响应到文件

## ⚙️ 配置文件

### Go 项目配置 (toolsconfig.json)

```json
{
  "ProjectName": "my-go-app",
  "Config": {
    "PublishPath": "./bin",
    "TargetPlatforms": ["linux/amd64", "windows/amd64"]
  }
}
```

### .NET 项目配置 (toolsconfig.json)

```json
{
  "ProjectName": "MyDotNetApp",
  "Config": {
    "DotNetHosts": {
      "Api": {
        "CsProjectPath": "D:\\Projects\\MyApp\\src\\MyApp.Api\\MyApp.Api.csproj",
        "DockerfilePath": "D:\\Projects\\MyApp\\src\\MyApp.Api\\Dockerfile",
        "PublishPath": "D:\\Projects\\MyApp\\publish\\Api"
      }
    },
    "TargetPlatforms": ["linux-x64", "win-x64"]
  }
}
```

## 🏗️ 技术架构

### 核心技术

- **CLI 框架**: [Cobra](https://github.com/spf13/cobra)
- **配置管理**: 基于 Go 泛型的类型安全配置系统
- **跨平台编译**: 利用 Go 和 .NET 的交叉编译能力
- **代理支持**: SOCKS5 代理集成

### 项目结构

```
go-cli/
├── cmd/                    # 命令行入口
├── config/                 # 配置管理（泛型实现）
│   └── config.go
├── internal/
│   ├── commands/           # Cobra 命令定义
│   │   ├── project.go      # 项目创建/初始化
│   │   ├── run.go          # 运行命令
│   │   ├── build.go        # 编译/发布
│   │   ├── proxy.go        # 代理管理
│   │   ├── http.go         # HTTP 工具
│   │   └── completion.go   # 自动补全
│   ├── dotnets/            # .NET 项目处理
│   ├── golangs/            # Go 项目处理
│   ├── networks/           # 网络工具
│   └── utils/              # 工具函数
├── pkg/                    # 公共包
└── main.go                 # 入口文件
```

## 💡 使用场景

### 1. 快速初始化新项目

```bash
# 创建并初始化 Go 微服务项目
tools project create go my-service
cd my-service
tools run
```

### 2. 多平台发布

```bash
# 一键编译 Linux 和 Windows 版本
tools publish

# 自动输出到 bin/linux 和 bin/windows 目录
```

### 3. 容器化部署

```bash
# 构建 Docker 镜像
tools build-image

# 推送到镜像仓库
docker push my-image:latest
```

### 4. 代理环境下开发

```bash
# 配置代理
tools proxy set socks5://127.0.0.1:7890

# 所有依赖拉取通过代理
tools proxy run go mod download
tools proxy run git clone https://github.com/...
```

## 📝 依赖包

- [github.com/spf13/cobra](https://github.com/spf13/cobra) - CLI 框架
- Go 标准库

## 📄 License

本项目采用 MIT License 许可证，详见 [LICENSE](LICENSE) 文件。

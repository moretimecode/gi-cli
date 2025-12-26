# Tools CLI

一个基于 Cobra 开发的多功能命令行工具，用于项目管理、代码分析、HTTP 请求、代理管理等。

## ✨ 功能特性

- 🚀 **项目管理**: 快速创建/初始化 Go 与 .NET 项目
- 🔍 **代码分析**:
  - 代码行数统计（自动忽略 .gitignore 配置）
  - 僵尸代码检测（Go & .NET）
- 🏗️ **构建发布**: 多平台交叉编译、Docker 镜像构建
- 🌐 **网络工具**: HTTP 请求、SOCKS5 代理管理
- 🔧 **开发工具**: Shell 自动补全、Git 工作区整理

## 📦 安装

### 从源码编译

```bash
git clone <repository-url>
cd go-cli
go build -o tools.exe
```

将可执行文件加入 PATH 后即可使用 `tools` 命令。

### 自动补全

**PowerShell:**

```powershell
tools completion powershell > $HOME\tools-completion.ps1
. $HOME\tools-completion.ps1
# 添加到配置文件实现永久生效
echo ". `$HOME\tools-completion.ps1" >> $PROFILE
```

**Bash:**

```bash
tools completion bash > ~/.tools-completion.bash
echo "source ~/.tools-completion.bash" >> ~/.bashrc
source ~/.bashrc
```

## 📚 命令参考

### 项目管理

#### 创建新项目

```bash
# 创建 Go 项目
tools create go <项目名>

# 创建 .NET 项目（DDD 分层架构）
tools create dotnet <项目名>
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
tools init <项目名>
```

自动检测项目类型并生成配置文件 `toolsconfig.json`

### 运行与发布

#### 运行项目

```bash
# 运行 Go 项目（自动检测）
tools run

# 运行 .NET 项目并指定启动项（支持 Tab 补全）
tools run -t <启动项名称>
tools run --hosttype <启动项名称>
```

#### 发布项目（交叉编译）

```bash
# 自动检测项目类型并发布到配置的目标平台
tools publish

# .NET 项目指定启动项
tools publish -t Api
```

**支持的平台：**

- Go: `linux/amd64`, `windows/amd64`, `darwin/amd64`, `darwin/arm64`
- .NET: `linux-x64`, `win-x64`, `osx-x64`, `osx-arm64`

**输出结构：**

```
bin/
├── linux/
│   └── myapp
├── windows/
│   └── myapp.exe
└── darwin/
    └── myapp
```

#### 构建 Docker 镜像

```bash
# 自动检测项目类型并构建 Docker 镜像
tools build-image

# .NET 项目指定启动项
tools build-image -t Api
```

### 代码分析

#### 代码统计

```bash
# 统计代码行数（自动忽略 .gitignore 配置）
tools code-summary

# 指定目录
tools code-summary <目录路径>
```

**输出信息：**

- 各文件代码行总数
- Top 10 最多行数文件
- Top 10 最少行数文件
- 按语言分类统计

#### 僵尸代码检测 🔍

检测 Go 和 .NET 项目中未使用的代码，包括：

- 未使用的函数/方法
- 未使用的变量和常量
- 未使用的类型定义
- 未使用的结构体字段
- 未使用的接口定义

```bash
# 检测当前目录（Go 项目）
tools deadcode

# 检测当前目录及所有子目录（Go 项目）
tools deadcode ./...

# 检测指定 .NET 解决方案
tools deadcode path/to/solution.sln

# 检测 .NET 项目文件
tools deadcode path/to/project.csproj

# 检测目录（自动查找 .sln 或 .csproj）
tools deadcode path/to/project-dir
```

**检测规则：**

Go 项目：

- ✅ 跳过 `main`、`init`、`Test*`、`Benchmark*`、`Example*` 函数
- ✅ 跳过自动生成文件（`wire_gen.go`、`*_gen.go`、`*.pb.go`、`mock_*`）
- ✅ 导出符号在 `/pkg/` 目录下不检查（库代码）

.NET 项目：

- ✅ 跳过 Controller、Middleware、Filter 中的成员
- ✅ 跳过接口定义和接口实现方法（如 `IHostedService.StartAsync`）
- ✅ 跳过带 Attribute 标注的成员（框架使用）
- ✅ 跳过 `Startup` 类的约定方法（`ConfigureServices`、`Configure`）
- ✅ 跳过 `override` 和 `virtual` 方法
- ✅ 支持多项目解决方案分析

**示例输出：**

```
🔍 开始检测僵尸代码: ./...
📦 检测到 Go 项目，正在加载...
🔬 正在分析代码...

============================================================
⚠️  发现 10 个未使用的符号:

  D:\Project\utils.go:107:6: 未使用的func 'MapSlice'
  D:\Project\utils.go:57:6: 未使用的func 'Select'
  D:\Project\handler.go:23:2: 未使用的var 'unusedVar'

总计: 10 个僵尸代码
============================================================
```

### 网络工具

#### HTTP 请求

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

#### 代理管理

```bash
# 设置 SOCKS5 代理环境变量
$env:socks5 = "socks5://127.0.0.1:1080"

# 查看代理设置命令（输出 http_proxy/https_proxy 设置语句）
tools proxy set

# 通过代理运行命令
tools proxy run git pull
tools proxy run go mod download
tools proxy run dotnet restore
```

### Git 工具

#### Git 工作区整理

```bash
tools git refresh
```

自动处理 Windows 下 Git 对文件修改和移动不敏感的问题，相当于执行：

- `git status`
- `git add .`
- `git rm` (对已删除的文件)

## ⚙️ 配置文件

### Go 项目配置 (toolsconfig.json)

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

### .NET 项目配置 (toolsconfig.json)

```json
{
  "ProjectName": "MyDotNetApp",
  "Config": {
    "DotNetHosts": {
      "Api": {
        "CsProjectPath": "D:\\Projects\\MyApp\\src\\MyApp.Api\\MyApp.Api.csproj",
        "DockerfilePath": "D:\\Projects\\MyApp\\src\\MyApp.Api\\Dockerfile",
        "PublishPath": "D:\\Projects\\MyApp\\publish\\Api",
        "TargetPlatform": "win-x64"
      }
    },
    "TargetPlatforms": ["linux-x64", "win-x64", "osx-x64", "osx-arm64"]
  }
}
```

## 🏗️ 技术架构

### 核心技术

- **CLI 框架**: [Cobra](https://github.com/spf13/cobra)
- **Go 分析**: golang.org/x/tools/go/analysis
- **.NET 分析**: Microsoft.CodeAnalysis (Roslyn)
- **配置管理**: 基于 Go 泛型的类型安全配置系统
- **跨平台编译**: 利用 Go 和 .NET 的交叉编译能力

### 项目结构

```
go-cli/
├── cmd/                    # 命令行入口
├── config/                 # 配置管理（泛型实现）
├── internal/
│   ├── commands/           # Cobra 命令定义
│   │   ├── deadcode.go     # 僵尸代码检测命令
│   │   ├── project.go      # 项目创建/初始化
│   │   ├── run.go          # 运行命令
│   │   ├── build.go        # 编译/发布
│   │   ├── proxy.go        # 代理管理
│   │   ├── http.go         # HTTP 工具
│   │   └── git.go          # Git 工具
│   ├── dotnets/            # .NET 项目处理
│   │   ├── deadcode.go     # .NET 僵尸代码检测
│   │   └── DeadCodeAnalyzer/ # C# Roslyn 分析器
│   ├── golangs/            # Go 项目处理
│   │   └── deadcode.go     # Go 僵尸代码检测
│   ├── statistics/         # 代码统计
│   ├── networks/           # 网络工具
│   └── utils/              # 工具函数
└── main.go                 # 入口文件
```

## 💡 使用场景

### 1. 代码质量提升

```bash
# 检测未使用代码
tools deadcode

# 查看代码统计
tools code-summary

# 清理僵尸代码后再次检测确认
```

### 2. 快速初始化新项目

```bash
# 创建并初始化 Go 微服务项目
tools create go my-service
cd my-service
tools run
```

### 3. 多平台发布

```bash
# 一键编译 Linux、Windows、macOS 版本
tools publish

# 自动输出到 bin/ 目录的各个平台子目录
```

### 4. 代理环境下开发

```bash
# 配置代理
$env:socks5 = "socks5://127.0.0.1:7890"

# 所有依赖拉取通过代理
tools proxy run go mod download
tools proxy run git clone https://github.com/...
```

## 📝 依赖说明

### Go 依赖

- [github.com/spf13/cobra](https://github.com/spf13/cobra) - CLI 框架
- golang.org/x/tools/go/packages - Go 代码分析
- golang.org/x/tools/go/analysis - AST 分析

### .NET 依赖（DeadCodeAnalyzer）

- Microsoft.CodeAnalysis.CSharp.Workspaces 5.0.0
- Microsoft.CodeAnalysis.Workspaces.MSBuild 5.0.0
- Microsoft.Build.Locator 1.11.2

## 🔧 开发指南

### 编译 .NET 分析器

```powershell
# 使用提供的脚本编译所有平台
.\build_dotnet.ps1

# 手动编译（Windows）
cd internal\dotnets\DeadCodeAnalyzer
dotnet publish -c Release -r win-x64 --self-contained true
```

### 运行测试

```bash
# 运行所有测试
go test ./... -v

# 运行带覆盖率的测试
go test ./... -v -cover
```

## 📄 License

本项目采用 MIT License 许可证，详见 [LICENSE](LICENSE) 文件。

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

### 5. 代码统计

```bash
# 代码统计总览，个文件代码行总数，列出最多的和最少的，自动忽略.gitignore里面配置的
tools code-summary
```

### 5. 僵尸代码检测

```bash
# 检测go 和dotnet 的僵尸代码，没有被直接引用的字段，属性，方法，接口，结构体
tools deadcode [路径]
```

## 📝 依赖包

- [github.com/spf13/cobra](https://github.com/spf13/cobra) - CLI 框架
- Go 标准库

## 📄 License

本项目采用 MIT License 许可证，详见 [LICENSE](LICENSE) 文件。

```

```

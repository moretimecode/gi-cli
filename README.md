# Tools CLI

多功能开发工具集合，包含项目管理、代码分析、Git 仓库分析、网络工具等。

## 安装

```bash
git clone <repository-url>
cd go-cli
go build -o tools.exe
```

### 自动补全

```powershell
# PowerShell
tools completion powershell > $HOME\tools-completion.ps1
echo ". `$HOME\tools-completion.ps1" >> $PROFILE
```

```bash
# Bash
tools completion bash > ~/.tools-completion.bash
echo "source ~/.tools-completion.bash" >> ~/.bashrc
```

## 命令参考

### 项目管理

#### 创建项目

```bash
# Go 项目
tools create go <项目名>

# .NET 项目（DDD 分层架构）
tools create dotnet <项目名>
```

**Go 项目结构：**

```
项目名/
├── cmd/main.go
├── internal/
├── pkg/
├── configs/
├── api/
└── go.mod
```

**\.NET 项目结构：**

```
项目名/
├── src/
│   ├── 项目名.Domain/
│   ├── 项目名.Application/
│   ├── 项目名.Infrastructure/
│   ├── 项目名.Api/
│   └── 项目名.Shared/
└── 项目名.sln
```

#### 初始化现有项目

```bash
tools init
```

自动检测项目类型并生成 `toolsconfig.json` 配置文件。

#### 运行项目

```bash
# Go 项目
tools run

# .NET 项目（需指定启动项，支持 Tab 补全）
tools run -t <启动项名称>
```

#### 发布项目

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
├── linux/myapp
├── windows/myapp.exe
└── darwin/myapp
```

#### 构建 Docker 镜像

```bash
# 构建镜像
tools build-image

# .NET 项目指定启动项
tools build-image -t Api
```

### 代码分析

#### 代码行数统计

```bash
# 统计当前目录（自动忽略 .gitignore 配置）
tools code-summary

# 指定目录
tools code-summary <目录路径>
```

**输出信息：**

- 各文件代码行总数
- Top 10 最多行数文件
- Top 10 最少行数文件
- 按语言分类统计

#### 僵尸代码检测

检测 Go 和 .NET 项目中未使用的代码。

```bash
# Go 项目
tools deadcode
tools deadcode ./...

# .NET 项目
tools deadcode path/to/solution.sln
tools deadcode path/to/project.csproj
tools deadcode path/to/project-dir
```

**检测内容：**

- 未使用的函数/方法
- 未使用的变量和常量
- 未使用的类型定义
- 未使用的结构体字段
- 未使用的接口定义

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

### Git 仓库分析

所有 Git 分析命令都支持 `--engine` 参数来选择仓储引擎：

```bash
# 自动选择（默认，优先使用原生 git）
tools git rank

# 强制使用原生 git 命令（速度更快）
tools git rank --engine native

# 强制使用 go-git 库（纯 Go 实现，无需安装 git）
tools git rank --engine go-git
```

#### 贡献者排行榜

```bash
# 按提交数排名
tools git rank

# 按代码量排名
tools git rank --type code

# 显示 Top 20
tools git rank --top 20

# 分析指定仓库
tools git rank /path/to/repo

# 使用 go-git 引擎分析
tools git rank --engine go-git
```

#### 夜猫子程序员

找出喜欢在深夜（22:00-06:00）提交代码的开发者。

```bash
tools git night-owl

# 设置夜间提交比例阈值
tools git night-owl --threshold 0.3
```

#### 查看某个作者的提交模式

分析指定开发者的提交时间分布（24 小时、星期、月度）。

```bash
tools git author <作者名>
tools git author /path/to/repo <作者名>
```

#### 作者热点文件分析

查看指定作者修改频率最高的文件。

```bash
# 查看作者的热点文件
tools git author-hotfile "张三"

# 指定仓库路径
tools git author-hotfile /path/to/repo "张三"

# 显示前 10 个文件
tools git author-hotfile "张三" --top 10
```

**输出内容：**

- 该作者修改最频繁的文件列表
- 每个文件的修改次数和总修改行数
- 汇总统计信息（总文件数、总修改次数、总修改行数）

#### 提交时间线分析

```bash
# 按小时统计
tools git timeline --view hour

# 按月统计
tools git timeline --view month
```

#### 热点文件分析

找出最频繁修改的文件，可能是代码热点或不稳定区域。

```bash
tools git hotfile

# 显示 Top 30
tools git hotfile --top 30
```

#### 代码稳定性分析

找出长时间未修改的稳定代码区域。

```bash
# 找出超过 180 天未修改的文件
tools git stability --days 180
```

#### 开发者协作分析

找出协作密切的开发者对（共同修改文件较多）。

```bash
# 显示 Top 10 协作关系
tools git collab

# 显示 Top 20 协作关系
tools git collab --top 20
```

#### 详细协作文件分析

查看特定作者之间共同修改的文件详情。

```bash
# 查看两个作者的协作文件
tools git collab-detail "张三" "李四"

# 查看多个作者的协作文件
tools git collab-detail "张三" "李四" "王五"

# 指定仓库路径
tools git collab-detail "张三" "李四" --path /path/to/repo
```

**输出内容：**

- 所有协作者共同修改的文件列表
- 每个作者在各文件的提交次数
- 文件总提交次数
- 汇总统计信息

#### Git 工作区整理

```bash
tools git refresh
```

自动处理 Windows 下 Git 对文件修改和移动不敏感的问题，执行：

- `git status`
- `git add .`
- `git rm`（对已删除的文件）

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
# 设置环境变量
$env:socks5 = "socks5://127.0.0.1:1080"

# 查看代理设置命令
tools proxy set

# 通过代理运行命令
tools proxy run git pull
tools proxy run go mod download
tools proxy run dotnet restore
```

## 配置文件

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

## License

MIT License，详见 [LICENSE](LICENSE)。

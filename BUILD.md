# 构建说明

## 工作流程

### 1. 当你修改了 .NET 代码后

运行 .NET 编译脚本（二选一）：

```bash
# 方式 1: 批处理脚本
build_dotnet.bat

# 方式 2: PowerShell 脚本（推荐，有彩色输出）
.\build_dotnet.ps1
```

如果 PowerShell 显示“无法加载文件，因为在此系统上禁止运行脚本”，请先执行一次：

```powershell
Set-ExecutionPolicy -Scope CurrentUser RemoteSigned
```

或者临时绕过执行策略：

```powershell
powershell.exe -ExecutionPolicy Bypass -File .\build_dotnet.ps1
```

这会：

- 编译 Windows、Linux、macOS 三个平台的 C# 分析器
- 自动复制到 `internal/dotnets/` 目录下
- 文件名：
  - `deadcode_windows_amd64.exe`
  - `deadcode_linux_amd64`
  - `deadcode_darwin_amd64`

### 2. 然后编译 Go 项目

使用你现有的构建脚本：

```bash
# 编译 Windows 版本
build.bat

# 或者发布所有平台
# 你的现有发布脚本...
```

## 交叉编译原理

Go 编译时会根据目标平台**自动选择**对应的 C# 分析器：

- 编译 Windows 版本 → 只嵌入 `deadcode_windows_amd64.exe`
- 编译 Linux 版本 → 只嵌入 `deadcode_linux_amd64`
- 编译 macOS 版本 → 只嵌入 `deadcode_darwin_amd64`

这是通过 **Go Build Tags** 实现的：

- `deadcode_windows.go` - Windows 专用
- `deadcode_linux.go` - Linux 专用
- `deadcode_darwin.go` - macOS 专用

## 文件说明

```
go-cli/
├── build_dotnet.bat          # .NET 编译脚本（批处理）
├── build_dotnet.ps1          # .NET 编译脚本（PowerShell）
├── build.bat                 # Go 编译脚本（你的现有脚本）
└── internal/
    └── dotnets/
        ├── deadcode.go                   # 主逻辑
        ├── deadcode_windows.go           # Windows embed
        ├── deadcode_linux.go             # Linux embed
        ├── deadcode_darwin.go            # macOS embed
        ├── deadcode_windows_amd64.exe    # Windows 二进制
        ├── deadcode_linux_amd64          # Linux 二进制
        └── deadcode_darwin_amd64         # macOS 二进制
```

## 注意事项

1. **首次使用前**必须运行一次 `build_dotnet.bat/ps1` 生成所有平台的 C# 分析器
2. 只有修改了 C# 代码才需要重新运行 `build_dotnet`
3. 每次修改 Go 代码只需运行 `build.bat` 即可

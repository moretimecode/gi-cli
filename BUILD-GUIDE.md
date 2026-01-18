# gi-cli 构建说明

## 📦 构建输出目录
- **容器内**: `/usr/local/bin/custom-cli/tools`
- **宿主机**: `D:\developerTools\cli\tools.exe`

## 🔨 构建方法

### Makefile (推荐)
```bash
cd gi-cli
make build          # 构建所有平台
make build-linux    # 仅 Linux
make build-windows  # 仅 Windows
```

### VS Code 快捷键
按 `Ctrl+Shift+B` 选择构建任务

### 直接命令
```bash
# Linux 版本
GOOS=linux GOARCH=amd64 go build -o /usr/local/bin/custom-cli/tools .

# Windows 版本  
GOOS=windows GOARCH=amd64 go build -o /usr/local/bin/custom-cli/tools.exe .
```

## 💡 使用说明
构建后在容器内直接运行 `tools` 命令即可，Windows 版本会自动同步到宿主机的 `D:\developerTools\cli` 目录。

package inputmethod

import (
	"fmt"
	"gocli/internal/utils"
	"os/exec"
	"path/filepath"
	"runtime"
)

const exeFileName = "WeaselDeployer.exe"

// RestartInputMethod 按操作系统选择不同的重启/重部署方式：
//   - Windows: 小狼毫(Weasel)，用 WeaselDeployer.exe /deploy
//   - Linux:   fcitx5-rime，通过 DBus 触发 Rime 重新部署
func RestartInputMethod() error {
	switch runtime.GOOS {
	case "windows":
		return restartWindows()
	case "linux":
		return restartLinuxFcitx5()
	default:
		return fmt.Errorf("暂不支持当前系统: %s（目前支持 Windows/小狼毫、Linux/fcitx5-rime）", runtime.GOOS)
	}
}

// restartWindows 小狼毫(Weasel)：在词库所在目录里找 WeaselDeployer.exe 并执行 /deploy
func restartWindows() error {
	path := filepath.Dir(term_resource)

	exePaths, err := utils.FileFilesByFileNameInDir(exeFileName, path)
	if err != nil {
		return fmt.Errorf("查找 %s 失败: %w", exeFileName, err)
	}
	if len(exePaths) == 0 {
		return fmt.Errorf("在 %s 目录下未找到 %s", path, exeFileName)
	}

	cmd := exec.Command(exePaths[0], "/deploy")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("deploy failed: %w, output: %s", err, string(out))
	}
	return nil
}

// restartLinuxFcitx5 fcitx5-rime：通过 DBus 调用 fcitx5 的 Rime 重新部署动作
func restartLinuxFcitx5() error {
	// 先确认 fcitx5 是否在运行
	if !fcitx5Running() {
		return fmt.Errorf("未检测到 fcitx5 在运行，无法重新部署 Rime")
	}

	deployed, err := rimeDeployViaDBus()
	if err == nil && deployed {
		return nil
	}
	if err != nil {
		// DBus 方式失败，退化为重载 fcitx5 配置（轻量部署，不一定重建词库）
		cmd := exec.Command("fcitx5-remote", "-r")
		if out, runErr := cmd.CombinedOutput(); runErr != nil {
			return fmt.Errorf("fcitx5 rime 重新部署失败: %v, output: %s", runErr, string(out))
		}
		return nil
	}
	return nil
}

// fcitx5Running 通过 fcitx5-remote --check 判断 fcitx5 是否在运行
func fcitx5Running() bool {
	err := exec.Command("fcitx5-remote", "--check").Run()
	return err == nil
}

// rimeDeployViaDBus 调用 fcitx5 的 org.fcitx.Fcitx.Controller1.SetConfig
// 触发 Rime 全量部署（等价于 fcitx5 配置工具里的「重新部署」按钮）。
// 优先使用 gdbus，找不到再退回 dbus-send。
func rimeDeployViaDBus() (bool, error) {
	if _, err := exec.LookPath("gdbus"); err == nil {
		cmd := exec.Command("gdbus", "call", "--session", "--dest", "org.fcitx.Fcitx5",
			"--object-path", "/controller",
			"--method", "org.fcitx.Fcitx.Controller1.SetConfig",
			"fcitx://config/addon/rime/deploy", "<''>")
		if out, err := cmd.CombinedOutput(); err == nil {
			return true, nil
		} else {
			return false, fmt.Errorf("gdbus 调用失败: %v, output: %s", err, string(out))
		}
	}

	if _, err := exec.LookPath("dbus-send"); err == nil {
		cmd := exec.Command("dbus-send", "--session", "--dest=org.fcitx.Fcitx5", "--type=method_call",
			"/controller", "org.fcitx.Fcitx.Controller1.SetConfig",
			"string:fcitx://config/addon/rime/deploy", "variant:string:")
		if out, err := cmd.CombinedOutput(); err == nil {
			return true, nil
		} else {
			return false, fmt.Errorf("dbus-send 调用失败: %v, output: %s", err, string(out))
		}
	}

	return false, fmt.Errorf("未找到 gdbus 或 dbus-send，无法通过 DBus 重新部署 Rime")
}

package inputmethod

import (
	"fmt"
	"gocli/internal/utils"
	"os/exec"
	"path/filepath"
)

const exeFileName = "WeaselDeployer.exe"

func RestartInputMethod() error {
	path := filepath.Dir(term_resource)

	exePaths, error := utils.FileFilesByFileNameInDir(exeFileName, path)
	if error != nil {
		return error
	}

	cmd := exec.Command(exePaths[0], "/deploy")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("deploy failed: %w, output: %s", err, string(out))
	}
	return nil

}

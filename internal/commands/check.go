package commands

import (
	"strings"

	"gocli/internal/dotnets"
	"gocli/internal/golangs"
	"gocli/internal/utils"
)

type ProjectType int

const (
	ProjectTypeUnknown ProjectType = 0
	ProjectTypeGo      ProjectType = 1
	ProjectTypeDotNet  ProjectType = 2
)

var ProjectTypeMap = map[string]ProjectType{
	"go":     ProjectTypeGo,
	"dotnet": ProjectTypeDotNet,
}

func ParseProjectType(typeStr string) (ProjectType, bool) {
	pt, ok := ProjectTypeMap[strings.ToLower(typeStr)]
	return pt, ok
}

type ProjectInfo struct {
	Type          ProjectType
	MainFilePaths []string
}

func DetectProjectType() *ProjectInfo {
	info := &ProjectInfo{
		Type: ProjectTypeUnknown,
	}

	isGolangProject, goMainFiles := utils.FindFileInCurrentAndSubDirs(golangs.GoDefaultMainFile)
	if isGolangProject {
		info.Type = ProjectTypeGo
		info.MainFilePaths = goMainFiles
		return info
	}

	isDotNetProject, dotnetFiles := utils.FindFileInCurrentAndSubDirs(dotnets.DotNetDefaultProjectFile)
	if isDotNetProject {
		info.Type = ProjectTypeDotNet
		info.MainFilePaths = dotnetFiles
		return info
	}

	return info
}

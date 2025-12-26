package golangs

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"
	"unicode"

	"golang.org/x/tools/go/packages"
)

// symbolInfo 存储符号信息
type symbolInfo struct {
	name    string
	pos     token.Position // 符号位置
	kind    string         // "func", "var", "const", "type", "field"
	pkgPath string         // 所在包的路径
}

type DeadCodeDetector struct {
	// 所有定义的符号，key = "包路径.符号名"
	defined map[string]*symbolInfo

	// 所有被使用的符号名
	used map[string]bool

	// 包信息
	pkgs []*packages.Package
}

func NewDeadCodeDetector() *DeadCodeDetector {
	return &DeadCodeDetector{
		defined: make(map[string]*symbolInfo),
		used:    make(map[string]bool),
	}
}

func (d *DeadCodeDetector) LoadProject(pattern string) error {
	// pattern 通常是 "./..." 表示当前目录及所有子目录
	cfg := &packages.Config{
		Mode: packages.NeedName |
			packages.NeedFiles |
			packages.NeedSyntax |
			packages.NeedTypes |
			packages.NeedTypesInfo,
	}

	pkgs, err := packages.Load(cfg, pattern)
	if err != nil {
		return fmt.Errorf("加载项目失败: %w", err)
	}

	// 检查是否有错误
	for _, pkg := range pkgs {
		if len(pkg.Errors) > 0 {
			return fmt.Errorf("包 %s 有错误: %v", pkg.PkgPath, pkg.Errors)
		}
	}

	d.pkgs = pkgs
	return nil
}

func (d *DeadCodeDetector) Analyze() error {
	d.collectDefinitions()
	d.collectUsages()
	return nil
}

// collectDefinitions 收集所有定义
func (detector *DeadCodeDetector) collectDefinitions() {
	for _, pkg := range detector.pkgs {
		isLibrary := isLibrarySymbol(pkg.PkgPath)
		for _, file := range pkg.Syntax {
			// 跳过自动生成的文件
			if shouldSkipFile(pkg.Fset.Position(file.Pos()).Filename) {
				continue
			}

			for _, decl := range file.Decls {
				switch d := decl.(type) {
				case *ast.FuncDecl:
					funcName := d.Name.Name
					if shouldSkip(funcName) {
						continue
					}

					if d.Recv != nil {
						continue
					}

					if isLibrary && isExported(funcName) {
						continue
					}

					key := makeKey(pkg.PkgPath, funcName)
					detector.defined[key] = &symbolInfo{
						name:    funcName,
						pos:     pkg.Fset.Position(d.Name.Pos()), //包含文件名、行号
						kind:    "func",
						pkgPath: pkg.PkgPath,
					}
				case *ast.GenDecl:
					// 1. 跳过 import 声明
					if d.Tok == token.IMPORT {
						continue
					}

					for _, spec := range d.Specs {
						{
							switch s := spec.(type) {
							case *ast.ValueSpec:
								for _, name := range s.Names {
									if name.Name == "_" {
										continue
									}
									// 库的导出符号不检查 供外部使用的变量或者常量
									if isLibrary && isExported(name.Name) {
										continue
									}

									kind := "var"
									if d.Tok == token.CONST {
										kind = "const"
									}

									key := makeKey(pkg.PkgPath, name.Name)
									detector.defined[key] = &symbolInfo{
										name:    name.Name,
										pos:     pkg.Fset.Position(name.Pos()),
										kind:    kind,
										pkgPath: pkg.PkgPath,
									}
								}

							case *ast.TypeSpec:
								typeName := s.Name.Name

								if typeName == "_" {
									continue
								}

								if isLibrary && isExported(typeName) {
									continue
								}

								key := makeKey(pkg.PkgPath, typeName)
								detector.defined[key] = &symbolInfo{
									name:    typeName,
									pos:     pkg.Fset.Position(s.Name.Pos()),
									kind:    "type",
									pkgPath: pkg.PkgPath,
								}
							}
						}
					}
				}
			}
		}
	}
}

// collectUsages 收集所有使用
func (d *DeadCodeDetector) collectUsages() {
	for _, pkg := range d.pkgs {
		for _, obj := range pkg.TypesInfo.Uses {
			if obj.Pkg() == nil {
				continue
			}
			pkgPath := obj.Pkg().Path()
			name := obj.Name()
			key := makeKey(pkgPath, name)
			d.used[key] = true
		}
	}
}

// Report 报告僵尸代码
func (d *DeadCodeDetector) Report() []string {
	var results []string

	for key, info := range d.defined {
		if !d.used[key] {
			results = append(results,
				fmt.Sprintf("%s: 未使用的%s '%s'",
					info.pos, info.kind, info.name))
		}
	}

	return results
}

// makeKey 生成唯一标识符
func makeKey(pkgPath, name string) string {
	return pkgPath + "." + name
}

// isExported 判断是否导出
func isExported(name string) bool {
	if len(name) == 0 {
		return false
	}
	runes := []rune(name)
	return unicode.IsUpper(runes[0])
}

func shouldSkip(name string) bool {
	return name == "main" || name == "init" || strings.HasPrefix(name, "Test") ||
		strings.HasPrefix(name, "Benchmark") || strings.HasPrefix(name, "Example")
}

func isLibrarySymbol(pkgPath string) bool {
	return strings.Contains(pkgPath, "/pkg/")
}

// shouldSkipFile 判断是否应该跳过某些文件（自动生成的文件）
func shouldSkipFile(filename string) bool {
	// 获取文件名（不含路径）
	parts := strings.Split(filename, "\\")
	if len(parts) == 0 {
		parts = strings.Split(filename, "/")
	}
	name := parts[len(parts)-1]

	// 跳过的文件模式
	skipPatterns := []string{
		"wire_gen.go", // Wire 依赖注入生成
		"_gen.go",     // 通用生成文件
		".pb.go",      // Protocol Buffers 生成
		".gen.go",     // 通用生成文件
		"mock_",       // Mock 文件
		"_mock.go",    // Mock 文件
	}

	for _, pattern := range skipPatterns {
		if strings.Contains(name, pattern) {
			return true
		}
	}

	return false
}

// +build mage

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

const (
	appName    = "neverbroke"
	mainPath   = "./cmd/server"
	binDir     = "./bin"
	version    = "v0.1.0"
)

var (
	// Default target when running `mage` without arguments
	Default = Build
)

// Build 编译二进制文件
func Build() error {
	mg.Deps(Install)
	
	fmt.Println("Building", appName, "...")
	
	// 创建 bin 目录
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return fmt.Errorf("failed to create bin directory: %w", err)
	}
	
	// 获取构建信息
	buildTime := time.Now().Format(time.RFC3339)
	gitCommit := getGitCommit()
	
	// 构建标志
	ldflags := fmt.Sprintf(
		"-X main.version=%s -X main.buildTime=%s -X main.gitCommit=%s",
		version, buildTime, gitCommit,
	)
	
	// 构建当前平台
	binaryName := appName
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}
	
	outputPath := filepath.Join(binDir, binaryName)
	
	args := []string{
		"build",
		"-ldflags", ldflags,
		"-o", outputPath,
		mainPath,
	}
	
	if err := sh.Run("go", args...); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}
	
	fmt.Printf("✅ Build successful: %s\n", outputPath)
	return nil
}

// BuildAll 为多个平台交叉编译
func BuildAll() error {
	mg.Deps(Install)
	
	fmt.Println("Building for all platforms...")
	
	platforms := []struct {
		os   string
		arch string
	}{
		{"linux", "amd64"},
		{"linux", "arm64"},
		{"darwin", "amd64"},
		{"darwin", "arm64"},
		{"windows", "amd64"},
	}
	
	buildTime := time.Now().Format(time.RFC3339)
	gitCommit := getGitCommit()
	
	for _, platform := range platforms {
		if err := buildForPlatform(platform.os, platform.arch, buildTime, gitCommit); err != nil {
			return err
		}
	}
	
	fmt.Println("✅ All platforms built successfully")
	return nil
}

// buildForPlatform 为指定平台编译
func buildForPlatform(goos, goarch, buildTime, gitCommit string) error {
	binaryName := appName
	if goos == "windows" {
		binaryName += ".exe"
	}
	
	outputName := fmt.Sprintf("%s-%s-%s", appName, goos, goarch)
	if goos == "windows" {
		outputName += ".exe"
	}
	
	outputPath := filepath.Join(binDir, outputName)
	
	ldflags := fmt.Sprintf(
		"-X main.version=%s -X main.buildTime=%s -X main.gitCommit=%s",
		version, buildTime, gitCommit,
	)
	
	env := map[string]string{
		"GOOS":   goos,
		"GOARCH": goarch,
	}
	
	args := []string{
		"build",
		"-ldflags", ldflags,
		"-o", outputPath,
		mainPath,
	}
	
	fmt.Printf("Building %s/%s...\n", goos, goarch)
	if err := sh.RunWith(env, "go", args...); err != nil {
		return fmt.Errorf("build failed for %s/%s: %w", goos, goarch, err)
	}
	
	fmt.Printf("✅ Built: %s\n", outputPath)
	return nil
}

// Run 运行开发服务器
func Run() error {
	fmt.Println("Running development server...")
	return sh.Run("go", "run", mainPath)
}

// Test 运行测试
func Test() error {
	fmt.Println("Running tests...")
	return sh.Run("go", "test", "./...")
}

// TestCoverage 运行测试并生成覆盖率报告
func TestCoverage() error {
	fmt.Println("Running tests with coverage...")
	
	coverageFile := "coverage.out"
	
	// 运行测试并生成覆盖率
	if err := sh.Run("go", "test", "-coverprofile", coverageFile, "./..."); err != nil {
		return err
	}
	
	// 显示覆盖率
	fmt.Println("\nCoverage report:")
	return sh.Run("go", "tool", "cover", "-func", coverageFile)
}

// TestCoverageHTML 生成 HTML 格式的覆盖率报告
func TestCoverageHTML() error {
	mg.Deps(TestCoverage)
	
	fmt.Println("Generating HTML coverage report...")
	coverageFile := "coverage.out"
	htmlFile := "coverage.html"
	
	return sh.Run("go", "tool", "cover", "-html", coverageFile, "-o", htmlFile)
}

// Clean 清理构建产物
func Clean() error {
	fmt.Println("Cleaning build artifacts...")
	
	// 删除 bin 目录
	if err := sh.Rm(binDir); err != nil {
		fmt.Printf("Warning: failed to remove %s: %v\n", binDir, err)
	}
	
	// 删除覆盖率文件
	coverageFiles := []string{"coverage.out", "coverage.html"}
	for _, file := range coverageFiles {
		if err := os.Remove(file); err != nil && !os.IsNotExist(err) {
			fmt.Printf("Warning: failed to remove %s: %v\n", file, err)
		}
	}
	
	// 清理测试二进制文件
	if err := sh.Run("go", "clean", "-testcache"); err != nil {
		fmt.Printf("Warning: failed to clean test cache: %v\n", err)
	}
	
	fmt.Println("✅ Clean completed")
	return nil
}

// Install 安装依赖
func Install() error {
	fmt.Println("Installing dependencies...")
	
	if err := sh.Run("go", "mod", "download"); err != nil {
		return fmt.Errorf("failed to download dependencies: %w", err)
	}
	
	if err := sh.Run("go", "mod", "tidy"); err != nil {
		return fmt.Errorf("failed to tidy dependencies: %w", err)
	}
	
	fmt.Println("✅ Dependencies installed")
	return nil
}

// Format 格式化代码
func Format() error {
	fmt.Println("Formatting code...")
	
	// 运行 gofmt
	if err := sh.Run("gofmt", "-s", "-w", "."); err != nil {
		return fmt.Errorf("gofmt failed: %w", err)
	}
	
	// 尝试运行 goimports（如果可用）
	if err := sh.Run("goimports", "-w", "."); err != nil {
		// goimports 可能未安装，这不是致命错误
		fmt.Println("Note: goimports not found, skipping import formatting")
	}
	
	fmt.Println("✅ Code formatted")
	return nil
}

// Lint 运行代码检查（需要安装 golangci-lint）
func Lint() error {
	fmt.Println("Running linter...")
	
	// 检查 golangci-lint 是否安装
	if _, err := exec.LookPath("golangci-lint"); err != nil {
		return fmt.Errorf("golangci-lint not found. Install it from https://golangci-lint.run/usage/install/")
	}
	
	return sh.Run("golangci-lint", "run", "./...")
}

// Vet 运行 go vet
func Vet() error {
	fmt.Println("Running go vet...")
	return sh.Run("go", "vet", "./...")
}

// getGitCommit 获取当前 git commit hash
func getGitCommit() string {
	commit, err := sh.Output("git", "rev-parse", "--short", "HEAD")
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(commit)
}

// Version 显示版本信息
func Version() error {
	fmt.Printf("Version: %s\n", version)
	fmt.Printf("Git Commit: %s\n", getGitCommit())
	fmt.Printf("Go Version: %s\n", runtime.Version())
	return nil
}


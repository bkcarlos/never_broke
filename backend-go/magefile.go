//go:build mage
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
	appName  = "neverbroke"
	mainPath = "./cmd/server"
	binDir   = "./bin"
	version  = "v0.1.0"
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

// Test namespace for all test-related commands
type Test mg.Namespace

// Unit 运行单元测试
func (Test) Unit() error {
	fmt.Println("🧪 Running unit tests...")
	return sh.RunV("go", "test", "./internal/...", "-v", "-short")
}

// Integration 运行集成测试
func (Test) Integration() error {
	fmt.Println("🧪 Running integration tests...")
	return sh.RunV("go", "test", "./tests/integration/...", "-v", "-timeout", "10m")
}

// IntegrationQuick 快速运行集成测试（不显示详细输出）
func (Test) IntegrationQuick() error {
	fmt.Println("🧪 Running integration tests (quick)...")
	return sh.RunV("go", "test", "./tests/integration/...", "-timeout", "10m")
}

// All 运行所有测试（单元测试 + 集成测试）
func (Test) All() error {
	fmt.Println("🧪 Running all tests...")
	fmt.Println("")
	fmt.Println("1/2: Unit Tests")
	if err := sh.RunV("go", "test", "./internal/...", "-v", "-short"); err != nil {
		return err
	}

	fmt.Println("")
	fmt.Println("2/2: Integration Tests")
	return sh.RunV("go", "test", "./tests/integration/...", "-v", "-timeout", "10m")
}

// Coverage 运行测试并生成覆盖率报告
func (Test) Coverage() error {
	fmt.Println("📊 Running tests with coverage...")

	// 单元测试覆盖率
	if err := sh.RunV("go", "test", "./internal/...", "-cover", "-coverprofile=coverage_unit.out"); err != nil {
		return err
	}

	// 集成测试覆盖率
	if err := sh.RunV("go", "test", "./tests/integration/...", "-cover", "-coverprofile=coverage_integration.out", "-timeout", "10m"); err != nil {
		return err
	}

	fmt.Println("")
	fmt.Println("✅ Coverage reports generated:")
	fmt.Println("  - Unit tests: coverage_unit.out")
	fmt.Println("  - Integration tests: coverage_integration.out")
	fmt.Println("")
	fmt.Println("To view HTML report:")
	fmt.Println("  go tool cover -html=coverage_unit.out")
	fmt.Println("  go tool cover -html=coverage_integration.out -o coverage_integration.html")

	return nil
}

// CoverageHTML 生成 HTML 格式的覆盖率报告
func (Test) CoverageHTML() error {
	mg.Deps(Test.Coverage)

	fmt.Println("📊 Generating HTML coverage report...")
	if err := sh.RunV("go", "tool", "cover", "-html=coverage_unit.out", "-o", "coverage_unit.html"); err != nil {
		return err
	}

	fmt.Println("✅ HTML coverage report generated: coverage_unit.html")
	return nil
}

// Race 运行测试并启用竞态检测
func (Test) Race() error {
	fmt.Println("🏃 Running tests with race detector...")

	// 单元测试竞态检测
	if err := sh.RunV("go", "test", "./internal/...", "-race", "-short"); err != nil {
		return err
	}

	// 集成测试竞态检测
	return sh.RunV("go", "test", "./tests/integration/...", "-race", "-timeout", "15m")
}

// Auth 只运行认证相关测试
func (Test) Auth() error {
	fmt.Println("🔐 Running auth tests...")

	// 单元测试
	if err := sh.RunV("go", "test", "./internal/auth/...", "-v"); err != nil {
		return err
	}

	// 集成测试
	return sh.RunV("go", "test", "./tests/integration/...", "-v", "-run", "TestAuthFlow", "-timeout", "10m")
}

// Ledger 只运行账本相关测试
func (Test) Ledger() error {
	fmt.Println("📝 Running ledger tests...")

	// 单元测试
	if err := sh.RunV("go", "test", "./internal/ledger/...", "-v"); err != nil {
		return err
	}

	// 集成测试
	return sh.RunV("go", "test", "./tests/integration/...", "-v", "-run", "TestTransactionFlow", "-timeout", "10m")
}

// Concurrent 运行多用户并发测试
func (Test) Concurrent() error {
	fmt.Println("👥 Running concurrent users tests...")
	return sh.RunV("go", "test", "./tests/integration/...", "-v", "-run", "TestMultiUser", "-timeout", "10m")
}

// Stress 运行压力测试和真实场景测试
func (Test) Stress() error {
	fmt.Println("🔥 Running stress and real-world scenario tests...")
	return sh.RunV("go", "test", "./tests/integration/...", "-v", "-run", "TestRealWorldScenario", "-timeout", "15m")
}

// Benchmark 运行性能基准测试
func (Test) Benchmark() error {
	fmt.Println("⚡ Running benchmark tests...")
	return sh.RunV("go", "test", "./internal/...", "-bench=.", "-benchmem")
}

// SetupDB 设置测试数据库
func (Test) SetupDB() error {
	fmt.Println("🔧 Setting up test database...")

	// 创建测试数据库
	fmt.Println("Creating test database...")
	if err := sh.Run("psql", "-U", "postgres", "-c", "CREATE DATABASE IF NOT EXISTS neverbroke_test;"); err != nil {
		// 如果失败，尝试先删除再创建
		sh.Run("psql", "-U", "postgres", "-c", "DROP DATABASE IF EXISTS neverbroke_test;")
		if err := sh.Run("psql", "-U", "postgres", "-c", "CREATE DATABASE neverbroke_test;"); err != nil {
			return fmt.Errorf("failed to create test database: %w", err)
		}
	}

	// 创建 .env.test 文件
	envContent := `DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=
DB_NAME=neverbroke_test
DB_SSLMODE=disable

JWT_SECRET=test-secret-key-for-testing-only
`
	envPath := "tests/integration/.env.test"
	if err := os.WriteFile(envPath, []byte(envContent), 0644); err != nil {
		return fmt.Errorf("failed to create .env.test: %w", err)
	}

	fmt.Println("✅ Test database setup completed")
	fmt.Println("  - Database: neverbroke_test")
	fmt.Println("  - Config: tests/integration/.env.test")

	return nil
}

// CleanDB 清理测试数据库
func (Test) CleanDB() error {
	fmt.Println("🧹 Cleaning test database...")

	// 使用 psql 删除测试数据库
	if err := sh.Run("psql", "-U", "postgres", "-c", "DROP DATABASE IF EXISTS neverbroke_test;"); err != nil {
		fmt.Printf("Warning: failed to drop test database: %v\n", err)
	}

	fmt.Println("✅ Test database cleaned")
	return nil
}

// Demo 演示所有复杂测试场景
func (Test) Demo() error {
	fmt.Println("🎭 Running test scenarios demo...")
	fmt.Println("")

	scenarios := []struct {
		name    string
		pattern string
		desc    string
	}{
		{"同一用户多设备登录", "TestMultiUser_SameUserDifferentDevices", "5个设备同时登录"},
		{"多设备并发操作", "TestMultiUser_SameUserDifferentDevices/MultiDevicesConcurrentOperations", "9笔并发交易"},
		{"家庭共享账户", "TestRealWorldScenario_FamilySharedAccount", "夫妻同时消费"},
		{"企业多部门运营", "TestRealWorldScenario_BusinessDayOperations", "18笔并发交易"},
		{"移动端网络不稳定", "TestRealWorldScenario_MobileAppUsage", "网络延迟场景"},
		{"高并发压力测试", "TestRealWorldScenario_HighConcurrencyStress", "300+并发交易"},
	}

	for i, scenario := range scenarios {
		fmt.Printf("\n%d/%d: %s - %s\n", i+1, len(scenarios), scenario.name, scenario.desc)
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

		if err := sh.RunV("go", "test", "./tests/integration/...", "-v", "-run", scenario.pattern, "-timeout", "5m"); err != nil {
			fmt.Printf("⚠️  Scenario '%s' failed\n", scenario.name)
			// 继续运行其他场景
		}
	}

	fmt.Println("\n✅ Demo completed")
	return nil
}

// Clean 清理构建产物
func Clean() error {
	fmt.Println("🧹 Cleaning build artifacts...")

	// 删除 bin 目录
	if err := sh.Rm(binDir); err != nil {
		fmt.Printf("Warning: failed to remove %s: %v\n", binDir, err)
	}

	// 删除覆盖率文件
	coverageFiles := []string{
		"coverage.out",
		"coverage.html",
		"coverage_unit.out",
		"coverage_unit.html",
		"coverage_integration.out",
		"coverage_integration.html",
		"tests/integration/coverage.out",
		"tests/integration/coverage.html",
	}
	for _, file := range coverageFiles {
		if err := os.Remove(file); err != nil && !os.IsNotExist(err) {
			fmt.Printf("Note: %s not found (skipped)\n", file)
		}
	}

	// 清理测试二进制文件
	if err := sh.Run("go", "clean", "-testcache"); err != nil {
		fmt.Printf("Warning: failed to clean test cache: %v\n", err)
	}

	fmt.Println("✅ Clean completed")
	return nil
}

// CleanAll 清理所有（包括测试数据库）
func CleanAll() error {
	fmt.Println("🧹 Cleaning everything...")

	// 清理构建产物
	if err := Clean(); err != nil {
		return err
	}

	// 清理测试数据库
	t := Test{}
	return t.CleanDB()
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

// Database namespace for database-related commands
type Database mg.Namespace

// Migrate 运行数据库迁移
func (Database) Migrate() error {
	fmt.Println("🔄 Running database migrations...")
	return sh.RunV("go", "run", "cmd/migrate/main.go")
}

// MigrateDown 回滚数据库迁移
func (Database) MigrateDown() error {
	fmt.Println("⏪ Rolling back database migrations...")
	return sh.RunV("go", "run", "cmd/migrate/main.go", "down")
}

// Version 显示版本信息
func Version() error {
	fmt.Printf("Version: %s\n", version)
	fmt.Printf("Git Commit: %s\n", getGitCommit())
	fmt.Printf("Go Version: %s\n", runtime.Version())
	return nil
}

// Info 显示项目信息
func Info() error {
	fmt.Println("ℹ️  Project Information")
	fmt.Println("")

	// Go 版本
	goVersion := runtime.Version()
	fmt.Printf("  Go version:    %s\n", goVersion)

	// 项目路径
	pwd, _ := os.Getwd()
	fmt.Printf("  Project path:  %s\n", pwd)

	// 模块名
	modName, err := sh.Output("go", "list", "-m")
	if err == nil {
		fmt.Printf("  Module:        %s\n", modName)
	}

	fmt.Println("")
	fmt.Println("📊 Code Statistics")

	// Go 文件数
	goFiles, _ := sh.Output("bash", "-c", "find . -name '*.go' | wc -l | tr -d ' '")
	fmt.Printf("  Go files:      %s\n", goFiles)

	// 测试文件数
	testFiles, _ := sh.Output("bash", "-c", "find . -name '*_test.go' | wc -l | tr -d ' '")
	fmt.Printf("  Test files:    %s\n", testFiles)

	// 总行数
	totalLines, _ := sh.Output("bash", "-c", "find . -name '*.go' | xargs wc -l | tail -1 | awk '{print $1}'")
	fmt.Printf("  Total lines:   %s\n", totalLines)

	fmt.Println("")
	return nil
}

// Dev 使用 air 热重载运行开发服务器
func Dev() error {
	fmt.Println("🔥 Starting with hot reload...")

	// 检查 air 是否安装
	if _, err := exec.LookPath("air"); err != nil {
		fmt.Println("⚠️  air not installed. Installing...")
		if err := sh.Run("go", "install", "github.com/cosmtrek/air@latest"); err != nil {
			return fmt.Errorf("failed to install air: %w", err)
		}
	}

	return sh.RunV("air")
}

// InstallTools 安装开发工具
func InstallTools() error {
	fmt.Println("🔧 Installing development tools...")

	tools := []struct {
		name string
		pkg  string
	}{
		{"air", "github.com/cosmtrek/air@latest"},
		{"golangci-lint", "github.com/golangci/golangci-lint/cmd/golangci-lint@latest"},
		{"goimports", "golang.org/x/tools/cmd/goimports@latest"},
	}

	for _, tool := range tools {
		fmt.Printf("Installing %s...\n", tool.name)
		if err := sh.Run("go", "install", tool.pkg); err != nil {
			fmt.Printf("Warning: failed to install %s: %v\n", tool.name, err)
		}
	}

	fmt.Println("✅ Tools installed")
	return nil
}

// CI CI 流程（验证、格式化、检查、测试）
func CI() error {
	fmt.Println("🚀 Running CI pipeline...")

	// 1. 验证依赖
	fmt.Println("\n1/5: Verifying dependencies...")
	if err := sh.RunV("go", "mod", "verify"); err != nil {
		return fmt.Errorf("dependency verification failed: %w", err)
	}

	// 2. 格式化检查
	fmt.Println("\n2/5: Checking code format...")
	if err := Format(); err != nil {
		return fmt.Errorf("format check failed: %w", err)
	}

	// 3. 运行 vet
	fmt.Println("\n3/5: Running go vet...")
	if err := Vet(); err != nil {
		return fmt.Errorf("vet failed: %w", err)
	}

	// 4. 运行 linter（如果可用）
	fmt.Println("\n4/5: Running linter...")
	if err := Lint(); err != nil {
		// Linter 失败不阻止 CI
		fmt.Printf("Warning: linter failed: %v\n", err)
	}

	// 5. 运行所有测试
	fmt.Println("\n5/5: Running tests...")
	t := Test{}
	if err := t.All(); err != nil {
		return fmt.Errorf("tests failed: %w", err)
	}

	fmt.Println("\n✅ CI checks passed")
	return nil
}

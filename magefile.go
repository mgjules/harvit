//go:build mage

package main

import (
	"os"

	"github.com/imdario/mergo"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

const (
	binPath     = "./bin"
	projectName = "harvit"
	ldFlags     = "-s -w"
	tags        = "jsoniter"
	opts        = "-trimpath"
)

var defaultEnvs = map[string]string{
	"CGO_ENABLED": "0",
}

func init() {
	os.Setenv("GO111MODULE", "on")
}

// Run tests
func Test() error {
	return sh.Run("go", "test", "./...")
}

// Run tests with race detector
func TestRace() error {
	return sh.Run("go", "test", "-race", "./...")
}

// Run go vet linter
func Vet() error {
	return sh.Run("go", "vet", "./...")
}

// Run go mod tidy
func Tidy() error {
	return sh.Run("go", "mod", "tidy")
}

// Generates docs
func Docs() error {
	return sh.Run("swag", "init", "--parseDependency", "--parseInternal")
}

// Lint
func Lint() error {
	return sh.Run("golangci-lint", "run", "./...", "--fast")
}

type Build mg.Namespace

// Builds for all supported popular OS/Arch
func (b Build) All() error {
	mg.Deps(
		b.LinuxAmd64,
		b.LinuxArm64,
		b.MacOSAmd64,
		b.MacOSArm64,
		b.WinAmd64,
	)
	return nil
}

// Builds for Linux 64bit
func (Build) LinuxAmd64() error {
	env := map[string]string{
		"GOOS":   "linux",
		"GOARCH": "amd64",
	}
	return sh.RunWith(flagEnvs(env), "go", "build", opts, "-tags", tags, "-ldflags", ldFlags, "-o", binPath+"/"+projectName+"-linux-amd64")
}

// Builds for Linux ARM 64bit
func (Build) LinuxArm64() error {
	env := map[string]string{
		"GOOS":   "linux",
		"GOARCH": "arm64",
	}
	return sh.RunWith(flagEnvs(env), "go", "build", opts, "-tags", tags, "-ldflags", ldFlags, "-o", binPath+"/"+projectName+"-linux-arm64")
}

// Builds for MacOS 64bit
func (Build) MacOSAmd64() error {
	env := map[string]string{
		"GOOS":   "darwin",
		"GOARCH": "amd64",
	}
	return sh.RunWith(flagEnvs(env), "go", "build", opts, "-tags", tags, "-ldflags", ldFlags, "-o", binPath+"/"+projectName+"-macos-amd64")
}

// Builds for MacOS M1
func (Build) MacOSArm64() error {
	env := map[string]string{
		"GOOS":   "darwin",
		"GOARCH": "arm64",
	}
	return sh.RunWith(flagEnvs(env), "go", "build", opts, "-tags", tags, "-ldflags", ldFlags, "-o", binPath+"/"+projectName+"-macos-arm64")
}

// Builds for Windows 64bit
func (Build) WinAmd64() error {
	env := map[string]string{
		"GOOS":   "windows",
		"GOARCH": "amd64",
	}
	return sh.RunWith(flagEnvs(env), "go", "build", opts, "-tags", tags, "-ldflags", ldFlags, "-o", binPath+"/"+projectName+"-win-amd64.exe")
}

func flagEnvs(env map[string]string) map[string]string {
	mergo.Merge(&env, defaultEnvs)
	return env
}

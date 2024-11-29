//go:build mage
// +build mage

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	// mg contains helpful utility functions, like Deps
)

// Build namespace
type Build mg.Namespace

// Test namespace
type Test mg.Namespace

// Run namespace
type Run mg.Namespace

var archTargets = map[string]map[string]string{
	"darwin_amd64": {
		"CGO_ENABLED": "0",
		"GO111MODULE": "on",
		"GOARCH":      "amd64",
		"GOOS":        "darwin",
	},
	"darwin_arm64": {
		"CGO_ENABLED": "0",
		"GO111MODULE": "on",
		"GOARCH":      "arm64",
		"GOOS":        "darwin",
	},
	"linux_amd64": {
		"CGO_ENABLED": "0",
		"GO111MODULE": "on",
		"GOARCH":      "amd64",
		"GOOS":        "linux",
	},
	"linux_arm64": {
		"CGO_ENABLED": "0",
		"GO111MODULE": "on",
		"GOARCH":      "arm64",
		"GOOS":        "linux",
	},
	"linux_386": {
		"CGO_ENABLED": "0",
		"GO111MODULE": "on",
		"GOARCH":      "386",
		"GOOS":        "linux",
	},
	"linux_armv5": {
		"CGO_ENABLED": "0",
		"GO111MODULE": "on",
		"GOARCH":      "arm",
		"GOARM":       "5",
		"GOOS":        "linux",
	},
	"linux_armv6": {
		"CGO_ENABLED": "0",
		"GO111MODULE": "on",
		"GOARCH":      "arm",
		"GOARM":       "6",
		"GOOS":        "linux",
	},
	"linux_armv7": {
		"CGO_ENABLED": "0",
		"GO111MODULE": "on",
		"GOARCH":      "arm",
		"GOARM":       "7",
		"GOOS":        "linux",
	},
	"windows_amd64": {
		"CGO_ENABLED": "0",
		"GO111MODULE": "on",
		"GOARCH":      "amd64",
		"GOOS":        "windows",
	},
}

func getVersion() string {
	out, err := exec.Command("git", "describe", "--tags").Output()
	if err != nil {
		return "unknown"
	}
	version := strings.TrimRight(string(out), "\r\n")
	return version
}

// Default target to run when none is specified
// If not set, running mage will list available targets
var Default = Build.Local

func buildCommand(command string, arch string) error {
	env, ok := archTargets[arch]
	if !ok {
		return fmt.Errorf("unknown arch %s", arch)
	}
	log.Printf("Building %s/%s\n", arch, command)
	outDir := fmt.Sprintf("./bin/%s/%s", arch, command)
	cmdDir := fmt.Sprintf("./pkg/cmd/%s", command)
	if err := sh.RunWith(
		env,
		"go",
		"build",
		"-ldflags",
		fmt.Sprintf("-X main.Version=%s", getVersion()),
		"-o", outDir, cmdDir); err != nil {
		return err
	}

	// intentionally igores errors
	sh.RunV("chmod", "+x", outDir)
	return nil
}

func kioskCmd() error {
	return buildCommand("grafana-kiosk", "darwin_amd64")
}

func buildCmdAll() error {
	for anArch := range archTargets {
		if err := buildCommand("grafana-kiosk", anArch); err != nil {
			return err
		}
	}
	return nil
}

func testVerbose() error {
	os.Setenv("GO111MODULE", "on")
	os.Setenv("CGO_ENABLED", "0")
	if err := sh.RunV("go", "test", "-v", "-coverpkg=./...", "--coverprofile=coverage.out", "./pkg/..."); err != nil {
		return err
	}
	if err := sh.RunV("go", "tool", "cover", "-func", "coverage.out"); err != nil {
		return err
	}
	return sh.RunV("go", "tool", "cover", "-html=coverage.out", "-o", "coverage.html")
}

func test() error {
	os.Setenv("GO111MODULE", "on")
	os.Setenv("CGO_ENABLED", "0")
	if err := sh.RunV("go", "test", "-coverpkg=./...", "--coverprofile=coverage.out", "./pkg/..."); err != nil {
		return err
	}
	if err := sh.RunV("go", "tool", "cover", "-func", "coverage.out"); err != nil {
		return err
	}
	return sh.RunV("go", "tool", "cover", "-html=coverage.out", "-o", "coverage.html")
}

// Format Formats the source files
func (Build) Format() error {
	if err := sh.RunV("gofmt", "-w", "./pkg"); err != nil {
		return err
	}
	return nil
}

// Local Minimal build
func (Build) Local(ctx context.Context) {
	mg.Deps(
		Clean,
		kioskCmd,
	)
}

// CI Lint/Format/Test/Build
func (Build) CI(ctx context.Context) {
	mg.Deps(
		Build.Lint,
		Build.Format,
		Test.Verbose,
		Clean,
		buildCmdAll,
	)
}

// All build all
func (Build) All(ctx context.Context) {
	mg.Deps(
		Build.Lint,
		Build.Format,
		Test.Verbose,
		buildCmdAll,
	)
}

// Build a docker image, call with image name as an argument: mage -v build:dockerArm64 "slimbean/grafana-kiosk:2024-11-29"
func (Build) DockerArm64(ctx context.Context, image string) error {
	log.Printf("Building docker...")
	return sh.RunV("docker", "build", "--build-arg", "TARGET_PLATFORM=linux/arm64", "--build-arg", "COMPILE_GOARCH=arm64", "-t", image, "-f", "build/Dockerfile", ".")
}

// Lint Run linter against codebase
func (Build) Lint() error {
	os.Setenv("GO111MODULE", "on")
	log.Printf("Linting...")
	return sh.RunV("golangci-lint", "--timeout", "5m", "run", "./pkg/...")
}

// Verbose Run tests in verbose mode
func (Test) Verbose() {
	mg.Deps(
		testVerbose,
	)
}

// Default Run tests in normal mode
func (Test) Default() {
	mg.Deps(
		test,
	)
}

// Clean Removes built files
func Clean() {
	log.Printf("Cleaning all")
	os.RemoveAll("./bin/linux_386")
	os.RemoveAll("./bin/linux_amd64")
	os.RemoveAll("./bin/linux_arm64")
	os.RemoveAll("./bin/linux_armv5")
	os.RemoveAll("./bin/linux_armv6")
	os.RemoveAll("./bin/linux_armv7")
	os.RemoveAll("./bin/darwin_amd64")
	os.RemoveAll("./bin/darwin_arm64")
	os.RemoveAll("./bin/windows_amd64")
}

// Local Build and Run
func (Run) Local() error {
	mg.Deps(Build.Local)
	return sh.RunV(
		"./bin/"+runtime.GOOS+"_"+runtime.GOARCH+"/grafana-kiosk",
		"-config",
		"config-example.yaml",
	)
}

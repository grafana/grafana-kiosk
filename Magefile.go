//go:build mage

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
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

func buildCommandWithVersion(command, arch, version string) error {
	env, ok := archTargets[arch]
	if !ok {
		return fmt.Errorf("unknown arch %s", arch)
	}
	log.Printf("Building %s/%s\n", arch, command)
	binary := command
	if env["GOOS"] == "windows" {
		binary += ".exe"
	}
	outDir := fmt.Sprintf("./bin/%s/%s", arch, binary)
	cmdDir := fmt.Sprintf("./pkg/cmd/%s", command)
	start := time.Now()
	err := sh.RunWith(
		env,
		"go",
		"build",
		"-ldflags",
		fmt.Sprintf("-X main.Version=%s", version),
		"-o", outDir, cmdDir)
	log.Printf("Built %s/%s in %s\n", arch, binary, time.Since(start).Round(time.Millisecond))
	return err
}

func kioskCmd() error {
	return buildCommandWithVersion("grafana-kiosk", runtime.GOOS+"_"+runtime.GOARCH, getVersion())
}

func buildCmdAll() error {
	version := getVersion()
	errs := make(chan error, len(archTargets))
	for anArch := range archTargets {
		go func(arch string) {
			errs <- buildCommandWithVersion("grafana-kiosk", arch, version)
		}(anArch)
	}
	for range archTargets {
		if err := <-errs; err != nil {
			return err
		}
	}
	return nil
}

func runTests(verbose bool) error {
	os.Setenv("GO111MODULE", "on")
	os.Setenv("CGO_ENABLED", "0")
	args := []string{"test"}
	if verbose {
		args = append(args, "-v")
	}
	args = append(args, "-coverpkg=./...", "--coverprofile=coverage.out", "./pkg/...")
	if err := sh.RunV("go", args...); err != nil {
		return err
	}
	if err := sh.RunV("go", "tool", "cover", "-func", "coverage.out"); err != nil {
		return err
	}
	return sh.RunV("go", "tool", "cover", "-html=coverage.out", "-o", "coverage.html")
}

// Format Formats the source files
func (Build) Format() error {
	return sh.RunV("gofmt", "-w", "./pkg")
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
	mg.Deps(Build.Format)
	mg.Deps(Build.Lint, Test.Verbose)
	mg.Deps(Clean, buildCmdAll)
}

// All build all
func (Build) All(ctx context.Context) {
	mg.SerialDeps(
		Build.Lint,
		Build.Format,
		Test.Verbose,
	)
	mg.Deps(
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
func (Test) Verbose() error {
	return runTests(true)
}

// Default Run tests in normal mode
func (Test) Default() error {
	return runTests(false)
}

// Clean Removes built files
func Clean() {
	log.Printf("Cleaning all")
	for arch := range archTargets {
		os.RemoveAll("./bin/" + arch)
	}
}

// Local Build and Run
func (Run) Local() error {
	mg.Deps(Build.Local)
	return sh.RunV(
		"./bin/"+runtime.GOOS+"_"+runtime.GOARCH+"/grafana-kiosk",
		"-c",
		"config-example.yaml",
	)
}

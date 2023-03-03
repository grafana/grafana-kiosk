//go:build mage
// +build mage

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	// mg contains helpful utility functions, like Deps
)

type Build mg.Namespace
type Test mg.Namespace
type Run mg.Namespace

var archTargets = map[string]map[string]string{
	"darwin_amd64": {
		"CGO_ENABLED": "0",
		"GO111MODULE": "on",
		"GOARCH":      "amd64",
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
	if err := sh.RunWith(env, "go", "build", "-o", outDir, cmdDir); err != nil {
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
	return sh.RunV("go", "test", "-v", "./pkg/...")
}

func test() error {
	os.Setenv("GO111MODULE", "on")
	os.Setenv("CGO_ENABLED", "0")
	return sh.RunV("go", "test", "./pkg/...")
}

// Formats the source files
func (Build) Format() error {
	if err := sh.RunV("gofmt", "-w", "./pkg"); err != nil {
		return err
	}
	return nil
}

// Minimal build
func (Build) Local(ctx context.Context) {
	mg.Deps(
		Clean,
		kioskCmd,
	)
}

// Lint/Format/Test/Build
func (Build) CI(ctx context.Context) {
	mg.Deps(
		Build.OSVScanner,
		Build.Lint,
		Build.Format,
		Test.Verbose,
		Clean,
		buildCmdAll,
	)
}

func (Build) All(ctx context.Context) {
	mg.Deps(
		Build.Lint,
		Build.Format,
		Test.Verbose,
		buildCmdAll,
	)
}

// Run linter against codebase
func (Build) Lint() error {
	os.Setenv("GO111MODULE", "on")
	log.Printf("Linting...")
	return sh.RunV("golangci-lint", "-v", "run", "./pkg/...")
}

func (Build) OSVScanner() error {
	log.Printf("Scanning...")
	return sh.RunV("osv-scanner", "--lockfile", "./go.mod")
}

// Run tests in verbose mode
func (Test) Verbose() {
	mg.Deps(
		testVerbose,
	)
}

// Run tests in normal mode
func (Test) Default() {
	mg.Deps(
		test,
	)
}

// Removes built files
func Clean() {
	log.Printf("Cleaning all")
	os.RemoveAll("./bin/linux_386")
	os.RemoveAll("./bin/linux_amd64")
	os.RemoveAll("./bin/linux_arm64")
	os.RemoveAll("./bin/linux_armv5")
	os.RemoveAll("./bin/linux_armv6")
	os.RemoveAll("./bin/linux_armv7")
	os.RemoveAll("./bin/darwin_amd64")
	os.RemoveAll("./bin/windows_amd64")
}

// Build and Run
func (Run) Local() error {
	mg.Deps(Build.Local)
	return sh.RunV(
		"./bin/"+runtime.GOOS+"_"+runtime.GOARCH+"/grafana-kiosk",
		"-config",
		"config-example.yaml",
	)
}

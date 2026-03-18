package initialize

import (
	"log"
	"os"
	"os/exec"
	"strings"
)

// sanitize removes newlines and carriage returns to prevent log injection
func sanitize(s string) string {
	s = strings.ReplaceAll(s, "\n", "")
	s = strings.ReplaceAll(s, "\r", "")
	return s
}

// allowedCommands defines the set of commands that runCommand may execute
var allowedCommands = map[string]bool{
	"/usr/bin/lxpanel":   true,
	"/usr/bin/pcmanfm":   true,
	"/usr/bin/xset":      true,
	"/usr/bin/unclutter": true,
}

// LXDE runs shell commands to setup LXDE for kiosk mode.
func LXDE(path string) {
	var command = "/usr/bin/lxpanel"

	args := []string{"--profile", "LXDE"}
	runCommand(path, command, args, true)
	command = "/usr/bin/pcmanfm"
	args = []string{"--desktop", "--profile", "LXDE"}
	runCommand(path, command, args, true)
	command = "/usr/bin/xset"
	runCommand(path, command, args, true)
	args = []string{"s", "off"}
	runCommand(path, command, args, true)
	args = []string{"-dpms"}
	runCommand(path, command, args, true)
	args = []string{"s", "noblank"}
	runCommand(path, command, args, true)
	command = "/usr/bin/unclutter"
	displayEnv := sanitize(os.Getenv("DISPLAY"))
	args = []string{"-display", displayEnv, "-idle", "5"}

	go runCommand(path, command, args, true)
}

func runCommand(path string, command string, args []string, waitForEnd bool) {
	if !allowedCommands[command] {
		log.Printf("command not allowed: %v", sanitize(command))
		return
	}

	log.Printf("path: %v", sanitize(path))
	log.Printf("command: %v", sanitize(command))
	log.Printf("arg0: %v", sanitize(args[0])) // #nosec G706 -- sanitized before logging

	cmd := exec.Command(command, args...) // #nosec G204 G702 -- command is validated against allowedCommands allowlist

	cmd.Env = append(os.Environ(),
		"DISPLAY=:0.0",
		"XAUTHORITY="+path+"/.Xauthority",
	)
	err := cmd.Start()

	if err != nil {
		log.Printf("Error in output, ignoring...")
	}

	if waitForEnd {
		log.Printf("Waiting for command to finish...")

		err = cmd.Wait()

		if err != nil {
			log.Printf("Command finished with error: %v", err)
		}
	}
}

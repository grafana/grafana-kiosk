package initialize

import (
	"log"
	"os"
	"os/exec"
)

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
	displayEnv := os.Getenv("DISPLAY")
	args = []string{"-display", displayEnv, "-idle", "5"}

	go runCommand(path, command, args, true)
}

func runCommand(path string, command string, args []string, waitForEnd bool) {
	// check if command exists
	log.Printf("path: %v", path)
	log.Printf("command: %v", command)
	log.Printf("arg0: %v", args[0])
	cmd := exec.Command(command, args...)

	cmd.Env = append(os.Environ(),
		"DISPLAY=:0.0",
		"XAUTHORITY="+path+"/.Xauthority",
	)
	err := cmd.Start()

	if err != nil {
		// log.Printf(err)
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

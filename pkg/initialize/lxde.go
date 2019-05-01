package initialize

import (
	"log"
	"os"
	"os/exec"
)

// LXDE runs shell commands to setup LXDE for kiosk mode
func LXDE(path string) {
	var command = "lxpanel"
	args := []string{"--profile", "LXDE"}
	runCommand(path, command, args)
	command = "pcmanfm"
	args = []string{"--desktop", "--profile", "LXDE"}
	runCommand(path, command, args)
	command = "xset"
	runCommand(path, command, args)
	args = []string{"s", "off"}
	runCommand(path, command, args)
	args = []string{"xset", "-dpms"}
	runCommand(path, command, args)
	args = []string{"xset", "s", "noblank"}
	runCommand(path, command, args)
	command = "unclutter"
	args = []string{"-display", ":0.0", "-idle", "5"}
	runCommand(path, command, args)
}

func runCommand(path string, command string, args []string) {
	cmd := exec.Command(command, args...)
	cmd.Env = append(os.Environ(),
		"DISPLAY=:0.0",
		"XAUTHORITY="+path+"/.Xauthority",
	)
	if err := cmd.Run(); err != nil {
		log.Println(err)
	}
}

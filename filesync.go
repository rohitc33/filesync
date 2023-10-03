package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
)

const remoteHostname = "raspberrypi"
const remotePathBase = "/media/user0/vol0/filesync"

func main() {
	currentUser, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	destination := currentUser.Username + "@" + remoteHostname

	filesyncDir := currentUser.HomeDir + "/Documents/.filesync"
	keyfile := filesyncDir + "/id_rsa"
	_, err = os.Stat(keyfile)
	if err != nil {
		exec.Command("/bin/rm", "-rf", filesyncDir).Run()
		exec.Command("/bin/mkdir", filesyncDir).Run()
		err := exec.Command("/usr/bin/ssh-keygen", "-t", "rsa", "-f", keyfile, "-q", "-N", "").Run()
		if err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}

	if len(os.Args) == 2 && os.Args[1] == "-i" {
		out, err := exec.Command("/usr/bin/osascript", "-e", "display dialog \"Start interactive SSH session?\" buttons {\"Yes\", \"No\"} default button \"No\"").Output()
		if err != nil {
			log.Fatal(err)
		}

		if string(out) == "button returned:Yes\n" {
			cmd := exec.Command("/usr/bin/ssh", "-i", keyfile, "-L", "5900:localhost:5900", destination)
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			err := cmd.Run()
			if err != nil {
				log.Fatal(err)
			}
		}
	} else {
		out, err := exec.Command("/bin/bash", "-c", "/usr/sbin/system_profiler SPHardwareDataType | awk '/Serial/ {printf $4}'").Output()
		if err != nil {
			log.Fatal(err)
		}
		serialNumber := string(out)
		dstPath := fmt.Sprintf("%s:%s/%s/%s", destination, remotePathBase, serialNumber, currentUser.Username)
		srcPath := currentUser.HomeDir + "/"
		sshCmd := fmt.Sprintf("/usr/bin/ssh -i \"%s\"", keyfile)

		cmd := exec.Command("/usr/bin/rsync", "-a", "-e", sshCmd, "--exclude", "Library", "--exclude", ".Trash", srcPath, dstPath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			log.Fatal(err)
		}
	}
}

package helper

import (
	"fmt"
	"os/exec"
	"runtime"
)

// Get the underlying OS command shell
func getOSC() string {

	osc := "sh"
	if runtime.GOOS == "windows" {
		osc = "cmd"
	}

	return osc
}

// Get the shell/command startup option to execute commands
func getOSE() string {

	ose := "-c"
	if runtime.GOOS == "windows" {
		ose = "/c"
	}
	return ose
}

func CheckCommand(command string) {
	path, err := exec.LookPath(command)
	if err != nil {
		fmt.Printf("didn't find '%s' executable\n", command)
	} else {
		fmt.Printf("'%s' executable is in '%s'\n", command, path)
	}
}

func Exec(command string) error {
	osc := getOSC()
	ose := getOSE()
	//cmd := exec.Command(command)
	//cmd.Env = append(os.Environ())

	err := exec.Command(osc, ose, command).Run()
	if err != nil {
		return err
	}
	return nil
}

func ExecFromDir(command, dir string) error {
	osc := getOSC()
	ose := getOSE()

	cmd := exec.Command(osc, ose, command)
	cmd.Dir = dir
	bytes, err := cmd.Output()
	fmt.Printf("OUTPUT: %s\n", string(bytes))
	//err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func ExecWithOut(command string) (string, error) {
	osc := getOSC()
	ose := getOSE()
	//cmd := exec.Command(command)
	//cmd.Env = append(os.Environ())

	bytes, err := exec.Command(osc, ose, command).Output()
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func MultiExecWithOut(outCmd string, cmds ...string) (string, error) {
	osc := getOSC()
	ose := getOSE()
	//cmd := exec.Command(command)
	//cmd.Env = append(os.Environ())

	for _, cmd := range cmds {
		fmt.Println(cmd)
		exec.Command(osc, ose, cmd).Run()
	}

	bytes, err := exec.Command(osc, ose, outCmd).Output()
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

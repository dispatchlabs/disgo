package helper

import (
	"github.com/dispatchlabs/disgo/commons/utils"
	"github.com/dispatchlabs/tools/common-util/util"
	"fmt"
	"os/user"

)

var disgoDir string

func Update(dir, versionNbr, password string) error {

	err := refreshCode()
	if err != nil {
		utils.Error(err)
		return err
	}
	err = buildDisgoExecutable(versionNbr, password)
	if err != nil {
		utils.Error(err)
		return err
	}
	err = updateDisgoExecutable(dir)
	if err != nil {
		utils.Error(err)
		return err
	}
	return err
}

func refreshCode() error {
	cmd := "go get -u github.com/dispatchlabs/disgo"
	output, err := ExecWithOut(cmd)
	if err != nil {
		utils.Error(err)
		return err
	}
	utils.Info(output)
	return nil
}

func buildDisgoExecutable(versionNbr, password string) error {
	util.DeleteFile(fmt.Sprintf("%s/disgo", GetDisgoDirectory()))
	CheckCommand("go")

	buildCmd := "go build -ldflags"
	versionArg := fmt.Sprintf("-X main.version=%s", versionNbr)
	dateArg := "-X main.date=`date \"+%Y-%m-%d-%H:%M:%S\"`"
	pwArg := fmt.Sprintf("-X go/src/github.com/dispatchlabs/types/Password=%s", password)

	cmd := fmt.Sprintf("%s \"%s %s %s\"", buildCmd, versionArg, dateArg, pwArg)

	fmt.Printf("CMD: %s\n", cmd)

	utils.Info("CMD:  ", cmd)
	err := ExecFromDir(cmd, GetDisgoDirectory())
	if err != nil {
		utils.Error(err)
		return err
	}
	return nil
}

func updateDisgoExecutable(dir string) error {
	cmd := fmt.Sprintf("cp %s/disgo %s", GetDisgoDirectory(), dir)
	fmt.Println(cmd)

	utils.Debug("Command: " + cmd)
	output, err := ExecWithOut(cmd)
	if err != nil {
		utils.Error(err)
		return err
	}
	utils.Info(output)
	return nil
}

func GetDisgoDirectory() string {
	if disgoDir == "" {
		usr, err := user.Current()
		if err != nil {
			utils.Fatal(err)
		}
		disgoDir = usr.HomeDir + "/go/src/github.com/dispatchlabs/disgo"
	}
	return disgoDir
}
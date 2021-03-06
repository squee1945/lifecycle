package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"

	"github.com/buildpack/lifecycle"
	"github.com/buildpack/lifecycle/cmd"
)

var (
	buildpacksDir string
	appDir        string
	platformDir   string
	orderPath     string
	groupPath     string
	planPath      string
	printVersion  bool
)

func init() {
	cmd.FlagBuildpacksDir(&buildpacksDir)
	cmd.FlagAppDir(&appDir)
	cmd.FlagPlatformDir(&platformDir)
	cmd.FlagOrderPath(&orderPath)
	cmd.FlagGroupPath(&groupPath)
	cmd.FlagPlanPath(&planPath)
	cmd.FlagVersion(&printVersion)
}

func main() {
	// suppress output from libraries, lifecycle will not use standard logger
	log.SetOutput(ioutil.Discard)

	flag.Parse()

	if printVersion {
		cmd.ExitWithVersion()
	}
	cmd.Exit(detect())
}

func detect() error {
	buildpacks, err := lifecycle.NewBuildpackMap(buildpacksDir)
	if err != nil {
		return cmd.FailErr(err, "read buildpack directory")
	}
	order, err := buildpacks.ReadOrder(orderPath)
	if err != nil {
		return cmd.FailErr(err, "read buildpack order file")
	}

	info, group := order.Detect(&lifecycle.DetectConfig{
		AppDir:      appDir,
		PlatformDir: platformDir,
		Out:         log.New(os.Stdout, "", 0),
		Err:         log.New(os.Stderr, "", 0),
	})

	if group == nil {
		return cmd.FailCode(cmd.CodeFailedDetect, "detect")
	}

	if err := group.Write(groupPath); err != nil {
		return cmd.FailErr(err, "write buildpack group")
	}

	if err := ioutil.WriteFile(planPath, info, 0666); err != nil {
		return cmd.FailErr(err, "write detect info")
	}

	return nil
}

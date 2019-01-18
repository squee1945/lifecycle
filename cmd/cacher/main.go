package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/BurntSushi/toml"

	"github.com/buildpack/lifecycle"
	"github.com/buildpack/lifecycle/cmd"
	"github.com/buildpack/lifecycle/image"
)


var (
	cacheImageTag string
	layersDir   string
	groupPath   string
	uid         int
	gid         int
)

func init() {
	cmd.FlagLayersDir(&layersDir)
	cmd.FlagGroupPath(&groupPath)
	cmd.FlagUID(&uid)
	cmd.FlagGID(&gid)
}

func main() {
	flag.Parse()
	if flag.NArg() > 1 || flag.Arg(0) == "" {
		args := map[string]interface{}{"narg": flag.NArg(), "layersDir": layersDir}
		cmd.Exit(cmd.FailCode(cmd.CodeInvalidArgs, "parse arguments", fmt.Sprintf("%+v", args)))
	}
	cacheImageTag = flag.Arg(0)
	cmd.Exit(cache())
}

func cache() error {
	var group lifecycle.BuildpackGroup
	if _, err := toml.DecodeFile(groupPath, &group); err != nil {
		return cmd.FailErr(err, "read group")
	}
	artifactsDir, err := ioutil.TempDir("", "lifecycle.exporter.layer")
	if err != nil {
		return cmd.FailErr(err, "create temp directory")
	}
	defer os.RemoveAll(artifactsDir)

	cacher := &lifecycle.Cacher{
		Buildpacks:   group.Buildpacks,
		ArtifactsDir: artifactsDir,
		Out:          log.New(os.Stdout, "", log.LstdFlags),
		Err:          log.New(os.Stderr, "", log.LstdFlags),
		UID:          uid,
		GID:          gid,
	}

	factory, err := image.DefaultFactory()
	if err != nil {
		return err
	}

	origCacheImage, err := factory.NewLocal(cacheImageTag, false)
	if err != nil {
		return err
	}

	cacher.Cache(layersDir, origCacheImage, factory.NewEmptyLocal(cacheImageTag))
	return nil
}

package main

import (
	"catchpole.net/bender/pkg/args"
	"catchpole.net/bender/pkg/bender"
	"os"
	"path"
)

func main() {
	o := bender.Options{}
	a := args.Args{}

	a.StringArg('j', "job", "", true, "job name.", nil, &o.Job)
	a.StringArg('p', "profile", "", true, "profile python file.", nil, &o.Profile)
	a.StringArg('b', "blend", "", true, "blend file.", nil, &o.Blend)
	a.StringArg('t', "target", "", true, "target directory.", nil, &o.Target)
	a.IntArg('s', "start", "start frame.", 1, &o.Start)
	a.IntArg('e', "end", "end frame.", 1, &o.End)
	a.IntArg('l', "samples", "cycles samples count.", 64, &o.Samples)
	a.StringArg('c', "camera", "", false, "Camera name", nil, &o.Camera)
	a.StringArg('x', "executable", "", false, "blender executable.", nil, &o.Blender)
	home, _ := os.UserHomeDir()
	a.LoadGlobalDefaults(path.Join(home, ".bender_defaults"))
	a.Process(os.Args, false, "", "")

	bender.Bender(o)
}

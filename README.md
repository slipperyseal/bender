
# Bender

## The slender Blender render defender

![bender](https://github.com/slipperyseal/bender/blob/main/doc/bender.jpg "bender")

### What it does

Bender is a simple Blender render job manager for creating animation
image sequences or single frames.

Bender uses python based configuration files, or `profiles`,
to set up the render including any standard settings you'd like to
apply. This means you don't need to check these are set correctly
in your `blend` files, or have to keep updating them if you are
targeting different preview or production configurations. 

If a job gets interrupted, Bender recovers from the last frame rendered.

### How it works

Bender is written on `go` 

#### Build

  `go build -o ./bender cmd/bender/main.go`

#### Run to view options

  `./bender`

```
Usage:
    bender [options] 

Options:
    -j  --job        <job>         job name. (required)
    -p  --profile    <profile>     profile python file. (required)
    -b  --blend      <blend>       blend file. (required)
    -t  --target     <target>      target directory. (required)
    -s  --start      [number]      start frame. default [ 1 ]
    -e  --end        [number]      end frame. default [ 1 ]
    -l  --samples    [number]      cycles samples count. default [ 64 ]
    -x  --executable <executable>  blender executable.
```

By default Bender will use the Blender executable path for MacOS.
You can use the `--executable` option to override it. 

`/Applications/Blender.app/Contents/MacOS/Blender`

#### Setting up a profile

This example template shows a 4K 2:1 aspect ratio configuration
with motion blur and output to ZIP encoded EXR files.
Options can be changed or removed as per your needs.

See the Blender python API documentation:
https://docs.blender.org/api/current/index.html

  `example.py`

```
import bpy

for scene in bpy.data.scenes:
    scene.cycles.device = "GPU"
    scene.cycles.samples = {samples}
    scene.frame_start = {start}
    scene.frame_end = {end}
    scene.frame_step = 1
    scene.render.filepath = {outpath}
    scene.render.resolution_x = 3840
    scene.render.resolution_y = 1920
    scene.render.resolution_percentage = 100
    scene.render.use_motion_blur = True
    scene.render.motion_blur_shutter = 0.5
    scene.render.use_overwrite = True
    scene.render.use_border = False
    scene.render.image_settings.file_format = "OPEN_EXR"
    scene.render.image_settings.exr_codec = "ZIP" # NONE, PXR24, ZIP, PIZ, RLE, ZIPS, B44, B44A, DWAA, DWAB
    scene.render.image_settings.color_depth = "16"
    scene.render.image_settings.color_mode='RGB'
```

#### Example job

  `./bender --job pants --target jobs --profile example.py --blend mypants.blend --start 1 --end 5 --samples 128`
  
Bender will then create the following directory structure,
creating the `profile` python file and run Blender which will
create the frames in that directory.

```
jobs/
    pants/
         pants.py
         pants0001.exr
         pants0002.exr
         pants0003.exr
         pants0004.exr
         pants0005.exr
```

If a job gets interrupted, Bender recovers from the last frame rendered.
If a job is complete Bender will do nothing. If you need to re-start a job
simply delete off the output files and start again. 

![progress](https://github.com/slipperyseal/bender/blob/main/doc/progress.png "progress")

Bender parses the log output of Blender and attempts to produce a nice
updating summary in your terminal.  It's a bit flakey, and may break as Blender
changes. But it's easier to read than Blender logs.


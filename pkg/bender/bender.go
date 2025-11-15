package bender

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

type Options struct {
	Blend     string
	Job       string
	Profile   string
	Target    string
	Overwrite bool
	Blender   string
	Start     int
	End       int
	Samples   int
	Percent   int
	Camera    string
}

type Paths struct {
	jobDir            string
	jobProfile        string
	outputPatternPath string
}

type RenderUpdate struct {
	Frame     int
	Sample    string
	Time      string
	Remaining string
}

var currentFrame int

func Bender(o Options) {
	if o.Blender == "" {
		o.Blender = "/Applications/Blender.app/Contents/MacOS/Blender"
	}
	paths := setupPaths(o.Job, o.Target)
	o.Start = skipFrames(paths.jobDir, o.Job, o.Start, o.Overwrite)
	if o.Start > o.End {
		log.Fatalln("No frames left to render")
	}
	currentFrame = o.Start
	
	profileBytes, err := os.ReadFile(o.Profile)
	if err != nil {
		log.Fatalf("Unable to read %s - %s\n", o.Profile, err)
	}
	createProfile(string(profileBytes), paths, o)
	fmt.Printf("Starting blender for %s\n", o.Job)
	startBlender(o, paths.jobProfile)
	fmt.Printf("Job complete: %s\n", o.Job)
}

func setupPaths(job string, dirname string) Paths {
	jobDir, err := filepath.Abs(filepath.Join(dirname, job))
	if err != nil {
		log.Fatalf("Path not valid %s - %s", dirname, err)
	}
	derr := os.MkdirAll(jobDir, os.ModePerm)
	if derr != nil {
		log.Fatalf("Unable to create %s - %s", jobDir, derr)
	}
	return Paths{
		jobDir:            jobDir,
		jobProfile:        filepath.Join(jobDir, job+".py"),
		outputPatternPath: filepath.Join(jobDir, job+"_####"),
	}
}

func frameExists(entries []os.DirEntry, job string, num int) bool {
	// we don't know the file extension so match on "example0001."
	match := fmt.Sprintf("%s_%04d.", job, num)
	for _, e := range entries {
		if !e.IsDir() && strings.HasPrefix(e.Name(), match) {
			return true
		}
	}
	return false
}

func skipFrames(jobDir string, job string, num int, overwrite bool) int {
	if overwrite {
		return num
	}
	entries, err := os.ReadDir(jobDir)
	if err != nil {
		log.Fatal(err)
	}
	for frameExists(entries, job, num) {
		num++
	}
	return num
}

func startBlender(o Options, jobProfile string) {
	cmd := exec.Command(o.Blender, "--background", o.Blend, "--python", jobProfile, "--render-anim")
	stdout, err := cmd.StdoutPipe()
	cmd.Stderr = cmd.Stdout
	if err != nil {
		log.Fatal(err)
	}
	if err = cmd.Start(); err != nil {
		log.Fatal(err)
	}
	scanner := bufio.NewScanner(stdout)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		processLine(scanner.Text(), o)
	}
	cmdErr := cmd.Wait()
	if cmdErr != nil {
		log.Fatal(cmdErr)
	}
}

// Blender 4 log format
// Fra:1 Mem:2845.68M (Peak 2845.69M) | Time:00:14.42 | Remaining:03:01.65 | Mem:4360.26M, Peak:4360.26M | Scene, View Layer | Sample 17/256

// Blender 5 log format
// 00:04.845  render           | Mem: 1614M | Sample 0/512 (Using optimized kernels)
// 00:34.042  render           | Remaining: 07:03.80 | Mem: 2007M | Sample 33/512 (Using optimized kernels)
func processLine(line string, o Options) {
	cols := strings.Split(line, "|")
	for i := range cols {
	    cols[i] = strings.TrimSpace(cols[i])
	}

	var update RenderUpdate
	var sample string

	if len(cols) == 3 {
		sample = cols[2]
		if strings.HasPrefix(sample, "Finished") {
			currentFrame++
			return
		}
	} else if len(cols) == 4 {
		update.Remaining = first(strings.TrimPrefix(cols[1], "Remaining: "), '.')
		sample = cols[3]
	} else {
		return
	}

	if !strings.HasPrefix(sample, "Sample ") {
		return
	}

	update.Frame = currentFrame
	update.Sample = first(strings.TrimPrefix(sample, "Sample "),' ')
	update.Time = first(first(cols[0], ' '), '.')

	printScreen(update, o)
}

func first(str string, sep byte) string {
    if i := strings.IndexByte(str, sep); i >= 0 {
        return str[:i]
    }
    return str
}

func createProfile(profile string, paths Paths, o Options) {
	profile = strings.ReplaceAll(profile, "{outpath}", paths.outputPatternPath)
	profile = strings.ReplaceAll(profile, "{samples}", strconv.Itoa(o.Samples))
	profile = strings.ReplaceAll(profile, "{start}", strconv.Itoa(o.Start))
	profile = strings.ReplaceAll(profile, "{end}", strconv.Itoa(o.End))
	profile = strings.ReplaceAll(profile, "{percent}", strconv.Itoa(o.Percent))
	profile = strings.ReplaceAll(profile, "{camera}", o.Camera)

	f, err := os.Create(paths.jobProfile)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	_, err2 := f.WriteString(profile)
	if err2 != nil {
		log.Fatal(err2)
	}
}

func printScreen(r RenderUpdate, o Options) {
	fmt.Printf("\033[2J\033[H\r\n\033[0;32m"+
		"Bender\033[0m\r\n\r\n"+
		"\033[7m%16s\033[0m"+
		"\033[0;32m\033[7m%8s\033[0m"+
		"\033[0;36m\033[7m%8s\033[0m"+
		"\033[0;33m\033[7m%8s\033[0m"+
		"\033[0;32m\033[7m%12s\033[0m"+
		"\033[0;33m\033[7m%12s\033[0m"+
		"\033[7m%12s\033[0m\n",
		"job", "start", "frame", "end", "sample", "time", "remaining")
	fmt.Printf("%16s%8d%8d%8d%12s%12s%12s\n\n",
		o.Job, o.Start, r.Frame, o.End, r.Sample, r.Time, r.Remaining)
}

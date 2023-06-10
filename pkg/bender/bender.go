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
	Blend   string
	Job     string
	Profile string
	Target  string
	Blender string
	Start   int
	End     int
	Samples int
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

func Bender(o Options) {
	if o.Blender == "" {
		o.Blender = "/Applications/Blender.app/Contents/MacOS/Blender"
	}
	paths := setupPaths(o.Job, o.Target)
	o.Start = skipFrames(paths.jobDir, o.Job, o.Start)
	if o.Start > o.End {
		log.Fatalln("No frames left to render")
	}
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
		outputPatternPath: filepath.Join(jobDir, job+"####"),
	}
}

func frameExists(entries []os.DirEntry, job string, num int) bool {
	// we don't know the file extension so match on "example0001."
	match := fmt.Sprintf("%s%04d.", job, num)
	for _, e := range entries {
		if !e.IsDir() && strings.HasPrefix(e.Name(), match) {
			return true
		}
	}
	return false
}

func skipFrames(jobDir string, job string, num int) int {
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

// Fra:1 Mem:2845.68M (Peak 2845.69M) | Time:00:14.42 | Remaining:03:01.65 | Mem:4360.26M, Peak:4360.26M | Scene, View Layer | Sample 17/256
func processLine(line string, o Options) {
	if strings.HasPrefix(line, "Fra:") {
		f, _ := strconv.Atoi(strings.TrimPrefix(strings.Split(line, " ")[0], "Fra:"))
		line = strings.ReplaceAll(line, " ", "")
		cols := strings.Split(line, "|")
		if len(cols) == 6 {
			printScreen(
				RenderUpdate{
					Frame:     f,
					Sample:    strings.TrimPrefix(cols[5], "Sample"),
					Time:      strings.TrimPrefix(cols[1], "Time:"),
					Remaining: strings.TrimPrefix(cols[2], "Remaining:"),
				}, o)
		}
	}
}

func createProfile(profile string, paths Paths, o Options) {
	profile = strings.ReplaceAll(profile, "{outpath}", "\""+paths.outputPatternPath+"\"")
	profile = strings.ReplaceAll(profile, "{samples}", strconv.Itoa(o.Samples))
	profile = strings.ReplaceAll(profile, "{start}", strconv.Itoa(o.Start))
	profile = strings.ReplaceAll(profile, "{end}", strconv.Itoa(o.End))

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
		"\033[7m%10s\033[0m"+
		"\033[0;32m\033[7m%6s\033[0m"+
		"\033[0;36m\033[7m%6s\033[0m"+
		"\033[0;33m\033[7m%6s\033[0m"+
		"\033[0;33m\033[7m%10s\033[0m"+
		"\033[7m%10s\033[0m\n",
		"job", "start", "end", "frame", "time", "remaining")
	fmt.Printf("%10s%6d%6d%6d%10s%10s\n\n%48s\n\n",
		o.Job, o.Start, o.End, r.Frame, r.Time, r.Remaining, r.Sample)
}

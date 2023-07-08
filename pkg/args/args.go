package args

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

type Args struct {
	Header    string
	argSlice  []arg
	program   string
	tailLabel string
	globals   string
}

type arg struct {
	letter       byte
	name         string
	required     bool
	dephault     string
	description  string
	choices      []string
	outputString *string
	outputBool   *bool
	outputInt    *int
	resolved     bool
	queen        bool
}

func (a *Args) Process(osargs []string, requireTail bool, tailLabel string, tailSuffix string) []string {
	ss := strings.Split(osargs[0], "/")
	a.program = ss[len(ss)-1]
	a.tailLabel = tailLabel

	argMap := make(map[string]*arg, 32)
	for i, _ := range a.argSlice {
		arg := &a.argSlice[i]
		argMap[fmt.Sprintf("-%c", arg.letter)] = arg
		argMap["--"+arg.name] = arg
	}

	osargs = append(osargs, strings.Split(a.globals, " ")...)

	queen := false
	for i, osarg := range osargs {
		if i == 0 {
			continue
		}
		arg, found := argMap[osarg]
		if found {
			if arg.resolved {
				a.FailWith(fmt.Sprintf("Duplicate option defined: %s", osarg))
			}
			if arg.outputBool != nil {
				*arg.outputBool = true
				if arg.queen {
					queen = true
				}
			} else {
				// end of slice or next element is an option
				if i == len(osargs)-1 || strings.HasPrefix(osargs[i+1], "-") {
					a.FailWith(fmt.Sprintf("Value expected for: %s", osarg))
				}
				value := osargs[i+1]

				if arg.outputString != nil {
					*arg.outputString = value
				}
				if arg.outputInt != nil {
					i, err := strconv.Atoi(value)
					if err != nil {
						a.FailWith(fmt.Sprintf("Option must be a number: %s", osarg))
					}
					*arg.outputInt = i
				}
				osargs[i+1] = "" // remove value
			}
			osargs[i] = "" // remove option
			arg.resolved = true
		} else if strings.HasPrefix(osarg, "-") {
			a.FailWith(fmt.Sprintf("Unknown option: %s", osarg))
		}
	}

	// set defaults and check mandatory Options
	for i, _ := range a.argSlice {
		arg := &a.argSlice[i]
		if !arg.resolved && arg.required && !queen {
			a.FailWith(fmt.Sprintf("Value expected for: %s", arg.name))
		}
		if !arg.resolved && arg.outputBool == nil {
			if arg.outputString != nil {
				*arg.outputString = arg.dephault
			}
			if arg.outputInt != nil {
				i, _ := strconv.Atoi(arg.dephault)
				*arg.outputInt = i
			}
		}
	}

	// as options are processed they are removed from the slice. whatever remains is the tail.
	tail := []string{}
	for i, osarg := range osargs {
		if i == 0 {
			continue
		}
		if len(osarg) != 0 {
			tail = append(tail, osarg)
		}
	}

	if requireTail && len(tail) == 0 {
		a.FailWith("")
	}

	if tailSuffix != "" {
		for _, t := range tail {
			if !strings.HasSuffix(t, tailSuffix) {
				a.FailWith(fmt.Sprintf("%s must have extension %s", tailLabel, tailSuffix))
			}
		}
	}

	return tail
}

func (a *Args) StringArg(letter byte, name string, dephault string, required bool, description string, choices []string, output *string) {
	if dephault != "" && required {
		log.Fatalf("Naughty Args config: required and default are mutually exclusive: %s\n", name)
	}
	arg := arg{
		letter:       letter,
		name:         name,
		required:     required,
		dephault:     dephault,
		description:  description,
		choices:      choices,
		outputString: output,
	}
	a.argSlice = append(a.argSlice, arg)
}

func (a *Args) BoolArg(letter byte, name string, description string, output *bool, queen bool) {
	arg := arg{
		letter:      letter,
		name:        name,
		description: description,
		outputBool:  output,
		queen:       queen,
	}
	a.argSlice = append(a.argSlice, arg)
}

func (a *Args) IntArg(letter byte, name string, description string, dephault int, output *int) {
	arg := arg{
		letter:      letter,
		name:        name,
		description: description,
		dephault:    strconv.Itoa(dephault),
		outputInt:   output,
	}
	a.argSlice = append(a.argSlice, arg)
}

func (a *Args) FailWith(error string) {
	a.PrintUsage()
	if error != "" {
		fmt.Printf("\n%s\n\n", error)
	}
	os.Exit(1)
}

func (a *Args) LoadGlobalDefaults(filename string) {
	bytes, err := os.ReadFile(filename) // just pass the file name
	if err != nil {
		// no defaults file
		return
	}
	a.globals = strings.ReplaceAll(strings.ReplaceAll(string(bytes), "\n", " "), "\r", " ")
}

func (a *Args) PrintUsage() {
	fmt.Println(a.Header)
	fmt.Printf("\nUsage:\n    %s [options] %s\n\nOptions:\n", a.program, a.tailLabel)

	// setup format based on the longest option name
	l := 8
	for _, a := range a.argSlice {
		if len(a.name) > l {
			l = len(a.name)
		}
	}
	format := fmt.Sprintf("    -%%c  --%%-%ds %%-%ds  %%s", l, l+2)

	for _, a := range a.argSlice {
		if a.queen {
			fmt.Println()
		}
		i := ""
		if a.outputString != nil {
			i = fmt.Sprintf("<%s>", a.name)
		} else if a.outputInt != nil {
			i = "[number]"
		}
		fmt.Printf(format, a.letter, a.name, i, a.description)
		if len(a.choices) != 0 {
			fmt.Print(" [ ")
			for i, c := range a.choices {
				if i != 0 {
					fmt.Print(", ")
				}
				fmt.Print(c)
			}
			fmt.Print(" ]")
		}
		if a.dephault != "" {
			fmt.Print(" default [ ")
			fmt.Print(a.dephault)
			fmt.Print(" ]")
		}
		if a.required {
			fmt.Print(" (required)")
		}
		fmt.Println()
	}
	fmt.Println()
}

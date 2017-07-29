package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type commandArgs struct {
	help       bool
	gingerFile string
	ninjaFile  string
}

type sourceFile struct {
	name      string
	path      string
	extension string
}

type gingerContext struct {
	buildDirectory string
	compiler       string
	compilerFlags  []string
	linker         string
	linkerFlags    []string
	sourceFiles    []sourceFile
	includePaths   []string
	target         string
}

var sourceFileRegexp = regexp.MustCompile(`^\.((c)|(cpp))$`)
var headerFileRegexp = regexp.MustCompile(`^\.(h)$`)

func main() {
	//override default usage
	flag.Usage = usage

	args := parseArgs()

	if args.help {
		usage()
		return
	}

	//use custom log to remove timestamp
	errorLog := log.New(os.Stderr, "", 0)

	var context gingerContext
	context.buildDirectory = "."
	context.compiler = ""
	context.compilerFlags = make([]string, 0)
	context.linker = ""
	context.linkerFlags = make([]string, 0)
	context.sourceFiles = make([]sourceFile, 0)
	context.includePaths = make([]string, 0)
	context.target = ""

	err := parseGingerFile(args.gingerFile, &context)

	if err != nil {
		errorLog.Fatal(err)
	}

	err = validateGingerContext(&context)

	if err != nil {
		errorLog.Fatal(err)
	}

	err = writeNinjaFile(args.ninjaFile, &context)

	if err != nil {
		errorLog.Fatal(err)
	}
}

func usage() {
	fmt.Println("\nOVERVIEW: ginger ninja build file generator")
	fmt.Println("\nUSAGE: ginger.exe [options]")
	fmt.Println("\nOPTIONS:\n")
	fmt.Println("  -i=<file>      Input ginger file")
	fmt.Println("  -o=<file>      Output ninja file")
	fmt.Println("  -h             Help")
}

func parseArgs() commandArgs {
	var args commandArgs

	flag.StringVar(&args.gingerFile, "i", "build.ginger", "")
	flag.StringVar(&args.ninjaFile, "o", "build.ninja", "")
	flag.BoolVar(&args.help, "h", false, "help")
	flag.Parse()

	return args
}

func parseGingerFile(fileName string, context *gingerContext) error {
	file, err := os.Open(fileName)

	if err != nil {
		return err
	}

	defer file.Close()
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		parseLine(line, context)
	}

	err = walkProject(".", &context.sourceFiles, &context.includePaths)

	if err != nil {
		return err
	}

	return nil
}

func parseLine(line string, context *gingerContext) {
	if len(line) < 1 {
		return
	}
	if line[0] == '#' {
		return
	}

	tokens := strings.Split(line, " ")

	if len(tokens) < 2 {
		return
	}

	contextType := tokens[0]
	contextEntry := strings.TrimPrefix(line, tokens[0]+" ")

	if strings.EqualFold(contextType, "-builddir") {
		context.buildDirectory = contextEntry
	}
	if strings.EqualFold(contextType, "-cc") {
		context.compiler = contextEntry
	}
	if strings.EqualFold(contextType, "-cf") {
		context.compilerFlags = append(context.compilerFlags, contextEntry)
	}
	if strings.EqualFold(contextType, "-ll") {
		context.linker = contextEntry
	}
	if strings.EqualFold(contextType, "-lf") {
		context.linkerFlags = append(context.linkerFlags, contextEntry)
	}
	if strings.EqualFold(contextType, "-target") {
		context.target = contextEntry
	}
}

func walkProject(root string, sourceFiles *[]sourceFile, includePaths *[]string) error {
	files, err := ioutil.ReadDir(root)

	if err != nil {
		return err
	}

	for _, file := range files {
		path := root + string(filepath.Separator) + file.Name()
		if file.IsDir() {
			err = walkProject(path, sourceFiles, includePaths)
			if err != nil {
				return err
			}
		} else if sourceFileRegexp.MatchString(filepath.Ext(file.Name())) {
			var sourceFile sourceFile
			sourceFile.name = strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
			sourceFile.path = root + string(filepath.Separator)
			sourceFile.extension = filepath.Ext(file.Name())
			*sourceFiles = append(*sourceFiles, sourceFile)
		} else if headerFileRegexp.MatchString(filepath.Ext(file.Name())) && !containsPath(root, includePaths) {
			*includePaths = append(*includePaths, root)
		}
	}

	return nil
}

func containsPath(path string, includePaths *[]string) bool {
	for _, includePath := range *includePaths {
		if strings.EqualFold(path, includePath) {
			return true
		}
	}

	return false
}

func validateGingerContext(context *gingerContext) error {
	if context.compiler == "" {
		return errors.New("invalid ginger file: -cc not defined")
	} else if context.linker == "" {
		return errors.New("invalid ginger file: -ll not defined")
	} else if len(context.sourceFiles) == 0 {
		return errors.New("no source files detected")
	}

	return nil
}

func writeNinjaFile(ninjaBuildFile string, context *gingerContext) error {
	file, err := os.OpenFile(ninjaBuildFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)

	if err != nil {
		return err
	}

	defer file.Close()

	file.WriteString("#ginger ninja file\n\n")
	file.WriteString("target = " + context.target + "\n")
	file.WriteString("builddir = " + context.buildDirectory + "\n")
	file.WriteString("cc = " + context.compiler + "\n")

	if len(context.compilerFlags) > 0 {
		file.WriteString("cf =")

		for _, flag := range context.compilerFlags {
			file.WriteString(" " + flag)
		}
		for _, includePath := range context.includePaths {
			file.WriteString(" -I \"" + includePath + "\"")
		}

		file.WriteString("\n")
	}

	file.WriteString("ll = " + context.linker + "\n")

	if len(context.linkerFlags) > 0 {
		file.WriteString("lf =")

		for _, flag := range context.linkerFlags {
			file.WriteString(" " + flag)
		}

		file.WriteString("\n")
	}

	file.WriteString("\nrule compile\n")
	file.WriteString("  command = $cc $cf -c $in -o $out\n")
	file.WriteString("\nrule link\n")
	file.WriteString("  command = $ll $lf $in -o $out\n")

	for _, sourceFile := range context.sourceFiles {
		compilerInput := sourceFile.path + sourceFile.name + sourceFile.extension
		compilerOutput := strings.TrimPrefix(sourceFile.path, ".") + sourceFile.name + ".o"
		file.WriteString(fmt.Sprintf("\nbuild $builddir"+compilerOutput+": $\n") + "  compile " + compilerInput + "\n")
	}

	for i, sourceFile := range context.sourceFiles {
		if i == 0 {
			file.WriteString("\nbuild $target : link ")
		} else {
			file.WriteString("                     ")
		}

		file.WriteString("$builddir" + strings.TrimPrefix(sourceFile.path, ".") + sourceFile.name + ".o")

		if i == len(context.sourceFiles)-1 {
			file.WriteString("\n")
		} else {
			file.WriteString(" $\n")
		}
	}

	file.WriteString("\ndefault $target\n")
	return nil
}

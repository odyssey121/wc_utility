/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var mapFn = map[string]any{"max-line-length": getMaxLineLen, "lines": getLines, "words": getWords, "chars": getChars, "bytes": getBytes}
var flagsArr = [5]string{"max-line-length", "lines", "words", "chars", "bytes"}

type getFuncParam struct {
	f            *os.File
	withLastLine bool
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "wc [file]",
	Short: "Usage: wc [OPTION]... [FILE]...",
	Long: `The options below may be used to select which counts are printed, always in
  the following order: newline, word, character, byte, maximum line length.
  -b, --bytes            print the byte counts
  -c, --chars            print the character counts
  -l, --lines            print the newline counts
      --files0-from=F    read input from the files specified by
                           NUL-terminated names in file F;
                           If F is - then read names from standard input
  -L, --max-line-length  print the maximum display width
  -w, --words            print the word counts
      --help     display this help and exit
      --version  output version information and exit`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.MinimumNArgs(1)(cmd, args); err != nil {
			return fmt.Errorf("Please enter the file as the first argument")
		} else if _, err := os.Stat(args[0]); err != nil && os.IsNotExist(err) {
			return fmt.Errorf("File path %s not found", args[0])
		} else if err != nil {
			return err
		}
		return cobra.OnlyValidArgs(cmd, args)
	},
	Run: func(cmd *cobra.Command, args []string) {
		defer timer("main")()
		path := args[0]
		fd, _ := os.Open(path)
		changedFlags := []string{}

		for _, flag := range flagsArr {
			if cmd.Flags().Changed(flag) {
				changedFlags = append(changedFlags, flag)
			}
		}

		if len(changedFlags) == 0 {
			changedFlags = append(changedFlags, flagsArr[:]...)
		}

		var resultSlice = make([]int, len(changedFlags))
		// main logic
		for i, f := range changedFlags {
			fn, ok := mapFn[f]
			if ok {
				resultSlice[i] = fn.(func(getFuncParam) int)(getFuncParam{fd, true})
			}
		}

		var result string

		for i, v := range resultSlice {
			result = fmt.Sprintf("%s%s: %d, ", result, changedFlags[i], v)
		}
		fmt.Println(strings.TrimRight(result, ", "), " ", filepath.Base(path))

	},
}

func timer(name string) func() {
	start := time.Now()
	return func() {
		fmt.Printf("%s took %v\n", name, time.Since(start))
	}
}

func getChars(p getFuncParam) int {
	if getCurPos(p.f) != 0 {
		p.f.Seek(0, io.SeekStart)
	}
	scanner := bufio.NewScanner(p.f)
	scanner.Split(bufio.ScanRunes)
	count := 0
	for scanner.Scan() {
		txt := scanner.Text()
		if txt != "" {
			count += len(txt)
		}
	}
	return count
}

func getWords(p getFuncParam) int {
	if getCurPos(p.f) != 0 {
		p.f.Seek(0, io.SeekStart)
	}
	scanner := bufio.NewScanner(p.f)
	count := 0
	re := regexp.MustCompile(`[^\s]+`)
	for scanner.Scan() {
		matchesB := re.FindAll(scanner.Bytes(), -1)
		count += len(matchesB)
	}
	return count
}

func getBytes(p getFuncParam) int {
	if getCurPos(p.f) != 0 {
		p.f.Seek(0, io.SeekStart)
	}
	fS, _ := p.f.Stat()
	scanner := bufio.NewScanner(p.f)
	scanner.Split(bufio.ScanBytes)
	bSlice := make([]byte, 0, fS.Size())
	for scanner.Scan() {
		b := scanner.Bytes()
		bSlice = append(bSlice, b...)
	}
	return len(bSlice)
}

func getCurPos(f *os.File) int64 {
	offset, err := f.Seek(0, io.SeekCurrent)
	if err != nil {
		fmt.Println("getCurPos error:", err)

	}
	return offset

}

func getMaxLineLen(p getFuncParam) int {
	if getCurPos(p.f) != 0 {
		p.f.Seek(0, io.SeekStart)
	}
	rd := bufio.NewReader(p.f)
	max := 0
	for {
		b, _, err := rd.ReadLine()
		lineLen := len(string(b))

		if err != nil {
			if err == io.EOF {
				if lineLen > 0 && p.withLastLine && (max < lineLen) {
					max = lineLen

				}
				break
			}
			fmt.Printf("read file max line length error: %v", err)
		}

		if max < lineLen {
			max = lineLen
		}

	}

	return max
}

func getLines(p getFuncParam) int {
	if getCurPos(p.f) != 0 {
		p.f.Seek(0, io.SeekStart)
	}
	rd := bufio.NewReader(p.f)
	count := 0
	for {
		l, err := rd.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				if len(l) != 0 && p.withLastLine {
					count++
				}
				break
			}
			fmt.Printf("read file line error: %v", err)
		}
		count++
	}
	return count
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}

}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.wc.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.

	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.Flags().BoolP("bytes", "b", false, "Print the byte counts")
	rootCmd.Flags().BoolP("chars", "c", false, "Print the character counts")
	rootCmd.Flags().BoolP("lines", "l", false, "print the newline counts")
	rootCmd.Flags().BoolP("max-line-length", "L", false, "Print the maximum display width")
	rootCmd.Flags().BoolP("words", "w", false, "Print the word counts")
}

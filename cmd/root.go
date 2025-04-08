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

var mapCh = make(map[string](chan int))
var flagsArr = [5]string{"max-line-length", "lines", "words", "chars", "bytes"}

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

		changedFlags := []string{}

		for _, flag := range flagsArr {
			if cmd.Flags().Changed(flag) {
				changedFlags = append(changedFlags, flag)
			}

		}

		if len(changedFlags) == 0 {
			changedFlags = append(changedFlags, flagsArr[:]...)
		}

		// init chan
		for _, f := range changedFlags {
			mapCh[f] = make(chan int)
		}

		for _, flag := range changedFlags {
			switch flag {
			case "max-line-length":
				go getMaxLineLen(path)
			case "lines":
				go getLines(getLinesParam{path, false})
			case "words":
				go getWords(path)
			case "chars":
				go getChars(path)
			case "bytes":
				go getBytes(path)
			default:
				fmt.Println("flag not found => ", flag)
			}

		}

		var result string

		for _, flag := range changedFlags {
			result = fmt.Sprintf("%s%s: %d, ", result, flag, <-mapCh[flag])
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

func getChars(fp string) {
	f, err := os.Open(fp)
	if err != nil {
		fmt.Println("getChars err:", err)
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanRunes)
	count := 0
	for scanner.Scan() {
		txt := scanner.Text()
		if txt != "" {
			count += len(txt)
		}
	}
	mapCh["chars"] <- count
}

func getWords(fp string) {

	f, err := os.Open(fp)
	if err != nil {
		fmt.Println("getWords err:", err)
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	count := 0
	re := regexp.MustCompile(`[^\s]+`)
	for scanner.Scan() {
		matchesB := re.FindAll(scanner.Bytes(), -1)
		count += len(matchesB)
	}
	mapCh["words"] <- count
}

func getBytes(fp string) {
	f, err := os.Open(fp)
	if err != nil {
		fmt.Println("getBytes err:", err)
	}
	defer f.Close()

	fS, _ := f.Stat()
	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanBytes)
	bSlice := make([]byte, 0, fS.Size())
	for scanner.Scan() {
		b := scanner.Bytes()
		bSlice = append(bSlice, b...)
	}

	mapCh["bytes"] <- len(bSlice)
}

func getMaxLineLen(fp string) {
	f, err := os.Open(fp)
	if err != nil {
		fmt.Println("getMaxLineLen err:", err)
	}
	defer f.Close()

	rd := bufio.NewReader(f)
	max := 0
	for {
		b, _, err := rd.ReadLine()
		if err != nil {
			if err == io.EOF && len(b) == 0 {
				break
			}
			fmt.Println("getMaxLineLen err readline: ", err)
		}
		lineLen := len(string(b))

		if max < lineLen {
			max = lineLen
		}

	}
	mapCh["max-line-length"] <- max
}

type getLinesParam struct {
	fp        string
	noNewLine bool
}

func getLines(p getLinesParam) {
	f, err := os.Open(p.fp)
	if err != nil {
		fmt.Println("getLines err:", err)
	}
	defer f.Close()

	rd := bufio.NewReader(f)
	count := 0
	for {
		l, err := rd.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				if len(l) != 0 && p.noNewLine {
					count++
				}
				break
			}
			fmt.Printf("read file line error: %v", err)
		}
		count++
	}
	mapCh["lines"] <- count
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

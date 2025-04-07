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

	"github.com/spf13/cobra"
)

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
		path := args[0]
		filename := filepath.Base(path)
		f, _ := os.Open(path)
		defer f.Close()

		var result []string

		if cmd.Flags().Changed("max-line-length") {
			result = append(result, fmt.Sprintf("max-line-length: %d", getMaxLineLen(f)))
		}

		if cmd.Flags().Changed("lines") {
			result = append(result, fmt.Sprintf("lines: %d", getLines(getLinesParam{f, false})))

		}

		if cmd.Flags().Changed("words") {
			result = append(result, fmt.Sprintf("words: %d", getWords(f)))
		}

		if cmd.Flags().Changed("chars") {
			result = append(result, fmt.Sprintf("chars: %d", getChars(f)))
		}

		if cmd.Flags().Changed("bytes") {
			result = append(result, fmt.Sprintf("bytes: %d", getBytes(f)))
		}

		if result == nil {
			result = append(
				result,
				fmt.Sprintf("max-line-length: %d", getMaxLineLen(f)),
				fmt.Sprintf("lines: %d", getLines(getLinesParam{f, false})),
				fmt.Sprintf("words: %d", getWords(f)),
				fmt.Sprintf("chars: %d", getChars(f)),
				fmt.Sprintf("bytes: %d", getBytes(f)))

		}

		resultStr := strings.Join(result, ", ")

		fmt.Println(fmt.Sprint(resultStr, "\t", filename))

	},
}

func getChars(f *os.File) int {
	if getCurPos(f) != 0 {
		f.Seek(0, io.SeekStart)
	}
	scanner := bufio.NewScanner(f)
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

func getWords(f *os.File) int {
	if getCurPos(f) != 0 {
		f.Seek(0, io.SeekStart)
	}
	scanner := bufio.NewScanner(f)
	count := 0
	re := regexp.MustCompile(`[^\s]+`)
	for scanner.Scan() {
		matchesB := re.FindAll(scanner.Bytes(), -1)
		count += len(matchesB)
	}
	return count
}

func getBytes(f *os.File) int {
	if getCurPos(f) != 0 {
		f.Seek(0, io.SeekStart)
	}
	fS, _ := f.Stat()
	scanner := bufio.NewScanner(f)
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

func getMaxLineLen(f *os.File) int {
	if getCurPos(f) != 0 {
		f.Seek(0, io.SeekStart)
	}
	rd := bufio.NewReader(f)
	max := 0
	for {
		b, _, err := rd.ReadLine()
		lineLen := len(string(b))

		if max < lineLen {
			max = lineLen
		}

		if err == io.EOF {
			break
		}
	}

	return max
}

type getLinesParam struct {
	f         *os.File
	noNewLine bool
}

func getLines(p getLinesParam) int {
	if getCurPos(p.f) != 0 {
		p.f.Seek(0, io.SeekStart)
	}
	rd := bufio.NewReader(p.f)
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

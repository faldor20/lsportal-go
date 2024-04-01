package main

import (
	"fmt"
	"main/lsportal"
	"os/exec"
	"regexp"
	"sync"

	"github.com/spf13/cobra"
	"github.com/tliron/commonlog"
	_ "github.com/tliron/commonlog/simple"
)

type Config struct {
	regex          string
	exclusionRegex string
	extension      string
	lsCmd          string
	lsArgs         []string
}

var config Config

var rootCmd = &cobra.Command{
	Use:   "lsportal <extension> <regex> <cmd> [-- lsArgs...]",
	Short: "LSPortal is a language server portal",
	Args:  cobra.MinimumNArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		config.extension = args[0]
		config.regex = args[1]
		config.lsCmd = args[2]
		// Find the index of "--" separator
		sepIndex := cmd.ArgsLenAtDash()
		if sepIndex != -1 {
			config.lsArgs = args[sepIndex:]
		}

		commonlog.Initialize(3, "./lsportalLog.log")

		err := validateInputs(&config)
		if err != nil {
			panic(err)
		}

		fromClient, fromInclusion := lsportal.InitForwarders(true, config.regex, config.exclusionRegex, config.extension)

		readWrite, err := lsportal.StartLanguageServer(config.lsCmd, config.lsArgs)
		if err != nil {
			panic(fmt.Errorf("error starting language server: %v", err))
		}

		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			fromClient.RunStdio()
		}()
		go func() {
			defer wg.Done()
			fromInclusion.ServeStream(readWrite, commonlog.GetLogger("fromInclusion"))
		}()
		wg.Wait()
	},
}

func init() {
	rootCmd.Flags().StringVar(&config.exclusionRegex, "exclusion", `;([\s\S]*);`, "Regular expression for exclusion")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		return
	}
}

func validateInputs(config *Config) error {

	// Validate regex
	if _, err := regexp.Compile(config.regex); err != nil {
		return fmt.Errorf("Invalid regex: %v\n", err)

	}

	// Validate cmd
	if _, err := exec.LookPath(config.lsCmd); err != nil {
		return fmt.Errorf("Command not found: %v\n", err)

	}
	return nil
}

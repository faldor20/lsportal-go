package main

import (
	"fmt"
	"main/lsportal"
	"sync"

	"github.com/tliron/commonlog"

	// Must include a backend implementation
	// See CommonLog for other options: https://github.com/tliron/commonlog
	_ "github.com/tliron/commonlog/simple"
)

const lsName = "my language"

type Config struct {
	regex          string
	exclusionRegex string
	extension      string
	lsCmd          string
	lsArgs         []string
}

func main() {
	debug := true
	commonlog.Initialize(3, "./lsportalLog.log")

	//Setup servers and connection between them

	config := Config{
		regex:          `~([\s\S]*)~`,
		exclusionRegex: `;([\s\S]*);`,
		extension:      "html",
		lsCmd:          "vscode-html-language-server",
		lsArgs:         []string{"--stdio"}}
	fromClient, fromInclusion := lsportal.InitForwarders(debug, config.regex, config.exclusionRegex, config.extension)

	//Start the other language server as a subprocess
	readWrite, err := lsportal.StartLanguageServer(config.lsCmd, config.lsArgs)
	if err != nil {
		panic(fmt.Errorf("error starting language server: %v", err))
	}
	var wg sync.WaitGroup
	//Start
	wg.Add(1)
	go func() {

		fromClient.RunStdio()
		wg.Done()
	}()
	go func() {

		fromInclusion.ServeStream(readWrite, commonlog.GetLogger("fromInclusion"))
		wg.Done()
	}()
	wg.Wait()
}

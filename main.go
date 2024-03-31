package main

import (
	"fmt"
	"main/lsportal"
	"sync"

	"github.com/tliron/commonlog"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
	"github.com/tliron/glsp/server"

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

var (
	version string = "0.0.1"
	handler protocol.Handler
)

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
	fromClient, fromInclusion := initForwarders(debug, config.regex, config.exclusionRegex, config.extension)

	//Start the other language server as a subprocess
	readWrite, err := lsportal.StartLanguageServer(config.lsCmd, config.lsArgs)
	if err != nil {
		panic(fmt.Errorf("error starting language server: %v", err))
	}
	var wg sync.WaitGroup
	//Start
	wg.Add(1)
	go func() {
		defer wg.Done()
		fromClient.RunStdio()
	}()
	go func() {
		defer wg.Done()
		fromInclusion.ServeStream(readWrite, commonlog.GetLogger("fromInclusion"))
	}()
	wg.Wait()
}

func initForwarders(debug bool, regex string, exclusionRegex string, extension string) (*server.Server, *server.Server) {
	//toInclusion
	fromClientTrans := lsportal.NewFromClientTransformer(regex, exclusionRegex, extension)
	fromClientForwarder := lsportal.ForwarderHandler{Transformer: &fromClientTrans}
	fromClient := server.NewServer(&fromClientForwarder, "toInclusion", debug)

	//client
	fromInclusionTrans := lsportal.FromInclusionTransformer{ServerTransformer: fromClientTrans}
	fromInclusionForwarder := lsportal.ForwarderHandler{Transformer: &fromInclusionTrans}
	fromInclusion := server.NewServer(&fromInclusionForwarder, "toClient", debug)

	//connect the two servers so they can send messages in between
	fromClientForwarder.InclusionServer = fromInclusion
	fromInclusionForwarder.InclusionServer = fromClient
	return fromClient, fromInclusion
}

func initialize(context *glsp.Context, params *protocol.InitializeParams) (any, error) {
	capabilities := handler.CreateServerCapabilities()

	return protocol.InitializeResult{
		Capabilities: capabilities,
		ServerInfo: &protocol.InitializeResultServerInfo{
			Name:    lsName,
			Version: &version,
		},
	}, nil
}

func initialized(context *glsp.Context, params *protocol.InitializedParams) error {
	return nil
}

func shutdown(context *glsp.Context) error {
	protocol.SetTraceValue(protocol.TraceValueOff)
	return nil
}

func setTrace(context *glsp.Context, params *protocol.SetTraceParams) error {
	protocol.SetTraceValue(params.Value)
	return nil
}

// modify the uri of the textDocument to have the extension we want
func changeUriExtension(context *glsp.Context, params *protocol.TextDocumentIdentifier) {

}

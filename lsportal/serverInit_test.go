package lsportal

import (
	"io"
	"testing"

	"github.com/sourcegraph/jsonrpc2"
	"github.com/tliron/commonlog"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

func TestInitForwarders(t *testing.T) {
	debug := true
	regex := "~(.*)~"
	exclusionRegex := ";(.*);"
	extension := "html"
	t.Log("starting test")
	commonlog.Initialize(3, "lsportalTest.log")
	commonlog.SetBackend(&testBackend{verbosity: 3, t: t})

	commonlog.GetLogger("test").Info("runnning a test")
	commonlog.GetWriter().Write([]byte("hello world"))

	// Create two pairs of pipes for bidirectional communication between the servers
	// Start serving the streams on both servers
	client, inclusion, closer := newFunction(debug, regex, exclusionRegex, extension)

	//NOTE: It is important to use jsonrpc2.VSCodeObjectCodec{} for the codecs or the data will never be read by the server
	clientRpc := jsonrpc2.NewBufferedStream(client, jsonrpc2.VSCodeObjectCodec{})
	inclusionRpc := jsonrpc2.NewBufferedStream(inclusion, jsonrpc2.VSCodeObjectCodec{})
	// clientRpc, inclusionRpc := makejsonStreams()

	// Send a message from the client to the inclusion server
	clientMessage := lspTest[fakeLspReq]{Method: "test", Params: fakeLspReq{TextDocument: protocol.TextDocumentItem{URI: "file://this/is/a.go"}}}
	go func() {
		err := clientRpc.WriteObject(clientMessage)
		if err != nil {
			t.Fatalf("Failed to write message from client: %v", err)
		}
	}()

	// Read the message on the inclusion server

	var receivedMessage lspTest[fakeLspReq]
	err := inclusionRpc.ReadObject(&receivedMessage)
	if err != nil {
		t.Fatalf("Failed to read message on inclusion server: %v", err)
	}
	//The extension should have changed
	clientMessage.Params.TextDocument.URI = "file://this/is/a.html"
	if receivedMessage != clientMessage {
		t.Errorf("Received message on inclusion server doesn't match. Got: %v, Want: %v", receivedMessage, clientMessage)
	}

	// Close the connections
	closer()

}

type lspTest[T any] struct {
	Method string `json:"method"`
	Params T      `json:"params"`
}
type fakeLspReq struct {
	TextDocument protocol.TextDocumentItem `json:"textDocument"`
}

func makejsonStreams() (jsonrpc2.ObjectStream, jsonrpc2.ObjectStream) {

	// Create two pairs of pipes for bidirectional communication between the servers
	clientWriteO, clientWriteI := io.Pipe()

	inclusionReadO, inclusionReadI := io.Pipe()

	client := struct {
		io.Reader
		io.WriteCloser
	}{inclusionReadO, clientWriteI}
	inclusion := struct {
		io.Reader
		io.WriteCloser
	}{clientWriteO, inclusionReadI}

	clientRpc := jsonrpc2.NewPlainObjectStream(client)
	inclusionRpc := jsonrpc2.NewPlainObjectStream(inclusion)
	return clientRpc, inclusionRpc
}

func newFunction(debug bool, regex string, exclusionRegex string, extension string) (io.ReadWriteCloser, io.ReadWriteCloser, func()) {
	fromClient, fromInclusion := InitForwarders(debug, regex, exclusionRegex, extension)

	// Create two pairs of pipes for bidirectional communication between the servers
	clientWriteO, clientWriteI := io.Pipe()
	clientReadO, clientReadI := io.Pipe()
	inclusionWriteO, inclusionWriteI := io.Pipe()
	inclusionReadO, inclusionReadI := io.Pipe()

	// Start serving the streams on both servers
	go fromClient.ServeStream(struct {
		io.Reader
		io.WriteCloser
	}{clientWriteO, clientReadI}, nil)

	go fromInclusion.ServeStream(struct {
		io.Reader
		io.WriteCloser
	}{inclusionWriteO, inclusionReadI}, nil)
	closer := func() {
		clientWriteO.Close()
		clientReadO.Close()
		inclusionWriteO.Close()
		inclusionReadO.Close()
	}
	client := struct {
		io.Reader
		io.WriteCloser
	}{clientReadO, clientWriteI}
	inclusion := struct {
		io.Reader
		io.WriteCloser
	}{inclusionReadO, inclusionWriteI}
	return &client, &inclusion, closer
}

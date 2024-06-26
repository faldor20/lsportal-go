package lsportal

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/tliron/commonlog"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

func validateChanges(t *testing.T, before string, after string) {
	//same length
	if len(before) != len(after) {
		t.Errorf("Expected: %q, Got: %q", len(before), len(after))
	}
	//same number of newlines
	beforeCount := strings.Count(before, "\n")
	afterCount := strings.Count(after, "\n")
	if beforeCount != afterCount {
		t.Errorf("Expected: %q, Got: %q", beforeCount, afterCount)
	}
}

func TestTransformer_Transform(t *testing.T) {
	// Create a new Transformer instance
	trans := &FromClientTransformer{
		Regex:          `(~[\s\S]*~)`,
		ExclusionRegex: `(;[\s\S]*;)`,
		Extension:      "txt",
		Documents:      make(map[string]TextDocument),
		logger:         commonlog.GetLogger("FromClientTransformer"),
	}

	// Create a test context
	context := &glsp.Context{
		Method: protocol.MethodTextDocumentDidChange,
		Params: []byte(`{
            "textDocument": {
                "uri": "file:///path/to/document.md"
            },
            "contentChanges": [
                {
                    "text": "~Hello,\n ;world;!~\n This is a test."
                }
            ]
        }`),
	}

	// Call the Transform method
	err := trans.TransformRequest(context)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Check if the document URI was changed
	var params protocol.DidChangeTextDocumentParams
	err = json.Unmarshal(context.Params, &params)
	if err != nil {
		t.Errorf("Failed to unmarshal params: %v", err)
	}
	expectedURI := "file:///path/to/document.txt"
	if params.TextDocument.URI != expectedURI {
		t.Errorf("Expected URI: %q, Got: %q", expectedURI, params.TextDocument.URI)
	}

	// Check if the document content was transformed correctly
	expectedContent := "~Hello,\n        !~\n                "
	text := params.ContentChanges[0].(protocol.TextDocumentContentChangeEventWhole).Text

	if text != expectedContent {
		t.Errorf("Expected content: %q, Got: %q", expectedContent, text)
	}
}
func TestTransform_completion(t *testing.T) {
	// Create a new Transformer instance
	trans := &FromClientTransformer{
		Regex:          `(~[\s\S]*~)`,
		ExclusionRegex: `(;[\s\S]*;)`,
		Extension:      "go",
		Documents:      make(map[string]TextDocument),
	}

	// Create a test context
	context := &glsp.Context{
		Method: protocol.MethodTextDocumentCompletion,
		Params: []byte(`{
            "textDocument": {
                "uri": "file:///path/to/document.md"
            },
        	"position":{}
        }`),
	}

	// Call the Transform method
	err := trans.TransformRequest(context)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Check if the document URI was changed
	var params protocol.DidChangeTextDocumentParams
	err = json.Unmarshal(context.Params, &params)
	if err != nil {
		t.Errorf("Failed to unmarshal params: %v", err)
	}
	expectedURI := "file:///path/to/document.go"
	if params.TextDocument.URI != expectedURI {
		t.Errorf("Expected URI: %q, Got: %q", expectedURI, params.TextDocument.URI)
	}
}

func TestTransformer_Transform_MultipleChangeEvents(t *testing.T) {
	// Create a new Transformer instance
	trans := &FromClientTransformer{
		Regex:          `(~[\s\S]*~)`,
		ExclusionRegex: `(;[\s\S]*;)`,
		Extension:      "md",
		Documents:      make(map[string]TextDocument),
		logger:         commonlog.GetLogger("FromClientTransformer"),
	}

	// Create a test context with multiple change events
	context := &glsp.Context{
		Method: protocol.MethodTextDocumentDidChange,
		Params: []byte(`{
            "textDocument": {
                "uri": "file:///path/to/document.txt",
                "version": 1
            },
            "contentChanges": [
                {
                    "range": {
                        "start": {"line": 0, "character": 0},
                        "end": {"line": 0, "character": 3}
                    },
                    "text": "Hello"
                },
                {
                    "range": {
                        "start": {"line": 1, "character": 0},
                        "end": {"line": 1, "character": 4}
                    },
                    "text": "~World"
                },
                {
                    "range": {
                        "start": {"line": 2, "character": 0},
                        "end": {"line": 2, "character": 3}
                    },
                    "text": "Test"
                }
            ]
        }`),
	}

	// Initialize the document in the transformer
	trans.Documents["file:///path/to/document.txt"] = TextDocument{
		Text: "Old text\n~Old line\nOld content~",
	}

	// Call the Transform method
	err := trans.TransformRequest(context)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Check if the document URI was changed
	var params protocol.DidChangeTextDocumentParams
	err = json.Unmarshal(context.Params, &params)
	if err != nil {
		t.Errorf("Failed to unmarshal params: %v", err)
	}
	expectedURI := "file:///path/to/document.md"
	if params.TextDocument.URI != expectedURI {
		t.Errorf("Expected URI: %q, Got: %q", expectedURI, params.TextDocument.URI)
	}

	expectedChanges := "          \n~World line\nTest content~"
	expectedContent := "Hello text\n~World line\nTest content~"

	// Check if the document content was transformed correctly

	doc, ok := trans.Documents["file:///path/to/document.txt"]
	if !ok {
		t.Errorf("Document not found in the transformer")
	}
	if doc.Text != expectedContent {
		t.Errorf("Expected content: %q, Got: %q", expectedContent, doc.Text)
	}
	// Check if the document content was transformed correctly
	change := params.ContentChanges[0].(protocol.TextDocumentContentChangeEvent).Text

	if change != expectedChanges {
		t.Errorf("Expected change: %q, Got: %q", expectedChanges, change)
	}
}

func TestTransformer_TransformResponse(t *testing.T) {
	// Create a new Transformer instance
	trans := &FromClientTransformer{
		Extension: "go",
		UriMap: map[string]string{
			"file:///path/to/doc.go": "file:///path/to/doc.txt",
		},
	}

	// Define a sample response
	response := any(map[string]interface{}{
		"textDocument": map[string]interface{}{
			"uri": "file:///path/to/doc.go",
		},
	})

	// Call the TransformResponse method
	trans.TransformResponse(&response)

	// Assert the transformed response
	expectedResponse := map[string]interface{}{
		"textDocument": map[string]interface{}{
			"uri": "file:///path/to/doc.txt",
		},
	}
	if !reflect.DeepEqual(expectedResponse, response) {
		t.Errorf("Expected transformed response: %v, but got: %v", expectedResponse, response)
	}
}

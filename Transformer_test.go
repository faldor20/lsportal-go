package lsportal

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

func TestGetOnlyInclusions(t *testing.T) {
	// Test cases
	testCases := []struct {
		name           string
		text           string
		inclusionRegex string
		exclusionRegex string
		expected       string
	}{
		{
			name:           "Simple inclusion",
			text:           "Hello, world! This is a test.",
			inclusionRegex: `(Hello|test)`,
			exclusionRegex: "",
			expected:       "Hello                   test ",
		},
		{
			name:           "Inclusion with newlines",
			text:           "Line 1\nLine 2\nLine 3",
			inclusionRegex: `(Line \d)`,
			exclusionRegex: "",
			expected:       "Line 1\nLine 2\nLine 3",
		},
		{
			name:           "Inclusion and exclusion",
			text:           "~Hello,\n ;world;!~\n This is a test.",
			inclusionRegex: `(~[\s\S]*~)`,
			exclusionRegex: `(;[\s\S]*;)`,
			expected:       "~Hello,\n        !~\n                ",
		},
		{
			name:           "No matches",
			text:           "Hello, world!",
			inclusionRegex: `(nothing)`,
			exclusionRegex: "",
			expected:       "             ",
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := whitespaceExceptInclusions(tc.text, tc.inclusionRegex, tc.exclusionRegex)
			validateChanges(t, tc.text, result)

			if result != tc.expected {
				t.Errorf("Expected: %q, Got: %q", tc.expected, result)
			}
		})
	}
}

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
	trans := &Transformer{
		regex:          `(~[\s\S]*~)`,
		exclusionRegex: `(;[\s\S]*;)`,
		extension:      "txt",
		documents:      make(map[string]TextDocument),
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
	err := trans.Transform(context)
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
	doc, ok := trans.documents[params.TextDocument.URI]
	if !ok {
		t.Errorf("Document not found in the transformer")
	}
	if doc.Text != expectedContent {
		t.Errorf("Expected content: %q, Got: %q", expectedContent, doc.Text)
	}
}

func TestTransformer_Transform_MultipleChangeEvents(t *testing.T) {
	// Create a new Transformer instance
	trans := &Transformer{
		regex:          `(~[\s\S]*~)`,
		exclusionRegex: `(;[\s\S]*;)`,
		extension:      "md",
		documents:      make(map[string]TextDocument),
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
                        "end": {"line": 1, "character": 3}
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
	trans.documents["file:///path/to/document.md"] = TextDocument{
		Text: "Old text\n~Old line\nOld content~",
	}

	// Call the Transform method
	err := trans.Transform(context)
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

	// Check if the document content was transformed correctly
	expectedContent := "~Hello text\nWorld line\nTest content~"
	doc, ok := trans.documents[params.TextDocument.URI]
	if !ok {
		t.Errorf("Document not found in the transformer")
	}
	if doc.Text != expectedContent {
		t.Errorf("Expected content: %q, Got: %q", expectedContent, doc.Text)
	}
}

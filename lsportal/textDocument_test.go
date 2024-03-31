package lsportal

import (
	"testing"

	. "github.com/tliron/glsp/protocol_3_16"
)

// Useful for converting a specific []A into an []any
func deepCopyArray[T any](a []T, b []any) []any {
	for i, v := range a {
		b[i] = v
	}
	return b
}
func TestApplychanges(t *testing.T) {
	// Initialize a TextDocument
	doc := TextDocument{Text: "Hello, World!", URI: "test.txt"}

	// Define the changes
	changes := []TextDocumentContentChangeEvent{
		{
			Range: &Range{
				Start: Position{0, 7},
				End:   Position{0, 12},
			},
			Text: "Gophers",
		},
	}

	// Define the params
	params := &DidChangeTextDocumentParams{
		TextDocument: VersionedTextDocumentIdentifier{
			TextDocumentIdentifier: TextDocumentIdentifier{URI: doc.URI},
		},
		ContentChanges: make([]interface{}, len(changes)),
	}

	//write the any type in
	deepCopyArray(changes, params.ContentChanges)

	// Apply the changes
	newDoc, err := doc.applychanges(params)
	if err != nil {
		t.Fatalf("Applychanges returned an error: %v", err)
	}

	// Check that the content was updated correctly
	expectedContent := "Hello, Gophers!"
	if newDoc.Text != expectedContent {
		t.Errorf("Expected content to be %s, but got %s", expectedContent, newDoc.Text)
	}
}

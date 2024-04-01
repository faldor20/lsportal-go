package lsportal

import (
	"bytes"
	"fmt"
	"strings"

	. "github.com/tliron/glsp/protocol_3_16"
)

type TextDocument struct {
	Text       string
	URI        URI
	Inclusions []Range
}

func (textDocument TextDocument) UpdateAndGetChanges(params DidChangeTextDocumentParams, regex string, exclusionRegex string) (TextDocument, DidChangeTextDocumentParams, error) {
	newDoc, err := textDocument.applychanges(&params)
	if err != nil {
		return textDocument, params, err
	}
	isolatedText, inclusions := whitespaceExceptInclusions(newDoc.Text, regex, exclusionRegex)
	//replace any content not in inclusions with whitespace
	inclusionDoc := TextDocument{
		URI:  textDocument.URI,
		Text: isolatedText,
	}
	newDoc.Inclusions = inclusions
	//update the content changes to reflect the whitespaced textDocument
	params.ContentChanges = inclusionDoc.NewChangeEventText(&params)

	return newDoc, params, nil
}

func (textDocument TextDocument) applychanges(params *DidChangeTextDocumentParams) (TextDocument, error) {
	ret, err := handleWholeOrPartialChanges(params,
		func(changes []TextDocumentContentChangeEvent) (*TextDocument, error) {
			newContent, err := textDocument.applyChangeEvents(changes)

			if err != nil {
				return &textDocument, err
			}
			textDocument.Text = string(newContent)

			return &textDocument, nil
		},
		func(change TextDocumentContentChangeEventWhole) (*TextDocument, error) {
			textDocument.Text = change.Text
			return &textDocument, nil
		},
	)
	return *ret, err
}

func (text TextDocument) applyChangeEvents(changes []TextDocumentContentChangeEvent) ([]byte, error) {
	content := []byte(text.Text)
	for _, change := range changes {

		// TODO the gopls code says you can use their diff type for this , worth looking into

		if change.Range == nil {
			return nil, fmt.Errorf(" unexpected nil range for change")
		}
		start, end := change.Range.IndexesIn(content)

		if end < start {
			return nil, fmt.Errorf("invalid range for content change")
		}
		var buf bytes.Buffer
		buf.Write(content[:start])
		buf.WriteString(change.Text)
		buf.Write(content[end:])
		content = buf.Bytes()

	}
	return content, nil
}

func handleWholeOrPartialChanges[T any](params *DidChangeTextDocumentParams, handlePartial func([]TextDocumentContentChangeEvent) (T, error), handleWhole func(TextDocumentContentChangeEventWhole) (T, error)) (T, error) {
	//If we only recieve a single param and it's a whole document change, we can just replace the whole thing
	length := len(params.ContentChanges)
	if length >= 1 {

		switch params.ContentChanges[length-1].(type) {
		case TextDocumentContentChangeEventWhole:
			return handleWhole(params.ContentChanges[length-1].(TextDocumentContentChangeEventWhole))
		}
	}
	var changes []TextDocumentContentChangeEvent

	for i := range params.ContentChanges {
		switch change := params.ContentChanges[i].(type) {
		case TextDocumentContentChangeEvent:
			changes = append(changes, change)

		}
	}
	return handlePartial(changes)
}

// Create new change events from the text within the text document
func (doc TextDocument) NewChangeEventText(params *DidChangeTextDocumentParams) []any {
	ret, _ := handleWholeOrPartialChanges(params,
		func(changes []TextDocumentContentChangeEvent) ([]any, error) {
			// We are just faking this by sending the whole document each time, we should probably be able to
			// figure out the correct range
			// that covers all changes made by the list of changes we got so we can send a single change event.
			return []any{makeFullDocumentChange(doc)}, nil
		},
		func(change TextDocumentContentChangeEventWhole) ([]any, error) {
			change.Text = doc.Text
			return []any{change}, nil
		},
	)
	return ret
}

// Gets the substring of the text from the range supplied
func textFromRange(text []byte, range_ *Range) []byte {
	start, end := range_.IndexesIn(text)
	return text[start:end]
}

// Returns a change event for the entire document
// This can be used so we don't have to bother with incremental changes
func makeFullDocumentChange(doc TextDocument) TextDocumentContentChangeEvent {
	lastLineLength := len(doc.Text) - strings.LastIndex(doc.Text, "\n")
	return TextDocumentContentChangeEvent{
		Range: &Range{
			Start: Position{Line: 0, Character: 0},
			End:   Position{Line: uint32(strings.Count(doc.Text, "\n")), Character: uint32(lastLineLength)},
		},
		Text: doc.Text,
	}

}

//TODO: We should try to generate a correct changeEvent for each changeEvent coming in, this would require us the run the inclusion extraction code after every change,
// then copy the change text from the document and put that into the event
// This cannot be done at the end because changes can delete content

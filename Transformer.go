package lsportal

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/tliron/glsp"
	. "github.com/tliron/glsp/protocol_3_16"
)

type Transformer struct {
	regex          string
	exclusionRegex string
	extension      string
	documents      map[string]TextDocument
}

func (trans *Transformer) Transform(context *glsp.Context) error {
	switch context.Method {
	case MethodTextDocumentHover:
		runParamsTransform(context, func(params *HoverParams) error {
			params.TextDocument.URI = trans.changeExtension(params.TextDocument.URI)
			return nil
		})
	case MethodTextDocumentCompletion:
		runParamsTransform(context, func(params *CompletionParams) error {
			params.TextDocument.URI = trans.changeExtension(params.TextDocument.URI)
			return nil
		})
	case MethodTextDocumentDidChange:
		runParamsTransform(context, func(params *DidChangeTextDocumentParams) error {
			params.TextDocument.URI = trans.changeExtension(params.TextDocument.URI)

			newDoc, newParams, err := trans.documents[params.TextDocument.URI].UpdateAndGetChanges(*params, trans.regex, trans.exclusionRegex)
			//TODO: figure out error handling
			if err != nil {
				return fmt.Errorf("Error applying changes to document: %v", err)
			}
			trans.documents[params.TextDocument.URI] = newDoc
			params = &newParams
			return nil

		})

	}
	return nil
}

// Process the text of the forwarder to replace anything except newlines not within the regexs with a space
// inclusionRegex: A multiline regex that should match the text you want to keep within its first match group, it is expected to match many times
// exclusionRegex: A multiline regex that should match text you want remove from within an inclusion
func whitespaceExceptInclusions(text string, inclusionRegex string, exclusionRegex string) string {
	// Compile the inclusion regex
	incRegex := regexp.MustCompile(inclusionRegex)

	// Convert the text to a rune slice
	runes := []rune(text)

	// Create a slice to store the result
	result := make([]rune, len(runes))

	// Initialize the result slice with spaces
	for i := range result {
		if runes[i] == '\n' {
			result[i] = '\n'
		} else {
			result[i] = ' '
		}
	}

	// Find all matches of the inclusion regex
	matches := incRegex.FindAllStringSubmatchIndex(text, -1)

	// Iterate over the matches
	for _, match := range matches {
		// Check if there is a capturing group
		if len(match) >= 4 {
			start, end := match[2], match[3]
			// Copy the captured text to the result slice
			copy(result[start:end], runes[start:end])
		}
	}

	// Convert the result slice back to a string
	processedText := string(result)

	// Compile the exclusion regex if provided
	if exclusionRegex != "" {
		excRegex := regexp.MustCompile(exclusionRegex)
		// Replace any matches of the exclusion regex with spaces
		processedText = excRegex.ReplaceAllStringFunc(processedText, func(match string) string {
			return strings.Map(func(r rune) rune {
				if r == '\n' {
					return '\n'
				}
				return ' '
			}, match)
		})
	}

	return processedText
}

// unmarshals into your format
func runParamsTransform[P any](context *glsp.Context, transform func(params *P) error) error {
	params := new(P)
	err := json.Unmarshal(context.Params, &params)
	if err != nil {
		return err
	}
	//Perform the transformation
	err = transform(params)
	if err != nil {
		return err
	}

	newParams, err := json.Marshal(params)
	if err != nil {
		return err
	}
	context.Params = newParams
	return nil
}

func (trans *Transformer) changeExtension(uri URI) string {
	strs := strings.Split(uri, ".")
	strs[len(strs)-1] = trans.extension
	return strings.Join(strs, ".")
}

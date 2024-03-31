package lsportal

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/tliron/commonlog"
	"github.com/tliron/glsp"
	. "github.com/tliron/glsp/protocol_3_16"
)

type Transformer interface {
	TransformRequest(context *glsp.Context) error
	TransformResponse(response *any)
}

// Proves that ServerTransformer implements Transformer
var _ Transformer = &FromClientTransformer{}

type FromClientTransformer struct {
	logger         commonlog.Logger
	Regex          string
	ExclusionRegex string
	Extension      string
	UriMap         map[string]string
	Documents      map[string]TextDocument
}

// New
func NewFromClientTransformer(regex string, exclusionRegex string, extension string) FromClientTransformer {
	return FromClientTransformer{
		Regex:          regex,
		ExclusionRegex: exclusionRegex,
		Extension:      extension,
		UriMap:         make(map[string]string),
		Documents:      make(map[string]TextDocument),
		logger:         commonlog.GetLogger("FromClientTransformer")}
}

// Transform requests from the client so that the inclusion server is happy
func (trans *FromClientTransformer) TransformRequest(context *glsp.Context) error {
	switch context.Method {
	case MethodTextDocumentDidChange:
		runParamsTransform(context, func(params *DidChangeTextDocumentParams) error {
			originalUri := params.TextDocument.URI
			params.TextDocument.URI = trans.changeExtension(params.TextDocument.URI)

			newDoc, newParams, err := trans.Documents[originalUri].UpdateAndGetChanges(*params, trans.Regex, trans.ExclusionRegex)
			trans.logger.Debugf("Updated document: %s", newDoc)
			//TODO: figure out error handling
			if err != nil {
				return fmt.Errorf("Error applying changes to document: %v", err)
			}
			//Newdoc has the changes applied but doesn't have the inclusions isolated
			trans.Documents[originalUri] = newDoc
			params.ContentChanges = newParams.ContentChanges
			return nil

		})
	case MethodTextDocumentDidOpen:
		runParamsTransform(context, func(params *DidOpenTextDocumentParams) error {
			originalUri := params.TextDocument.URI

			params.TextDocument.URI = trans.changeExtension(originalUri)
			//We need to save this so we can change the URI back to the original in the response
			trans.UriMap[params.TextDocument.URI] = originalUri
			trans.Documents[originalUri] = TextDocument{
				Text: params.TextDocument.Text,
				URI:  params.TextDocument.URI,
			}
			trans.logger.Debugf("Added document: %s", originalUri)
			return nil
		})
	default:
		runParamsTransform(context, func(params *any) error {
			params2 := (*params).(map[string]interface{})
			if reqMap, ok := params2["textDocument"].(map[string]interface{}); ok {
				// reqMap is the object in "request: {...}"
				if uri, ok := reqMap["uri"].(string); ok {
					// uri is the URI of the document
					reqMap["uri"] = trans.changeExtension(uri)
				}
			}
			return nil
		})

	}
	return nil
}

// Transfrom Responses from the inclusion server so that they are recognizable by the client
func (trans *FromClientTransformer) TransformResponse(response *any) {

	//Change url back to original
	if *response == nil {
		return
	}
	response2 := (*response).(map[string]interface{})
	if reqMap, ok := response2["textDocument"].(map[string]interface{}); ok {
		// reqMap is the object in "request: {...}"
		if uri, ok := reqMap["uri"].(string); ok {
			if strings.HasSuffix(uri, trans.Extension) {
				// uri is the URI of the document
				reqMap["uri"] = trans.UriMap[uri]

			}
		}
	}

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

func (trans *FromClientTransformer) changeExtension(uri URI) string {
	strs := strings.Split(uri, ".")
	strs[len(strs)-1] = trans.Extension
	return strings.Join(strs, ".")
}

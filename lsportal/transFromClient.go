package lsportal

import (
	"encoding/json"
	"fmt"
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
			var foundUri string
			if reqMap, ok := params2["textDocument"].(map[string]interface{}); ok {
				// reqMap is the object in "request: {...}"
				if uri, ok := reqMap["uri"].(string); ok {
					// uri is the URI of the document
					foundUri = uri
					reqMap["uri"] = trans.changeExtension(uri)
				}
			}
			//check to make sure we are within the inclusion area
			if foundUri != "" {
				if position, ok := params2["position"].(map[string]interface{}); ok {
					if position["line"] != nil && position["character"] != nil {
						line := uint32(position["line"].(float64))
						character := uint32(position["character"].(float64))
						//find if the position is within an inclusion
						for _, inclusion := range trans.Documents[foundUri].Inclusions {
							if isInRange(inclusion, Position{Line: line, Character: character}) {
								return nil
							}
						}
						trans.logger.Infof("Request from outside of inclusions: %v", trans.Documents[foundUri].Inclusions)
						return fmt.Errorf("Request from outside of inclusion: %v", trans.Documents[foundUri].Inclusions)
					}
				}
			}
			return nil
		})

	}
	return nil
}
func isInRange(r Range, pos Position) bool {
	if pos.Line < r.Start.Line || pos.Line > r.End.Line {
		return false
	}
	if pos.Line == r.Start.Line && pos.Character < r.Start.Character {
		return false
	}
	if pos.Line == r.End.Line && pos.Character > r.End.Character {
		return false
	}
	return true
}

// Transfrom Responses from the inclusion server so that they are recognizable by the client
func (trans *FromClientTransformer) TransformResponse(response *any) {

	//Change url back to original
	switch (*response).(type) {
	case map[string]interface{}:
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
	default:
		return
	}

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

package lsportal

import "github.com/tliron/glsp"

// This transformer specifically handles messages from client to the server
type FromInclusionTransformer struct {
	ServerTransformer FromClientTransformer
}

//TODO: the server and client transformer should be symmetric and their contets should be interchangable

// Transform requests from the inclusionServer so that the client is happy
func (trans *FromInclusionTransformer) TransformRequest(context *glsp.Context) error {
	return runParamsTransform(context, func(params *any) error {
		params2 := (*params).(map[string]interface{})
		if reqMap, ok := params2["textDocument"].(map[string]interface{}); ok {
			// reqMap is the object in "request: {...}"
			if uri, ok := reqMap["uri"].(string); ok {
				// uri is the URI of the document

				//TODO:this is likely wrong
				reqMap["uri"] = trans.ServerTransformer.UriMap[uri]
			}
		}
		return nil
	})

}

// Transform responses from the client so that the inclusion server is happy
func (trans *FromInclusionTransformer) TransformResponse(response any) any {
	//Change url back to original
	response2 := response.(map[string]interface{})
	if reqMap, ok := response2["textDocument"].(map[string]interface{}); ok {
		// reqMap is the object in "request: {...}"
		if uri, ok := reqMap["uri"].(string); ok {
			// uri is the URI of the document
			//TODO:This is likely wrong
			reqMap["uri"] = trans.ServerTransformer.changeExtension(uri)

		}
	}

	return response
}

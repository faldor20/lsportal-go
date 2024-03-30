package lsportal

// Forwarder handles sending all requests and responses from the client to our inclusion-server
// and vice versa.
// We will also run the transformer over all messages so we can adjust content and uris and such so that the inclusion-server
// thinks it's operating on a file of it's own language

import (
	"github.com/tliron/glsp"
	"github.com/tliron/glsp/server"
)

type MyHandler struct {
	// Base Protocol
	transformer     Transformer
	inclusionServer *server.Server
}

// ([glsp.Handler] interface)
func (self *MyHandler) Handle(context *glsp.Context) (r any, validMethod bool, validParams bool, err error) {

	//special case for exit
	if context.Method == "exit" {
		return nil, true, true, nil
	}
	//forward to transformer+
	self.transformer.Transform(context)
	res, err := self.forwardMessage(context)

	//TODO: return proper params and method validation
	return res, true, true, err

}

func (self *MyHandler) forwardMessage(context *glsp.Context) (any, error) {
	res, err := self.inclusionServer.Connection.CallRaw(context.Context, context.Method, context.Params)
	if err != nil {
		return nil, err
	}
	return res, err
}

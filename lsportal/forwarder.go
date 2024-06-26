package lsportal

// Forwarder handles sending all requests and responses from the client to our inclusion-server
// and vice versa.
// We will also run the transformer over all messages so we can adjust content and uris and such so that the inclusion-server
// thinks it's operating on a file of it's own language

import (
	contextpkg "context"
	"time"

	"github.com/tliron/commonlog"
	"github.com/tliron/glsp"
	"github.com/tliron/glsp/server"
)

type ForwarderHandler struct {
	logger commonlog.Logger
	// Base Protocol
	Transformer Transformer
	otherServer *server.Server
}

// Proves that ForwarderHandler implements glsp.Handler
var _ glsp.Handler = &ForwarderHandler{}

// TODO: I should make sure the handler is running in a goroutine
// ([glsp.Handler] interface)
func (self *ForwarderHandler) Handle(context *glsp.Context) (r any, validMethod bool, validParams bool, err error) {

	//special case for exit
	if context.Method == "exit" {
		return nil, true, true, nil
	}
	//forward to transformer+
	self.Transformer.TransformRequest(context)
	res, err := self.forwardMessage(context)
	if err != nil {
		self.logger.Errorf("error forwarding message: %v", err)
		return nil, true, true, err
	}

	//this means we sent a request with a response
	if *res != nil {
		self.Transformer.TransformResponse(res)
		//TODO: return proper params and method validation
		return res, true, true, nil
	}
	return nil, true, true, nil

}

func (self *ForwarderHandler) forwardMessage(context *glsp.Context) (*any, error) {

	var res any
	var err error
	ctx, cancel := contextpkg.WithTimeout(context.Context, 2*time.Second)
	defer cancel()
	if context.Notification {

		err = self.otherServer.Connection.Notify(ctx, context.Method, context.Params)
		res = nil

	} else {
		err = self.otherServer.Connection.Call(ctx, context.Method, context.Params, &res)
	}
	if err != nil {
		return nil, err
	}
	//unmarshal to any

	return &res, err
}

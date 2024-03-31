package lsportal

import (
	"github.com/tliron/commonlog"
	"github.com/tliron/glsp/server"
)

func InitForwarders(debug bool, regex string, exclusionRegex string, extension string) (*server.Server, *server.Server) {
	//toInclusion
	fromClientTrans := NewFromClientTransformer(regex, exclusionRegex, extension)
	fromClientForwarder := ForwarderHandler{Transformer: &fromClientTrans, logger: commonlog.GetLogger("fromClientForwader")}
	fromClient := server.NewServer(&fromClientForwarder, "fromCLient", debug)

	//client
	fromInclusionTrans := FromInclusionTransformer{ServerTransformer: &fromClientTrans}
	fromInclusionForwarder := ForwarderHandler{Transformer: &fromInclusionTrans, logger: commonlog.GetLogger("fromInclusionForwader")}
	fromInclusion := server.NewServer(&fromInclusionForwarder, "fromInclusion", debug)

	//connect the two servers so they can send messages in between
	fromClientForwarder.otherServer = fromInclusion
	fromInclusionForwarder.otherServer = fromClient
	return fromClient, fromInclusion
}

package main

import (
	"io"
	"testing"
)

func TestInitForwarders(t *testing.T) {
	debug := true
	regex := ".*\\.go"
	exclusionRegex := "_test\\.go"
	extension := ".go"

	fromClient, fromInclusion := initForwarders(debug, regex, exclusionRegex, extension)

	// Create pipes for communication between the servers
	clientToInclusion, inclusionToClient := io.Pipe()

	// Start serving the streams on both servers
	go fromClient.ServeStream(clientToInclusion, nil)
	go fromInclusion.ServeStream(inclusionToClient, nil)

	// Send a message from the client to the inclusion server
	clientMessage := "Hello from client"
	_, err := clientToInclusion.Write([]byte(clientMessage))
	if err != nil {
		t.Fatalf("Failed to write message from client: %v", err)
	}

	// Read the message on the inclusion server
	inclusionBuffer := make([]byte, len(clientMessage))
	_, err = inclusionToClient.Read(inclusionBuffer)
	if err != nil {
		t.Fatalf("Failed to read message on inclusion server: %v", err)
	}
	receivedMessage := string(inclusionBuffer)
	if receivedMessage != clientMessage {
		t.Errorf("Received message on inclusion server doesn't match. Got: %s, Want: %s", receivedMessage, clientMessage)
	}

	// Send a message from the inclusion server to the client
	inclusionMessage := "Hello from inclusion"
	_, err = inclusionToClient.Write([]byte(inclusionMessage))
	if err != nil {
		t.Fatalf("Failed to write message from inclusion server: %v", err)
	}

	// Read the message on the client
	clientBuffer := make([]byte, len(inclusionMessage))
	_, err = clientToInclusion.Read(clientBuffer)
	if err != nil {
		t.Fatalf("Failed to read message on client: %v", err)
	}
	receivedMessage = string(clientBuffer)
	if receivedMessage != inclusionMessage {
		t.Errorf("Received message on client doesn't match. Got: %s, Want: %s", receivedMessage, inclusionMessage)
	}

	// Close the connections
	clientToInclusion.Close()
	inclusionToClient.Close()
}

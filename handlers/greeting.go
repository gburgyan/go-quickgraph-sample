package handlers

import "github.com/gburgyan/go-quickgraph"

type GreetingResponse struct {
	Greeting string `graphy:"greeting,description=The greeting message"`
}

// GraphTypeExtension implements the GraphTypeExtension interface for GreetingResponse
func (GreetingResponse) GraphTypeExtension() quickgraph.GraphTypeInfo {
	return quickgraph.GraphTypeInfo{
		Description: "Response containing a personalized greeting message",
	}
}

func Greeting(name string) GreetingResponse {
	return GreetingResponse{
		Greeting: "Hello, " + name,
	}
}

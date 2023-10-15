package handlers

type GreetingResponse struct {
	Greeting string
}

func Greeting(name string) GreetingResponse {
	return GreetingResponse{
		Greeting: "Hello, " + name,
	}
}

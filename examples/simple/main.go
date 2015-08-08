package main

import (
	"log"
	"net/http"

	"github.com/codykrieger/jeeves"
	"github.com/codykrieger/jeeves/ask"
)

func main() {
	http.Handle("/skills/hello", jeeves.RegisterSkill(&jeeves.Skill{
		ApplicationID: "amzn1.echo-sdk-ams.app.000000-d0ed-0000-ad00-000000d00ebe",
		Handler:       helloHandler,
	}))

	log.Println("Listening...")
	log.Fatal(http.ListenAndServe(":3000", nil))
}

func helloHandler(skill *jeeves.Skill, req *ask.Request) *ask.Response {
	resp := ask.NewResponse(req)

	if req.IsLaunchRequest() {
		resp.Body.OutputSpeech = ask.NewOutputSpeech("Hello there!")
	} else if req.IsIntentRequest() {
	} else if req.IsSessionEndedRequest() {
	}

	return resp
}

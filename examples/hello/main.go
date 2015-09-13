package main

import (
	"log"
	"net/http"
	"os"

	"github.com/codykrieger/jeeves"
	"github.com/codykrieger/jeeves/ask"
)

func main() {
	http.Handle("/skills/hello", jeeves.RegisterSkill(&jeeves.Skill{
		ApplicationID: os.Getenv("ASK_APP_ID"), // e.g. "amzn1.echo-sdk-ams.app.000000-d0ed-0000-ad00-000000d00ebe"
		Handler:       helloHandler,
	}))

	log.Println("Listening...")
	log.Fatal(http.ListenAndServe(":3000", nil))
}

func helloHandler(skill *jeeves.Skill, req *ask.Request) *ask.Response {
	resp := ask.NewResponse(req)

	if req.IsLaunchRequest() {
		log.Println("Launch request")
		resp.Body.OutputSpeech = ask.NewOutputSpeech("Your friendly neighborhood hello service is ready for commands.")
	} else if req.IsIntentRequest() {
		intentName := req.Body.Intent.Name
		log.Printf("Intent request: %v", intentName)

		switch intentName {
		case "SayHello":
			resp.Body.OutputSpeech = ask.NewOutputSpeech("Hi there!")
			resp.Body.Card = ask.NewCard("Hi there", "You asked me to say hello.")
		default:
			resp.Body.OutputSpeech = ask.NewOutputSpeech("Unknown command.")
		}
	} else if req.IsSessionEndedRequest() {
		log.Println("Session Ended request!")
	}

	return resp
}

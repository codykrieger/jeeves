package ask

import (
	"encoding/json"
	"io"
)

type Session struct {
	New         bool   `json:"new"`
	SessionID   string `json:"sessionId"`
	Application struct {
		ApplicationID string `json:"applicationId"`
	} `json:"application"`
	Attributes map[string]interface{} `json:"attributes"`
	User       struct {
		UserID string `json:"userId"`
	} `json:"user"`
}

type Intent struct {
	Name  string          `json:"name"`
	Slots map[string]Slot `json:"slots"`
}

type Request struct {
	Version string   `json:"version"`
	Session *Session `json:"session"`
	Body    struct {
		Type      string `json:"type"`
		RequestID string `json:"requestId"`
		Timestamp string `json:"timestamp"`
		Intent    Intent `json:"intent,omitempty"`
		Reason    string `json:"reason,omitempty"`
	} `json:"request"`
}

type ResponseBody struct {
	OutputSpeech     *OutputSpeech `json:"outputSpeech,omitempty"`
	Card             *Card         `json:"card,omitempty"`
	Reprompt         *Reprompt     `json:"reprompt,omitempty"`
	ShouldEndSession bool          `json:"shouldEndSession"`
}

type Response struct {
	Version           string                 `json:"version"`
	SessionAttributes map[string]interface{} `json:"sessionAttributes,omitempty"`
	Body              *ResponseBody          `json:"response"`
}

type OutputSpeech struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type Card struct {
	Type    string `json:"type"`
	Title   string `json:"title,omitempty"`
	Content string `json:"content,omitempty"`
}

type Slot struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Reprompt struct {
	OutputSpeech *OutputSpeech `json:"outputSpeech,omitempty"`
}

// Card methods

func NewCard(title string, content string) *Card {
	return &Card{
		// As of v1.0 of the ASK, "Simple" is the only supported Card type.
		Type:    "Simple",
		Title:   title,
		Content: content,
	}
}

// OutputSpeech methods

func NewOutputSpeech(text string) *OutputSpeech {
	return &OutputSpeech{
		// As of v1.0 of the ASK, "PlainText" is the only supported OutputSpeech type.
		Type: "PlainText",
		Text: text,
	}
}

// Reprompt methods

func NewReprompt(text string) *Reprompt {
	return &Reprompt{
		OutputSpeech: NewOutputSpeech(text),
	}
}

// Request methods

func NewRequestFromJSON(reader io.Reader) (*Request, error) {
	var req *Request
	if err := json.NewDecoder(reader).Decode(&req); err != nil {
		return nil, err
	}
	return req, nil
}

func (r *Request) IsLaunchRequest() bool {
	return r.Body.Type == "LaunchRequest"
}

func (r *Request) IsIntentRequest() bool {
	return r.Body.Type == "IntentRequest"
}

func (r *Request) IsSessionEndedRequest() bool {
	return r.Body.Type == "SessionEndedRequest"
}

func (r *Request) SessionTerminationWasUserInitiated() bool {
	return r.Body.Reason == "USER_INITIATED"
}

func (r *Request) SessionTerminationIsDueToError() bool {
	return r.Body.Reason == "ERROR"
}

func (r *Request) SessionTerminationIsDueToMaxRepromptLimitExceeded() bool {
	return r.Body.Reason == "EXCEEDED_MAX_REPROMPTS"
}

func (r *Request) String() (string, error) {
	bytes, err := json.Marshal(r)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// Response methods

func NewResponse(req *Request) *Response {
	return &Response{
		Version:           "1.0",
		SessionAttributes: req.Session.Attributes,
		Body: &ResponseBody{
			ShouldEndSession: true,
		},
	}
}

func (r *Response) String() (string, error) {
	bytes, err := json.Marshal(r)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (r *Response) Bytes() ([]byte, error) {
	bytes, err := json.Marshal(r)
	if err != nil {
		return []byte{}, err
	}
	return bytes, nil
}

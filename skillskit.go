package jeeves

import (
	"encoding/json"
	"io"
)

type ASKSession struct {
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

type ASKIntent struct {
	Name  string             `json:"name"`
	Slots map[string]ASKSlot `json:"slots"`
}

type ASKRequest struct {
	Version string      `json:"version"`
	Session *ASKSession `json:"session"`
	Body    struct {
		Type      string    `json:"type"`
		RequestID string    `json:"requestId"`
		Timestamp string    `json:"timestamp"`
		Intent    ASKIntent `json:"intent,omitempty"`
		Reason    string    `json:"reason,omitempty"`
	} `json:"request"`
}

type ASKResponseBody struct {
	OutputSpeech     *ASKOutputSpeech `json:"outputSpeech,omitempty"`
	Card             *ASKCard         `json:"card,omitempty"`
	Reprompt         *ASKReprompt     `json:"reprompt,omitempty"`
	ShouldEndSession bool             `json:"shouldEndSession"`
}

type ASKResponse struct {
	Version           string                 `json:"version"`
	SessionAttributes map[string]interface{} `json:"sessionAttributes,omitempty"`
	Body              *ASKResponseBody       `json:"response"`
}

type ASKOutputSpeech struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type ASKCard struct {
	Type    string `json:"type"`
	Title   string `json:"title,omitempty"`
	Content string `json:"content,omitempty"`
}

type ASKSlot struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type ASKReprompt struct {
	OutputSpeech *ASKOutputSpeech `json:"outputSpeech,omitempty"`
}

// Card methods

func NewASKCard(title string, content string) *ASKCard {
	return &ASKCard{
		// As of v1.0 of the ASK, "Simple" is the only supported Card type.
		Type:    "Simple",
		Title:   title,
		Content: content,
	}
}

// OutputSpeech methods

func NewASKOutputSpeech(text string) *ASKOutputSpeech {
	return &ASKOutputSpeech{
		// As of v1.0 of the ASK, "PlainText" is the only supported OutputSpeech type.
		Type: "PlainText",
		Text: text,
	}
}

// Reprompt methods

func NewASKReprompt(text string) *ASKReprompt {
	return &ASKReprompt{
		OutputSpeech: NewASKOutputSpeech(text),
	}
}

// Request methods

func NewASKRequestFromJSON(reader io.Reader) (*ASKRequest, error) {
	var req *ASKRequest
	if err := json.NewDecoder(reader).Decode(&req); err != nil {
		return nil, err
	}
	return req, nil
}

func (r *ASKRequest) IsLaunchRequest() bool {
	return r.Body.Type == "LaunchRequest"
}

func (r *ASKRequest) IsIntentRequest() bool {
	return r.Body.Type == "IntentRequest"
}

func (r *ASKRequest) IsSessionEndedRequest() bool {
	return r.Body.Type == "SessionEndedRequest"
}

func (r *ASKRequest) SessionTerminationWasUserInitiated() bool {
	return r.Body.Reason == "USER_INITIATED"
}

func (r *ASKRequest) SessionTerminationIsDueToError() bool {
	return r.Body.Reason == "ERROR"
}

func (r *ASKRequest) SessionTerminationIsDueToMaxRepromptLimitExceeded() bool {
	return r.Body.Reason == "EXCEEDED_MAX_REPROMPTS"
}

func (r *ASKRequest) String() (string, error) {
	bytes, err := json.Marshal(r)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// Response methods

func NewASKResponse(req *ASKRequest) *ASKResponse {
	return &ASKResponse{
		Version:           "1.0",
		SessionAttributes: req.Session.Attributes,
		Body: &ASKResponseBody{
			ShouldEndSession: true,
		},
	}
}

func (r *ASKResponse) String() (string, error) {
	bytes, err := json.Marshal(r)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (r *ASKResponse) Bytes() ([]byte, error) {
	bytes, err := json.Marshal(r)
	if err != nil {
		return []byte{}, err
	}
	return bytes, nil
}

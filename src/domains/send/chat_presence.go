package send

// ChatPresenceRequest represents a request to send chat presence (typing indicator)
// Action: "start" (typing), "stop" (stop typing), "recording" (recording audio)
type ChatPresenceRequest struct {
	BaseRequest
	Phone  string `json:"phone" validate:"required"`
	Action string `json:"action" validate:"required"`
}

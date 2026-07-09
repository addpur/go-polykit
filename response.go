package polykit

// StandardResponse is the canonical structure for all outgoing responses across protocols.
type StandardResponse struct {
	ResponseCode string      `json:"response_code"`
	Message      string      `json:"message"`
	Data         interface{} `json:"data,omitempty"`
}

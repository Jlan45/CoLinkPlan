package protocol

// MessageType defines the type of message sent over WebSocket
type MessageType string

const (
	MsgTypeRegister MessageType = "REGISTER"
	MsgTypeCall     MessageType = "CALL"
	MsgTypeStream   MessageType = "STREAM"
	MsgTypeError    MessageType = "ERROR"
	MsgTypeFinish   MessageType = "FINISH"
)

// WSPayload represents the base structure for WebSocket communication
type WSPayload struct {
	Type MessageType `json:"type"`
	Data interface{} `json:"data"`
}

// RegisterData is sent by the client upon connection
type RegisterData struct {
	MaxParallel int      `json:"max_parallel"`
	Models      []string `json:"models"` // e.g., ["pro-model", "ultra-model"]
}

// CallData is sent by the server to the client
type CallData struct {
	RequestID string      `json:"request_id"`
	Model     string      `json:"model"`
	Payload   interface{} `json:"payload"` // OpenAI API ChatCompletion payload mapped as interface{}
}

// StreamData is sent by the client back to the server
type StreamData struct {
	RequestID string      `json:"request_id"`
	Chunk     interface{} `json:"chunk"` // OpenAI API ChatCompletion stream chunk map
}

// ErrorData is sent by either side upon an error
type ErrorData struct {
	RequestID string `json:"request_id,omitempty"` // Empty if connection-level error
	Code      int    `json:"code"`
	Message   string `json:"message"`
}

// FinishData is sent by the client when a stream is complete
type FinishData struct {
	RequestID string `json:"request_id"`
}

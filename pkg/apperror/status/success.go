package status


type SuccessCode int

// Reserved ranges by domain:
//   1000-1999: AI Copilot
//   2000-2999: AI Assistant

// AI Copilot success codes (1000-1999)
const (
	OK SuccessCode = 200
)

type FiberSuccessMessage struct {
	Code       SuccessCode `json:"code"`
	Message    string      `json:"message"`
	TrackingID string      `json:"tracking_id"`
	Data       any         `json:"data"`
}


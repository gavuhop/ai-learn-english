package status

// ErrorCode is a numeric code to classify API errors in a stable way
type ErrorCode int

// Reserved ranges by domain:
//   1000-1999: FileUpload
//   2000-2999: AI Assistant

// FileUpload error codes (1000-1999)
const (
	BadRequestBase    ErrorCode = 0
	InternalErrorBase ErrorCode = 1000
)

// FileUpload client/validation errors start at *000
const (
	FileUploadInvalidRequestBody         ErrorCode = BadRequestBase + iota // 0
	FileUploadMissingParams                                                // 1
)

// FileUpload internal errors start at 1000
const (
	FileUploadInternal                         ErrorCode = InternalErrorBase + iota // 1000
	FileUploadMarshalRequestFailed                                                  // 1001
	FileUploadEnqueueTaskFailed                                                     // 1002
)

// Deprecated: prefer domain-specific internal codes above
const (
	ErrorCodeInternal ErrorCode = 9000
)

// CodedError represents an error with an associated ErrorCode
type CodedError interface {
	error
	ErrorCode() ErrorCode
}

type codedError struct {
	code ErrorCode
	err  error
}

func (e codedError) Error() string        { return e.err.Error() }
func (e codedError) Unwrap() error        { return e.err }
func (e codedError) ErrorCode() ErrorCode { return e.code }

// New creates a new CodedError with the given code and underlying error
func New(code ErrorCode, err error) error {
	if err == nil {
		return nil
	}
	return codedError{code: code, err: err}
}

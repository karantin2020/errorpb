package errorpb

import (
	"encoding/json"
	"fmt"
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type statusError struct {
	s *Status
}

func (se *statusError) Error() string {
	return fmt.Sprintf("rpc error: code = %s desc = %s", se.Code(), se.Message())
}

// GRPCStatus implements interface{ GRPCStatus() *status.Status }
func (se *statusError) GRPCStatus() *status.Status {
	return status.New(se.Code(), se.Message())
}

// Code returns codes.Code
func (se *statusError) Code() codes.Code {
	m := (*Status)(se.s)
	return codes.Code(m.Code)
}

// Message returns codes.Code
func (se *statusError) Message() string {
	m := (*Status)(se.s)
	return m.Message
}

// Err returns error interface
func (m *Status) Err() error {
	if codes.Code(m.Code) == codes.OK {
		return nil
	}
	return &statusError{m}
}

// WithDetails adds detail info to m *Status
func (m *Status) WithDetails(d string) *Status {
	m.Details = append(m.Details, d)
	return m
}

// FromError returns *Status from err
func FromError(err error) *Status {
	if err == nil {
		return &Status{Code: int32(codes.OK), Details: []string{}}
	}
	if se, ok := err.(*statusError); ok {
		return se.s
	}
	return &Status{Code: int32(codes.Unknown), Message: err.Error(), Details: []string{}}
}

// New returns new error
func New(c codes.Code, msg string, details ...string) error {
	if c == codes.OK {
		return nil
	}
	err := &Status{Code: int32(c), Message: msg, Details: append([]string{}, details...)}
	return &statusError{err}
}

// WriteError writes json encoded error response
// Header "Content-Type" is set to "application/json"
func WriteError(req *http.Request, w http.ResponseWriter, err error) {
	// fmt.Printf("err to convert: %#v\n", err)
	w.Header().Set("Content-Type", "application/json")
	errn := FromError(err)
	var status int
	switch codes.Code(errn.Code) {
	case codes.OK:
		status = http.StatusOK
	case codes.Canceled:
		status = http.StatusRequestTimeout
	case codes.Unknown:
		status = http.StatusInternalServerError
	case codes.InvalidArgument:
		status = http.StatusBadRequest
	case codes.DeadlineExceeded:
		status = http.StatusGatewayTimeout
	case codes.NotFound:
		status = http.StatusNotFound
	case codes.AlreadyExists:
		status = http.StatusConflict
	case codes.PermissionDenied:
		status = http.StatusForbidden
	case codes.Unauthenticated:
		status = http.StatusUnauthorized
	case codes.ResourceExhausted:
		status = http.StatusTooManyRequests
	case codes.FailedPrecondition:
		status = http.StatusPreconditionFailed
	case codes.Aborted:
		status = http.StatusConflict
	case codes.OutOfRange:
		status = http.StatusBadRequest
	case codes.Unimplemented:
		status = http.StatusNotImplemented
	case codes.Internal:
		status = http.StatusInternalServerError
	case codes.Unavailable:
		status = http.StatusServiceUnavailable
	case codes.DataLoss:
		status = http.StatusInternalServerError
	}
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	errn.Code = int32(status)
	enc.Encode(errn)
}

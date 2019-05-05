package errorpb

import (
	context "context"
	"encoding/json"
	"fmt"
	"net/http"

	"google.golang.org/grpc/codes"
)

type statusError Error

func (se *statusError) Error() string {
	p := (*Error)(se)
	return fmt.Sprintf("rpc error: code = %s desc = %s", codes.Code(p.Code), p.Message)
}

func (se *statusError) GRPCStatus() *Error {
	return (*Error)(se)
}

// Err returns Error as error interface
func (m *Error) Err() error {
	if codes.Code(m.Code) == codes.OK {
		return nil
	}
	return (*statusError)(m)
}

// WithDetails adds detail info to m *Error
func (m *Error) WithDetails(d string) *Error {
	m.Details = append(m.Details, d)
	return m
}

// FromError returns *Error from err
func FromError(err error) (*Error, bool) {
	if err == nil {
		return &Error{Code: int32(codes.OK)}, true
	}
	if se, ok := err.(interface {
		GRPCStatus() *Error
	}); ok {
		return se.GRPCStatus(), true
	}
	return &Error{Code: int32(codes.Unknown), Message: err.Error(), Details: []string{}}, false
}

// New returns new error
func New(c codes.Code, msg string, details ...string) error {
	if c == codes.OK {
		return nil
	}
	err := &Error{Code: int32(c), Message: msg, Details: append([]string{}, details...)}
	return (*statusError)(err)
}

// WriteError writes json encoded error response
// Header "Content-Type" is set to "application/json"
func WriteError(ctx context.Context, req *http.Request, w http.ResponseWriter, err error) {
	// fmt.Printf("err to convert: %#v\n", err)
	w.Header().Set("Content-Type", "application/json")
	errt, ok := err.(*statusError)
	var status int
	if ok {
		errn := (*Error)(errt)
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
		enc.Encode(*errn)
		return
	}
	w.WriteHeader(500)
	enc := json.NewEncoder(w)
	enc.Encode(Error{Code: int32(codes.Internal), Message: err.Error()})
}

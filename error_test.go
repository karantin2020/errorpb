package errorpb

import (
	"errors"
	"reflect"
	"testing"

	"google.golang.org/grpc/codes"
)

func TestNewError(t *testing.T) {
	type args struct {
		c       codes.Code
		msg     string
		details []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Return error",
			args: args{
				c:       codes.InvalidArgument,
				msg:     "Invalid argiments",
				details: []string{"error details"},
			},
			wantErr: true,
		},
		{
			name: "Return no error",
			args: args{
				c:       codes.OK,
				msg:     "No error",
				details: []string{"no error"},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := New(tt.args.c, tt.args.msg, tt.args.details...); (err != nil) != tt.wantErr {
				t.Errorf("NewError() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFromError(t *testing.T) {
	perr := errors.New("New error")
	nerr := &Status{Code: int32(codes.Unknown), Message: perr.Error(), Details: []string{}}
	noerr := &Status{Code: int32(codes.OK), Details: []string{}}
	type args struct {
		err error
	}
	tests := []struct {
		name  string
		args  args
		want  *Status
		want1 bool
	}{
		{
			name: "Return error",
			args: args{
				err: perr,
			},
			want:  nerr,
			want1: false,
		},
		{
			name: "Return no error",
			args: args{
				err: nil,
			},
			want:  noerr,
			want1: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := FromError(tt.args.err)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FromError() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("FromError() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

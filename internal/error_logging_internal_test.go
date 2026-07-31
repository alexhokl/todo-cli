package internal

import (
	"context"
	"errors"
	"testing"

	"google.golang.org/grpc"
)

func TestErrorLoggingInterceptor(t *testing.T) {
	tests := []struct {
		name        string
		handlerResp any
		handlerErr  error
	}{
		{"success", "ok", nil},
		{"failure", nil, errors.New("something went wrong")},
	}

	info := &grpc.UnaryServerInfo{FullMethod: "/todo.TodoService/ListTodos"}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			handler := func(_ context.Context, _ any) (any, error) {
				return test.handlerResp, test.handlerErr
			}

			resp, err := ErrorLoggingInterceptor(context.Background(), nil, info, handler)
			if !errors.Is(err, test.handlerErr) {
				t.Errorf("expected error %v but got %v", test.handlerErr, err)
			}
			if resp != test.handlerResp {
				t.Errorf("expected response %v but got %v", test.handlerResp, resp)
			}
		})
	}
}

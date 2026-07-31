package internal

import (
	"context"
	"testing"

	"google.golang.org/grpc"
)

func TestDummyAuthenticationInterceptorInjectsUserID(t *testing.T) {
	var gotUserID uint
	var handlerCalled bool
	handler := func(ctx context.Context, _ any) (any, error) {
		handlerCalled = true
		gotUserID, _ = ctx.Value(contextKeyUser{}).(uint)
		return "ok", nil
	}

	resp, err := DummyAuthenticationInterceptor(
		context.Background(),
		nil,
		&grpc.UnaryServerInfo{FullMethod: "/item.ItemService/ListItems"},
		handler,
	)
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if !handlerCalled {
		t.Fatalf("expected the handler to be called")
	}
	if gotUserID != 1 {
		t.Errorf("expected userID 1 but got %d", gotUserID)
	}
	if resp != "ok" {
		t.Errorf("expected the handler response to pass through but got %v", resp)
	}
}

func TestDummyStreamAuthenticationInterceptorInjectsUserID(t *testing.T) {
	var gotUserID uint
	var handlerCalled bool
	handler := func(_ any, ss grpc.ServerStream) error {
		handlerCalled = true
		gotUserID, _ = ss.Context().Value(contextKeyUser{}).(uint)
		return nil
	}

	err := DummyStreamAuthenticationInterceptor(
		nil,
		&fakeServerStream{ctx: context.Background()},
		&grpc.StreamServerInfo{},
		handler,
	)
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if !handlerCalled {
		t.Fatalf("expected the handler to be called")
	}
	if gotUserID != 1 {
		t.Errorf("expected userID 1 but got %d", gotUserID)
	}
}

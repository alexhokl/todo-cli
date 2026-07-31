package internal

import (
	"context"
	"errors"
	"net"
	"testing"

	"github.com/alexhokl/todo-cli/database"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"tailscale.com/client/tailscale/apitype"
	"tailscale.com/tailcfg"
)

// fakeIdentityLookup is a stub callerIdentityLookup returning a canned
// WhoIsResponse (or an error) so the interceptor can be exercised without a
// live tailscaled.
type fakeIdentityLookup struct {
	resp     *apitype.WhoIsResponse
	err      error
	calls    int
	lastAddr string
}

func (f *fakeIdentityLookup) GetCallerIdentityFromRemoteIPAddress(_ context.Context, remoteAddr string) (*apitype.WhoIsResponse, error) {
	f.calls++
	f.lastAddr = remoteAddr
	return f.resp, f.err
}

func setupInterceptorDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}
	if err := database.AutoMigrate(db); err != nil {
		t.Fatalf("failed to migrate test database: %v", err)
	}
	return db
}

// peerContext returns a context carrying the given peer address, mimicking what
// grpc injects on an incoming RPC. ip should be a bare IP address; the port is
// appended so the interceptor's net.SplitHostPort can extract it.
func peerContext(ip string) context.Context {
	return peer.NewContext(context.Background(), &peer.Peer{
		Addr: &net.TCPAddr{IP: net.ParseIP(ip), Port: 12345},
	})
}

func TestInterceptCachedAddressSkipsWhoIs(t *testing.T) {
	db := setupInterceptorDB(t)
	user := &database.User{Username: "cached@example.com"}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("failed to seed the user: %v", err)
	}
	if err := db.Create(&database.TailscaleAddress{Address: "100.64.0.1", UserID: user.ID}).Error; err != nil {
		t.Fatalf("failed to cache the address: %v", err)
	}

	lookup := &fakeIdentityLookup{}
	interceptor := NewTailscaleAuthenticationInterceptor(db, lookup)

	var gotUserID uint
	var handlerCalled bool
	handler := func(ctx context.Context, _ any) (any, error) {
		handlerCalled = true
		gotUserID, _ = ctx.Value(contextKeyUser{}).(uint)
		return nil, nil
	}

	_, err := interceptor.Intercept(peerContext("100.64.0.1"), nil, &grpc.UnaryServerInfo{}, handler)
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if !handlerCalled {
		t.Fatalf("expected the handler to be called")
	}
	if gotUserID != user.ID {
		t.Errorf("expected userID %d in context, got %d", user.ID, gotUserID)
	}
	if lookup.calls != 0 {
		t.Errorf("expected WhoIs to be skipped on a cache hit but got %d calls", lookup.calls)
	}
}

func TestInterceptWhoIsSuccessCreatesUserAndCachesAddress(t *testing.T) {
	db := setupInterceptorDB(t)
	lookup := &fakeIdentityLookup{
		resp: &apitype.WhoIsResponse{
			UserProfile: &tailcfg.UserProfile{LoginName: "new@example.com"},
		},
	}
	interceptor := NewTailscaleAuthenticationInterceptor(db, lookup)

	var gotUserID uint
	handler := func(ctx context.Context, _ any) (any, error) {
		gotUserID, _ = ctx.Value(contextKeyUser{}).(uint)
		return nil, nil
	}

	_, err := interceptor.Intercept(peerContext("100.64.0.2"), nil, &grpc.UnaryServerInfo{}, handler)
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if gotUserID == 0 {
		t.Fatalf("expected a non-zero userID in context")
	}

	// The user and address must have been persisted.
	var user database.User
	if err := db.Where("username = ?", "new@example.com").First(&user).Error; err != nil {
		t.Fatalf("expected the user to have been created: %v", err)
	}
	if user.ID != gotUserID {
		t.Errorf("expected cached userID %d to match context userID %d", user.ID, gotUserID)
	}

	var addr database.TailscaleAddress
	if err := db.Where("address = ?", "100.64.0.2").First(&addr).Error; err != nil {
		t.Fatalf("expected the address to have been cached: %v", err)
	}
	if addr.UserID != user.ID {
		t.Errorf("expected cached address to point at user %d but got %d", user.ID, addr.UserID)
	}
}

func TestInterceptWhoIsFailureReturnsUnauthenticated(t *testing.T) {
	db := setupInterceptorDB(t)
	lookup := &fakeIdentityLookup{err: errors.New("not connected")}
	interceptor := NewTailscaleAuthenticationInterceptor(db, lookup)

	handler := func(context.Context, any) (any, error) { return nil, nil }

	_, err := interceptor.Intercept(peerContext("100.64.0.3"), nil, &grpc.UnaryServerInfo{}, handler)
	if got := status.Code(err); got != codes.Unauthenticated {
		t.Errorf("expected %v but got %v (%v)", codes.Unauthenticated, got, err)
	}
}

func TestInterceptStreamResolvesUserID(t *testing.T) {
	db := setupInterceptorDB(t)
	lookup := &fakeIdentityLookup{
		resp: &apitype.WhoIsResponse{
			UserProfile: &tailcfg.UserProfile{LoginName: "stream@example.com"},
		},
	}
	interceptor := NewTailscaleAuthenticationInterceptor(db, lookup)

	var gotUserID uint
	handler := func(_ any, ss grpc.ServerStream) error {
		gotUserID, _ = ss.Context().Value(contextKeyUser{}).(uint)
		return nil
	}

	err := interceptor.InterceptStream(nil, &fakeServerStream{ctx: peerContext("100.64.0.4")}, &grpc.StreamServerInfo{}, handler)
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if gotUserID == 0 {
		t.Errorf("expected a non-zero userID in the stream context")
	}
}

// fakeServerStream is a minimal grpc.ServerStream that returns a fixed context.
type fakeServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (f *fakeServerStream) Context() context.Context { return f.ctx }

func TestInterceptNoPeerReturnsInternal(t *testing.T) {
	db := setupInterceptorDB(t)
	lookup := &fakeIdentityLookup{}
	interceptor := NewTailscaleAuthenticationInterceptor(db, lookup)

	handler := func(context.Context, any) (any, error) { return nil, nil }

	// A context with no peer information mimics a non-Tailscale transport.
	_, err := interceptor.Intercept(context.Background(), nil, &grpc.UnaryServerInfo{}, handler)
	if got := status.Code(err); got != codes.Internal {
		t.Errorf("expected %v but got %v (%v)", codes.Internal, got, err)
	}
}

func TestInterceptStreamNoPeerReturnsInternal(t *testing.T) {
	db := setupInterceptorDB(t)
	lookup := &fakeIdentityLookup{}
	interceptor := NewTailscaleAuthenticationInterceptor(db, lookup)

	handler := func(any, grpc.ServerStream) error { return nil }

	err := interceptor.InterceptStream(nil, &fakeServerStream{ctx: context.Background()}, &grpc.StreamServerInfo{}, handler)
	if got := status.Code(err); got != codes.Internal {
		t.Errorf("expected %v but got %v (%v)", codes.Internal, got, err)
	}
}

func TestGetAddressInfoMissReturnsFalse(t *testing.T) {
	db := setupInterceptorDB(t)
	userID, ok := getAddressInfo(db, "100.64.0.99")
	if ok {
		t.Errorf("expected ok=false on a cache miss but got true")
	}
	if userID != 0 {
		t.Errorf("expected userID 0 on a miss but got %d", userID)
	}
}

func TestGetAddressInfoHitReturnsUserID(t *testing.T) {
	db := setupInterceptorDB(t)
	user := &database.User{Username: "cached@example.com"}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("failed to seed the user: %v", err)
	}
	if err := db.Create(&database.TailscaleAddress{Address: "100.64.0.1", UserID: user.ID}).Error; err != nil {
		t.Fatalf("failed to cache the address: %v", err)
	}

	userID, ok := getAddressInfo(db, "100.64.0.1")
	if !ok {
		t.Fatalf("expected ok=true on a cache hit")
	}
	if userID != user.ID {
		t.Errorf("expected userID %d but got %d", user.ID, userID)
	}
}

func TestGetOrCreateUserCreatesThenReuses(t *testing.T) {
	db := setupInterceptorDB(t)

	created, err := getOrCreateUser(db, "new@example.com")
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if created.ID == 0 {
		t.Fatalf("expected a non-zero id on first sight")
	}

	reused, err := getOrCreateUser(db, "new@example.com")
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if reused.ID != created.ID {
		t.Errorf("expected the existing user %d to be reused but got %d", created.ID, reused.ID)
	}
}

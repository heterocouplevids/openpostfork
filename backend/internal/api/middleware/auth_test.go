package middleware

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	_ "github.com/mattn/go-sqlite3"
	"github.com/openpost/backend/internal/models"
	"github.com/openpost/backend/internal/services/apitokens"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
)

// fakeAuthenticator stands in for the real composite authenticator.
// Each test sets nextErr/nextPrincipal to drive the middleware's
// behavior.
type fakeAuthenticator struct {
	principal *Principal
	err       error
	gotToken  string
}

func (f *fakeAuthenticator) AuthenticateBearer(_ context.Context, token string) (*Principal, error) {
	f.gotToken = token
	return f.principal, f.err
}

func newEchoAuthed(auth Authenticator) *echo.Echo {
	e := echo.New()
	e.GET("/x", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	}, BearerMiddleware(auth))
	return e
}

func TestBearerMiddleware_Success_AttachesPrincipal(t *testing.T) {
	want := &Principal{UserID: "u_42", Email: "rodrigo@example.com"}
	auth := &fakeAuthenticator{principal: want}

	e := newEchoAuthed(auth)
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/x", nil)
	req.Header.Set("Authorization", "Bearer op_cli_abc_secret")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (body=%q)", rec.Code, rec.Body.String())
	}
	if auth.gotToken != "op_cli_abc_secret" {
		t.Errorf("middleware did not pass raw token to authenticator, got %q", auth.gotToken)
	}
}

func TestBearerMiddleware_MissingHeader(t *testing.T) {
	auth := &fakeAuthenticator{principal: &Principal{UserID: "u"}}
	e := newEchoAuthed(auth)
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/x", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for missing header, got %d", rec.Code)
	}
	if auth.gotToken != "" {
		t.Errorf("authenticator should not be called on missing header, got token %q", auth.gotToken)
	}
}

func TestBearerMiddleware_BadFormat(t *testing.T) {
	auth := &fakeAuthenticator{principal: &Principal{UserID: "u"}}
	e := newEchoAuthed(auth)

	for _, h := range []string{"op_cli_abc", "Basic abc123", "Bearer"} {
		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/x", nil)
		req.Header.Set("Authorization", h)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("expected 401 for header %q, got %d", h, rec.Code)
		}
	}
}

func TestBearerMiddleware_InvalidToken_Returns401(t *testing.T) {
	// The media upload 401 the user hit: a valid Bearer header whose
	// token the authenticator rejects. This is the regression guard.
	auth := &fakeAuthenticator{err: errors.New("not found")}
	e := newEchoAuthed(auth)
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/x", nil)
	req.Header.Set("Authorization", "Bearer op_cli_bogus")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for rejected token, got %d (body=%q)", rec.Code, rec.Body.String())
	}
}

type rejectingAuthenticator struct{}

func (rejectingAuthenticator) AuthenticateBearer(_ context.Context, _ string) (*Principal, error) {
	return nil, errors.New("invalid jwt")
}

func TestCompositeServicePreservesAPITokenScope(t *testing.T) {
	ctx := context.Background()
	sqldb, err := sql.Open("sqlite3", fmt.Sprintf("file:%s?mode=memory&cache=private", strings.ReplaceAll(t.Name(), "/", "_")))
	if err != nil {
		t.Fatal(err)
	}
	sqldb.SetMaxOpenConns(1)
	db := bun.NewDB(sqldb, sqlitedialect.New())
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Fatal(err)
		}
	})
	for _, model := range []interface{}{(*models.User)(nil), (*models.APIToken)(nil)} {
		if _, err := db.NewCreateTable().Model(model).IfNotExists().Exec(ctx); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := db.NewInsert().Model(&models.User{
		ID:           "user-1",
		Email:        "user@example.com",
		PasswordHash: "hash",
	}).Exec(ctx); err != nil {
		t.Fatal(err)
	}

	tokenService := apitokens.NewService(db)
	generated, err := tokenService.GenerateToken(ctx, "user-1", "ChatGPT", apitokens.ScopeMCP, nil)
	if err != nil {
		t.Fatal(err)
	}
	composite := &CompositeService{jwt: rejectingAuthenticator{}, apiTokens: tokenService}

	principal, err := composite.AuthenticateBearer(ctx, generated.Token)
	if err != nil {
		t.Fatal(err)
	}
	if principal.Scope != apitokens.ScopeMCP {
		t.Fatalf("expected scope %q, got %q", apitokens.ScopeMCP, principal.Scope)
	}
}

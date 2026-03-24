package middleware

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/sessions"
)

type contextKey string

const UserIDKey contextKey = "user_id"
const UserRoleKey contextKey = "user_role"

var Store *sessions.CookieStore

func InitSessionStore(secretKey string) {
	Store = sessions.NewCookieStore([]byte(secretKey))
	Store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 7 days
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
}

func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := Store.Get(r, "session")
		if err != nil || session.IsNew {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "не авторизован"})
			return
		}

		userID, ok := session.Values["user_id"].(int64)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "не авторизован"})
			return
		}

		role, _ := session.Values["role"].(string)

		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		ctx = context.WithValue(ctx, UserRoleKey, role)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserID(ctx context.Context) int64 {
	id, _ := ctx.Value(UserIDKey).(int64)
	return id
}

func GetUserRole(ctx context.Context) string {
	role, _ := ctx.Value(UserRoleKey).(string)
	return role
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

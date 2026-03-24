package handler

import (
	"net/http"

	"golang.org/x/crypto/bcrypt"

	"simplestock/internal/dto"
	"simplestock/internal/middleware"
	"simplestock/internal/repository"
)

type AuthHandler struct {
	userRepo *repository.UserRepo
}

func NewAuthHandler(userRepo *repository.UserRepo) *AuthHandler {
	return &AuthHandler{userRepo: userRepo}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "неверный формат запроса")
		return
	}

	if req.Username == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "логин и пароль обязательны")
		return
	}

	user, err := h.userRepo.GetByUsername(r.Context(), req.Username)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "неверный логин или пароль")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		writeError(w, http.StatusUnauthorized, "неверный логин или пароль")
		return
	}

	session, _ := middleware.Store.Get(r, "session")
	session.Values["user_id"] = user.ID
	session.Values["role"] = user.Role
	if err := session.Save(r, w); err != nil {
		writeError(w, http.StatusInternalServerError, "ошибка сессии")
		return
	}

	writeJSON(w, http.StatusOK, dto.LoginResponse{
		ID:       user.ID,
		Username: user.Username,
		FullName: user.FullName,
		Role:     user.Role,
	})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	session, _ := middleware.Store.Get(r, "session")
	session.Options.MaxAge = -1
	session.Save(r, w)
	writeJSON(w, http.StatusOK, map[string]string{"message": "выход выполнен"})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	user, err := h.userRepo.GetByID(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "пользователь не найден")
		return
	}

	writeJSON(w, http.StatusOK, dto.LoginResponse{
		ID:       user.ID,
		Username: user.Username,
		FullName: user.FullName,
		Role:     user.Role,
	})
}

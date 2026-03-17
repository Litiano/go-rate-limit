package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Litiano/go-rate-limit/internal/dto"
	"github.com/go-chi/jwtauth"
)

type AuthHandler struct {
	Jwt          *jwtauth.JWTAuth
	JwtExpiresIn int
}

type Error struct {
	Message string `json:"message"`
}

func NewAuthHandler(jwt *jwtauth.JWTAuth, jwtExpiresIn int) *AuthHandler {
	return &AuthHandler{Jwt: jwt, JwtExpiresIn: jwtExpiresIn}
}

func (h *AuthHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var user dto.LoginInput
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	_, tokenString, err := h.Jwt.Encode(map[string]interface{}{
		"sub": user.Username,
		"exp": time.Now().Add(time.Second * time.Duration(h.JwtExpiresIn)).Unix(),
	})
	accessToken := dto.LoginOutput{AccessToken: tokenString}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(accessToken)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

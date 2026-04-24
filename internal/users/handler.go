package users

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gustavomzina/chirpy/internal/auth"
	"github.com/gustavomzina/chirpy/internal/database"
	"github.com/gustavomzina/chirpy/internal/webutil"
)

type Handler struct {
	DB          *database.Queries
	TokenSecret string
}

const accessTokenDuration = time.Hour
const refreshTokenDuration = time.Duration(24 * 60 * time.Hour)

func (h *Handler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	type returnVals struct {
		Id        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	in := parameters{}
	err := decoder.Decode(&in)
	if err != nil {
		webutil.RespondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	if len(in.Password) < 5 {
		webutil.RespondWithError(w, http.StatusBadRequest, "Password minimum length = 5", errors.New("minimum length"))
		return
	}

	hash, err := auth.HashPassword(in.Password)
	if err != nil {
		webutil.RespondWithError(w, http.StatusInternalServerError, "Couldn't hash password", err)
		return
	}

	userParams := database.CreateUserParams{
		Email:          in.Email,
		HashedPassword: hash,
	}

	user, err := h.DB.CreateUser(r.Context(), userParams)
	if err != nil {
		webutil.RespondWithError(w, http.StatusInternalServerError, "Couldn't create user", err)
		return
	}

	ret := returnVals{
		Id:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}

	webutil.RespondWithJson(w, http.StatusCreated, ret)
}

func (h *Handler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	type userResponse struct {
		Id        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
	}

	type response struct {
		userResponse
		AccessToken  string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}

	decoder := json.NewDecoder(r.Body)
	in := parameters{}
	err := decoder.Decode(&in)
	if err != nil {
		webutil.RespondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	if len(in.Password) < 5 {
		webutil.RespondWithError(w, http.StatusBadRequest, "Password minimum length = 5", errors.New("minimum length"))
		return
	}

	user, err := h.DB.GetUser(r.Context(), in.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			webutil.RespondWithError(w, http.StatusNotFound, "No user found", err)
			return
		}
		webutil.RespondWithError(w, http.StatusInternalServerError, "Couldn't get the user", err)
		return
	}

	check, err := auth.CheckPasswordHash(in.Password, user.HashedPassword)
	if err != nil || !check {
		webutil.RespondWithError(w, http.StatusUnauthorized, "Incorrect email or password", err)
		return
	}

	acccessToken, err := auth.MakeJWT(user.ID, h.TokenSecret, accessTokenDuration)
	if err != nil {
		webutil.RespondWithError(w, http.StatusInternalServerError, "Failed to create JWT", err)
		return
	}

	createRefreshTokenParams := database.CreateRefreshTokenParams{
		Token:     auth.MakeRefreshToken(),
		UserID:    user.ID,
		ExpiresAt: time.Now().UTC().Add(refreshTokenDuration),
	}

	refreshToken, err := h.DB.CreateRefreshToken(r.Context(), createRefreshTokenParams)
	if err != nil {
		webutil.RespondWithError(w, http.StatusInternalServerError, "Failed to create refresh token", err)
		return
	}

	userResp := userResponse{
		Id:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}
	ret := response{
		userResponse: userResp,
		AccessToken:  acccessToken,
		RefreshToken: refreshToken.Token,
	}

	webutil.RespondWithJson(w, http.StatusOK, ret)
}

func (h *Handler) HandleRefresh(w http.ResponseWriter, r *http.Request) {
	type response struct {
		AccessToken string `json:"token"`
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		webutil.RespondWithError(w, http.StatusInternalServerError, "Couldn't get token", err)
		return
	}

	refreshToken, err := h.DB.GetRefreshToken(r.Context(), token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			webutil.RespondWithError(w, http.StatusUnauthorized, "No refresh token found", err)
		}
		webutil.RespondWithError(w, http.StatusUnauthorized, "unauthorized", err)
		return
	}

	if time.Now().After(refreshToken.ExpiresAt) {
		webutil.RespondWithError(w, http.StatusUnauthorized, "unauthorized", errors.New("expired refresh token"))
		return
	}

	if refreshToken.RevokedAt.Valid {
		webutil.RespondWithError(w, http.StatusUnauthorized, "unauthorized", errors.New("revoked refresh token"))
		return
	}

	acccessToken, err := auth.MakeJWT(refreshToken.UserID, h.TokenSecret, accessTokenDuration)
	if err != nil {
		webutil.RespondWithError(w, http.StatusInternalServerError, "Failed to create JWT", err)
		return
	}

	webutil.RespondWithJson(w, http.StatusOK, response{AccessToken: acccessToken})
}

func (h *Handler) HandleRevoke(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		webutil.RespondWithError(w, http.StatusInternalServerError, "Couldn't get token", err)
		return
	}

	refreshToken, err := h.DB.GetRefreshToken(r.Context(), token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			webutil.RespondWithError(w, http.StatusUnauthorized, "No refresh token found", err)
		}
		webutil.RespondWithError(w, http.StatusUnauthorized, "unauthorized", err)
		return
	}

	if time.Now().After(refreshToken.ExpiresAt) {
		webutil.RespondWithError(w, http.StatusUnauthorized, "unauthorized", errors.New("expired refresh token"))
		return
	}

	if refreshToken.RevokedAt.Valid {
		webutil.RespondWithError(w, http.StatusUnauthorized, "unauthorized", errors.New("revoked refresh token"))
		return
	}

	err = h.DB.RevokeRefreshToken(r.Context(), refreshToken.Token)
	if err != nil {
		webutil.RespondWithError(w, http.StatusInternalServerError, "Couldn't revoke refresh token", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

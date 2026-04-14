package users

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gustavomzina/chirpy/internal/auth"
	"github.com/gustavomzina/chirpy/internal/database"
	"github.com/gustavomzina/chirpy/internal/webutil"
)

type Handler struct {
	DB *database.Queries
}

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

	if len(in.Password) < 6 {
		webutil.RespondWithError(w, http.StatusBadRequest, "Password minimum length = 6", errors.New("minimum length"))
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

	if len(in.Password) < 6 {
		webutil.RespondWithError(w, http.StatusBadRequest, "Password minimum length = 6", errors.New("minimum length"))
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
	if err != nil {
		webutil.RespondWithError(w, http.StatusInternalServerError, "Hash verification failed", err)
		return
	}
	fmt.Printf("password: %v\n", in.Password)
	fmt.Printf("hash: %v\n", user.HashedPassword)
	fmt.Printf("Password matches? %v\n", check)
	if !check {
		webutil.RespondWithError(w, http.StatusUnauthorized, "Unauthorized", errors.New("unauthorized"))
		return
	}

	ret := returnVals{
		Id:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}

	webutil.RespondWithJson(w, http.StatusOK, ret)
}

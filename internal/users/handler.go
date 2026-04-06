package users

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gustavomzina/chirpy/internal/database"
	"github.com/gustavomzina/chirpy/internal/webutil"
)

type Handler struct {
	DB *database.Queries
}

func (h *Handler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email string `json:"email"`
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

	user, err := h.DB.CreateUser(r.Context(), in.Email)
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

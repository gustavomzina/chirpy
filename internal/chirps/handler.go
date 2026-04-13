package chirps

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
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
		Body   string    `json:"body"`
		UserId uuid.UUID `json:"user_id"`
	}

	type returnVals struct {
		Id        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body      string    `json:"body"`
		UserId    uuid.UUID `json:"user_id"`
	}

	decoder := json.NewDecoder(r.Body)
	in := parameters{}
	err := decoder.Decode(&in)
	if err != nil {
		webutil.RespondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	cleanedBody, err := validateChirpBody(in.Body)
	if err != nil {
		webutil.RespondWithError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	chirp, err := h.DB.CreateChirp(r.Context(), database.CreateChirpParams{UserID: in.UserId, Body: cleanedBody})
	if err != nil {
		webutil.RespondWithError(w, http.StatusInternalServerError, "Couldn't create chirp", err)
		return
	}

	webutil.RespondWithJson(w, http.StatusCreated, returnVals{
		Id:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserId:    chirp.UserID,
	})
}

func (h *Handler) HandleGetAll(w http.ResponseWriter, r *http.Request) {
	type returnVal struct {
		Id        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body      string    `json:"body"`
		UserId    uuid.UUID `json:"user_id"`
	}

	chirps, err := h.DB.GetChirps(r.Context())
	if err != nil {
		webutil.RespondWithError(w, http.StatusInternalServerError, "Couldn't get chirps", err)
		return
	}

	ret := make([]returnVal, 0, len(chirps))

	for _, chirp := range chirps {
		val := returnVal{
			Id:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserId:    chirp.UserID,
		}

		ret = append(ret, val)
	}

	webutil.RespondWithJson(w, http.StatusOK, ret)
}

func (h *Handler) HandleGet(w http.ResponseWriter, r *http.Request) {
	type returnVal struct {
		Id        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body      string    `json:"body"`
		UserId    uuid.UUID `json:"user_id"`
	}

	id := r.PathValue("id")
	if strings.TrimSpace(id) == "" {
		webutil.RespondWithError(w, http.StatusBadRequest, "Chirp id is required", errors.New("chirp id is required"))
		return
	}

	uid, err := uuid.Parse(id)
	if err != nil {
		webutil.RespondWithError(w, http.StatusInternalServerError, "Invalid chirp id", err)
		return
	}

	chirp, err := h.DB.GetChirp(r.Context(), uid)
	if err != nil {
		if strings.Contains(err.Error(), "sql: no rows in result set") {
			webutil.RespondWithError(w, http.StatusNotFound, "No chirp found", errors.New("no chirp found"))
			return
		}

		webutil.RespondWithError(w, http.StatusInternalServerError, "Couldn't get chirp", err)
		return
	}

	ret := returnVal{
		Id:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserId:    chirp.UserID,
	}

	webutil.RespondWithJson(w, http.StatusOK, ret)
}

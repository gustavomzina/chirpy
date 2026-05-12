package chirps

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"slices"
	"strings"
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

func (h *Handler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	type returnVals struct {
		Id        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body      string    `json:"body"`
		UserId    uuid.UUID `json:"user_id"`
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		webutil.RespondWithError(w, http.StatusInternalServerError, "Couldn't get token", err)
		return
	}

	userId, err := auth.ValidateJWT(token, h.TokenSecret)
	if err != nil {
		webutil.RespondWithError(w, http.StatusUnauthorized, "unauthorized", err)
		return
	}

	decoder := json.NewDecoder(r.Body)
	in := parameters{}
	err = decoder.Decode(&in)
	if err != nil {
		webutil.RespondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	cleanedBody, err := validateChirpBody(in.Body)
	if err != nil {
		webutil.RespondWithError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	chirp, err := h.DB.CreateChirp(r.Context(), database.CreateChirpParams{UserID: userId, Body: cleanedBody})
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

func authorIDFromRequest(r *http.Request) (uuid.UUID, error) {
	authorIDString := r.URL.Query().Get("author_id")
	if authorIDString == "" {
		return uuid.Nil, nil
	}
	authorID, err := uuid.Parse(authorIDString)
	if err != nil {
		return uuid.Nil, err
	}
	return authorID, nil
}

func (h *Handler) HandleGetAll(w http.ResponseWriter, r *http.Request) {
	type returnVal struct {
		Id        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body      string    `json:"body"`
		UserId    uuid.UUID `json:"user_id"`
	}

	authorUUID, err := authorIDFromRequest(r)
	if err != nil {
		webutil.RespondWithError(w, http.StatusBadRequest, "Invalid author id", err)
		return
	}

	var chirps []database.Chirp

	if authorUUID != uuid.Nil {
		chirps, err = h.DB.GetChirpsForUserID(r.Context(), authorUUID)
	} else {
		chirps, err = h.DB.GetChirps(r.Context())
	}

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusNoContent)
			return
		}
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

	sortOrder := "asc"
	sortQueryParam := r.URL.Query().Get("sort")
	if sortQueryParam == "desc" {
		sortOrder = "desc"
	}

	slices.SortFunc(ret, func(a, b returnVal) int {
		var ret = -1
		if a.CreatedAt.After(b.CreatedAt) {
			ret = 1
		}
		if sortOrder == "desc" {
			ret = ret * (-1)
		}
		return ret
	})

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
		if errors.Is(err, sql.ErrNoRows) {
			webutil.RespondWithError(w, http.StatusNotFound, "No chirp found", err)
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

func (h *Handler) HandleDelete(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		webutil.RespondWithError(w, http.StatusUnauthorized, "Couldn't get token", err)
		return
	}

	userId, err := auth.ValidateJWT(token, h.TokenSecret)
	if err != nil {
		webutil.RespondWithError(w, http.StatusUnauthorized, "unauthorized", err)
		return
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
		if errors.Is(err, sql.ErrNoRows) {
			webutil.RespondWithError(w, http.StatusNotFound, "No chirp found", err)
			return
		}

		webutil.RespondWithError(w, http.StatusInternalServerError, "Couldn't get chirp", err)
		return
	}

	if chirp.UserID != userId {
		webutil.RespondWithError(w, http.StatusForbidden, "Not chirp author", errors.New("not chirp author"))
		return
	}

	err = h.DB.DeleteChirp(r.Context(), database.DeleteChirpParams{ID: chirp.ID, UserID: userId})
	if err != nil {
		webutil.RespondWithError(w, http.StatusInternalServerError, "Could not delete chirp", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

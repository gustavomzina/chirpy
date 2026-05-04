package polka

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/gustavomzina/chirpy/internal/database"
	"github.com/gustavomzina/chirpy/internal/webutil"
)

type Handler struct {
	DB          *database.Queries
	TokenSecret string
}

func (h *Handler) HandleUpgradeUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Event string `json:"event"`
		Data  struct {
			UserId string `json:"user_id"`
		} `json:"data"`
	}

	decoder := json.NewDecoder(r.Body)
	in := parameters{}
	err := decoder.Decode(&in)
	if err != nil {
		webutil.RespondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	if in.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	userUUID, err := uuid.Parse(in.Data.UserId)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	_, err = h.DB.UpgradeUser(r.Context(), userUUID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		webutil.RespondWithError(w, http.StatusInternalServerError, "Couldn't update user", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

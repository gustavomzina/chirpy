package chirps

import (
	"encoding/json"
	"net/http"

	"github.com/gustavomzina/chirpy/internal/database"
	"github.com/gustavomzina/chirpy/internal/webutil"
)

type Handler struct {
	DB *database.Queries
}

func (h *Handler) HandleValidate(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	type returnVals struct {
		Valid        bool   `json:"valid"`
		Cleaned_body string `json:"cleaned_body"`
	}

	decoder := json.NewDecoder(r.Body)
	in := parameters{}
	err := decoder.Decode(&in)
	if err != nil {
		webutil.RespondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	const maxChirpLength = 140
	if len(in.Body) > maxChirpLength {
		webutil.RespondWithError(w, http.StatusBadRequest, "Chirp is too long", err)
		return
	}

	cleanedBody := filterBadWords(in.Body)

	webutil.RespondWithJson(w, http.StatusOK, returnVals{Cleaned_body: cleanedBody, Valid: true})
}

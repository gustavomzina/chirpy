package chirps

import (
	"errors"
	"strings"
)

func filterBadWords(s string) string {
	badWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}

	words := strings.Fields(s)

	for i, word := range words {
		_, isBadWord := badWords[strings.ToLower(word)]
		if isBadWord {
			words[i] = "****"
		}
	}

	return strings.Join(words, " ")
}

func validateChirpBody(body string) (string, error) {
	const maxChirpLength = 140
	if len(body) > maxChirpLength {
		return "", errors.New("Chirp is too long")
	}

	return filterBadWords(body), nil
}

package chirps

import "strings"

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

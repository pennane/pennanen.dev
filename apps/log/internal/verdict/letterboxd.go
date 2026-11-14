package verdict

import "encoding/json"

type LetterboxdReaction string

const (
	NoReaction LetterboxdReaction = ""
	Liked      LetterboxdReaction = "liked"
)

type LetterboxdTake struct {
	Rating      *float64           `json:"rating,omitempty"`
	Reaction    LetterboxdReaction `json:"reaction,omitempty"`
	Review      string             `json:"review,omitempty"`
	LastWatched string             `json:"lastWatched,omitempty"`
}

func readLetterboxdTake(raw json.RawMessage) (map[string]any, error) {
	var take LetterboxdTake
	if err := json.Unmarshal(raw, &take); err != nil {
		return nil, err
	}
	values := map[string]any{
		"reaction":     string(take.Reaction),
		"review":       take.Review,
		"last_watched": take.LastWatched,
	}
	if take.Rating != nil {
		values["rating"] = *take.Rating
	}
	return withoutEmpty(values), nil
}

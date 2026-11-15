package record

import (
	"fmt"
	"strings"
)

type Identity struct {
	Source     string
	ExternalID string
}

func (i Identity) String() string {
	return i.Source + ":" + i.ExternalID
}

func ParseIdentity(s string) (Identity, error) {
	source, externalID, found := strings.Cut(s, ":")
	if !found || source == "" || externalID == "" {
		return Identity{}, fmt.Errorf("identity %q is not source:external-id", s)
	}
	return Identity{Source: source, ExternalID: externalID}, nil
}

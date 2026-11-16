package store_test

import (
	"context"
	"testing"

	"github.com/pennane/pennanen.dev/apps/log/internal/record"
	"github.com/pennane/pennanen.dev/apps/log/internal/rule"
	"github.com/pennane/pennanen.dev/apps/log/internal/world/kind"
)

const measuringSource = "weigher"

func TestOverrideBeatsMeasure(t *testing.T) {
	ctx := context.Background()
	var measured int64
	weigher := rule.Proposer{
		Source:  measuringSource,
		Propose: func(context.Context, record.Queryer) ([]rule.Proposal, error) { return nil, nil },
		Measure: func(context.Context, record.Queryer) ([]rule.Measure, error) {
			return []rule.Measure{{ItemID: measured, Field: kind.FieldListening, Value: 42}}, nil
		},
	}
	s, _ := openStoreProposing(t, weigher)
	artist := record.Identity{Source: record.SourceSpotifyArtist, ExternalID: "X"}
	if err := s.PutClaim(ctx, artist, map[string]any{"id": "X", "name": "Xena"}); err != nil {
		t.Fatal(err)
	}
	id, err := s.FindByIdentity(ctx, artist.Source, artist.ExternalID)
	if err != nil {
		t.Fatal(err)
	}
	measured = id
	if err := s.ProposeAll(ctx); err != nil {
		t.Fatal(err)
	}
	v, err := s.GetItem(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if v.Verdicts.Values[kind.FieldListening] != int64(42) || v.Verdicts.SourceOf[kind.FieldListening] != measuringSource {
		t.Fatalf("measured verdicts = %+v", v.Verdicts)
	}
	if err := s.ChangeOverride(ctx, id, map[string]any{kind.FieldListening: 7.0}); err != nil {
		t.Fatal(err)
	}
	v, err = s.GetItem(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if v.Verdicts.Values[kind.FieldListening] != int64(7) || v.Verdicts.SourceOf[kind.FieldListening] != record.SourceAP {
		t.Fatalf("overridden verdicts = %+v", v.Verdicts)
	}
}

func TestAMeasureOfAFieldTheKindLacksIsNotAValue(t *testing.T) {
	ctx := context.Background()
	s, d := openStoreProposing(t)
	movie := record.Identity{Source: record.SourceTMDB, ExternalID: "949"}
	if err := s.PutClaim(ctx, movie, map[string]any{"id": 949, "title": "Heat"}); err != nil {
		t.Fatal(err)
	}
	id, err := s.FindByIdentity(ctx, movie.Source, movie.ExternalID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := d.Exec(`INSERT INTO measure (item_id, source, field, value) VALUES (?, ?, ?, 3)`, id, measuringSource, kind.FieldListening); err != nil {
		t.Fatal(err)
	}
	v, err := s.GetItem(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if _, present := v.Verdicts.Values[kind.FieldListening]; present {
		t.Fatalf("a movie got listening: %+v", v.Verdicts)
	}
}

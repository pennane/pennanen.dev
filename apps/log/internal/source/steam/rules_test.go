package steam

import (
	"context"
	"database/sql"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/pennane/pennanen.dev/apps/log/internal/platform/db"
	"github.com/pennane/pennanen.dev/apps/log/internal/rule"
	"github.com/pennane/pennanen.dev/apps/log/internal/take"
	"github.com/pennane/pennanen.dev/apps/log/internal/world/kind"
)

func minutes(n int64) *int64 { return &n }

func untriaged(itemID int64, signs gameSigns) loggedGame {
	return loggedGame{itemID: itemID, status: kind.Untriaged, signs: signs}
}

func TestSteamOwnedLogsTheGame(t *testing.T) {
	got := propose([]int64{7}, nil)
	want := []rule.Proposal{{ItemID: 7, Rule: RuleOwned, Subject: rule.SubjectEntry, Value: rule.EntryLogged, Strength: rule.Applying}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("proposals = %+v, want %+v", got, want)
	}
}

func TestHintsSuggest(t *testing.T) {
	cases := []struct {
		name  string
		signs gameSigns
		want  []rule.Proposal
	}{
		{"never played", gameSigns{playtimeMinutes: minutes(0)}, []rule.Proposal{
			{ItemID: 1, Rule: RuleLikelyNeverPlayed, Subject: rule.SubjectStatus, Value: statusNeverPlayed, Strength: rule.Suggesting},
		}},
		{"idled", gameSigns{playtimeMinutes: minutes(100), steamCategories: []string{steamTradingCardsCategory}}, []rule.Proposal{
			{ItemID: 1, Rule: RuleLikelyIdled, Subject: rule.SubjectTag, Value: kind.TagIdled, Strength: rule.Suggesting},
			{ItemID: 1, Rule: RuleLikelyIdled, Subject: rule.SubjectStatus, Value: statusNeverPlayed, Strength: rule.Suggesting},
		}},
		{"not a game", gameSigns{playtimeMinutes: minutes(500), genres: []string{"Utilities"}}, []rule.Proposal{
			{ItemID: 1, Rule: RuleLikelyNotAGame, Subject: rule.SubjectDismissal, Value: kind.Dismissed, Strength: rule.Suggesting},
		}},
		{"played", gameSigns{playtimeMinutes: minutes(200)}, []rule.Proposal{
			{ItemID: 1, Rule: RuleLikelyPlayed, Subject: rule.SubjectStatus, Value: statusPlayed, Strength: rule.Suggesting},
		}},
		{"played for achievements", gameSigns{playtimeMinutes: minutes(30), achievementsUnlocked: 1, steamCategories: []string{steamTradingCardsCategory}}, []rule.Proposal{
			{ItemID: 1, Rule: RuleLikelyPlayed, Subject: rule.SubjectStatus, Value: statusPlayed, Strength: rule.Suggesting},
		}},
		{"short play says nothing", gameSigns{playtimeMinutes: minutes(60)}, nil},
		{"unknown playtime says nothing", gameSigns{}, nil},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := propose(nil, []loggedGame{untriaged(1, c.signs)})
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("proposals = %+v, want %+v", got, c.want)
			}
		})
	}
}

func TestHintsOnlyForUntriagedGames(t *testing.T) {
	triaged := loggedGame{itemID: 2, status: statusPlayed, signs: gameSigns{playtimeMinutes: minutes(0)}}
	if got := propose(nil, []loggedGame{triaged}); len(got) != 0 {
		t.Errorf("proposals = %+v, want none for a triaged game", got)
	}
}

func openLog(t *testing.T) (context.Context, *sql.DB) {
	t.Helper()
	ctx := context.Background()
	d, err := db.Open(ctx, filepath.Join(t.TempDir(), "log.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { d.Close() })
	if err := take.WriteKindStatuses(ctx, d, kind.All()); err != nil {
		t.Fatal(err)
	}
	return ctx, d
}

func exec(t *testing.T, d *sql.DB, query string) {
	t.Helper()
	if _, err := d.Exec(query); err != nil {
		t.Fatal(err)
	}
}

func TestSteamProposerReadsStoredRecords(t *testing.T) {
	ctx, d := openLog(t)
	exec(t, d, `
		INSERT INTO item (id, kind, slug) VALUES (1, 'game', 'idler'), (2, 'game', 'hades'), (3, 'game', 'wallpaper'), (4, 'game', 'unlogged');
		INSERT INTO record_claim (source, external_id, data, fetched_at, changed_at) VALUES
			('steam', '10', '{}', '2026-01-01T00:00:00Z', '2026-01-01T00:00:00Z'),
			('steam', '20', '{}', '2026-01-01T00:00:00Z', '2026-01-01T00:00:00Z'),
			('steam', '30', '{}', '2026-01-01T00:00:00Z', '2026-01-01T00:00:00Z'),
			('steam', '40', '{}', '2026-01-01T00:00:00Z', '2026-01-01T00:00:00Z');
		INSERT INTO record_reading (source, external_id, version, reading) VALUES
			('steam', '10', 5, '{}'), ('steam', '20', 5, '{}'), ('steam', '30', 5, '{}'), ('steam', '40', 5, '{}');
		INSERT INTO item_identity (item_id, source, external_id) VALUES (1, 'steam', '10'), (2, 'steam', '20'), (3, 'steam', '30'), (4, 'steam', '40');
		INSERT INTO item_verdict (item_id, own, verdicts) VALUES
			(1, '{}', '{"values":{"playtime_minutes":100,"steam_categories":["Steam Trading Cards"]},"sourceOf":{"playtime_minutes":"steam","steam_categories":"steam-store"}}'),
			(2, '{}', '{"values":{"playtime_minutes":0},"sourceOf":{"playtime_minutes":"steam"}}'),
			(3, '{}', '{"values":{"playtime_minutes":0},"sourceOf":{"playtime_minutes":"steam"}}'),
			(4, '{}', '{"values":{"playtime_minutes":0},"sourceOf":{"playtime_minutes":"steam"}}');
		INSERT INTO entry (item_id, logged_by) VALUES (1, 'steam-owned'), (2, 'steam-owned'), (3, 'steam-owned');
		INSERT INTO status_choice (item_id, status, chosen_at) VALUES (2, 'finished', '2026-01-02T00:00:00Z');
		INSERT INTO dismissal (item_id, dismissed_at) VALUES (3, '2026-01-02T00:00:00Z');`)
	got, err := Proposer.Propose(ctx, d)
	if err != nil {
		t.Fatal(err)
	}
	entry := func(id int64) rule.Proposal {
		return rule.Proposal{ItemID: id, Rule: RuleOwned, Subject: rule.SubjectEntry, Value: rule.EntryLogged, Strength: rule.Applying}
	}
	want := []rule.Proposal{
		entry(1), entry(2), entry(3), entry(4),
		{ItemID: 1, Rule: RuleLikelyIdled, Subject: rule.SubjectTag, Value: kind.TagIdled, Strength: rule.Suggesting},
		{ItemID: 1, Rule: RuleLikelyIdled, Subject: rule.SubjectStatus, Value: statusNeverPlayed, Strength: rule.Suggesting},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("proposals = %+v\nwant %+v", got, want)
	}
	if Proposer.Source != "steam" {
		t.Errorf("source = %s", Proposer.Source)
	}
	if _, err := rule.Run(ctx, d, Proposer, "public"); err != nil {
		t.Errorf("running the proposals = %v", err)
	}
}

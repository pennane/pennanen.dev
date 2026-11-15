package kind_test

import (
	"github.com/pennane/pennanen.dev/apps/log/internal/record"
	"github.com/pennane/pennanen.dev/apps/log/internal/world/kind"
	"testing"
)

func TestAnOrderedLinkTypeIsPartOfOnly(t *testing.T) {
	partOf, known := kind.LinkTypeOf(kind.LinkPartOf)
	if !known || partOf.Order != kind.Ordered {
		t.Fatalf("part-of = %+v known %v", partOf, known)
	}
	for _, name := range []string{kind.LinkBy, kind.LinkWith} {
		linkType, known := kind.LinkTypeOf(name)
		if !known || linkType.Order != kind.Unordered {
			t.Fatalf("%s = %+v known %v", name, linkType, known)
		}
	}
	if _, known := kind.LinkTypeOf("nope"); known {
		t.Fatal("nope is a link type")
	}
}

func TestAKindIsFoundByItsIdentitySource(t *testing.T) {
	album, found := kind.KindOfIdentitySource(record.SourceSpotifyAlbum)
	if !found || album.Name != kind.Album {
		t.Fatalf("album = %+v found %v", album, found)
	}
	if _, found := kind.KindOfIdentitySource(record.SourceSteam); found {
		t.Fatal("steam is an identity source")
	}
}

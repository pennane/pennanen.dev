package kind

const (
	LinkBy     = "by"
	LinkWith   = "with"
	LinkPartOf = "part-of"

	SourceMatched        = "matched"
	SourceLegacy         = "legacy"
	SourceSpotifyTrackID = "spotify-track-id"
)

type Order string

const (
	Ordered   Order = "ordered"
	Unordered Order = "unordered"
)

type LinkType struct {
	Name  string
	Order Order
}

var linkTypes = []LinkType{
	{Name: LinkBy, Order: Unordered},
	{Name: LinkWith, Order: Unordered},
	{Name: LinkPartOf, Order: Ordered},
}

func LinkTypeOf(name string) (LinkType, bool) {
	for _, linkType := range linkTypes {
		if linkType.Name == name {
			return linkType, true
		}
	}
	return LinkType{}, false
}

func KindOfIdentitySource(source string) (Kind, bool) {
	for _, k := range All() {
		if k.IdentitySource == source {
			return k, true
		}
	}
	return Kind{}, false
}

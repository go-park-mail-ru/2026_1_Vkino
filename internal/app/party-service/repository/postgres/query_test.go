package postgres

import (
	"strings"
	"testing"
)

func TestOverviewFeaturedRoomsQueryReturnsOnlyPublicRooms(t *testing.T) {
	t.Parallel()

	if !strings.Contains(sqlGetOverviewFeaturedRooms, "where r.visibility = 'public'") {
		t.Fatalf("featured rooms query must filter public rooms: %s", sqlGetOverviewFeaturedRooms)
	}
}

package core

import "testing"

func TestParseCommunityDiscoveryTimeline(t *testing.T) {
	communities, cursor := parseCommunityDiscovery(readFixture(t, "communities/discovery.json"))
	if len(communities) != 1 || communities[0].ID != "700" || communities[0].Name != "Go Community" || communities[0].MemberCount != 1234 || cursor != "communities-next" {
		t.Fatalf("communities=%#v cursor=%q", communities, cursor)
	}
}

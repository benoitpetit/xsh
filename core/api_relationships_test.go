package core

import (
	"testing"
)

func TestParseFollowersYouKnow(t *testing.T) {
	users, cursor, err := parseRelationshipUsers(readFixture(t, "relationships/followers_you_know.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(users) != 1 || users[0].ID != "101" || users[0].Handle != "known_follower" || cursor != "followers-next" {
		t.Fatalf("users=%#v cursor=%q", users, cursor)
	}
}

func TestParseBlueVerifiedFollowers(t *testing.T) {
	users, cursor, err := parseRelationshipUsers(readFixture(t, "relationships/blue_verified_followers.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(users) != 1 || users[0].ID != "102" || !users[0].Verified || cursor != "blue-next" {
		t.Fatalf("users=%#v cursor=%q", users, cursor)
	}
}

func TestParseBlockedAccounts(t *testing.T) {
	users, cursor, err := parseRelationshipUsers(readFixture(t, "relationships/blocked.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(users) != 1 || users[0].ID != "103" || cursor != "blocked-next" {
		t.Fatalf("users=%#v cursor=%q", users, cursor)
	}
}

func TestParseMutedAccounts(t *testing.T) {
	users, cursor, err := parseRelationshipUsers(readFixture(t, "relationships/muted.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(users) != 1 || users[0].ID != "104" || cursor != "muted-next" {
		t.Fatalf("users=%#v cursor=%q", users, cursor)
	}
}

func TestParseRelationshipUsersEmpty(t *testing.T) {
	users, cursor, err := parseRelationshipUsers(map[string]interface{}{})
	if err != nil {
		t.Fatal(err)
	}
	if users == nil || len(users) != 0 || cursor != "" {
		t.Fatalf("users=%#v cursor=%q", users, cursor)
	}
}

func TestParseListDetails(t *testing.T) {
	list := parseListInfo(readFixture(t, "lists/list_details.json"))
	if list == nil || list.ID != "555" || list.Name != "Go Builders" || list.MemberCount != 42 || !list.IsPinned {
		t.Fatalf("list=%#v", list)
	}
}

func TestParseListMemberships(t *testing.T) {
	lists, cursor, err := parseListsFromTimeline(readFixture(t, "lists/memberships.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(lists) != 1 || lists[0].ID != "556" || lists[0].Name != "CLI Builders" || cursor != "memberships-next" {
		t.Fatalf("lists=%#v cursor=%q", lists, cursor)
	}
}

func TestGetListMembershipsIncludesUserID(t *testing.T) {
	fixture := readFixture(t, "lists/memberships.json")
	client := &XClient{requestWithOperationHook: func(_, _ string, params, _ map[string]interface{}, _ int, _, operation string) (map[string]interface{}, error) {
		if operation != "ListMemberships" {
			t.Fatalf("operation = %q", operation)
		}
		variables, ok := params["variables"].(map[string]interface{})
		if !ok || variables["userId"] != "viewer-1" {
			t.Fatalf("variables = %#v", variables)
		}
		return fixture, nil
	}}

	lists, cursor, err := GetListMemberships(client, "viewer-1", 10, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(lists) != 1 || lists[0].ID != "556" || cursor != "memberships-next" {
		t.Fatalf("lists=%#v cursor=%q", lists, cursor)
	}
}

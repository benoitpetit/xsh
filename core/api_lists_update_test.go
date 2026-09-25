package core

import "testing"

func TestUpdateListBuildsExpectedVariables(t *testing.T) {
	name := "Updated name"
	private := false
	client := &XClient{requestWithOperationHook: func(method, _ string, _ map[string]interface{}, jsonData map[string]interface{}, _ int, _, operation string) (map[string]interface{}, error) {
		if method != "POST" || operation != "UpdateList" {
			t.Fatalf("method=%q operation=%q", method, operation)
		}
		variables, ok := jsonData["variables"].(map[string]interface{})
		if !ok {
			t.Fatalf("jsonData=%#v", jsonData)
		}
		if variables["listId"] != "555" || variables["name"] != name || variables["isPrivate"] != private {
			t.Fatalf("variables=%#v", variables)
		}
		if _, ok := variables["description"]; ok {
			t.Fatalf("description should be omitted: %#v", variables)
		}
		return map[string]interface{}{"data": map[string]interface{}{"list": map[string]interface{}{"rest_id": "555"}}}, nil
	}}

	_, err := UpdateList(client, "555", ListUpdate{Name: &name, IsPrivate: &private})
	if err != nil {
		t.Fatal(err)
	}
}

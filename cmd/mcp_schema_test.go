package cmd

import (
	"encoding/json"
	"sort"
	"testing"

	"github.com/mark3labs/mcp-go/server"
)

// newTestMCPServer builds a server with every tool the MCP command registers.
func newTestMCPServer() *server.MCPServer {
	s := server.NewMCPServer("xsh", "1.0.0")
	registerReadTools(s)
	registerWriteTools(s)
	registerInfoTools(s)
	registerExtendedTools(s)
	return s
}

func TestCreateToolEmitsFlatInputSchema(t *testing.T) {
	tool := createTool("get_tweet", "Fetch a tweet", map[string]interface{}{
		"id": map[string]interface{}{
			"type":        "string",
			"description": "The tweet ID",
			"required":    true,
		},
		"thread": map[string]interface{}{
			"type":        "boolean",
			"description": "Return the conversation thread",
			"default":     false,
		},
	})

	if tool.InputSchema.Type != "object" {
		t.Errorf("InputSchema.Type = %q, want \"object\"", tool.InputSchema.Type)
	}

	// A nested "properties" object is what strict MCP clients reject.
	if nested, ok := tool.InputSchema.Properties["properties"]; ok {
		t.Errorf("InputSchema.Properties contains a nested \"properties\" entry: %#v", nested)
	}

	for _, reserved := range []string{"required", "type"} {
		if _, ok := tool.InputSchema.Properties[reserved]; ok {
			t.Errorf("InputSchema.Properties must not contain the schema keyword %q as a tool argument", reserved)
		}
	}

	id, ok := tool.InputSchema.Properties["id"].(map[string]interface{})
	if !ok {
		t.Fatalf("Properties[\"id\"] = %#v, want a JSON Schema object", tool.InputSchema.Properties["id"])
	}
	if id["type"] != "string" || id["description"] != "The tweet ID" {
		t.Errorf("Properties[\"id\"] = %#v, want the argument description preserved", id)
	}
	if _, leaked := id["required"]; leaked {
		t.Error("the \"required\" marker must be hoisted out of the argument schema")
	}

	thread, ok := tool.InputSchema.Properties["thread"].(map[string]interface{})
	if !ok {
		t.Fatalf("Properties[\"thread\"] = %#v, want a JSON Schema object", tool.InputSchema.Properties["thread"])
	}
	if _, ok := thread["default"]; !ok {
		t.Error("optional argument defaults must be preserved")
	}

	if len(tool.InputSchema.Required) != 1 || tool.InputSchema.Required[0] != "id" {
		t.Errorf("InputSchema.Required = %v, want [\"id\"]", tool.InputSchema.Required)
	}
}

func TestCreateToolDoesNotMutateParams(t *testing.T) {
	params := map[string]interface{}{
		"id": map[string]interface{}{"type": "string", "required": true},
	}
	createTool("t", "d", params)

	fragment, _ := params["id"].(map[string]interface{})
	if marker, leaked := fragment["required"]; !leaked || marker != true {
		t.Errorf("createTool mutated the caller's params: %#v", fragment)
	}
}

func TestCreateToolWithoutRequiredOmitsTheField(t *testing.T) {
	tool := createTool("get_lists", "Fetch lists", map[string]interface{}{})

	if tool.InputSchema.Required != nil {
		t.Errorf("InputSchema.Required = %v, want nil", tool.InputSchema.Required)
	}
}

// TestAllMCPToolSchemasAreSpecCompliant guards the wire format of every tool
// the server advertises. Strict clients validate each entry of "properties" as
// a JSON Schema object and drop the whole server when one is a scalar or array.
func TestAllMCPToolSchemasAreSpecCompliant(t *testing.T) {
	s := newTestMCPServer()

	// The registered tools are unexported, so drive the real JSON-RPC handler.
	response := s.HandleMessage(t.Context(), json.RawMessage(`{"jsonrpc":"2.0","id":1,"method":"tools/list"}`))
	encoded, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("json.Marshal(response) error = %v", err)
	}

	var envelope struct {
		Result struct {
			Tools []struct {
				Name        string `json:"name"`
				InputSchema struct {
					Type       string                     `json:"type"`
					Properties map[string]json.RawMessage `json:"properties"`
					Required   []string                   `json:"required"`
				} `json:"inputSchema"`
			} `json:"tools"`
		} `json:"result"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(encoded, &envelope); err != nil {
		t.Fatalf("tools/list response is not valid JSON: %v; response=%s", err, encoded)
	}
	if envelope.Error != nil {
		t.Fatalf("tools/list returned an error: %s", envelope.Error.Message)
	}
	if len(envelope.Result.Tools) == 0 {
		t.Fatal("tools/list advertised no tools")
	}

	requiredSeen := 0
	for _, tool := range envelope.Result.Tools {
		if tool.InputSchema.Type != "object" {
			t.Errorf("%s: inputSchema.type = %q, want \"object\"", tool.Name, tool.InputSchema.Type)
		}
		if len(tool.InputSchema.Required) > 0 {
			requiredSeen++
		}

		propertyNames := make([]string, 0, len(tool.InputSchema.Properties))
		for name := range tool.InputSchema.Properties {
			propertyNames = append(propertyNames, name)
		}
		sort.Strings(propertyNames)

		for _, name := range propertyNames {
			var decoded any
			if err := json.Unmarshal(tool.InputSchema.Properties[name], &decoded); err != nil {
				t.Errorf("%s: properties[%q] is not valid JSON: %v", tool.Name, name, err)
				continue
			}
			// Every entry must itself be a JSON Schema object. A bare scalar
			// or array here is what makes strict clients reject the server.
			if _, ok := decoded.(map[string]any); !ok {
				t.Errorf("%s: properties[%q] = %s, want a JSON Schema object", tool.Name, name, tool.InputSchema.Properties[name])
			}
		}

		// Required entries must be declared arguments, and arguments marked
		// required must be listed.
		declared := make(map[string]bool, len(tool.InputSchema.Properties))
		for name := range tool.InputSchema.Properties {
			declared[name] = true
		}
		for _, name := range tool.InputSchema.Required {
			if !declared[name] {
				t.Errorf("%s: required entry %q is not declared in properties", tool.Name, name)
			}
		}
	}

	if requiredSeen == 0 {
		t.Error("no tool declares required arguments; the \"required\" marker is being dropped")
	}
}

// TestMCPToolsListAdvertisesToolsCapability checks the handshake advertises the
// tools capability, which clients require before they will call tools/list.
func TestMCPToolsListAdvertisesToolsCapability(t *testing.T) {
	s := newTestMCPServer()

	response := s.HandleMessage(t.Context(), json.RawMessage(
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1"}}}`,
	))
	encoded, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("json.Marshal(response) error = %v", err)
	}

	var envelope struct {
		Result struct {
			Capabilities struct {
				Tools *struct {
					ListChanged bool `json:"listChanged"`
				} `json:"tools"`
			} `json:"capabilities"`
		} `json:"result"`
	}
	if err := json.Unmarshal(encoded, &envelope); err != nil {
		t.Fatalf("initialize response is not valid JSON: %v; response=%s", err, encoded)
	}
	if envelope.Result.Capabilities.Tools == nil {
		t.Errorf("initialize did not advertise the tools capability; response=%s", encoded)
	}
}

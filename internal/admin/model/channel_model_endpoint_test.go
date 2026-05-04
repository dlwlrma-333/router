package model

import "testing"

func TestNormalizeProviderModelSupportedEndpointsFiltersByModelType(t *testing.T) {
	got := NormalizeProviderModelSupportedEndpoints(ProviderModelTypeText, []string{
		ChannelModelEndpointResponses,
		ChannelModelEndpointImages,
		ChannelModelEndpointChat,
		ChannelModelEndpointResponses,
	})
	if len(got) != 2 || got[0] != ChannelModelEndpointChat || got[1] != ChannelModelEndpointResponses {
		t.Fatalf("NormalizeProviderModelSupportedEndpoints = %#v, want chat+responses", got)
	}
}

func TestBuildChannelModelEndpointRowsUsesProviderCatalogCandidates(t *testing.T) {
	rows := []ChannelModel{
		{
			ChannelId:     "channel-1",
			Model:         "gpt-5.4",
			UpstreamModel: "gpt-5.4",
			Provider:      "openai",
			Type:          ProviderModelTypeText,
			Selected:      true,
		},
	}
	providerEndpoints := map[string][]string{
		buildProviderModelEndpointKey("openai", "gpt-5.4"): {
			ChannelModelEndpointChat,
			ChannelModelEndpointResponses,
		},
	}

	got := BuildChannelModelEndpointRowsWithProviderEndpoints(nil, rows, providerEndpoints)
	if len(got) != 2 {
		t.Fatalf("len(got)=%d, want 2", len(got))
	}
	if got[0].Endpoint != ChannelModelEndpointChat {
		t.Fatalf("got[0].Endpoint=%q, want %q", got[0].Endpoint, ChannelModelEndpointChat)
	}
	if got[1].Endpoint != ChannelModelEndpointResponses {
		t.Fatalf("got[1].Endpoint=%q, want %q", got[1].Endpoint, ChannelModelEndpointResponses)
	}
}

func TestBuildChannelModelEndpointRowsPreservesExistingDisabledEndpointState(t *testing.T) {
	existing := []ChannelModelEndpoint{
		{ChannelId: "channel-1", Model: "gpt-5.4", Endpoint: ChannelModelEndpointResponses, Enabled: false},
	}
	rows := []ChannelModel{
		{
			ChannelId:     "channel-1",
			Model:         "gpt-5.4",
			UpstreamModel: "gpt-5.4",
			Provider:      "openai",
			Type:          ProviderModelTypeText,
			Selected:      true,
		},
	}
	providerEndpoints := map[string][]string{
		buildProviderModelEndpointKey("openai", "gpt-5.4"): {ChannelModelEndpointResponses},
	}

	got := BuildChannelModelEndpointRowsWithProviderEndpoints(existing, rows, providerEndpoints)
	if len(got) != 1 {
		t.Fatalf("BuildChannelModelEndpointRows len = %d, want 1", len(got))
	}
	if got[0].Endpoint != ChannelModelEndpointResponses || got[0].Enabled {
		t.Fatalf("responses endpoint = (%q, %t), want (%q, false)", got[0].Endpoint, got[0].Enabled, ChannelModelEndpointResponses)
	}
}

func TestBuildChannelModelEndpointRowsDoesNotFallbackToChannelModelEndpoint(t *testing.T) {
	rows := []ChannelModel{
		{
			ChannelId:     "channel-1",
			Model:         "gpt-5.4",
			UpstreamModel: "gpt-5.4",
			Provider:      "openai",
			Type:          ProviderModelTypeText,
			Selected:      true,
			Endpoint:      ChannelModelEndpointResponses,
			Endpoints:     []string{ChannelModelEndpointResponses},
		},
	}

	got := BuildChannelModelEndpointRowsWithProviderEndpoints(nil, rows, nil)
	if len(got) != 0 {
		t.Fatalf("len(got)=%d, want 0 without provider catalog endpoint candidates", len(got))
	}
}

func TestBuildDisabledChannelModelEndpointRowsMarksOnlyTargetEndpoint(t *testing.T) {
	rows := []ChannelModelEndpoint{
		{ChannelId: "channel-1", Model: "gpt-5.4", Endpoint: ChannelModelEndpointChat, Enabled: true},
		{ChannelId: "channel-1", Model: "gpt-5.4", Endpoint: ChannelModelEndpointResponses, Enabled: true},
	}

	got, changed := buildDisabledChannelModelEndpointRows(rows, "channel-1", "gpt-5.4", ChannelModelEndpointResponses)
	if !changed {
		t.Fatalf("changed = false, want true")
	}
	if len(got) != 2 {
		t.Fatalf("len(got) = %d, want 2", len(got))
	}
	if !got[0].Enabled {
		t.Fatalf("chat endpoint enabled = false, want true")
	}
	if got[1].Enabled {
		t.Fatalf("responses endpoint enabled = true, want false")
	}
}

func TestNormalizeRequestedChannelModelEndpointMessagesMapsToMessages(t *testing.T) {
	if got := NormalizeRequestedChannelModelEndpoint("/v1/messages"); got != ChannelModelEndpointMessages {
		t.Fatalf("NormalizeRequestedChannelModelEndpoint(/v1/messages)=%q, want %q", got, ChannelModelEndpointMessages)
	}
	if got := NormalizeRequestedChannelModelEndpoint("/api/v1/public/messages"); got != ChannelModelEndpointMessages {
		t.Fatalf("NormalizeRequestedChannelModelEndpoint(/api/v1/public/messages)=%q, want %q", got, ChannelModelEndpointMessages)
	}
}

func TestIsChannelModelRequestEndpointSupportedByEndpointMapNoBridgeCompatibility(t *testing.T) {
	endpointMap := map[string]bool{
		ChannelModelEndpointResponses: true,
	}
	if supported, explicit := IsChannelModelRequestEndpointSupportedByEndpointMap(endpointMap, ChannelModelEndpointChat); !explicit || supported {
		t.Fatalf("chat request support via responses endpoint = (%t, %t), want (false, true)", supported, explicit)
	}
	if supported, explicit := IsChannelModelRequestEndpointSupportedByEndpointMap(endpointMap, ChannelModelEndpointMessages); !explicit || supported {
		t.Fatalf("messages request support via responses endpoint = (%t, %t), want (false, true)", supported, explicit)
	}
}

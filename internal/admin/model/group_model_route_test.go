package model

import "testing"

func TestNormalizeGroupModelRouteRowsPreserveOrder_DeduplicatesByPrimaryKey(t *testing.T) {
	rows := []GroupModelRoute{
		{
			Group:         " group-a ",
			Model:         "gpt-4.1",
			ChannelId:     " channel-1 ",
			UpstreamModel: "upstream-a",
			Enabled:       true,
		},
		{
			Group:         "group-a",
			Model:         "gpt-4.1",
			ChannelId:     "channel-1",
			UpstreamModel: "upstream-b",
			Enabled:       false,
		},
		{
			Group:     "group-a",
			Model:     "gpt-4.1",
			ChannelId: "channel-2",
			Enabled:   true,
		},
	}

	got := normalizeGroupModelRouteRowsPreserveOrder(rows)
	if len(got) != 2 {
		t.Fatalf("normalizeGroupModelRouteRowsPreserveOrder returned %d rows, want 2", len(got))
	}
	if got[0].Group != "group-a" || got[0].Model != "gpt-4.1" || got[0].ChannelId != "channel-1" {
		t.Fatalf("unexpected first row key: %#v", got[0])
	}
	if got[0].UpstreamModel != "upstream-a" {
		t.Fatalf("unexpected first row upstream model: %q", got[0].UpstreamModel)
	}
	if got[1].UpstreamModel != "gpt-4.1" {
		t.Fatalf("unexpected fallback upstream model: %q", got[1].UpstreamModel)
	}
}

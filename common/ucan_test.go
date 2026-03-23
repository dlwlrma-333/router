package common

import (
	"testing"

	"github.com/yeying-community/router/common/config"
)

func TestResolveUcanRequiredCapabilitySets_DefaultIncludesAppCompat(t *testing.T) {
	prevResource := config.UcanResource
	prevAction := config.UcanAction
	prevAud := config.UcanAud
	defer func() {
		config.UcanResource = prevResource
		config.UcanAction = prevAction
		config.UcanAud = prevAud
	}()

	config.UcanResource = ""
	config.UcanAction = ""
	config.UcanAud = "did:web:router.example.com"

	sets := ResolveUcanRequiredCapabilitySets()
	assertHasSingleCapabilitySet(t, sets, UcanCapability{
		Resource: "llm:router.example.com",
		Action:   config.DefaultUcanAction,
	})
	assertHasSingleCapabilitySet(t, sets, UcanCapability{
		Resource: config.AppCompatUcanResource,
		Action:   config.AppCompatUcanAction,
	})
	assertHasSingleCapabilitySet(t, sets, UcanCapability{
		Resource: config.CompatUcanResource,
		Action:   config.CompatUcanAction,
	})
	assertHasSingleCapabilitySet(t, sets, UcanCapability{
		Resource: config.LegacyUcanResource,
		Action:   config.LegacyUcanAction,
	})
}

func TestCapsAllow_AppWildcardRequiredMatchesExactAvailable(t *testing.T) {
	available := []UcanCapability{
		{Resource: "app:localhost-3020", Action: "invoke"},
	}
	required := []UcanCapability{
		{Resource: "app:*", Action: "invoke"},
	}
	if !capsAllow(available, required) {
		t.Fatalf("expected app wildcard requirement to match exact app capability")
	}
}

func TestCapsAllow_RequiredInvokeDoesNotMatchWrite(t *testing.T) {
	available := []UcanCapability{
		{Resource: "app:localhost-3020", Action: "write"},
	}
	required := []UcanCapability{
		{Resource: "app:*", Action: "invoke"},
	}
	if capsAllow(available, required) {
		t.Fatalf("expected invoke requirement to reject write-only capability")
	}
}

func assertHasSingleCapabilitySet(t *testing.T, sets [][]UcanCapability, target UcanCapability) {
	t.Helper()
	for _, set := range sets {
		if len(set) != 1 {
			continue
		}
		if capabilityEquals(set[0], target) {
			return
		}
	}
	t.Fatalf("missing capability set: %#v in %#v", target, sets)
}

package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSetupClaude(t *testing.T) {
	t.Parallel()

	pc := ProviderConfig{ExtraHeaders: map[string]string{"anthropic-beta": "foo"}}
	pc.SetupClaude()
	pc.SetupClaude()
	require.Equal(t, "foo,oauth-2025-04-20", pc.ExtraHeaders["anthropic-beta"])
	require.Contains(t, pc.SystemPromptPrefix, "Claude Code")

	var empty ProviderConfig
	empty.SetupClaude()
	require.Equal(t, "oauth-2025-04-20", empty.ExtraHeaders["anthropic-beta"])
}

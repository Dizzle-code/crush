package tools

import (
	"strings"
	"testing"

	"github.com/charmbracelet/crush/internal/config"
	"github.com/stretchr/testify/require"
)

func TestBashDescriptionIsNotHTMLEscaped(t *testing.T) {
	t.Parallel()
	desc := bashDescription(&config.Attribution{TrailerStyle: "assisted-by", GeneratedWith: true}, "test-model", []string{"curl"})
	require.Contains(t, desc, "cat <<'EOF'")
	require.NotContains(t, desc, "&lt;")
	require.NotContains(t, desc, "&#39;")
	require.True(t, strings.Contains(desc, "Assisted-by: Crush:test-model"))
	require.True(t, strings.HasSuffix(strings.TrimSpace(desc), "</git_commits>"))
}

package agents

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestManagedNPMRuntimeContracts(t *testing.T) {
	tests := []struct {
		name        string
		agent       ManagedNPMRuntimeAgent
		wantPackage string
		wantACPArgs []string
	}{
		{"claude", NewClaudeACP(), "@agentclientprotocol/claude-agent-acp", nil},
		{"codex", NewCodexACP(), "@agentclientprotocol/codex-acp", nil},
		{"opencode", NewOpenCodeACP(), "opencode-ai", []string{"acp", "--print-logs", "--log-level", "ERROR"}},
		{"copilot", NewCopilotACP(), "@github/copilot", []string{"--acp"}},
		{"gemini", NewGemini(), "@google/gemini-cli", []string{"--acp"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := tt.agent.ManagedNPMRuntime()
			if got := spec.Package; got != tt.wantPackage {
				t.Fatalf("Package = %q, want %q", got, tt.wantPackage)
			}
			if got := spec.ACPArgs; !slices.Equal(got, tt.wantACPArgs) {
				t.Fatalf("ACPArgs = %#v, want %#v", got, tt.wantACPArgs)
			}

			cached := spec.CachedACPCommand()
			wantCached := append([]string{"npx", "--yes", "--prefer-offline", tt.wantPackage}, tt.wantACPArgs...)
			if !slices.Equal(cached.Args(), wantCached) {
				t.Fatalf("CachedACPCommand = %#v, want %#v", cached.Args(), wantCached)
			}
			assertUnversionedPackage(t, cached.Args(), tt.wantPackage)

			update := spec.CacheUpdateCommand()
			wantUpdate := []string{
				"npm", "exec", "--yes", "--prefer-online",
				"--package=" + tt.wantPackage, "--", "node", "-e", "",
			}
			if !slices.Equal(update.Args(), wantUpdate) {
				t.Fatalf("CacheUpdateCommand = %#v, want %#v", update.Args(), wantUpdate)
			}
			if strings.Contains(strings.Join(update.Args(), " "), "latest") {
				t.Fatalf("CacheUpdateCommand contains explicit latest: %#v", update.Args())
			}
		})
	}
}

func TestManagedNPMRuntimeExecutionCacheKeyMatchesNPM(t *testing.T) {
	spec := ManagedNPMRuntimeSpec{Package: "opencode-ai"}
	if got := spec.ExecutionCacheKey(); got != "e2094862b59aac7b" {
		t.Fatalf("ExecutionCacheKey = %q, want npm key e2094862b59aac7b", got)
	}
}

// TestManagedNPMRuntimeNativeBinaryPreference covers the binary-first launch
// used by opencode-acp: direct-binary commands when the binary is on PATH,
// npx fallback everywhere else, and a stable execution-cache key regardless
// of the NativeBinary field (the npm key derives from the package only).
func TestManagedNPMRuntimeNativeBinaryPreference(t *testing.T) {
	nativeBin := "opencode"

	// Absent binary -> npx fallback with unchanged managed-runtime contract.
	absent := ManagedNPMRuntimeSpec{
		Package:      "opencode-ai",
		ACPArgs:      []string{"acp", "--print-logs", "--log-level", "ERROR"},
		NativeBinary: nativeBin,
	}
	t.Setenv("PATH", t.TempDir())
	if absent.NativeBinaryOnPath() {
		t.Fatal("NativeBinaryOnPath() = true with empty PATH dir, want false")
	}
	wantNpx := []string{"npx", "--yes", "--prefer-offline", "opencode-ai", "acp", "--print-logs", "--log-level", "ERROR"}
	if got := absent.RefreshCommand().Args(); !slices.Equal(got, wantNpx) {
		t.Fatalf("RefreshCommand (absent binary) = %#v, want %#v", got, wantNpx)
	}

	// Present binary -> direct-binary command with no npm on the hot path.
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, nativeBin), []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatalf("write fake binary: %v", err)
	}
	t.Setenv("PATH", dir)
	if !absent.NativeBinaryOnPath() {
		t.Fatal("NativeBinaryOnPath() = false with binary on PATH, want true")
	}
	wantNative := []string{"opencode", "acp", "--print-logs", "--log-level", "ERROR"}
	if got := absent.NativeCommand().Args(); !slices.Equal(got, wantNative) {
		t.Fatalf("NativeCommand() = %#v, want %#v", got, wantNative)
	}
	if got := absent.RefreshCommand().Args(); !slices.Equal(got, wantNative) {
		t.Fatalf("RefreshCommand (present binary) = %#v, want %#v", got, wantNative)
	}

	// RefreshCommand stays npx when no NativeBinary is configured at all.
	plain := ManagedNPMRuntimeSpec{Package: "opencode-ai", ACPArgs: []string{"acp"}}
	if got := plain.RefreshCommand().Args(); !slices.Equal(got, []string{"npx", "--yes", "--prefer-offline", "opencode-ai", "acp"}) {
		t.Fatalf("RefreshCommand (no native) = %#v, want npx", got)
	}

	// The npm execution-cache key must not depend on the NativeBinary field.
	withNative := ManagedNPMRuntimeSpec{Package: "opencode-ai", NativeBinary: nativeBin}
	if got := withNative.ExecutionCacheKey(); got != "e2094862b59aac7b" {
		t.Fatalf("ExecutionCacheKey with NativeBinary = %q, want e2094862b59aac7b", got)
	}
}

func assertUnversionedPackage(t *testing.T, argv []string, wantPackage string) {
	t.Helper()
	packageArg := argv[3]
	if packageArg != wantPackage {
		t.Fatalf("package argv = %q, want unversioned %q", packageArg, wantPackage)
	}
	if packageArg == wantPackage+"@latest" {
		t.Fatalf("package argv contains explicit latest: %q", packageArg)
	}
}

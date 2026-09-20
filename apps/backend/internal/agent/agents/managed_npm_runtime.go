package agents

import (
	"crypto/sha512"
	"encoding/hex"
	"os/exec"
)

// ManagedNPMRuntimeSpec defines a built-in npm-distributed ACP runtime.
// Package and ACPArgs must come from trusted agent metadata, never request
// input, because update jobs execute them directly.
//
// NativeBinary optionally names a standalone CLI that ships the same ACP
// interface as the npm package (e.g. "opencode" for opencode-ai). When that
// binary is resolvable on PATH, launch/probe/refresh commands run it directly
// and skip the per-launch npm resolution entirely — the same pattern used by
// local-first harnesses such as CodeNomad to launch opencode. npx stays the
// fallback for environments without the binary (fresh hosts, containers,
// remotes), so installing the runtime through the managed npm path keeps
// working unchanged.
type ManagedNPMRuntimeSpec struct {
	Package      string
	ACPArgs      []string
	NativeBinary string
}

// ExecutionCacheKey returns npm's deterministic _npx execution-tree key for
// this trusted package spec. npm derives it from the full package string using
// SHA-512 and the first 16 lowercase hexadecimal characters.
func (s ManagedNPMRuntimeSpec) ExecutionCacheKey() string {
	digest := sha512.Sum512([]byte(s.Package))
	return hex.EncodeToString(digest[:])[:16]
}

// CachedACPCommand returns the normal launch command. The package is
// intentionally unversioned and prefer-offline lets npm reuse a suitable
// execution-cache entry while still fetching when the cache is empty.
func (s ManagedNPMRuntimeSpec) CachedACPCommand() Command {
	args := []string{"npx", "--yes", "--prefer-offline", s.Package}
	args = append(args, s.ACPArgs...)
	return NewCommand(args...)
}

// CacheUpdateCommand returns the explicit cache-refresh command. npm exec
// installs the built-in unversioned package with online freshness before
// running a no-op under the prepared execution environment.
func (s ManagedNPMRuntimeSpec) CacheUpdateCommand() Command {
	return NewCommand(
		"npm",
		"exec",
		"--yes",
		"--prefer-online",
		"--package="+s.Package,
		"--",
		"node",
		"-e",
		"",
	)
}

// NativeCommand returns the direct-binary launch command
// ("<binary> <acpArgs...>") that bypasses npm resolution entirely. Callers
// must gate it on NativeBinaryOnPath() (or the lifecycle's
// CommandOptions.PreferNativeBinary probe) so containers and remotes without
// the binary keep using the managed npm runtime.
func (s ManagedNPMRuntimeSpec) NativeCommand() Command {
	args := []string{s.NativeBinary}
	args = append(args, s.ACPArgs...)
	return NewCommand(args...)
}

// NativeBinaryOnPath reports whether the optional native binary is resolvable
// from the host PATH. LookPath semantics match the lifecycle's standalone
// probe (preferNativeBinary) and the settings preview, so host-side callers
// reach the same conclusion without duplicating the probe.
func (s ManagedNPMRuntimeSpec) NativeBinaryOnPath() bool {
	if s.NativeBinary == "" {
		return false
	}
	_, err := exec.LookPath(s.NativeBinary)
	return err == nil
}

// RefreshCommand returns the command used to ACP-probe the runtime after an
// update job. Prefers the native binary when available so the refresh never
// depends on npm cache state — the failure mode behind "ACP initialize
// failed: peer disconnected before response" when the npm packument cache is
// stale or the execution-cache entry lacks its postinstall binary bootstrap.
func (s ManagedNPMRuntimeSpec) RefreshCommand() Command {
	if s.NativeBinaryOnPath() {
		return s.NativeCommand()
	}
	return s.CachedACPCommand()
}

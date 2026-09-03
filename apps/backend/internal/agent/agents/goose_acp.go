//nolint:dupl,goconst // Native-binary ACP agents (Kiro, Qoder, Hermes, Devin, ...) follow the same minimal scaffold; differences are the binary name, argv, and auth surface. Shared literals live in every peer file by convention.
package agents

import (
	"context"
	_ "embed"
	"time"

	"github.com/kandev/kandev/internal/agent/usage"
	"github.com/kandev/kandev/pkg/agent"
)

//go:embed logos/goose_light.svg
var gooseLogoLight []byte

//go:embed logos/goose_dark.svg
var gooseLogoDark []byte

const gooseBin = "goose"

var (
	_ Agent            = (*GooseACP)(nil)
	_ PassthroughAgent = (*GooseACP)(nil)
	_ InferenceAgent   = (*GooseACP)(nil)
	_ LoginAgent       = (*GooseACP)(nil)
)

// GooseACP implements Agent for the Goose coding agent (Agentic AI
// Foundation / AAIF, formerly Block) using native ACP over stdin/stdout
// (`goose acp`). Goose is installed as a standalone native binary
// (`download_cli.sh`, Homebrew `block-goose-cli`, or `pip install goose-ai`),
// so the launch command is the bare `goose` binary discovered on PATH.
//
// Auth is declarative: Goose reads its provider + credential config from
// ~/.config/goose (not environment variables), so `goose configure` is the
// interactive setup, and the primary remote-auth path is a config-file copy.
type GooseACP struct {
	StandardPassthrough
}

func NewGooseACP() *GooseACP {
	return &GooseACP{
		StandardPassthrough: StandardPassthrough{
			PermSettings: emptyPermSettings,
			Cfg: PassthroughConfig{
				Supported:      true,
				Label:          "CLI Passthrough",
				Description:    "Show terminal directly instead of chat interface",
				PassthroughCmd: NewCommand(gooseBin),
				IdleTimeout:    3 * time.Second,
				BufferMaxBytes: DefaultBufferMaxBytes,
			},
		},
	}
}

func (a *GooseACP) ID() string          { return "goose-acp" }
func (a *GooseACP) Name() string        { return "Goose ACP Agent" }
func (a *GooseACP) DisplayName() string { return "Goose" }
func (a *GooseACP) Description() string {
	return "AAIF Goose coding agent using the ACP protocol via goose acp. Local, extensible, open source."
}
func (a *GooseACP) Enabled() bool     { return true }
func (a *GooseACP) DisplayOrder() int { return 22 }

func (a *GooseACP) Logo(v LogoVariant) []byte {
	if v == LogoDark {
		return gooseLogoDark
	}
	return gooseLogoLight
}

func (a *GooseACP) IsInstalled(ctx context.Context) (*DiscoveryResult, error) {
	// `goose acp` is a blocking server, so detection is presence-based
	// (WithCommand) rather than a WithCommandCheck probe that could hang or
	// open a provider config during discovery.
	result, err := Detect(ctx, WithCommand(gooseBin))
	if err != nil {
		return result, err
	}
	result.SupportsMCP = true
	result.Capabilities = DiscoveryCapabilities{
		SupportsSessionResume: true,
	}
	return result, nil
}

func (a *GooseACP) BuildCommand(_ CommandOptions) Command {
	return Cmd(gooseBin, "acp").Build()
}

func (a *GooseACP) Runtime() *RuntimeConfig {
	canRecover := true
	return &RuntimeConfig{
		Cmd:            Cmd(gooseBin, "acp").Build(),
		WorkingDir:     "{workspace}",
		Env:            map[string]string{},
		ResourceLimits: DefaultResourceLimits,
		Protocol:       agent.ProtocolACP,
		// Goose supports Stdio and HTTP (StreamableHttp) MCP servers but
		// rejects SSE ("migrate to streamable_http"); do not force SSE.
		// Goose authenticates via its ~/.config/goose config files (not
		// provider API-key env vars), so no StripEnv is declared.
		SessionConfig: SessionConfig{
			NativeSessionResume: true,
			CanRecover:          &canRecover,
			SessionDirTemplate:  "{home}/.local/share/goose",
			SessionDirTarget:    "/root/.local/share/goose",
		},
	}
}

func (a *GooseACP) RemoteAuth() *RemoteAuth {
	return &RemoteAuth{
		Methods: []RemoteAuthMethod{
			{
				Type:  "files",
				Label: "Copy Goose config files",
				SourceFiles: map[string][]string{
					"darwin": {".config/goose/config.yaml", ".config/goose/extensions.yaml"},
					"linux":  {".config/goose/config.yaml", ".config/goose/extensions.yaml"},
				},
				// Provider + model config only. Sessions, checkpoints, and
				// history under ~/.local/share/goose are deliberately
				// excluded; credentials stored in the OS keyring are not
				// represented by config files and must be set up remotely.
				TargetRelDir: ".config/goose",
			},
		},
	}
}

// goose configure is the interactive provider/model setup wizard (OAuth
// flows, API keys). Credentials land in ~/.config/goose.
func (a *GooseACP) LoginCommand() *LoginCommand {
	return &LoginCommand{
		Cmd:         []string{gooseBin, "configure"},
		Description: "Configure Goose model provider credentials.",
	}
}

func (a *GooseACP) InstallScript() string {
	// CONFIGURE=false skips the interactive provider wizard so installs in
	// remote/headless environments are deterministic.
	return "curl -fsSL https://github.com/aaif-goose/goose/releases/download/stable/download_cli.sh | CONFIGURE=false bash"
}

func (a *GooseACP) PermissionSettings() map[string]PermissionSetting {
	return emptyPermSettings
}

func (a *GooseACP) InferenceConfig() *InferenceConfig {
	return &InferenceConfig{
		Supported: true,
		Command:   NewCommand(gooseBin, "acp"),
	}
}

func (a *GooseACP) BillingType() usage.BillingType { return defaultBillingType() }

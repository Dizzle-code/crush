package cmd

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/crush/internal/config"
	"github.com/charmbracelet/crush/internal/login"
	"github.com/charmbracelet/crush/internal/oauth/claude"
	"github.com/charmbracelet/crush/internal/oauth/copilot"
	"github.com/charmbracelet/crush/internal/workspace"
	"github.com/pkg/browser"
	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Aliases: []string{"auth"},
	Use:     "login [platform]",
	Short:   "Login Crush to a platform",
	Long: `Login Crush to a specified platform.
	The platform should be provided as an argument.
	Available platforms are: hyper, copilot, claude (anthropic), openai (chatgpt), grok (xai).`,
	Example: `
	# Authenticate with Charm Hyper
	crush login

	# Authenticate with GitHub Copilot
	crush login copilot

	# Authenticate with a Claude Pro/Max subscription
	crush login claude

	# Authenticate with a ChatGPT (OpenAI) account
	crush login openai

	# Authenticate with a Grok (xAI) account
	crush login grok

	# Force re-authentication even if already logged in
	crush login -f copilot
	`,
	ValidArgs: []cobra.Completion{
		"hyper",
		"copilot",
		"github",
		"github-copilot",
		"claude",
		"anthropic",
		"openai",
		"chatgpt",
		"grok",
		"xai",
	},
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, cleanup, err := setupWorkspaceWithProgressBar(cmd)
		if err != nil {
			return err
		}
		defer cleanup()

		provider := "hyper"
		if len(args) > 0 {
			provider = args[0]
		}
		force, _ := cmd.Flags().GetBool("force")
		switch provider {
		case "hyper":
			return loginHyper(ws, force)
		case "copilot", "github", "github-copilot":
			return loginCopilot(ws, force)
		case "claude", "anthropic":
			return loginClaude(ws, force)
		case "openai", "chatgpt":
			return loginOpenAI(ws, force)
		case "grok", "xai":
			return loginGrok(ws, force)
		default:
			return fmt.Errorf("unknown platform: %s", args[0])
		}
	},
}

func init() {
	loginCmd.Flags().BoolP("force", "f", false, "Force re-authentication even if already logged in")
}

func loginHyper(ws workspace.Workspace, force bool) error {
	if !force {
		cfg := ws.Config()
		if cfg != nil {
			if pc, ok := cfg.Providers.Get("hyper"); ok && pc.OAuthToken != nil {
				fmt.Println("You are already logged in to Hyper.")
				fmt.Println("Use --force to re-authenticate.")
				return nil
			}
		}
	}

	ctx := getLoginContext()
	token, err := login.Run(ctx, login.PlatformHyper)
	if err != nil {
		return err
	}

	if err := ws.SetProviderAPIKey(config.ScopeGlobal, "hyper", token); err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("You're now authenticated with Hyper!")
	return nil
}

func loginCopilot(ws workspace.Workspace, force bool) error {
	if !force {
		cfg := ws.Config()
		if cfg != nil {
			if pc, ok := cfg.Providers.Get("copilot"); ok && pc.OAuthToken != nil {
				fmt.Println("You are already logged in to GitHub Copilot.")
				fmt.Println("Use --force to re-authenticate.")
				return nil
			}
		}
	}

	ctx := getLoginContext()

	if diskToken, hasDiskToken := copilot.RefreshTokenFromDisk(); hasDiskToken {
		fmt.Println("Found existing GitHub Copilot token on disk. Using it to authenticate...")
		token, err := copilot.RefreshToken(ctx, diskToken)
		if err != nil {
			return fmt.Errorf("unable to refresh token from disk: %w", err)
		}
		if err := ws.SetProviderAPIKey(config.ScopeGlobal, "copilot", token); err != nil {
			return err
		}
		fmt.Println()
		fmt.Println("You're now authenticated with GitHub Copilot!")
		return nil
	}

	token, err := login.Run(ctx, login.PlatformCopilot)
	if err != nil {
		if errors.Is(err, copilot.ErrNotAvailable) {
			fmt.Println()
			fmt.Println("GitHub Copilot is unavailable for this account. To signup, go to the following page:")
			fmt.Println()
			lipgloss.Println(lipgloss.NewStyle().Hyperlink(copilot.SignupURL, "id=copilot-signup").Render(copilot.SignupURL))
			fmt.Println()
			fmt.Println("You may be able to request free access if eligible. For more information, see:")
			fmt.Println()
			lipgloss.Println(lipgloss.NewStyle().Hyperlink(copilot.FreeURL, "id=copilot-free").Render(copilot.FreeURL))
		}
		return err
	}

	if err := ws.SetProviderAPIKey(config.ScopeGlobal, "copilot", token); err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("You're now authenticated with GitHub Copilot!")
	return nil
}

func loginClaude(ws workspace.Workspace, force bool) error {
	if !force {
		if cfg := ws.Config(); cfg != nil {
			if pc, ok := cfg.Providers.Get("anthropic"); ok && pc.OAuthToken != nil {
				fmt.Println("You are already logged in to Claude with a subscription.")
				fmt.Println("Use --force to re-authenticate.")
				return nil
			}
		}
	}

	ctx := getLoginContext()
	flow, err := claude.NewFlow()
	if err != nil {
		return err
	}
	url := flow.AuthorizeURL()
	fmt.Println("Open this URL, sign in to Claude and copy the code shown afterwards:")
	fmt.Println()
	lipgloss.Println(lipgloss.NewStyle().Hyperlink(url, "id=claude").Render(url))
	fmt.Println()
	_ = browser.OpenURL(url)

	fmt.Print("Paste the code here and press enter: ")
	lines := make(chan string, 1)
	go func() {
		line, _ := bufio.NewReader(os.Stdin).ReadString('\n')
		lines <- line
	}()
	var input string
	select {
	case input = <-lines:
	case <-ctx.Done():
		return ctx.Err()
	}

	token, err := flow.Exchange(ctx, input)
	if err != nil {
		return err
	}
	if err := ws.SetProviderAPIKey(config.ScopeGlobal, "anthropic", token); err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("You're now authenticated with your Claude subscription!")
	return nil
}

func loginOpenAI(ws workspace.Workspace, force bool) error {
	if !force {
		cfg := ws.Config()
		if cfg != nil {
			if pc, ok := cfg.Providers.Get("openai"); ok && pc.OAuthToken != nil {
				fmt.Println("You are already logged in to OpenAI with a ChatGPT account.")
				fmt.Println("Use --force to re-authenticate.")
				return nil
			}
		}
	}

	ctx := getLoginContext()
	token, err := login.Run(ctx, login.PlatformOpenAI)
	if err != nil {
		return err
	}

	if err := ws.SetProviderAPIKey(config.ScopeGlobal, "openai", token); err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("You're now authenticated with your ChatGPT account!")
	return nil
}

func loginGrok(ws workspace.Workspace, force bool) error {
	if !force {
		cfg := ws.Config()
		if cfg != nil {
			if pc, ok := cfg.Providers.Get("xai"); ok && pc.OAuthToken != nil {
				fmt.Println("You are already logged in to xAI with a Grok account.")
				fmt.Println("Use --force to re-authenticate.")
				return nil
			}
		}
	}

	ctx := getLoginContext()
	token, err := login.Run(ctx, login.PlatformGrok)
	if err != nil {
		return err
	}

	if err := ws.SetProviderAPIKey(config.ScopeGlobal, "xai", token); err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("You're now authenticated with your Grok account!")
	return nil
}

func getLoginContext() context.Context {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	go func() {
		<-ctx.Done()
		cancel()
	}()
	return ctx
}

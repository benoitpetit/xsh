package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/benoitpetit/xsh/core"
	"github.com/benoitpetit/xsh/display"
)



// configCmd represents the config command
// By default shows current configuration (like Python version)
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Show current configuration",
	Long:  "Show current configuration. Use subcommands for advanced management.",
	Run: func(cmd *cobra.Command, args []string) {
		// Default behavior: show config (like Python)
		cfg, err := core.LoadConfig()
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Failed to load config: %v", err)))
			os.Exit(core.ExitError)
			return
		}

		output(cfg, func() {
			cv := &display.ConfigValues{
				DefaultCount:    cfg.DefaultCount,
				DefaultAccount:  cfg.DefaultAccount,
				Theme:           cfg.Display.Theme,
				ShowEngagement:  cfg.Display.ShowEngagement,
				ShowTimestamps:  cfg.Display.ShowTimestamps,
				MaxWidth:        cfg.Display.MaxWidth,
				Delay:           cfg.Request.Delay,
				Proxy:           cfg.Network.Proxy,
				Timeout:         cfg.Request.Timeout,
				MaxRetries:      cfg.Request.MaxRetries,
				LikesWeight:     cfg.Filter.LikesWeight,
				RetweetsWeight:  cfg.Filter.RetweetsWeight,
				RepliesWeight:   cfg.Filter.RepliesWeight,
				BookmarksWeight: cfg.Filter.BookmarksWeight,
				ViewsLogWeight:  cfg.Filter.ViewsLogWeight,
			}
			fmt.Println(display.FormatConfig(cv))
		})
	},
}

// configShowCmd shows current configuration
var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := core.LoadConfig()
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Failed to load config: %v", err)))
			return
		}

		output(cfg, func() {
			cv := &display.ConfigValues{
				DefaultCount:    cfg.DefaultCount,
				DefaultAccount:  cfg.DefaultAccount,
				Theme:           cfg.Display.Theme,
				ShowEngagement:  cfg.Display.ShowEngagement,
				ShowTimestamps:  cfg.Display.ShowTimestamps,
				MaxWidth:        cfg.Display.MaxWidth,
				Delay:           cfg.Request.Delay,
				Proxy:           cfg.Network.Proxy,
				Timeout:         cfg.Request.Timeout,
				MaxRetries:      cfg.Request.MaxRetries,
				LikesWeight:     cfg.Filter.LikesWeight,
				RetweetsWeight:  cfg.Filter.RetweetsWeight,
				RepliesWeight:   cfg.Filter.RepliesWeight,
				BookmarksWeight: cfg.Filter.BookmarksWeight,
				ViewsLogWeight:  cfg.Filter.ViewsLogWeight,
			}
			fmt.Println(display.FormatConfig(cv))
		})
	},
}

// configGetCmd gets a specific config value
var configGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get a specific configuration value",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := core.LoadConfig()
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Failed to load config: %v", err)))
			return
		}

		key := strings.ToLower(args[0])
		value := getConfigValue(cfg, key)

		if value == "" {
			fmt.Println(display.Error(fmt.Sprintf("Unknown config key: %s", key)))
			fmt.Println(display.Section("\nAvailable keys:"))
			printAvailableKeys()
			os.Exit(1)
			return
		}

		output(map[string]string{key: value}, func() {
			fmt.Println(display.Code(value))
		})
	},
}

// configSetCmd sets a config value
var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Long: `Set a configuration value.

Available keys:
  default_count           - Default number of tweets to fetch
  default_account         - Default account to use
  display.theme           - Display theme (default, dark, light)
  display.show_engagement - Show engagement stats (true/false)
  display.show_timestamps - Show timestamps (true/false)
  display.max_width       - Maximum display width
  request.delay           - Request delay in seconds
  request.timeout         - Request timeout in seconds
  request.max_retries     - Maximum retry attempts
  network.proxy           - Proxy URL
  filter.likes_weight     - Likes weight for scoring
  filter.retweets_weight  - Retweets weight for scoring
  filter.replies_weight   - Replies weight for scoring
  filter.bookmarks_weight - Bookmarks weight for scoring
  filter.views_log_weight - Views log weight for scoring`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := core.LoadConfig()
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Failed to load config: %v", err)))
			return
		}

		key := strings.ToLower(args[0])
		value := args[1]

		if err := setConfigValue(cfg, key, value); err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Failed to set config: %v", err)))
			os.Exit(1)
			return
		}

		if err := cfg.Save(); err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Failed to save config: %v", err)))
			os.Exit(1)
			return
		}

		output(map[string]string{"status": "ok", "key": key, "value": value}, func() {
			fmt.Println(display.Success(fmt.Sprintf("Set %s = %s", key, value)))
		})
	},
}

// configResetCmd resets config to defaults
var configResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset configuration to defaults",
	Run: func(cmd *cobra.Command, args []string) {
		force, _ := cmd.Flags().GetBool("force")

		if !force && !isJSONMode() {
			fmt.Print("Reset all configuration to defaults? [y/N] ")
			var confirm string
			fmt.Scanln(&confirm)
			if confirm != "y" && confirm != "Y" {
				fmt.Println(display.Warning("Aborted."))
				return
			}
		}

		cfg := core.DefaultConfig()
		if err := cfg.Save(); err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Failed to save config: %v", err)))
			os.Exit(1)
			return
		}

		output(map[string]string{"status": "reset"}, func() {
			fmt.Println(display.Success("Configuration reset to defaults"))
		})
	},
}

// configPathCmd shows config file path
var configPathCmd = &cobra.Command{
	Use:   "path",
	Short: "Show configuration file path",
	Run: func(cmd *cobra.Command, args []string) {
		path, err := core.GetConfigPath()
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Failed to get config path: %v", err)))
			return
		}

		output(map[string]string{"path": path}, func() {
			fmt.Println(display.Code(path))
		})
	},
}

// configEditCmd opens config in editor
var configEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Open configuration file in default editor",
	Run: func(cmd *cobra.Command, args []string) {
		path, err := core.GetConfigPath()
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Failed to get config path: %v", err)))
			return
		}

		// Ensure config exists
		if _, err := os.Stat(path); os.IsNotExist(err) {
			cfg := core.DefaultConfig()
			if err := cfg.Save(); err != nil {
				fmt.Println(display.Error(fmt.Sprintf("Failed to create config: %v", err)))
				return
			}
		}

		// Get editor from environment or use defaults
		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = os.Getenv("VISUAL")
		}
		if editor == "" {
			// Try common editors
			for _, e := range []string{"nano", "vim", "vi", "emacs", "code"} {
				if _, err := exec.LookPath(e); err == nil {
					editor = e
					break
				}
			}
		}
		if editor == "" {
			fmt.Println(display.Error("No editor found. Set EDITOR environment variable."))
			fmt.Println(display.Info(fmt.Sprintf("\nConfig file location: %s", path)))
			return
		}

		// Open editor
		cmd2 := exec.Command(editor, path)
		cmd2.Stdin = os.Stdin
		cmd2.Stdout = os.Stdout
		cmd2.Stderr = os.Stderr
		if err := cmd2.Run(); err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Failed to open editor: %v", err)))
			return
		}
	},
}

// Helper functions

func getConfigValue(cfg *core.Config, key string) string {
	switch key {
	case "default_count":
		return strconv.Itoa(cfg.DefaultCount)
	case "default_account":
		return cfg.DefaultAccount
	case "display.theme":
		return cfg.Display.Theme
	case "display.show_engagement":
		return strconv.FormatBool(cfg.Display.ShowEngagement)
	case "display.show_timestamps":
		return strconv.FormatBool(cfg.Display.ShowTimestamps)
	case "display.max_width":
		return strconv.Itoa(cfg.Display.MaxWidth)
	case "request.delay":
		return fmt.Sprintf("%.2f", cfg.Request.Delay)
	case "network.proxy":
		return cfg.Network.Proxy
	case "request.timeout":
		return strconv.Itoa(cfg.Request.Timeout)
	case "request.max_retries":
		return strconv.Itoa(cfg.Request.MaxRetries)
	case "filter.likes_weight":
		return fmt.Sprintf("%.1f", cfg.Filter.LikesWeight)
	case "filter.retweets_weight":
		return fmt.Sprintf("%.1f", cfg.Filter.RetweetsWeight)
	case "filter.replies_weight":
		return fmt.Sprintf("%.1f", cfg.Filter.RepliesWeight)
	case "filter.bookmarks_weight":
		return fmt.Sprintf("%.1f", cfg.Filter.BookmarksWeight)
	case "filter.views_log_weight":
		return fmt.Sprintf("%.1f", cfg.Filter.ViewsLogWeight)
	default:
		return ""
	}
}

func setConfigValue(cfg *core.Config, key, value string) error {
	switch key {
	case "default_count":
		v, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid integer value")
		}
		cfg.DefaultCount = v
	case "default_account":
		cfg.DefaultAccount = value
	case "display.theme":
		cfg.Display.Theme = value
	case "display.show_engagement":
		v, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid boolean value (use true/false)")
		}
		cfg.Display.ShowEngagement = v
	case "display.show_timestamps":
		v, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid boolean value (use true/false)")
		}
		cfg.Display.ShowTimestamps = v
	case "display.max_width":
		v, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid integer value")
		}
		cfg.Display.MaxWidth = v
	case "request.delay":
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("invalid float value")
		}
		cfg.Request.Delay = v
	case "network.proxy":
		cfg.Network.Proxy = value
	case "request.proxy":
		// Backwards compatibility
		cfg.Network.Proxy = value
	case "request.timeout":
		v, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid integer value")
		}
		cfg.Request.Timeout = v
	case "request.max_retries":
		v, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid integer value")
		}
		cfg.Request.MaxRetries = v
	case "filter.likes_weight":
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("invalid float value")
		}
		cfg.Filter.LikesWeight = v
	case "filter.retweets_weight":
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("invalid float value")
		}
		cfg.Filter.RetweetsWeight = v
	case "filter.replies_weight":
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("invalid float value")
		}
		cfg.Filter.RepliesWeight = v
	case "filter.bookmarks_weight":
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("invalid float value")
		}
		cfg.Filter.BookmarksWeight = v
	case "filter.views_log_weight":
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("invalid float value")
		}
		cfg.Filter.ViewsLogWeight = v
	default:
		return fmt.Errorf("unknown config key: %s", key)
	}
	return nil
}

func printAvailableKeys() {
	keys := []string{
		"default_count",
		"default_account",
		"display.theme",
		"display.show_engagement",
		"display.show_timestamps",
		"display.max_width",
		"request.delay",
		"request.timeout",
		"request.max_retries",
		"network.proxy",
		"filter.likes_weight",
		"filter.retweets_weight",
		"filter.replies_weight",
		"filter.bookmarks_weight",
		"filter.views_log_weight",
	}
	for _, k := range keys {
		fmt.Println(display.Bullet(k))
	}
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configResetCmd)
	configCmd.AddCommand(configPathCmd)
	configCmd.AddCommand(configEditCmd)

	configResetCmd.Flags().BoolP("force", "f", false, "Skip confirmation")
}

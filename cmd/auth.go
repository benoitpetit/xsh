package cmd

import (
	"fmt"
	"runtime"

	"github.com/benoitpetit/xsh/browser"
	"github.com/benoitpetit/xsh/core"
	"github.com/benoitpetit/xsh/display"
	"github.com/spf13/cobra"
)

var (
	authBrowser string
	authAccount string
	forceFlag   bool
)

// authCmd represents the auth command
var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage authentication",
	Long:  "Authenticate with Twitter/X using browser cookies or manual entry.",
}

// authStatusCmd checks authentication status
var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check authentication status",
	RunE: func(cmd *cobra.Command, args []string) error {
		creds, err := core.GetCredentials(account)
		if err != nil {
			output(map[string]bool{"authenticated": false}, func() {
				fmt.Println(display.Error("Not authenticated"))
			})
			abortCommand(core.ExitAuthError)
			return nil
		}

		info := map[string]interface{}{
			"authenticated": true,
			"auth_token":    creds.AuthToken[:8] + "...",
			"ct0":           creds.Ct0[:8] + "...",
			"account":       creds.AccountName,
		}

		output(info, func() {
			fmt.Println(display.Success(fmt.Sprintf("Authenticated (token: %s)", info["auth_token"])))
			if creds.AccountName != "" {
				fmt.Println(display.KeyValue("Account:", creds.AccountName))
			}
		})

		return nil
	},
}

// authLoginCmd extracts cookies from browser
var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Extract cookies from browser",
	Long: `Extract Twitter/X cookies automatically from your browser.

Supported browsers:
  - Chrome (including profiles)
  - Brave
  - Microsoft Edge
  - Chromium
  - Firefox

The command will try to extract cookies from all available browsers
and use the first valid credentials found.

Examples:
  # Auto-detect browser
  xsh auth login

  # Extract from specific browser
  xsh auth login --browser chrome
  xsh auth login --browser firefox
  xsh auth login --browser brave

  # Save to specific account
  xsh auth login --account work`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var creds *core.AuthCredentials
		var browserName string
		var err error

		if authBrowser != "" {
			// Extract from specific browser
			fmt.Println(display.Action("Extracting cookies from", authBrowser))
			creds, err = browser.ExtractFromBrowserVerbose(authBrowser, core.Verbose)
			browserName = authBrowser
		} else {
			// Auto-detect from all browsers
			fmt.Println(display.Action("Detecting browsers with", "Twitter/X cookies"))
			availableBrowsers := browser.ListAvailableBrowsers()

			if len(availableBrowsers) == 0 {
				fmt.Println(display.Error("No supported browser found"))
				fmt.Println(display.Section("Supported browsers"))
				fmt.Println(display.Bullet("Google Chrome / Chromium"))
				fmt.Println(display.Bullet("Brave"))
				fmt.Println(display.Bullet("Microsoft Edge"))
				fmt.Println(display.Bullet("Firefox"))
				fmt.Println(display.Bullet("Opera / Vivaldi / Safari (via kooky library)"))

				fmt.Println(display.Section("Troubleshooting"))
				switch runtime.GOOS {
				case "linux":
					fmt.Println(display.Bullet("Browsers are typically in ~/.config/<browser-name>/"))
					fmt.Println(display.Bullet("Make sure you have read permissions on the browser directories"))
				case "darwin":
					fmt.Println(display.Bullet("Browsers are in ~/Library/Application Support/"))
					fmt.Println(display.Bullet("Grant Full Disk Access to Terminal in System Preferences"))
				case "windows":
					fmt.Println(display.Bullet("Browsers are in %LOCALAPPDATA%"))
				}

				fmt.Println(display.Section("Recommended alternatives"))
				fmt.Println(display.Bullet("xsh auth import <cookies.json>  — Export from Cookie Editor extension"))
				fmt.Println(display.Bullet("xsh auth set                    — Manual token entry"))
				abortCommand(core.ExitAuthError)
				return nil
			}

			fmt.Println(display.Info(fmt.Sprintf("Found browsers: %v", availableBrowsers)))

			creds, browserName, err = browser.ExtractFromAllBrowsersVerbose(core.Verbose)
		}

		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Failed to extract cookies: %v", err)))

			// Provide OS-specific help
			fmt.Println(display.Section("Troubleshooting"))
			switch runtime.GOOS {
			case "darwin":
				fmt.Println(display.Numbered(1, "Make sure Chrome/Firefox is not running (locks the database)"))
				fmt.Println(display.Numbered(2, "Try granting Full Disk Access to Terminal in System Preferences > Security & Privacy"))
				fmt.Println(display.Numbered(3, "Chrome 80+ uses Keychain encryption which may require authentication"))
			case "windows":
				fmt.Println(display.Numbered(1, "Make sure Chrome/Firefox is not running"))
				fmt.Println(display.Numbered(2, "Try running as Administrator if access is denied"))
				fmt.Println(display.Numbered(3, "Windows Defender or antivirus may block cookie access"))
			case "linux":
				fmt.Println(display.Numbered(1, display.Warning("CLOSE CHROME COMPLETELY")+" (cookie database is locked when Chrome is running)"))
				fmt.Println(display.Bullet("Run: killall chrome"))
				fmt.Println(display.Numbered(2, "Check file permissions on browser config directories (~/.config/google-chrome)"))
				fmt.Println(display.Numbered(3, "Chrome 80+ uses system keyring (libsecret/gnome-keyring) for encryption"))
				fmt.Println(display.Numbered(4, "Install required packages:"))
				fmt.Println(display.Bullet("Fedora: sudo dnf install python3-secretstorage"))
				fmt.Println(display.Bullet("Ubuntu/Debian: sudo apt install python3-secretstorage"))
				fmt.Println(display.Bullet("Arch: sudo pacman -S python-secretstorage"))
			}

			fmt.Println(display.Section("Recommended alternative methods"))
			fmt.Println(display.Numbered(1, "xsh auth import <cookies.json>  — Export from Cookie Editor extension"))
			fmt.Println(display.Numbered(2, "xsh auth set                    — Enter tokens manually"))

			fmt.Println(display.Section("Cookie Editor method (most reliable)"))
			fmt.Println(display.Numbered(1, "Install 'Cookie Editor' extension in your browser"))
			fmt.Println(display.Numbered(2, "Go to x.com and log in"))
			fmt.Println(display.Numbered(3, "Open Cookie Editor, click 'Export' → 'JSON'"))
			fmt.Println(display.Numbered(4, "Save to a file and run: xsh auth import <file>"))

			abortCommand(core.ExitAuthError)
			return nil
		}

		if creds == nil || !creds.IsValid() {
			fmt.Println(display.Error("Extracted credentials are invalid"))
			fmt.Println(display.Info("Make sure you're logged into x.com in your browser"))
			abortCommand(core.ExitAuthError)
			return nil
		}

		acc := authAccount
		if acc == "" {
			acc = "default"
		}

		creds.AccountName = acc
		if err := core.SaveAuth(creds, acc); err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Failed to save credentials: %v", err)))
			abortCommand(core.ExitError)
			return nil
		}

		fmt.Println(display.Success(fmt.Sprintf("Authenticated using %s! Saved as account '%s'", browserName, acc)))
		fmt.Println(display.KeyValue("Token:", creds.AuthToken[:8]+"..."))

		return nil
	},
}

// authImportCmd imports cookies from file
var authImportCmd = &cobra.Command{
	Use:   "import [file]",
	Short: "Import cookies from Cookie Editor JSON export",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		creds, err := core.ImportCookiesFromFile(args[0])
		if err != nil || creds == nil || !creds.IsValid() {
			fmt.Println(display.Error("Could not find auth_token/ct0 in the file. Make sure you exported cookies from x.com with Cookie Editor"))
			abortCommand(core.ExitAuthError)
			return nil
		}

		acc := authAccount
		if acc == "" {
			acc = "default"
		}

		if err := core.SaveAuth(creds, acc); err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Failed to save credentials: %v", err)))
			abortCommand(core.ExitError)
			return nil
		}

		fmt.Println(display.Success(fmt.Sprintf("Imported cookies! Saved as account '%s'", acc)))

		return nil
	},
}

// authSetCmd manually sets credentials
var authSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Manually set authentication credentials",
	RunE: func(cmd *cobra.Command, args []string) error {
		var authToken, ct0 string

		fmt.Print("Enter auth_token: ")
		fmt.Scanln(&authToken)

		fmt.Print("Enter ct0: ")
		fmt.Scanln(&ct0)

		if authToken == "" || ct0 == "" {
			fmt.Println(display.Error("Both auth_token and ct0 are required"))
			abortCommand(core.ExitAuthError)
			return nil
		}

		creds := &core.AuthCredentials{
			AuthToken:   authToken,
			Ct0:         ct0,
			AccountName: authAccount,
		}

		acc := authAccount
		if acc == "" {
			acc = "default"
		}

		if err := core.SaveAuth(creds, acc); err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Failed to save credentials: %v", err)))
			abortCommand(core.ExitError)
			return nil
		}

		fmt.Println(display.Success(fmt.Sprintf("Credentials saved as account '%s'", acc)))

		return nil
	},
}

// authAccountsCmd lists stored accounts
var authAccountsCmd = &cobra.Command{
	Use:   "accounts",
	Short: "List stored accounts",
	RunE: func(cmd *cobra.Command, args []string) error {
		accounts, err := core.ListAccounts()
		if err != nil {
			accounts = []string{}
		}

		output(map[string]interface{}{"accounts": accounts}, func() {
			if len(accounts) == 0 {
				fmt.Println(display.Warning("No accounts stored"))
			} else {
				for _, acc := range accounts {
					fmt.Println(display.Bullet(acc))
				}
			}
		})

		return nil
	},
}

// authSwitchCmd switches default account
var authSwitchCmd = &cobra.Command{
	Use:   "switch [account]",
	Short: "Switch default account",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := core.SetDefaultAccount(args[0]); err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Account '%s' not found", args[0])))
			abortCommand(core.ExitError)
			return nil
		}
		fmt.Println(display.Success(fmt.Sprintf("Switched to account '%s'", args[0])))

		return nil
	},
}

// authLogoutCmd removes stored credentials
var authLogoutCmd = &cobra.Command{
	Use:   "logout [account]",
	Short: "Remove stored credentials",
	RunE: func(cmd *cobra.Command, args []string) error {
		acc := "default"
		if len(args) > 0 {
			acc = args[0]
		}

		if !forceFlag {
			fmt.Printf("Remove account '%s'? [y/N] ", acc)
			var confirm string
			fmt.Scanln(&confirm)
			if confirm != "y" && confirm != "Y" {
				fmt.Println(display.Warning("Aborted."))
				return nil
			}
		}

		if err := core.RemoveAuth(acc); err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Failed to remove account: %v", err)))
			abortCommand(core.ExitError)
			return nil
		}

		fmt.Println(display.Success(fmt.Sprintf("Removed account '%s'", acc)))

		return nil
	},
}

// authWhoamiCmd shows current user info
var authWhoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Show current user information",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient("")
		if err != nil {
			fmt.Println(display.Error(err.Error()))
			abortCommand(core.ExitAuthError)
			return nil
		}
		defer client.Close()

		// Get current user by verifying credentials
		creds, _ := core.GetCredentials(account)
		if creds == nil {
			fmt.Println(display.Error("Not authenticated"))
			abortCommand(core.ExitAuthError)
			return nil
		}

		// Try to get user info from API
		// We'll use the home timeline to verify
		response, err := core.GetHomeTimeline(client, "for-you", 1, "")
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Failed to verify credentials: %v", err)))
			abortCommand(core.ExitAuthError)
			return nil
		}

		output(map[string]interface{}{
			"authenticated": true,
			"account":       creds.AccountName,
			"auth_token":    creds.AuthToken[:8] + "...",
		}, func() {
			fmt.Println(display.Success("Authenticated"))
			fmt.Println(display.KeyValue("Account:", creds.AccountName))
			fmt.Println(display.KeyValue("Token:", creds.AuthToken[:8]+"..."))
			if len(response.Tweets) > 0 {
				fmt.Println(display.KeyValue("API Status:", "OK (timeline accessible)"))
			}
		})

		return nil
	},
}

func init() {
	rootCmd.AddCommand(authCmd)
	authCmd.AddCommand(authStatusCmd)
	authCmd.AddCommand(authLoginCmd)
	authCmd.AddCommand(authImportCmd)
	authCmd.AddCommand(authSetCmd)
	authCmd.AddCommand(authAccountsCmd)
	authCmd.AddCommand(authSwitchCmd)
	authCmd.AddCommand(authLogoutCmd)
	authCmd.AddCommand(authWhoamiCmd)

	// Flags
	authLoginCmd.Flags().StringVar(&authBrowser, "browser", "", "Browser to extract from (chrome, firefox, brave, edge, chromium)")
	authLoginCmd.Flags().StringVar(&authAccount, "account", "default", "Account name")
	authImportCmd.Flags().StringVar(&authAccount, "account", "default", "Account name")
	authSetCmd.Flags().StringVar(&authAccount, "account", "default", "Account name")
	authLogoutCmd.Flags().BoolVarP(&forceFlag, "force", "f", false, "Skip confirmation")
}

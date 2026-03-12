package cmd

import (
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/benoitpetit/xsh/browser"
	"github.com/benoitpetit/xsh/core"
	"github.com/benoitpetit/xsh/display"
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
	Run: func(cmd *cobra.Command, args []string) {
		creds, err := core.GetCredentials(account)
		if err != nil {
			if isJSONMode() {
				outputJSON(map[string]bool{"authenticated": false})
			} else {
				fmt.Println(display.PrintError("Not authenticated"))
			}
			os.Exit(core.ExitAuthError)
			return
		}

		info := map[string]interface{}{
			"authenticated": true,
			"auth_token":    creds.AuthToken[:8] + "...",
			"ct0":           creds.Ct0[:8] + "...",
			"account":       creds.AccountName,
		}

		if isJSONMode() {
			outputJSON(info)
		} else {
			fmt.Println(display.PrintSuccess(fmt.Sprintf("Authenticated (token: %s)", info["auth_token"])))
			if creds.AccountName != "" {
				fmt.Printf("  Account: %s\n", creds.AccountName)
			}
		}
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
	Run: func(cmd *cobra.Command, args []string) {
		var creds *core.AuthCredentials
		var browserName string
		var err error

		if authBrowser != "" {
			// Extract from specific browser
			fmt.Printf("Extracting cookies from %s...\n", authBrowser)
			creds, err = browser.ExtractFromBrowserVerbose(authBrowser, core.Verbose)
			browserName = authBrowser
		} else {
			// Auto-detect from all browsers
			fmt.Println("Detecting browsers with Twitter/X cookies...")
			availableBrowsers := browser.ListAvailableBrowsers()
			
			if len(availableBrowsers) == 0 {
				fmt.Println(display.PrintError("No supported browser found."))
				fmt.Println("\nSupported browsers:")
				fmt.Println("  - Google Chrome / Chromium")
				fmt.Println("  - Brave")
				fmt.Println("  - Microsoft Edge")
				fmt.Println("  - Firefox")
				fmt.Println("  - Opera / Vivaldi / Safari (via kooky library)")
				
				fmt.Println("\nTroubleshooting:")
				switch runtime.GOOS {
				case "linux":
					fmt.Println("  - Browsers are typically in ~/.config/<browser-name>/")
					fmt.Println("  - Make sure you have read permissions on the browser directories")
				case "darwin":
					fmt.Println("  - Browsers are in ~/Library/Application Support/")
					fmt.Println("  - Grant Full Disk Access to Terminal in System Preferences")
				case "windows":
					fmt.Println("  - Browsers are in %LOCALAPPDATA%")
				}
				
				fmt.Println("\nRecommended alternative:")
				fmt.Println("  xsh auth import <cookies.json>  # Export from Cookie Editor extension")
				fmt.Println("  xsh auth set                    # Manual token entry")
				os.Exit(core.ExitAuthError)
				return
			}
			
			fmt.Printf("Found browsers: %v\n", availableBrowsers)
			
			creds, browserName, err = browser.ExtractFromAllBrowsersVerbose(core.Verbose)
		}

		if err != nil {
			fmt.Println(display.PrintError(fmt.Sprintf("Failed to extract cookies: %v", err)))
			
			// Provide OS-specific help
			fmt.Println("\n" + display.StyleBold.Render("Troubleshooting:"))
			switch runtime.GOOS {
			case "darwin":
				fmt.Println("  1. Make sure Chrome/Firefox is not running (locks the database)")
				fmt.Println("  2. Try granting Full Disk Access to Terminal in System Preferences > Security & Privacy")
				fmt.Println("  3. Chrome 80+ uses Keychain encryption which may require authentication")
			case "windows":
				fmt.Println("  1. Make sure Chrome/Firefox is not running")
				fmt.Println("  2. Try running as Administrator if access is denied")
				fmt.Println("  3. Windows Defender or antivirus may block cookie access")
			case "linux":
				fmt.Println("  1. ⚠️  CLOSE CHROME COMPLETELY (cookie database is locked when Chrome is running)")
				fmt.Println("     Run: killall chrome")
				fmt.Println("  2. Check file permissions on browser config directories (~/.config/google-chrome)")
				fmt.Println("  3. Chrome 80+ uses system keyring (libsecret/gnome-keyring) for encryption")
				fmt.Println("  4. Install required packages:")
				fmt.Println("     - Fedora: sudo dnf install python3-secretstorage")
				fmt.Println("     - Ubuntu/Debian: sudo apt install python3-secretstorage")
				fmt.Println("     - Arch: sudo pacman -S python-secretstorage")
			}
			
			fmt.Println("\n" + display.StyleBold.Render("Recommended alternative methods:"))
			fmt.Println("  1. xsh auth import <cookies.json>  # Export from Cookie Editor extension")
			fmt.Println("  2. xsh auth set                    # Enter tokens manually")
			
			fmt.Println("\n" + display.StyleBold.Render("Cookie Editor method (most reliable):"))
			fmt.Println("  1. Install 'Cookie Editor' extension in your browser")
			fmt.Println("  2. Go to x.com and log in")
			fmt.Println("  3. Open Cookie Editor, click 'Export' → 'JSON'")
			fmt.Println("  4. Save to a file and run: xsh auth import <file>")
			
			os.Exit(core.ExitAuthError)
			return
		}

		if creds == nil || !creds.IsValid() {
			fmt.Println(display.PrintError("Extracted credentials are invalid."))
			fmt.Println("Make sure you're logged into x.com in your browser.")
			os.Exit(core.ExitAuthError)
			return
		}

		acc := authAccount
		if acc == "" {
			acc = "default"
		}

		creds.AccountName = acc
		if err := core.SaveAuth(creds, acc); err != nil {
			fmt.Println(display.PrintError(fmt.Sprintf("Failed to save credentials: %v", err)))
			os.Exit(core.ExitError)
			return
		}

		fmt.Println(display.PrintSuccess(fmt.Sprintf("Authenticated using %s! Saved as account '%s'", browserName, acc)))
		fmt.Printf("  Token: %s...\n", creds.AuthToken[:8])
	},
}

// authImportCmd imports cookies from file
var authImportCmd = &cobra.Command{
	Use:   "import [file]",
	Short: "Import cookies from Cookie Editor JSON export",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		creds, err := core.ImportCookiesFromFile(args[0])
		if err != nil || creds == nil || !creds.IsValid() {
			fmt.Println(display.PrintError("Could not find auth_token/ct0 in the file. Make sure you exported cookies from x.com with Cookie Editor."))
			os.Exit(core.ExitAuthError)
			return
		}

		acc := authAccount
		if acc == "" {
			acc = "default"
		}

		if err := core.SaveAuth(creds, acc); err != nil {
			fmt.Println(display.PrintError(fmt.Sprintf("Failed to save credentials: %v", err)))
			os.Exit(core.ExitError)
			return
		}

		fmt.Println(display.PrintSuccess(fmt.Sprintf("Imported cookies! Saved as account '%s'", acc)))
	},
}

// authSetCmd manually sets credentials
var authSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Manually set authentication credentials",
	Run: func(cmd *cobra.Command, args []string) {
		var authToken, ct0 string

		fmt.Print("Enter auth_token: ")
		fmt.Scanln(&authToken)
		
		fmt.Print("Enter ct0: ")
		fmt.Scanln(&ct0)

		if authToken == "" || ct0 == "" {
			fmt.Println(display.PrintError("Both auth_token and ct0 are required"))
			os.Exit(core.ExitAuthError)
			return
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
			fmt.Println(display.PrintError(fmt.Sprintf("Failed to save credentials: %v", err)))
			os.Exit(core.ExitError)
			return
		}

		fmt.Println(display.PrintSuccess(fmt.Sprintf("Credentials saved as account '%s'", acc)))
	},
}

// authAccountsCmd lists stored accounts
var authAccountsCmd = &cobra.Command{
	Use:   "accounts",
	Short: "List stored accounts",
	Run: func(cmd *cobra.Command, args []string) {
		accounts, err := core.ListAccounts()
		if err != nil {
			accounts = []string{}
		}

		if isJSONMode() {
			outputJSON(map[string]interface{}{"accounts": accounts})
		} else {
			if len(accounts) == 0 {
				fmt.Println(display.PrintWarning("No accounts stored"))
			} else {
				for _, acc := range accounts {
					fmt.Printf("  %s\n", acc)
				}
			}
		}
	},
}

// authSwitchCmd switches default account
var authSwitchCmd = &cobra.Command{
	Use:   "switch [account]",
	Short: "Switch default account",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := core.SetDefaultAccount(args[0]); err != nil {
			fmt.Println(display.PrintError(fmt.Sprintf("Account '%s' not found", args[0])))
			os.Exit(core.ExitError)
			return
		}
		fmt.Println(display.PrintSuccess(fmt.Sprintf("Switched to account '%s'", args[0])))
	},
}

// authLogoutCmd removes stored credentials
var authLogoutCmd = &cobra.Command{
	Use:   "logout [account]",
	Short: "Remove stored credentials",
	Run: func(cmd *cobra.Command, args []string) {
		acc := "default"
		if len(args) > 0 {
			acc = args[0]
		}

		if !forceFlag {
			fmt.Printf("Remove account '%s'? [y/N] ", acc)
			var confirm string
			fmt.Scanln(&confirm)
			if confirm != "y" && confirm != "Y" {
				fmt.Println("Aborted.")
				return
			}
		}

		if err := core.RemoveAuth(acc); err != nil {
			fmt.Println(display.PrintError(fmt.Sprintf("Failed to remove account: %v", err)))
			os.Exit(core.ExitError)
			return
		}

		fmt.Println(display.PrintSuccess(fmt.Sprintf("Removed account '%s'", acc)))
	},
}

// authWhoamiCmd shows current user info
var authWhoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Show current user information",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := getClient("")
		if err != nil {
			fmt.Println(display.PrintError(err.Error()))
			os.Exit(core.ExitAuthError)
			return
		}
		defer client.Close()

		// Get current user by verifying credentials
		creds, _ := core.GetCredentials(account)
		if creds == nil {
			fmt.Println(display.PrintError("Not authenticated"))
			os.Exit(core.ExitAuthError)
			return
		}

		// Try to get user info from API
		// We'll use the home timeline to verify
		response, err := core.GetHomeTimeline(client, "for-you", 1, "")
		if err != nil {
			fmt.Println(display.PrintError(fmt.Sprintf("Failed to verify credentials: %v", err)))
			os.Exit(core.ExitAuthError)
			return
		}

		if isJSONMode() {
			outputJSON(map[string]interface{}{
				"authenticated": true,
				"account":       creds.AccountName,
				"auth_token":    creds.AuthToken[:8] + "...",
			})
		} else {
			fmt.Println(display.PrintSuccess("Authenticated"))
			fmt.Printf("  Account: %s\n", creds.AccountName)
			fmt.Printf("  Token: %s...\n", creds.AuthToken[:8])
			if len(response.Tweets) > 0 {
				fmt.Printf("  API Status: OK (timeline accessible)\n")
			}
		}
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

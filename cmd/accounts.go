package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/benoitpetit/xsh/core"
	"github.com/benoitpetit/xsh/display"

	"github.com/spf13/cobra"
)

var accountsCmd = &cobra.Command{
	Use:   "accounts",
	Short: "List stored accounts",
	Long:  `List all stored Twitter/X accounts in the authentication file.`,
	Run: func(cmd *cobra.Command, args []string) {
		accounts, err := core.ListAccounts()
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Failed to list accounts: %v", err)))
			os.Exit(core.ExitError)
		}

		if len(accounts) == 0 {
			fmt.Println(display.Warning("No stored accounts found"))
			fmt.Println(display.Muted("\nUse 'xsh auth' to add an account."))
			return
		}

		// Get default account
		defaultAccount := ""
		creds, err := core.GetCredentials("")
		if err == nil && creds != nil {
			defaultAccount = creds.AccountName
		}

		fmt.Println(display.Title("📂 Stored Accounts"))
		fmt.Println()

		for _, acc := range accounts {
			if acc == defaultAccount {
				fmt.Println(display.Bullet(
					display.Success("●") +
					display.Bold(acc+" (active)")))
			} else {
				fmt.Println(display.Bullet(
					display.Muted("○") +
					display.Bold(acc)))
			}
		}

		fmt.Println()
		fmt.Println(display.Muted("Use 'xsh switch <account>' to change accounts"))
	},
}

var switchCmd = &cobra.Command{
	Use:   "switch [account-name]",
	Short: "Switch default account",
	Long:  `Switch the default Twitter/X account for subsequent commands.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		acc := args[0]

		if err := core.SetDefaultAccount(acc); err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Failed to switch account: %v", err)))
			os.Exit(core.ExitAuthError)
		}

		fmt.Println(display.Success(fmt.Sprintf("Switched to account '%s'", acc)))
	},
}

var importCmd = &cobra.Command{
	Use:   "import [cookie-file] [account-name]",
	Short: "Import cookies from file",
	Long: `Import Twitter/X cookies from a JSON file exported by Cookie Editor browser extension.

The cookie file should be in JSON format exported from x.com with 'auth_token' and 'ct0' cookies.`,
	Args: cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		cookieFile := args[0]
		accountName := ""
		if len(args) > 1 {
			accountName = args[1]
		}

		// Resolve relative path
		if !filepath.IsAbs(cookieFile) {
			cwd, err := os.Getwd()
			if err != nil {
				fmt.Println(display.Error(fmt.Sprintf("Failed to get working directory: %v", err)))
				os.Exit(core.ExitError)
			}
			cookieFile = filepath.Join(cwd, cookieFile)
		}

		// Check file exists
		if _, err := os.Stat(cookieFile); os.IsNotExist(err) {
			fmt.Println(display.Error(fmt.Sprintf("Cookie file not found: %s", cookieFile)))
			os.Exit(core.ExitError)
		}

		// Import cookies
		creds, err := core.ImportCookiesFromFile(cookieFile)
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Failed to import cookies: %v", err)))
			os.Exit(core.ExitAuthError)
		}

		// Generate account name if not provided
		if accountName == "" {
			// Try to verify credentials and get username
			client, err := core.NewXClient(creds, "", "")
			if err == nil {
				// Try to get user info by searching for a known user
				// Since we don't have a direct "get me" endpoint, use the imported filename
				base := filepath.Base(cookieFile)
				ext := filepath.Ext(base)
				accountName = base[:len(base)-len(ext)]
			} else {
				// Use filename as fallback
				base := filepath.Base(cookieFile)
				ext := filepath.Ext(base)
				accountName = base[:len(base)-len(ext)]
			}
			_ = client // Suppress unused warning
		}

		creds.AccountName = accountName

		// Save auth
		if err := core.SaveAuth(creds, accountName); err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Failed to save auth: %v", err)))
			os.Exit(core.ExitAuthError)
		}

		fmt.Println(display.Success(fmt.Sprintf("Successfully imported account '%s'", accountName)))
	},
}

func init() {
	rootCmd.AddCommand(accountsCmd)
	rootCmd.AddCommand(switchCmd)
	rootCmd.AddCommand(importCmd)
}

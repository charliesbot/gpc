package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/charliesbot/gpc/internal/auth"
	"github.com/charliesbot/gpc/internal/client"
	"github.com/charliesbot/gpc/internal/config"
	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage authentication credentials",
	Long:  "Configure, inspect, and verify Google Play Developer API credentials.",
}

var authInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Set up credentials for the first time",
	Long:  "Guided setup that stores a service account JSON key in the system keyring.",
	RunE:  runAuthInit,
}

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current authentication state",
	Long:  "Display which credential source is active and show the current configuration.",
	RunE:  runAuthStatus,
}

var authTestCmd = &cobra.Command{
	Use:   "test",
	Short: "Verify that credentials are valid",
	Long:  "Resolve credentials and attempt to create a Google Play Developer API client.",
	RunE:  runAuthTest,
}

func init() {
	authCmd.AddCommand(authInitCmd)
	authCmd.AddCommand(authStatusCmd)
	authCmd.AddCommand(authTestCmd)
	rootCmd.AddCommand(authCmd)
}

func runAuthInit(cmd *cobra.Command, args []string) error {
	keyPath := flagKeyFile

	if keyPath == "" {
		fmt.Print("Enter path to service account JSON key file: ")
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			keyPath = strings.TrimSpace(scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			return fmt.Errorf("reading input: %w", err)
		}
	}

	if keyPath == "" {
		return fmt.Errorf("key file path is required")
	}

	data, err := os.ReadFile(keyPath)
	if err != nil {
		return fmt.Errorf("reading key file %q: %w", keyPath, err)
	}

	if err := auth.StoreCredential(data); err != nil {
		return fmt.Errorf("storing credential in keyring: %w", err)
	}
	fmt.Println("Credential stored in system keyring.")

	fmt.Print("Save key file path to config for future use? [y/N]: ")
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		answer := strings.TrimSpace(strings.ToLower(scanner.Text()))
		if answer == "y" || answer == "yes" {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("loading config: %w", err)
			}
			cfg.DefaultKeyFile = keyPath
			if err := cfg.Save(); err != nil {
				return fmt.Errorf("saving config: %w", err)
			}
			fmt.Printf("Key file path saved to %s\n", config.DefaultConfigPath())
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("reading input: %w", err)
	}

	fmt.Println("Authentication setup complete.")
	return nil
}

func runAuthStatus(cmd *cobra.Command, args []string) error {
	fmt.Println("Credential sources:")

	// 1. --key-file flag
	if flagKeyFile != "" {
		fmt.Printf("  [active] --key-file flag: %s\n", flagKeyFile)
	} else {
		fmt.Println("  [      ] --key-file flag: not set")
	}

	// 2. GPC_KEY_FILE env
	if v := os.Getenv("GPC_KEY_FILE"); v != "" {
		if flagKeyFile == "" {
			fmt.Printf("  [active] GPC_KEY_FILE env: %s\n", v)
		} else {
			fmt.Printf("  [  set  ] GPC_KEY_FILE env: %s\n", v)
		}
	} else {
		fmt.Println("  [      ] GPC_KEY_FILE env: not set")
	}

	// 3. GPC_KEY_JSON env
	if v := os.Getenv("GPC_KEY_JSON"); v != "" {
		if flagKeyFile == "" && os.Getenv("GPC_KEY_FILE") == "" {
			fmt.Println("  [active] GPC_KEY_JSON env: set")
		} else {
			fmt.Println("  [  set  ] GPC_KEY_JSON env: set")
		}
	} else {
		fmt.Println("  [      ] GPC_KEY_JSON env: not set")
	}

	// 4. Keyring
	_, err := auth.LoadCredential()
	if err == nil {
		noHigherSource := flagKeyFile == "" &&
			os.Getenv("GPC_KEY_FILE") == "" &&
			os.Getenv("GPC_KEY_JSON") == ""
		if noHigherSource {
			fmt.Println("  [active] System keyring: credential found")
		} else {
			fmt.Println("  [  set  ] System keyring: credential found")
		}
	} else {
		fmt.Println("  [      ] System keyring: no credential stored")
	}

	// 5. ADC
	fmt.Println("  [      ] Application Default Credentials: checked as fallback")

	// Config file
	fmt.Println()
	configPath := config.DefaultConfigPath()
	fmt.Printf("Config file: %s\n", configPath)

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "  Error loading config: %v\n", err)
	} else {
		if cfg.DefaultPackage != "" {
			fmt.Printf("  default_package:  %s\n", cfg.DefaultPackage)
		}
		if cfg.DefaultKeyFile != "" {
			fmt.Printf("  default_key_file: %s\n", cfg.DefaultKeyFile)
		}
		if cfg.DefaultOutput != "" {
			fmt.Printf("  default_output:   %s\n", cfg.DefaultOutput)
		}
		if cfg.DefaultPackage == "" && cfg.DefaultKeyFile == "" && cfg.DefaultOutput == "" {
			fmt.Println("  (empty)")
		}
	}

	return nil
}

func runAuthTest(cmd *cobra.Command, args []string) error {
	fmt.Println("Resolving credentials...")

	creds, err := auth.ResolveCredentials(auth.CredentialOptions{
		KeyFile:    flagKeyFile,
		UseKeyring: true,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Authentication failed: %v\n", err)
		return fmt.Errorf("credential resolution failed")
	}
	fmt.Println("Credentials resolved successfully.")

	fmt.Println("Creating API client...")
	ctx := context.Background()
	_, err = client.NewService(ctx, creds)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Client creation failed: %v\n", err)
		return fmt.Errorf("client creation failed")
	}

	fmt.Println("Authentication test passed. Credentials are valid.")
	return nil
}

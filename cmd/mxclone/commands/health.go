// Package commands contains the CLI commands for the MXToolbox clone.
package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"mxclone/pkg/dns"
	"mxclone/pkg/dnsbl"
	"mxclone/pkg/emailauth"
	"mxclone/pkg/smtp"
	"mxclone/pkg/types"
)

// HealthCmd represents the health command
var HealthCmd = &cobra.Command{
	Use:   "health [domain]",
	Short: "Perform comprehensive domain health check",
	Long: `Perform a comprehensive health check for a domain.
This combines DNS, blacklist, SMTP, and email authentication checks
to provide an overall assessment of the domain's health.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		domain := args[0]
		timeout, _ := cmd.Flags().GetInt("timeout")
		checkDNS, _ := cmd.Flags().GetBool("check-dns")
		checkBlacklist, _ := cmd.Flags().GetBool("check-blacklist")
		checkSMTP, _ := cmd.Flags().GetBool("check-smtp")
		checkAuth, _ := cmd.Flags().GetBool("check-auth")
		outputFormat, _ := cmd.Flags().GetString("output")

		fmt.Printf("Performing comprehensive health check for %s...\n", domain)

		ctx := context.Background()
		timeoutDuration := time.Duration(timeout) * time.Second

		report := &types.DomainHealthReport{
			Domain:    domain,
			Timestamp: time.Now(),
		}

		type result struct {
			name string
			val  interface{}
			err  error
		}

		resultsCh := make(chan result, 4)
		checks := 0

		if checkDNS {
			checks++
			go func() {
				fmt.Println("Performing DNS checks...")
				dnsResult, err := performDNSCheck(ctx, domain, timeoutDuration)
				resultsCh <- result{name: "dns", val: dnsResult, err: err}
			}()
		}
		if checkBlacklist {
			checks++
			go func() {
				fmt.Println("Performing blacklist checks...")
				blacklistResult, err := performBlacklistCheck(ctx, domain, timeoutDuration)
				resultsCh <- result{name: "blacklist", val: blacklistResult, err: err}
			}()
		}
		if checkSMTP {
			checks++
			go func() {
				fmt.Println("Performing SMTP checks...")
				smtpResult, err := performSMTPCheck(ctx, domain, timeoutDuration)
				resultsCh <- result{name: "smtp", val: smtpResult, err: err}
			}()
		}
		if checkAuth {
			checks++
			go func() {
				fmt.Println("Performing email authentication checks...")
				authResult, err := performAuthCheck(ctx, domain, timeoutDuration)
				resultsCh <- result{name: "auth", val: authResult, err: err}
			}()
		}

		for i := 0; i < checks; i++ {
			res := <-resultsCh
			switch res.name {
			case "dns":
				if res.err != nil {
					fmt.Fprintf(os.Stderr, "Error performing DNS check: %v\n", res.err)
				} else if dnsRes, ok := res.val.(*types.DNSResult); ok {
					report.DNS = dnsRes
				}
			case "blacklist":
				if res.err != nil {
					fmt.Fprintf(os.Stderr, "Error performing blacklist check: %v\n", res.err)
				} else if blRes, ok := res.val.(*types.BlacklistResult); ok {
					report.Blacklist = blRes
				}
			case "smtp":
				if res.err != nil {
					fmt.Fprintf(os.Stderr, "Error performing SMTP check: %v\n", res.err)
				} else if smtpRes, ok := res.val.(*types.SMTPResult); ok {
					report.SMTP = smtpRes
				}
			case "auth":
				if res.err != nil {
					fmt.Fprintf(os.Stderr, "Error performing email authentication check: %v\n", res.err)
				} else if authRes, ok := res.val.(*types.AuthResult); ok {
					report.Auth = authRes
				}
			}
		}

		report.OverallStatus = calculateOverallStatus(report)

		if outputFormat == "json" {
			jsonOutput, err := json.MarshalIndent(report, "", "  ")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error marshaling JSON: %v\n", err)
				os.Exit(1)
			}
			fmt.Println(string(jsonOutput))
		} else {
			fmt.Println(formatHealthReport(report))
		}
	},
}

// performDNSCheck performs DNS checks for a domain.
func performDNSCheck(ctx context.Context, domain string, timeout time.Duration) (*types.DNSResult, error) {
	// Perform DNS lookup for common record types
	return dns.LookupAll(ctx, domain)
}

// performBlacklistCheck performs blacklist checks for a domain.
func performBlacklistCheck(ctx context.Context, domain string, timeout time.Duration) (*types.BlacklistResult, error) {
	// First, get the IP addresses for the domain
	dnsResult, err := dns.LookupWithRetry(ctx, domain, "A", 2, timeout)
	if err != nil || len(dnsResult.Lookups["A"]) == 0 {
		return nil, fmt.Errorf("failed to resolve domain to IP: %v", err)
	}

	// Use the first IP address for blacklist check
	ip := dnsResult.Lookups["A"][0]

	// Default blacklist zones
	zones := []string{
		"bl.spamcop.net",
		"dnsbl.sorbs.net",
	}

	// Check the IP against the blacklists
	return dnsbl.CheckMultipleBlacklists(ctx, ip, zones, timeout), nil
}

// performSMTPCheck performs SMTP checks for a domain.
func performSMTPCheck(ctx context.Context, domain string, timeout time.Duration) (*types.SMTPResult, error) {
	// First, get MX records for the domain
	mxResult, err := dns.LookupWithRetry(ctx, domain, "MX", 2, timeout)
	if err != nil || len(mxResult.Lookups["MX"]) == 0 {
		return nil, fmt.Errorf("failed to get MX records: %v", err)
	}

	// Extract the hostname from the first MX record
	mxRecord := mxResult.Lookups["MX"][0]
	hostname := mxRecord
	if idx := strings.Index(mxRecord, " (priority:"); idx > 0 {
		hostname = mxRecord[:idx]
	}

	// Perform SMTP check on the MX server
	return smtp.CheckSMTP(ctx, hostname, smtp.DefaultPorts, timeout)
}

// performAuthCheck performs email authentication checks for a domain.
func performAuthCheck(ctx context.Context, domain string, timeout time.Duration) (*types.AuthResult, error) {
	// Perform email authentication check
	return emailauth.CheckEmailAuth(ctx, domain, timeout)
}

// calculateOverallStatus calculates the overall status of a domain based on the check results.
func calculateOverallStatus(report *types.DomainHealthReport) string {
	// Initialize with "Healthy" status
	status := "Healthy"
	issues := []string{}

	// Check DNS results
	if report.DNS != nil && report.DNS.Error != "" {
		issues = append(issues, "DNS issues found")
	}

	// Check blacklist results
	if report.Blacklist != nil && len(report.Blacklist.ListedOn) > 0 {
		issues = append(issues, "Domain is listed on blacklists")
	}

	// Check SMTP results
	if report.SMTP != nil {
		if !report.SMTP.ConnectSuccess {
			issues = append(issues, "SMTP connection failed")
		} else {
			// Check if STARTTLS is supported
			if report.SMTP.SupportsSTARTTLS != nil && !*report.SMTP.SupportsSTARTTLS {
				issues = append(issues, "SMTP server does not support STARTTLS")
			}
			// Check if it's an open relay
			if report.SMTP.IsOpenRelay != nil && *report.SMTP.IsOpenRelay {
				issues = append(issues, "SMTP server is an open relay")
			}
		}
	}

	// Check email authentication results
	if report.Auth != nil {
		// Check SPF
		if report.Auth.SPFError != "" {
			issues = append(issues, "SPF record issues found")
		}
		// Check DMARC
		if report.Auth.DMARCError != "" {
			issues = append(issues, "DMARC record issues found")
		}
		// Check DKIM
		if report.Auth.DKIMError != "" {
			issues = append(issues, "DKIM record issues found")
		}
	}

	// If there are any issues, update the status
	if len(issues) > 0 {
		status = "Issues Found"
	}

	return status
}

// formatHealthReport formats a health report as a string.
func formatHealthReport(report *types.DomainHealthReport) string {
	output := fmt.Sprintf("Domain Health Report for %s\n", report.Domain)
	output += fmt.Sprintf("Timestamp: %s\n", report.Timestamp.Format(time.RFC3339))
	output += fmt.Sprintf("Overall Status: %s\n\n", report.OverallStatus)

	// DNS results
	if report.DNS != nil {
		output += "DNS Results:\n"
		if report.DNS.Error != "" {
			output += fmt.Sprintf("  Error: %s\n", report.DNS.Error)
		} else {
			for recordType, records := range report.DNS.Lookups {
				output += fmt.Sprintf("  %s records:\n", recordType)
				for _, record := range records {
					output += fmt.Sprintf("    %s\n", record)
				}
			}
		}
		output += "\n"
	}

	// Blacklist results
	if report.Blacklist != nil {
		output += "Blacklist Results:\n"
		if len(report.Blacklist.ListedOn) == 0 {
			output += fmt.Sprintf("  IP %s is not listed on any blacklists\n", report.Blacklist.CheckedIP)
		} else {
			output += fmt.Sprintf("  IP %s is listed on %d blacklists:\n", report.Blacklist.CheckedIP, len(report.Blacklist.ListedOn))
			for zone, reason := range report.Blacklist.ListedOn {
				if reason != "" {
					output += fmt.Sprintf("    %s: %s\n", zone, reason)
				} else {
					output += fmt.Sprintf("    %s\n", zone)
				}
			}
		}
		if report.Blacklist.CheckError != "" {
			output += fmt.Sprintf("  Error: %s\n", report.Blacklist.CheckError)
		}
		output += "\n"
	}

	// SMTP results
	if report.SMTP != nil {
		output += "SMTP Results:\n"
		output += fmt.Sprintf("  Connection: %t\n", report.SMTP.ConnectSuccess)
		if !report.SMTP.ConnectSuccess {
			output += fmt.Sprintf("  Connection error: %s\n", report.SMTP.ConnectError)
		} else {
			output += fmt.Sprintf("  Response time: %s\n", report.SMTP.ResponseTime)
			if report.SMTP.SupportsSTARTTLS != nil {
				output += fmt.Sprintf("  Supports STARTTLS: %t\n", *report.SMTP.SupportsSTARTTLS)
				if report.SMTP.STARTTLSError != "" {
					output += fmt.Sprintf("  STARTTLS error: %s\n", report.SMTP.STARTTLSError)
				}
			}
			if report.SMTP.IsOpenRelay != nil {
				output += fmt.Sprintf("  Open relay: %t\n", *report.SMTP.IsOpenRelay)
				if report.SMTP.RelayCheckError != "" {
					output += fmt.Sprintf("  Relay check error: %s\n", report.SMTP.RelayCheckError)
				}
			}
		}
		output += "\n"
	}

	// Email authentication results
	if report.Auth != nil {
		output += "Email Authentication Results:\n"

		// SPF results
		output += "  SPF:\n"
		if report.Auth.SPFRecord != "" {
			output += fmt.Sprintf("    Record: %s\n", report.Auth.SPFRecord)
			if report.Auth.SPFResult != "" {
				output += fmt.Sprintf("    Result: %s\n", report.Auth.SPFResult)
			}
		} else {
			output += fmt.Sprintf("    Error: %s\n", report.Auth.SPFError)
		}

		// DMARC results
		output += "  DMARC:\n"
		if report.Auth.DMARCRecord != "" {
			output += fmt.Sprintf("    Record: %s\n", report.Auth.DMARCRecord)
			if report.Auth.DMARCPolicy != "" {
				output += fmt.Sprintf("    Policy: %s\n", report.Auth.DMARCPolicy)
			}
		} else {
			output += fmt.Sprintf("    Error: %s\n", report.Auth.DMARCError)
		}

		// DKIM results
		if report.Auth.DKIMRecord != "" || report.Auth.DKIMError != "" {
			output += "  DKIM:\n"
			if report.Auth.DKIMRecord != "" {
				output += fmt.Sprintf("    Record: %s\n", report.Auth.DKIMRecord)
				if report.Auth.DKIMResult != "" {
					output += fmt.Sprintf("    Result: %s\n", report.Auth.DKIMResult)
				}
			} else {
				output += fmt.Sprintf("    Error: %s\n", report.Auth.DKIMError)
			}
		}
	}

	return output
}

func init() {
	HealthCmd.Flags().IntP("timeout", "t", 30, "Timeout in seconds for each check")
	HealthCmd.Flags().BoolP("check-dns", "d", true, "Perform DNS checks")
	HealthCmd.Flags().BoolP("check-blacklist", "b", true, "Perform blacklist checks")
	HealthCmd.Flags().BoolP("check-smtp", "s", true, "Perform SMTP checks")
	HealthCmd.Flags().BoolP("check-auth", "a", true, "Perform email authentication checks")

	// Add the command to the root command
	rootCmd.AddCommand(HealthCmd)
}

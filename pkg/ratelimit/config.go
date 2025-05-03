// Package ratelimit provides rate limiting functionality for external service queries.
package ratelimit

import (
	"context"
)

// ServiceConfig defines the rate limiting configuration for a service.
type ServiceConfig struct {
	// Rate is the number of requests allowed per second.
	Rate float64
	// BucketSize is the maximum number of tokens that can be accumulated.
	BucketSize float64
}

// DefaultConfigs contains the default rate limiting configurations for various services.
var DefaultConfigs = map[string]ServiceConfig{
	// DNS services
	"dns":           {Rate: 10.0, BucketSize: 20.0},  // General DNS lookups
	"dns.advanced":  {Rate: 5.0, BucketSize: 10.0},   // Advanced DNS lookups (using miekg/dns)

	// DNSBL services
	"dnsbl":         {Rate: 5.0, BucketSize: 10.0},   // General DNSBL checks
	"dnsbl.health":  {Rate: 2.0, BucketSize: 5.0},    // DNSBL health checks

	// SMTP services
	"smtp":          {Rate: 2.0, BucketSize: 5.0},    // General SMTP checks
	"smtp.starttls": {Rate: 2.0, BucketSize: 5.0},    // SMTP STARTTLS checks
	"smtp.relay":    {Rate: 1.0, BucketSize: 2.0},    // SMTP relay checks (more conservative)

	// Email authentication services
	"emailauth":     {Rate: 5.0, BucketSize: 10.0},   // General email authentication checks
	"emailauth.spf": {Rate: 5.0, BucketSize: 10.0},   // SPF record checks
	"emailauth.dkim":{Rate: 5.0, BucketSize: 10.0},   // DKIM record checks
	"emailauth.dmarc":{Rate: 5.0, BucketSize: 10.0},  // DMARC record checks

	// Network tools
	"whois":         {Rate: 1.0, BucketSize: 2.0},    // WHOIS queries (very conservative)
	"ping":          {Rate: 5.0, BucketSize: 10.0},   // ICMP ping
	"traceroute":    {Rate: 2.0, BucketSize: 5.0},    // Traceroute
}

// GetConfig returns the rate limiting configuration for a service.
func GetConfig(service string) ServiceConfig {
	if config, ok := DefaultConfigs[service]; ok {
		return config
	}
	// Return a default conservative configuration if the service is not found
	return ServiceConfig{Rate: 1.0, BucketSize: 2.0}
}

// WaitForService blocks until the rate limit for the specified service allows an operation to proceed.
func WaitForService(ctx context.Context, service string) error {
	config := GetConfig(service)
	return Wait(ctx, service, config.Rate, config.BucketSize)
}

// AllowService returns true if the operation for the specified service is allowed to proceed.
func AllowService(service string) bool {
	config := GetConfig(service)
	return Allow(service, config.Rate, config.BucketSize)
}

// RateLimitedServiceFunc wraps a function with rate limiting for a specific service.
func RateLimitedServiceFunc(ctx context.Context, service string, fn func() error) error {
	config := GetConfig(service)
	return RateLimitedFunc(ctx, service, config.Rate, config.BucketSize, fn)
}

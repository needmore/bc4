// This file demonstrates the improved error handling in bc4
// It's not meant to be run, but shows the before/after comparison

package main

import (
	"fmt"
	"net/http"
	
	"github.com/needmore/bc4/internal/errors"
)

func main() {
	// Example 1: Authentication Error
	// Before: failed to fetch projects: API error: {"error": "invalid_token"} (status: 401)
	// After:
	authErr := errors.NewAuthenticationError(fmt.Errorf("invalid token"))
	fmt.Println("Authentication Error:")
	fmt.Println(errors.FormatError(authErr))
	fmt.Println()
	
	// Example 2: Not Found Error
	// Before: API error: {"error": "not found"} (status: 404)
	// After:
	notFoundErr := errors.NewNotFoundError("Project", "12345", nil)
	fmt.Println("Not Found Error:")
	fmt.Println(errors.FormatError(notFoundErr))
	fmt.Println()
	
	// Example 3: API Error (Rate Limit)
	// Before: API error: {"error": "rate limit exceeded"} (status: 429)
	// After:
	rateLimitErr := errors.NewAPIError(http.StatusTooManyRequests, "rate limit exceeded", nil)
	fmt.Println("Rate Limit Error:")
	fmt.Println(errors.FormatError(rateLimitErr))
	fmt.Println()
	
	// Example 4: Configuration Error
	// Before: no account specified and no default account set
	// After:
	configErr := errors.NewConfigurationError("no account specified and no default account set", nil)
	fmt.Println("Configuration Error:")
	fmt.Println(errors.FormatError(configErr))
	fmt.Println()
	
	// Example 5: Network Error
	// Before: request failed: dial tcp: lookup api.basecamp.com: no such host
	// After:
	networkErr := errors.NewNetworkError(fmt.Errorf("connection refused"))
	fmt.Println("Network Error:")
	fmt.Println(errors.FormatError(networkErr))
}

/* Output would be:

Authentication Error:
✗ Authentication failed

Your authentication token may have expired or been revoked.

→ Run 'bc4 auth login' to refresh your credentials

Not Found Error:
✗ Project not found

Could not find project with ID '12345'.

→ Check the ID for typos
→ Run 'bc4 project list' to see available items

Rate Limit Error:
✗ API request failed

Rate limit exceeded. Please wait before trying again.

→ Wait a few moments and try again

Configuration Error:
✗ No account selected

You need to select a default account or specify one with --account.

→ Run 'bc4 account select' to choose a default account
→ Or use '--account <id>' to specify an account

Network Error:
✗ Network connection failed

Unable to connect to Basecamp. This could be due to:
  • No internet connection
  • Firewall or proxy blocking the connection
  • Basecamp service temporarily unavailable

→ Check your internet connection and try again
*/
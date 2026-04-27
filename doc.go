// Package nendesktop provides a Go client for the Nen Desktop API.
//
// Create cloud desktops, execute computer-use tools, and manage RDP
// sessions programmatically.
//
//	client := nendesktop.New(nendesktop.Config{APIKey: "sk_nen_..."})
//	desktop, err := client.CreateDesktop(ctx, "sandbox")
package nendesktop

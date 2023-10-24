//go:build tools

package tools

import (
	// Documentation generation
	_ "github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs"
	// Linter
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
)

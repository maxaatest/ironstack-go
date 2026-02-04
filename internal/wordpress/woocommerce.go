package wordpress

import (
	"fmt"
)

// WooCommerce manages WooCommerce-specific operations
type WooCommerce struct {
	wp *WordPress
}

// NewWooCommerce creates a WooCommerce manager
func NewWooCommerce(wp *WordPress) *WooCommerce {
	return &WooCommerce{wp: wp}
}

// Install sets up WooCommerce with recommended settings
func (wc *WooCommerce) Install() error {
	// Install WooCommerce
	if err := wc.wp.InstallPlugin("woocommerce"); err != nil {
		return fmt.Errorf("failed to install WooCommerce: %w", err)
	}

	// Install recommended plugins
	recommended := []string{
		"woo-gutenberg-products-block",
		"woocommerce-gateway-stripe",
		"woocommerce-services",
	}

	for _, plugin := range recommended {
		wc.wp.InstallPlugin(plugin)
	}

	return nil
}

// OptimizeForWoo applies WooCommerce-specific optimizations
func (wc *WooCommerce) OptimizeForWoo() error {
	// Increase memory for WooCommerce
	wc.wp.run("config", "set", "WP_MEMORY_LIMIT", "'512M'", "--raw")

	// WooCommerce-specific settings
	configs := map[string]string{
		"WC_LOG_HANDLER":          "'WC_Log_Handler_DB'",
		"DISABLE_WP_CRON":         "true",
		"ALTERNATE_WP_CRON":       "true",
	}

	for key, value := range configs {
		wc.wp.run("config", "set", key, value, "--raw")
	}

	return nil
}

// SetupCron configures system cron for WooCommerce
func (wc *WooCommerce) SetupCron() string {
	publicPath := wc.wp.Path + "/public"
	return fmt.Sprintf("*/5 * * * * cd %s && wp cron event run --due-now > /dev/null 2>&1", publicPath)
}

// ClearTransients clears WooCommerce transients
func (wc *WooCommerce) ClearTransients() error {
	return wc.wp.run("transient", "delete", "--all")
}

// ReindexProducts regenerates product lookup tables
func (wc *WooCommerce) ReindexProducts() error {
	return wc.wp.run("wc", "tool", "run", "regenerate_product_lookup_tables")
}

package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// Caddy generates Caddyfile configurations
type Caddy struct {
	ConfigDir string
}

// NewCaddy creates Caddy config generator
func NewCaddy() *Caddy {
	return &Caddy{ConfigDir: "/etc/caddy/sites"}
}

// AddSite generates Caddyfile for a domain
func (c *Caddy) AddSite(domain string, enableVarnish bool) error {
	os.MkdirAll(c.ConfigDir, 0755)
	
	backend := "127.0.0.1:9000" // FrankenPHP
	if enableVarnish {
		backend = "127.0.0.1:6081" // Varnish
	}
	
	config := fmt.Sprintf(`%s {
    root * /var/www/%s/public
    
    # PHP handling via FrankenPHP
    php_fastcgi %s
    
    # Static file serving
    file_server
    
    # Compression
    encode gzip zstd
    
    # Security headers
    header {
        X-Content-Type-Options "nosniff"
        X-Frame-Options "SAMEORIGIN"
        X-XSS-Protection "1; mode=block"
        Referrer-Policy "strict-origin-when-cross-origin"
    }
    
    # Cache static assets
    @static {
        path *.css *.js *.ico *.gif *.jpg *.jpeg *.png *.svg *.woff *.woff2
    }
    header @static Cache-Control "public, max-age=31536000, immutable"
    
    # Block sensitive files
    @blocked {
        path /wp-config.php
        path /.git*
        path /readme.html
        path /license.txt
    }
    respond @blocked 404
    
    # Logs
    log {
        output file /var/log/caddy/%s-access.log
    }
}
`, domain, domain, backend, domain)

	return os.WriteFile(filepath.Join(c.ConfigDir, domain+".conf"), []byte(config), 0644)
}

// Varnish generates VCL configurations
type Varnish struct{}

// NewVarnish creates Varnish config generator
func NewVarnish() *Varnish {
	return &Varnish{}
}

// GenerateVCL creates WordPress-optimized VCL
func (v *Varnish) GenerateVCL() string {
	return `vcl 4.1;

backend default {
    .host = "127.0.0.1";
    .port = "8080";
    .connect_timeout = 5s;
    .first_byte_timeout = 90s;
    .between_bytes_timeout = 2s;
}

sub vcl_recv {
    # Skip cache for logged-in WordPress users
    if (req.http.Cookie ~ "wordpress_logged_in|wp-postpass|woocommerce_cart_hash|woocommerce_items_in_cart") {
        return (pass);
    }
    
    # Skip cache for POST requests and admin
    if (req.method == "POST" || req.url ~ "wp-admin|wp-login|xmlrpc.php|preview=true") {
        return (pass);
    }
    
    # Skip cache for WooCommerce dynamic pages
    if (req.url ~ "cart|checkout|my-account|add-to-cart|logout|lost-password") {
        return (pass);
    }
    
    # Remove cookies for static files
    if (req.url ~ "\.(css|js|jpg|jpeg|png|gif|ico|svg|woff|woff2|ttf|eot|webp|avif)(\?.*)?$") {
        unset req.http.Cookie;
        return (hash);
    }
    
    # Remove tracking cookies
    set req.http.Cookie = regsuball(req.http.Cookie, "(^|;\s*)(_ga|_gid|_gat|__utm[a-z]+|_fbp|_fbc)[^;]*", "");
    
    return (hash);
}

sub vcl_backend_response {
    # Cache static files for 30 days
    if (bereq.url ~ "\.(css|js|jpg|jpeg|png|gif|ico|svg|woff|woff2|ttf|eot|webp|avif)(\?.*)?$") {
        set beresp.ttl = 30d;
        unset beresp.http.Set-Cookie;
    }
    
    # Default cache time for HTML
    if (beresp.http.Content-Type ~ "text/html") {
        set beresp.ttl = 1h;
    }
    
    # Grace period for stale content
    set beresp.grace = 24h;
}

sub vcl_deliver {
    # Cache hit/miss header
    if (obj.hits > 0) {
        set resp.http.X-Cache = "HIT";
        set resp.http.X-Cache-Hits = obj.hits;
    } else {
        set resp.http.X-Cache = "MISS";
    }
    
    # Remove internal headers
    unset resp.http.X-Varnish;
    unset resp.http.Via;
}
`
}

// WriteVCL saves VCL to default location
func (v *Varnish) WriteVCL() error {
	return os.WriteFile("/etc/varnish/default.vcl", []byte(v.GenerateVCL()), 0644)
}

// WordPress generates wp-config optimizations
type WordPress struct{}

// NewWordPress creates WordPress config generator
func NewWordPress() *WordPress {
	return &WordPress{}
}

// OptimizeConfig returns performance constants for wp-config.php
func (w *WordPress) OptimizeConfig() string {
	return `
// IronStack Performance Optimizations
define('WP_MEMORY_LIMIT', '256M');
define('WP_MAX_MEMORY_LIMIT', '512M');
define('WP_POST_REVISIONS', 5);
define('AUTOSAVE_INTERVAL', 120);
define('EMPTY_TRASH_DAYS', 7);
define('DISABLE_WP_CRON', true);
define('WP_CACHE', true);

// DragonflyDB Object Cache
define('WP_REDIS_HOST', '127.0.0.1');
define('WP_REDIS_PORT', 6379);
define('WP_REDIS_DATABASE', 0);

// Security
define('DISALLOW_FILE_EDIT', true);
define('FORCE_SSL_ADMIN', true);
`
}

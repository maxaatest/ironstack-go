package modules

import (
	"fmt"
	"os/exec"
)

// Varnish manages Varnish cache
type Varnish struct{}

func NewVarnish() *Varnish {
	return &Varnish{}
}

func (v *Varnish) Install() error {
	return exec.Command("apt-get", "install", "-y", "varnish").Run()
}

func (v *Varnish) GenerateVCL(domain string) string {
	return `
vcl 4.1;

backend default {
    .host = "127.0.0.1";
    .port = "8080";
}

sub vcl_recv {
    # Skip cache for logged-in WordPress users
    if (req.http.Cookie ~ "wordpress_logged_in") {
        return (pass);
    }
    
    # Skip cache for POST requests
    if (req.method == "POST") {
        return (pass);
    }
    
    # Skip cache for WooCommerce cart/checkout
    if (req.url ~ "^/(cart|checkout|my-account)") {
        return (pass);
    }
    
    # Remove cookies for static files
    if (req.url ~ "\.(css|js|jpg|jpeg|png|gif|ico|svg|woff2)$") {
        unset req.http.Cookie;
    }
    
    return (hash);
}

sub vcl_backend_response {
    # Cache static files for 30 days
    if (bereq.url ~ "\.(css|js|jpg|jpeg|png|gif|ico|svg|woff2)$") {
        set beresp.ttl = 30d;
    }
    
    # Default cache time
    if (beresp.ttl <= 0s) {
        set beresp.ttl = 1h;
    }
}

sub vcl_deliver {
    # Add cache hit/miss header
    if (obj.hits > 0) {
        set resp.http.X-Cache = "HIT";
    } else {
        set resp.http.X-Cache = "MISS";
    }
}
`
}

func (v *Varnish) Purge(url string) error {
	return exec.Command("varnishadm", "ban", fmt.Sprintf("req.url ~ %s", url)).Run()
}

func (v *Varnish) PurgeAll() error {
	return exec.Command("varnishadm", "ban", "req.url ~ .").Run()
}

func (v *Varnish) Status() (string, error) {
	out, err := exec.Command("systemctl", "is-active", "varnish").Output()
	return string(out), err
}

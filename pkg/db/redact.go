package db

import (
	"net/url"
	"regexp"
	"strings"
)

var authorityPassRe = regexp.MustCompile(`(?i)(://[^:/@?]+:)([^@/\?]+)(@)`)
var kvPassRe = regexp.MustCompile(`(?i)((?:password|passwd|pwd)=)([^\s&;]+)`)
var dsnURIRe = regexp.MustCompile(`(?i)(postgres(?:ql)?://[^\s"'<>]+)`)

func RedactDSN(dsn string) string {
	s := strings.TrimSpace(dsn)
	if s == "" {
		return s
	}

	if u, err := url.Parse(s); err == nil {
		if u.User != nil {
			username := u.User.Username()
			if username != "" {
				u.User = url.UserPassword(username, "***")
			}
		}
		q := u.Query()
		changed := false
		for _, key := range []string{"password", "passwd", "pwd"} {
			if q.Has(key) {
				q.Set(key, "***")
				changed = true
			}
		}
		if changed {
			u.RawQuery = q.Encode()
		}
		s = u.String()
	}

	s = authorityPassRe.ReplaceAllString(s, `${1}***${3}`)
	s = kvPassRe.ReplaceAllString(s, `${1}***`)
	return s
}

func SafeDSNInfo(dsn string) string {
	s := strings.TrimSpace(dsn)
	if s == "" {
		return ""
	}
	u, err := url.Parse(s)
	if err != nil {
		return "unknown"
	}
	host := u.Host
	db := strings.TrimPrefix(u.Path, "/")
	if host == "" && db == "" {
		return "unknown"
	}
	if db == "" {
		return host
	}
	return host + "/" + db
}

func RedactText(s string) string {
	if strings.TrimSpace(s) == "" {
		return s
	}
	s = dsnURIRe.ReplaceAllStringFunc(s, RedactDSN)
	s = kvPassRe.ReplaceAllString(s, `${1}***`)
	s = authorityPassRe.ReplaceAllString(s, `${1}***${3}`)
	return s
}

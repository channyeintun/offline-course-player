package view

import (
	"embed"
	"fmt"
	"html/template"
	"net/url"
	"path/filepath"
	"strings"
)

//go:embed templates/*.html
var templatesFS embed.FS

// ParseTemplates returns the parsed HTML templates used by the application.
func ParseTemplates() (*template.Template, error) {
	tmpl := template.New("base").Funcs(template.FuncMap{
		"formatTitle": formatTitle,
		"urlencode":   url.QueryEscape,
		"dict": func(values ...interface{}) map[string]interface{} {
			if len(values)%2 != 0 {
				return nil
			}
			dict := make(map[string]interface{}, len(values)/2)
			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					continue
				}
				dict[key] = values[i+1]
			}
			return dict
		},
		"hash": func(s string) string {
			return fmt.Sprintf("%x", s)
		},
	})
	
	return tmpl.ParseFS(templatesFS, "templates/*.html")
}

func isNumeric(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func formatTitle(name string) string {
	title := strings.TrimSuffix(name, filepath.Ext(name))
	parts := strings.SplitN(title, "_", 2)
	if len(parts) == 2 && isNumeric(parts[0]) {
		title = parts[1]
	} else {
		parts := strings.SplitN(title, "-", 2)
		if len(parts) == 2 && isNumeric(parts[0]) {
			title = parts[1]
		}
	}
	return strings.ReplaceAll(title, "_", " ")
}

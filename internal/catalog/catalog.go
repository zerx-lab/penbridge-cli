// Package catalog provides a static, embedded list of every SiYuan kernel API
// endpoint. It is generated from the SiYuan source (kernel/api/router.go) and
// powers endpoint discovery (`penbridge api list`).
package catalog

import (
	"sort"
	"strings"
)

//go:generate go run ./gen ../../../siyuan/kernel/api/router.go catalog_data.go

// Endpoint describes a single kernel API route.
type Endpoint struct {
	// Method is the HTTP method (almost always POST).
	Method string `json:"method"`
	// Path is the absolute API path, e.g. /api/block/insertBlock.
	Path string `json:"path"`
	// Category is the path group, e.g. "block", "notebook".
	Category string `json:"category"`
	// Name is the final path segment, e.g. "insertBlock".
	Name string `json:"name"`
	// Auth indicates the endpoint requires authentication (CheckAuth).
	Auth bool `json:"auth"`
	// Admin indicates the endpoint requires the administrator role.
	Admin bool `json:"admin"`
	// Write indicates the endpoint mutates data (blocked in read-only mode).
	Write bool `json:"write"`
	// Deprecated indicates the endpoint is scheduled for removal.
	Deprecated bool `json:"deprecated"`
}

// All returns a copy of every known endpoint, sorted by path.
func All() []Endpoint {
	out := make([]Endpoint, len(endpoints))
	copy(out, endpoints)
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out
}

// Categories returns the sorted, unique list of categories.
func Categories() []string {
	seen := map[string]bool{}
	var cats []string
	for _, e := range endpoints {
		if !seen[e.Category] {
			seen[e.Category] = true
			cats = append(cats, e.Category)
		}
	}
	sort.Strings(cats)
	return cats
}

// Get returns the endpoint matching the given path (after normalization), or
// false if unknown.
func Get(path string) (Endpoint, bool) {
	path = normalize(path)
	for _, e := range endpoints {
		if e.Path == path {
			return e, true
		}
	}
	return Endpoint{}, false
}

// Filter returns endpoints matching the optional search term (substring of path
// or name) and category. Empty filters match everything.
func Filter(search, category string, writesOnly bool) []Endpoint {
	search = strings.ToLower(strings.TrimSpace(search))
	category = strings.ToLower(strings.TrimSpace(category))
	var out []Endpoint
	for _, e := range All() {
		if category != "" && strings.ToLower(e.Category) != category {
			continue
		}
		if writesOnly && !e.Write {
			continue
		}
		if search != "" {
			hay := strings.ToLower(e.Path + " " + e.Name)
			if !strings.Contains(hay, search) {
				continue
			}
		}
		out = append(out, e)
	}
	return out
}

func normalize(p string) string {
	p = strings.TrimSpace(p)
	if p == "" {
		return p
	}
	if strings.HasPrefix(p, "/") {
		return p
	}
	if strings.HasPrefix(p, "api/") {
		return "/" + p
	}
	return "/api/" + p
}

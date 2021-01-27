package go_vanity_urls

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"net/http"
	"sort"
	"strings"
)

type pathInfo struct {
	path    string
	repo    string
	display string
}

type pathInfos []pathInfo

func (p pathInfos) Len() int {
	return len(p)
}

func (p pathInfos) Less(i, j int) bool {
	return p[i].path < p[j].path
}

func (p pathInfos) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (p pathInfos) get(path string) (*pathInfo, string) {
	// Fast path with binary search to retrieve exact matches
	// e.g. given p ["/", "/abc", "/xyz"], path "/def" won't match.
	i := sort.Search(len(p), func(i int) bool {
		return p[i].path >= path
	})
	if i < len(p) && p[i].path == path {
		return &p[i], ""
	}
	if i > 0 && strings.HasPrefix(path, p[i-1].path+"/") {
		return &p[i-1], path[len(p[i-1].path)+1:]
	}

	// Slow path, now looking for the longest prefix/shortest subpath i.e.
	// e.g. given p ["/", "/abc/", "/abc/def/", "/xyz"/]
	//  * query "/abc/foo" returns "/abc/" with a subpath of "foo"
	//  * query "/x" returns "/" with a subpath of "x"
	lenShortestSubpath := len(path)
	var bestMatchConfig *pathInfo
	var subpath string

	// After binary search with the >= lexicographic comparison,
	// nothing greater than i will be a prefix of path.
	max := i
	for i := 0; i < max; i++ {
		ps := p[i]
		if len(ps.path) >= len(path) {
			// We previously didn't find the path by search, so any
			// route with equal or greater length is NOT a match.
			continue
		}
		sSubpath := strings.TrimPrefix(path, ps.path)
		if len(sSubpath) < lenShortestSubpath {
			subpath = sSubpath
			lenShortestSubpath = len(sSubpath)
			bestMatchConfig = &p[i]
		}
	}
	return bestMatchConfig, subpath
}

type handler struct {
	host         string
	cacheControl string
	paths        pathInfos
}

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	currentURL := r.URL.Path

	pc, subpath := h.paths.get(currentURL)

	if pc == nil && currentURL == "/" {
		host := h.getHost(r)
		handlers := make([]string, len(h.paths))
		for i, h := range h.paths {
			handlers[i] = host + h.path
		}
		if err := indexTmpl.Execute(w, struct {
			Host     string
			Handlers []string
		}{
			Host:     host,
			Handlers: handlers,
		}); err != nil {
			http.Error(w, "cannot render the page", http.StatusInternalServerError)
		}
	} else if pc == nil {
		http.NotFound(w, r)
	} else {
		w.Header().Set("Cache-Control", h.cacheControl)
		if err := vanityTmpl.Execute(w, struct {
			Import  string
			Subpath string
			Repo    string
			Display string
		}{
			Import:  h.getHost(r) + pc.path,
			Subpath: subpath,
			Repo:    pc.repo,
			Display: pc.display,
		}); err != nil {
			http.Error(w, "cannot render the page", http.StatusInternalServerError)
		}
	}
}

func (h *handler) getHost(r *http.Request) string {
	host := h.host
	if host == "" {
		host = r.Host
	}
	return host
}

func NewHandler(config []byte) (*handler, error) {
	var parsedConfig struct {
		Host     string `yaml:"host,omitempty"`
		CacheAge *int64 `yaml:"cache_max_age,omitempty"`
		Paths    map[string]struct {
			Repo   string `yaml:"repo,omitempty"`
			Branch string `yaml:"branch,omitempty"`
		} `yaml:"paths"`
	}

	if err := yaml.Unmarshal(config, &parsedConfig); err != nil {
		return nil, err
	}

	h := &handler{host: parsedConfig.Host}
	cacheAge := int64(86400)
	if parsedConfig.CacheAge != nil {
		cacheAge = *parsedConfig.CacheAge
		if cacheAge < 0 {
			return nil, errors.New("cache_max_age cannot be negative")
		}
	}

	h.cacheControl = fmt.Sprintf("public, max-age=%d", cacheAge)

	for p, e := range parsedConfig.Paths {
		pi := pathInfo{
			path: strings.TrimSuffix(p, "/"),
			repo: e.Repo,
		}

		if e.Branch == "" {
			e.Branch = "master"
		}

		pi.display = fmt.Sprintf("%s %s/tree/%s{/dir} %s/blob/%s{/dir}/{file}#L{line}", e.Repo, e.Repo, e.Branch, e.Repo, e.Branch)

		h.paths = append(h.paths, pi)
	}

	sort.Sort(h.paths)

	return h, nil
}

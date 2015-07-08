package cf_ignore

import (
	"io"
	"io/ioutil"
	"path"
	"strings"

	"github.com/cloudfoundry-incubator/lattice/ltc/droplet_runner/command_factory/cf_ignore/glob"
)

//go:generate counterfeiter -o fake_cf_ignore/fake_cf_ignore.go . CFIgnore
type CFIgnore interface {
	Parse(ignored io.Reader) error
	ShouldIgnore(path string) bool
}

type ignorePattern struct {
	exclude bool
	glob    glob.Glob
}

type cfIgnore struct {
	patterns []ignorePattern
}

var defaultIgnoreLines = []string{
	".cfignore",
	"/manifest.yml",
	".gitignore",
	".git",
	".hg",
	".svn",
	"_darcs",
	".DS_Store",
}

func New() CFIgnore {
	return &cfIgnore{
		parsePatterns(defaultIgnoreLines),
	}
}

func (c *cfIgnore) Parse(ignored io.Reader) error {
	ignoredBytes, err := ioutil.ReadAll(ignored)
	if err != nil {
		return err
	}

	lines := strings.Split(string(ignoredBytes), "\n")
	c.patterns = append(c.patterns, parsePatterns(lines)...)

	return nil
}

func (c *cfIgnore) ShouldIgnore(path string) bool {
	result := false

	for _, pattern := range c.patterns {
		if strings.HasPrefix(pattern.glob.String(), "/") && !strings.HasPrefix(path, "/") {
			path = "/" + path
		}

		if pattern.glob.Match(path) {
			result = pattern.exclude
		}
	}

	return result
}

func globsForPattern(pattern string) (globs []glob.Glob) {
	globs = append(globs, glob.MustCompileGlob(pattern))
	globs = append(globs, glob.MustCompileGlob(path.Join(pattern, "*")))
	globs = append(globs, glob.MustCompileGlob(path.Join(pattern, "**", "*")))

	if !strings.HasPrefix(pattern, "/") {
		globs = append(globs, glob.MustCompileGlob(path.Join("**", pattern)))
		globs = append(globs, glob.MustCompileGlob(path.Join("**", pattern, "*")))
		globs = append(globs, glob.MustCompileGlob(path.Join("**", pattern, "**", "*")))
	}

	return
}

func parsePatterns(ignoresLines []string) []ignorePattern {
	var patterns []ignorePattern

	for _, line := range ignoresLines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		ignore := true
		if strings.HasPrefix(line, "!") {
			line = line[1:]
			ignore = false
		}

		for _, p := range globsForPattern(path.Clean(line)) {
			patterns = append(patterns, ignorePattern{ignore, p})
		}
	}

	return patterns
}

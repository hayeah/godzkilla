package gozkilla

import "testing"

func TestParseRemote(t *testing.T) {
	cases := []struct {
		src      string
		wantRepo string
		wantSub  string
	}{
		{"github.com/hayeah/skills", "github.com/hayeah/skills", ""},
		{"github.com/hayeah/skills/foo", "github.com/hayeah/skills", "foo"},
		{"github.com/hayeah/skills/foo/bar", "github.com/hayeah/skills", "foo/bar"},
		{"gitlab.com/org/repo", "gitlab.com/org/repo", ""},
		{"gitlab.com/org/repo/deep/path", "gitlab.com/org/repo", "deep/path"},
	}
	for _, c := range cases {
		repo, sub := ParseRemote(c.src)
		if repo != c.wantRepo || sub != c.wantSub {
			t.Errorf("ParseRemote(%q) = (%q, %q); want (%q, %q)",
				c.src, repo, sub, c.wantRepo, c.wantSub)
		}
	}
}

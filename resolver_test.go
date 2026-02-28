package gozkilla

import "testing"

func TestNormalizeSource(t *testing.T) {
	cases := []struct {
		src  string
		want string
	}{
		// HTTPS URLs with /tree/<branch>/<path>
		{"https://github.com/hayeah/dotfiles/tree/master/skills", "github.com/hayeah/dotfiles/skills"},
		{"https://github.com/hayeah/dotfiles/tree/main/foo/bar", "github.com/hayeah/dotfiles/foo/bar"},
		// HTTPS URL with /tree/<branch> only (no subpath)
		{"https://github.com/hayeah/skills/tree/master", "github.com/hayeah/skills"},
		// HTTPS URL without /tree/
		{"https://github.com/hayeah/skills", "github.com/hayeah/skills"},
		// HTTP
		{"http://github.com/hayeah/skills/tree/main/foo", "github.com/hayeah/skills/foo"},
		// Already bare — unchanged
		{"github.com/hayeah/skills", "github.com/hayeah/skills"},
		{"github.com/hayeah/skills/foo", "github.com/hayeah/skills/foo"},
		// Local path — unchanged
		{"./my-skills", "./my-skills"},
		{"/abs/path", "/abs/path"},
	}
	for _, c := range cases {
		got := NormalizeSource(c.src)
		if got != c.want {
			t.Errorf("NormalizeSource(%q) = %q; want %q", c.src, got, c.want)
		}
	}
}

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

package skill

import (
	"testing"
)

func TestName(t *testing.T) {
	cases := []struct {
		baseName string
		relPath  string
		want     string
	}{
		{"github.com_hayeah_skills", "", "github.com_hayeah_skills"},
		{"github.com_hayeah_skills", "foo", "github.com_hayeah_skills_foo"},
		{"github.com_hayeah_skills", "bar/baz", "github.com_hayeah_skills_bar_baz"},
		{"github.com_hayeah_dotfiles_skills", "", "github.com_hayeah_dotfiles_skills"},
		{"github.com_hayeah_dotfiles_skills", "shell-helper", "github.com_hayeah_dotfiles_skills_shell-helper"},
		{"skills", "python", "skills_python"},
		{"skills", "", "skills"},
	}
	for _, c := range cases {
		got := Name(c.baseName, c.relPath)
		if got != c.want {
			t.Errorf("Name(%q, %q) = %q; want %q", c.baseName, c.relPath, got, c.want)
		}
	}
}

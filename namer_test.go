package gozkilla

import "testing"

func TestSkillName(t *testing.T) {
	cases := []struct {
		src     string
		relPath string
		want    string
	}{
		{"github.com/hayeah/skills", "", "github.com_hayeah_skills"},
		{"github.com/hayeah/skills", "foo", "github.com_hayeah_skills_foo"},
		{"github.com/hayeah/skills", "bar/baz", "github.com_hayeah_skills_bar_baz"},
		{"github.com/hayeah/dotfiles/skills", "shell-helper", "github.com_hayeah_dotfiles_skills_shell-helper"},
		{"skills", "python", "skills_python"},
		{"skills", "", "skills"},
	}
	for _, c := range cases {
		got := SkillName(c.src, c.relPath)
		if got != c.want {
			t.Errorf("SkillName(%q, %q) = %q; want %q", c.src, c.relPath, got, c.want)
		}
	}
}

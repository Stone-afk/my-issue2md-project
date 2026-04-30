package converter

import (
	"testing"

	"github.com/stoneafk/issue2md/internal/model"
)

func TestRenderUser(t *testing.T) {
	tests := []struct {
		name string
		user model.UserData
		opts Options
		want string
	}{
		{
			name: "plain text by default",
			user: model.UserData{Login: "octocat", URL: "https://github.com/octocat"},
			opts: Options{},
			want: "octocat",
		},
		{
			name: "markdown link when enabled",
			user: model.UserData{Login: "octocat", URL: "https://github.com/octocat"},
			opts: Options{EnableUserLinks: true},
			want: "[octocat](https://github.com/octocat)",
		},
		{
			name: "plain text fallback when url missing",
			user: model.UserData{Login: "octocat"},
			opts: Options{EnableUserLinks: true},
			want: "octocat",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderer := NewRenderer(tt.opts)
			if got := renderer.renderUser(tt.user); got != tt.want {
				t.Fatalf("renderUser() = %q, want %q", got, tt.want)
			}
		})
	}
}

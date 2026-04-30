package converter

import "github.com/stoneafk/issue2md/internal/model"

func (r Renderer) renderUser(user model.UserData) string {
	if !r.Options.EnableUserLinks || user.URL == "" {
		return user.Login
	}
	return "[" + user.Login + "](" + user.URL + ")"
}

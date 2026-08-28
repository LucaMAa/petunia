package dto

import "github.com/gin-gonic/gin"

type ListUsersParams struct {
	PageParams
	Search  string
	Enabled *bool
}

func ParseListUsersParams(c *gin.Context) ListUsersParams {
	p := ListUsersParams{PageParams: ParsePageParams(c)}
	p.Search = c.Query("search")
	if e := c.Query("enabled"); e != "" {
		v := e == "true"
		p.Enabled = &v
	}
	return p
}

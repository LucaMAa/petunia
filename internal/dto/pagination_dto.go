package dto

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

type PageParams struct {
	Limit  int
	Offset int
}

func ParsePageParams(c *gin.Context) PageParams {
	p := PageParams{Limit: 10, Offset: 0}
	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			p.Limit = v
		}
	}
	if o := c.Query("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			p.Offset = v
		}
	}
	return p
}

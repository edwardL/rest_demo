package orm

import (
	"strings"
)

// Common Table Expression

type Cte struct {
	withs []string
	as    []string
}

func WITH(tmpTable string) *Cte {
	cte := Cte{
		withs: []string{tmpTable},
	}
	return &cte
}

func (c *Cte) WITH(tmpTable string) *Cte {
	c.withs = append(c.withs, tmpTable)
	return c
}

func (c *Cte) AS(selectSQL string) *Cte {
	c.as = append(c.as, selectSQL)
	return c
}

func (c *Cte) SQL() string {
	var sb strings.Builder
	for i, with := range c.withs {
		sb.WriteString("WITH " + with + " AS (" + c.as[i] + ")")
		sb.WriteString("\n")
	}

	return sb.String()
}

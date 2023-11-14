package bot

import (
	"html/template"
	"strconv"
)

var Funcs = template.FuncMap{
	"plus": pl,
}

func pl(idx int) string {
	return strconv.Itoa(idx + 1)
}

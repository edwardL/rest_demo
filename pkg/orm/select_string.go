package orm

type SelectString struct {
	*selec[SelectString]
}

func SELECT(cols ...string) *SelectString {
	s := &SelectString{
		&selec[SelectString]{
			dbCommon: dbCommon{},
			cols:     cols,
		}}
	s.setT(s)
	return s
}

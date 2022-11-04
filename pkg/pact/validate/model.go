package validate

type ComparisonMap struct {
	Topics map[string]map[string]*Field
}

type Field struct {
	MatchingRule string
	PayloadValue interface{}
	Date         string
	Format       string
}

package routing

var (
	ConditionExpressionDivider = "::"
	ConditionExpressionAny     = "any"
	ConditionExpressionEqual   = "equal"
	ConditionExpressionPrefix  = "prefix"
)

func MatchAny() string {
	return ConditionExpressionEqual + ConditionExpressionDivider
}

func MatchEqual(value string) string {
	return ConditionExpressionEqual + ConditionExpressionDivider + value
}

func MatchPrefix(value string) string {
	return ConditionExpressionPrefix + ConditionExpressionDivider + value
}

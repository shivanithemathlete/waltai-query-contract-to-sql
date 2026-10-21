package internal

func BuildQueryPlan(
	contract Contract,
	deps Dependencies,
) QueryPlan {
	return QueryPlan{
		PrimaryModel: deps.PrimaryModel,
		Metrics:      contract.Metrics,
		Filters:      contract.Filters,
		GroupBy:      contract.GroupBy,
		Joins:        deps.Joins,
		Compare:      contract.Compare,
		Totals:       contract.Totals,
	}
}

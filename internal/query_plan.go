package internal

func BuildQueryPlan(
	contract Contract,
	deps Dependencies,
) QueryPlan {
	// default compare.primary to 2nd period, if not givev
	if contract.Compare != nil && contract.Compare.Primary == ""{
		contract.Compare.Primary = contract.Compare.Periods[0]
	}

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
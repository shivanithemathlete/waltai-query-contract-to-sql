package internal

import (
	"fmt"
	"slices"
	"strings"
)

func GenerateSQL(
	semModel SemanticModel,
	plan QueryPlan,
	dialect Dialect,
) (string, error) {
	if dialect != DuckDB && dialect != Postgres {
		return "", fmt.Errorf("unsupported dialect: %s", dialect)
	}

	var query strings.Builder

	// SELECT
	query.WriteString("SELECT\n")
	query.WriteString(generateSelectExpressions(semModel, plan, false))

	// FROM JOINS AND GLOBAL FILTER
	query.WriteString(generateFromWhere(plan))

	// GROUP BY
	if len(plan.GroupBy) > 0 {
		groupBy := identifyGroupBy(semModel, plan)
		query.WriteString("\nGROUP BY " + groupBy)
	}

	// GRAND TOTAL
	if plan.Totals == "grand" && len(plan.GroupBy) > 0 {
		var totalQuery strings.Builder

		totalQuery.WriteString("SELECT\n")
		totalQuery.WriteString(
			generateSelectExpressions(semModel, plan, true),
		)
		totalQuery.WriteString(generateFromWhere(plan))

		query.WriteString("\nUNION ALL\n")
		query.WriteString(totalQuery.String())
	}

	return query.String(), nil
}

func generateSelectExpressions(
	semModel SemanticModel,
	plan QueryPlan,
	grandTotal bool,
) string {
	var query strings.Builder

	// Dimensions
	for i, group := range plan.GroupBy {
		if i > 0 {
			query.WriteString(",\n")
		}

		if grandTotal {
			query.WriteString(fmt.Sprintf(
				"    'Total' AS %s", group,
			))
		} else {
			model := getDimensionModel(semModel.Dimensions, group)
			alias := modelAlias(model, plan)

			query.WriteString(fmt.Sprintf(
				"    %s.%s AS %s", alias, group, group,
			))
		}
	}

	if len(plan.GroupBy) > 0 && len(plan.Metrics) > 0 {
		query.WriteString(",\n")
	}

	// Comparison or regular metrics
	if plan.Compare != nil && len(plan.Compare.Periods) > 0 {
		query.WriteString(compareExpression(semModel, plan))
	} else {
		for i, metric := range plan.Metrics {
			if i > 0 {
				query.WriteString(",\n")
			}

			expression := generateMetricExpression(semModel, metric, plan)
			alias := metric.As
			if alias == "" {
				alias = metric.Name
			}

			query.WriteString(fmt.Sprintf(
				"    %s AS %s", expression, alias,
			))
		}
	}

	return query.String()
}

func generateFromWhere(plan QueryPlan) string {
	var query strings.Builder

	// FROM
	query.WriteString("\nFROM " + plan.PrimaryModel + " t0")

	// JOIN
	for i, join := range plan.Joins {
		alias := fmt.Sprintf("t%d", i+1)
		joinType := identifyJoinType(plan.Filters, join.RightModel)

		query.WriteString(fmt.Sprintf(
			"\n%s %s %s ON t0.%s = %s.%s",
			joinType, join.RightModel, alias,
			join.LeftColumn, alias, join.RightColumn,
		))
	}

	// WHERE
	for i, filter := range plan.Filters {
		alias := modelAlias(filter.Model, plan)

		if i == 0 {
			query.WriteString("\nWHERE ")
		} else {
			query.WriteString(" AND ")
		}

		query.WriteString(fmt.Sprintf(
			"%s.%s %s %s",
			alias, filter.Field, filter.Op, sqlLiteral(filter.Value),
		))
	}

	return query.String()
}

func generateMetricExpression(semModel SemanticModel, metric MetricRequest, plan QueryPlan) string {
	definition := getMetricDefinition(semModel, metric)

	var expression string
	switch definition.Agg {
	case "sum":
		expression = fmt.Sprintf("SUM(t0.%s)", definition.Expression)
	case "count_distinct":
		expression = fmt.Sprintf("COUNT(DISTINCT t0.%s)", definition.Expression)
	default:
		return ""
	}

	if len(metric.Filters) == 0 {
		return expression
	}

	var conditions []string
	for _, filter := range metric.Filters {
		alias := modelAlias(filter.Model, plan)
		conditions = append(conditions, fmt.Sprintf(
			"%s.%s %s %s",
			alias, filter.Field, filter.Op, sqlLiteral(filter.Value),
		))
	}

	switch definition.Agg {
	case "sum":
		return fmt.Sprintf(
			"SUM(CASE WHEN %s THEN t0.%s END)",
			strings.Join(conditions, " AND "), definition.Expression,
		)
	case "count_distinct":
		return fmt.Sprintf(
			"COUNT(DISTINCT CASE WHEN %s THEN t0.%s END)",
			strings.Join(conditions, " AND "), definition.Expression,
		)
	default:
		return ""
	}
}

// assuming we have exactly 2 periods
func compareExpression(semModel SemanticModel, plan QueryPlan) string {
	var expressions []string

	primaryPeriod := plan.Compare.Primary
	var baselinePeriod string

	for _, period := range plan.Compare.Periods {
		if period != primaryPeriod {
			baselinePeriod = period
			break
		}
	}

	if baselinePeriod == "" {
		return ""
	}

	for _, metric := range plan.Metrics {
		metricAlias := metric.As
		if metricAlias == "" {
			metricAlias = metric.Name
		}

		primaryExpr := generateCompareExpression(
			semModel, metric, primaryPeriod, plan,
		)
		baselineExpr := generateCompareExpression(
			semModel, metric, baselinePeriod, plan,
		)

		// Period values
		if slices.Contains(plan.Compare.Outputs, "values") {
			expressions = append(expressions,
				fmt.Sprintf("%s AS %s_%s", baselineExpr, metricAlias, baselinePeriod),
				fmt.Sprintf("%s AS %s_%s", primaryExpr, metricAlias, primaryPeriod),
			)
		}

		// Delta
		if slices.Contains(plan.Compare.Outputs, "delta") {
			expressions = append(expressions, fmt.Sprintf(
				"(%s - %s) AS %s_delta",
				primaryExpr, baselineExpr, metricAlias,
			))
		}

		// Percentage change
		if slices.Contains(plan.Compare.Outputs, "pct_change") {
			expressions = append(expressions, fmt.Sprintf(
				"(%s - %s) * 100.0 / NULLIF(%s, 0) AS %s_pct_change",
				primaryExpr, baselineExpr, baselineExpr, metricAlias,
			))
		}
	}

	return strings.Join(expressions, ",\n    ")
}

func generateCompareExpression(semModel SemanticModel, metric MetricRequest, period string, plan QueryPlan) string {
	definition := getMetricDefinition(semModel, metric)
	model := getDimensionModel(semModel.Dimensions, plan.Compare.Dimension)
	alias := modelAlias(model, plan)

	switch definition.Agg {
	case "sum":
		return fmt.Sprintf(
			"SUM(CASE WHEN %s.%s = %s THEN t0.%s END)",
			alias, plan.Compare.Dimension, sqlLiteral(period),
			definition.Expression,
		)
	case "count_distinct":
		return fmt.Sprintf(
			"COUNT(DISTINCT CASE WHEN %s.%s = %s THEN t0.%s END)",
			alias, plan.Compare.Dimension, sqlLiteral(period),
			definition.Expression,
		)
	default:
		return ""
	}
}

func getMetricDefinition(semModel SemanticModel, metric MetricRequest) MetricDefinition {
	var definition MetricDefinition

	for j := range semModel.Metrics {
		if semModel.Metrics[j].Name == metric.Name {
			definition = semModel.Metrics[j]
			break
		}
	}

	return definition
}

// these are global filters
func identifyJoinType(filters []Filter, joinModel string) string {
	var joinType = ""
	for _, filter := range filters {
		if filter.Model == joinModel {
			// global filter references the dimnesion
			joinType = "INNER JOIN"

			return joinType
		}
	}

	// might be for grouping or metric level calculation
	return "LEFT JOIN"
}

func identifyGroupBy(semModel SemanticModel, plan QueryPlan) string {
	groupBy := ""

	for i, group := range plan.GroupBy {
		if i > 0 {
			groupBy += ", "
		}
		model := getDimensionModel(semModel.Dimensions, group)
		alias := modelAlias(model, plan)
		groupBy += fmt.Sprintf("%s.%s", alias, group)
	}

	return groupBy
}

func getDimensionModel(dimensions []Dimension, dimension string) string {
	for _, dim := range dimensions {
		if dim.Name == dimension {
			return dim.Model
		}
	}

	return ""
}

func modelAlias(modelName string, plan QueryPlan) string {
	if modelName == plan.PrimaryModel {
		return "t0"
	}

	for i, join := range plan.Joins {
		if join.RightModel == modelName {
			return fmt.Sprintf("t%d", i+1)
		}
	}

	return ""
}

func sqlLiteral(value any) string {
	switch v := value.(type) {
	case string:
		return "'" + strings.ReplaceAll(v, "'", "''") + "'"
	case nil:
		return "NULL"
	default:
		return fmt.Sprint(v)
	}
}

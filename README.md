# Overview
This project compiles a typed JSON query contract into deterministic SQL using a semantic model.

# Out of Scope
- Metrics from multiple models in one query.
- Aggregations beyond the supported SUM and COUNT_DISTINCT.
- Transitive relationship traversal.
- Automatic database schema assumptions.

# Architecture
Semantic Model + Dialect + Contract -> Validation -> Dependency resolution -> Query plan -> SQL

Defined clear i/p o/p format for each level.

## Semantic model
The semantic model defines the vocabulary available to the compiler:
- Datasets and their grain.
- Relationships between datasets.
- Metrics, including aggregation, expression (or attribute), and source model.
- Dimensions and their source models.

If contract contains anything outside semantic model definition, it will considered as invalid.

## Contract Validation
This compiler takes below points into consideration for the provided contract.
- Metrics
    * at least one metric is required mandatorily in contract
    * validate each metric against metrics defined in semantic model

- Filter (both metric level and global)
    * field must be a valid dimension defined in semantic model
    * validate each model against model defined in semantic model
    * no validation logic for op and values in place

- Compare
    * dimension must be a valid dimension defined in semantic model
    * exactly two non-empty periods should be provided
    * if primary period is not provided, defaulting 2nd period to be primary
    * primary period must be one of the periods
    * at least one output must be provided
    * supporting only values, deltas and pct_change logic for compare outputs

- GroupBy
    * groupBy must be a valid dimension defined in semantic model
    * without group_by, the query returns one aggregate row.
    * with group_by, the query returns one aggregate row per group.
    
- Totals 
    * "grand" appends a grand-total row only when grouping is present.

## Dependency Resolution
- identify all models required by the contract, including those referenced by metrics, metric-level filters, global filters, group_by, and compare.dimension.
- **Primary Model:** The model associated with the first requested metric is selected as the anchor for SQL generation.
- check whether relationship between primary model and each referenced model exists
- only direct relationships are supported with respect to primary model. No transitive relationship is resolved
- error out if no valid relationship is found
- join direction is normalized so the primary model is represented as the left model in the generated join metadata.

### Output
ResolveDependencies returns a Dependencies object containing:

* PrimaryModel: the starting model, selected from the first requested metric.
* Models: the unique set of models required by the contract, including the primary model.
* Joins: the direct relationships connecting the primary model to each required model.
**Note:** If a required model has no direct relationship with the primary model, dependency resolution returns an error.

## Query Plan
This separates the meaning of the request from the mechanics of rendering SQL.
- combine the validated contract and resolved dependencies into a single QueryPlan.
- default compare.primary to the second period when it is not provided.
- output contains: primary model, requested metrics, filters, grouping dimensions, joins, comparison configuration, and totals configuration.

## SQL GENERATION
- consumes the output from query plan
- aliases and generated expressions use deterministic iteration order.
- metric-level filters are applied within conditional aggregate expressions.
- comparison values are calculated for each requested period.
- delta and percentage change are calculated from the period values.
- - Grand totals are computed by re-aggregating the rows instead of summing grouped results, to ensure correct totals for non-additive metrics like `COUNT_DISTINCT` (Contract C).
- JOIN Type
    * joined dimesion with a global filter - use *INNER JOIN* to eliminate unmatched rows
    * joined dimension without a global filter - use *LEFT JOIN* so unmatched fact rows are not dropped

## Error handling
Explicit failures for below scenarios:
- Unknown metric
- Unknown dimension
- Incorrect model for a filter
- Unsupported aggregation
- Unsupported dialect
- Invalid comparison configuration
- Empty metric list
- Multi model metrics
- Unsupported totals mode or other unsupported contract features

## Running

Run the compiler:

```bash
go run .
```

## Things to Investigate / Pending
Below things are not covered in current version. 
Will update as and when they are resolved.

- **Mixed-model metrics:** How to safely combine metrics from different fact tables without inflating aggregates.
- **Join fanout:** How relationship cardinality and duplicate dimension rows can cause incorrect aggregation.
- **Join-path resolution:** How to resolve transitive relationships and handle ambiguous join paths.
- **Metric-level filter validation:** How to validate filter references, model reachability, and join safety.
- **HAVING support:** How to filter results using aggregate conditions, distinct from row-level WHERE filters.
- **Time-series grouping:** How to support monthly trends, date bucketing, and missing periods.
- **Dialect differences:** How to handle database-specific syntax for dates, casts, quoting, and aggregation.
- **Type-aware SQL literals:** How to validate and render values according to column types instead of relying on implicit database coercion.
- **Scalability:** How the planner should evolve as metrics, dimensions, relationships, and supported operations grow significantly.
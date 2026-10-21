
metrics - check validity - metrics can have filters
filters - field, op, value, model
compare - 
    dimension - check validity [column]
    periods - values for comparison
    primary - operation order for comparison
    outputs - operation type [e.g. delta, %]
totals - grand


## Contract A

```
{
"metrics": [
    { "name": "total_revenue", "as": "online_rev",
    "filters": [ { "field": "channel", "op": "=", "value": "Online", "model": "fact_sales" } ] },
    { "name": "total_revenue" }
],
"group_by": [],
"filters": [
    { "field": "fiscal_year", "op": "=", "value": 2026, "model": "dim_calendar" }
]
}
```






metrics - total_revenue - model - fact_sales - with some filter for one
filter - fiscal_year - model - dim_calendar

two tables - some kind of join or IN/EXISTS
relationship in these two tables - store_id
filter on - fiscal_year

output needed - revenue - nothing from dim_cal table - IN/EXISTS - no join

now two metrices - additive - both on same column with different filters - sum with case
filter on dim_calendar as subquery


Contract → Validation → Dependency resolution → Query plan → SQL generation

clear i/p o/p format for each level.


## Validation
- check if metric exists in semanticModel.metrics - total_revenue - check - model - fact_sales
- metric level filter - channel - is a dimension - model - same as metrics - fact_sales
- global filter - fiscal_year - is a dimension - model - dim_calendar - different model


## Dependency Resolution
- any relationship between dim_cal and fact_sales - yes
- Join Condition: fact_sales.order_date = dim_calendar.date

## Query Plan

metrics
1. total_revenue: Sum(revenue)
2. online_rev: Sum(revenue where channel='Online')

Global Filter
1. fiscal_year=2026

## SQL generation

SELECT
SUM(t1revenue) as total_revenue,
SUM(t1revenue where t1channel='Online') as online_revenue
from fact_sales t1 
where
t1.order_date IN (select t2.date from dim_calendar t2 where t2.fiscal_year=2026)


so basically where has to be a dimension


package internal

type Dataset struct {
	Name  string `json:"name"`
	Grain string `json:"grain"`
}

type Relationship struct {
	From        string `json:"from"`
	FromColumn  string `json:"from_column"`
	To          string `json:"to"`
	ToColumn    string `json:"to_column"`
	Cardinality string `json:"cardinality"`
}

type MetricDefinition struct {
	Name         string `json:"name"`
	Agg          string `json:"agg"`
	Expression   string `json:"expression"`
	Model        string `json:"model"`
	MeasureClass string `json:"measure_class"`
}

type Dimension struct {
	Name  string `json:"name"`
	Model string `json:"model"`
}

type SemanticModel struct {
	Datasets      []Dataset          `json:"datasets"`
	Relationships []Relationship     `json:"relationships"`
	Metrics       []MetricDefinition `json:"metrics"`
	Dimensions    []Dimension        `json:"dimensions"`
}

type Filter struct {
	Field string `json:"field"`
	Op    string `json:"op"`
	Value any    `json:"value"`
	Model string `json:"model"`
}

type MetricRequest struct {
	Name    string   `json:"name"`
	As      string   `json:"as,omitempty"`
	Filters []Filter `json:"filters,omitempty"`
}

type CompareOutput string

const (
	OutputValues    CompareOutput = "values"
	OutputDelta     CompareOutput = "delta"
	OutputPctChange CompareOutput = "pct_change"
)

type Compare struct {
	Dimension string          `json:"dimension"`
	Periods   []string        `json:"periods"`
	Primary   string          `json:"primary"`
	Outputs   []CompareOutput `json:"outputs"`
}

type Contract struct {
	Metrics []MetricRequest `json:"metrics"`
	GroupBy []string        `json:"group_by"`
	Filters []Filter        `json:"filters"`
	Compare *Compare        `json:"compare,omitempty"`
	Totals  string          `json:"totals,omitempty"`
}

type Dialect string

const (
	DuckDB   Dialect = "duckdb"
	Postgres Dialect = "postgres"
)

type QueryPlan struct {
	PrimaryModel string
	Metrics      []MetricRequest
	Filters      []Filter
	GroupBy      []string
	Joins        []Join
	Compare      *Compare
	Totals       string
}

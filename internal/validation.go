package internal

import (
	"fmt"
	"slices"
	"strings"
)

func Validate(semModel SemanticModel, contract Contract) error {
	// Validate metrics
	if len(contract.Metrics) == 0 {
		return fmt.Errorf("at least one metric is required")
	}

	for _, requested := range contract.Metrics {
		var metric *MetricDefinition

		for i := range semModel.Metrics {
			if semModel.Metrics[i].Name == requested.Name {
				metric = &semModel.Metrics[i]
				break
			}
		}

		// input metrics need to be defined in the semantic model
		if metric == nil {
			return fmt.Errorf("unknown metric: %s", requested.Name)
		}

		// Validate metric-level filters
		for _, filter := range requested.Filters {
			if err := validateFilter(semModel.Dimensions, filter); err != nil {
				return err
			}
		}
	}

	// Validate global filters
	for _, filter := range contract.Filters {
		if err := validateFilter(semModel.Dimensions, filter); err != nil {
			return err
		}
	}

	// Validate compare
	if contract.Compare != nil {
		if err := validateCompare(*contract.Compare, semModel); err != nil {
			return err
		}
	}

	// Validate group_by
	for _, field := range contract.GroupBy {
		if err := validateFilter(semModel.Dimensions, Filter{
			Field: field,
		}); err != nil {
			return err
		}
	}

	return nil
}

func validateCompare(compare Compare, semModel SemanticModel) error {
	// Validate dimension
	if err := validateFilter(semModel.Dimensions, Filter{
		Field: compare.Dimension,
	}); err != nil {
		return err
	}

	// Validate periods
	if len(compare.Periods) != 2 {
		return fmt.Errorf("compare must contain exactly two periods")
	}

	// Validate period value: not accepting empty period
	for _, period := range compare.Periods {
		if strings.TrimSpace(period) == "" {
			return fmt.Errorf("compare periods must not be empty")
		}
	}

	// Validate primary period - if its present, it should be one of the two periods
	if compare.Primary != "" &&
		!slices.Contains(compare.Periods, compare.Primary) {
		return fmt.Errorf("compare primary must be one of the requested periods")
	}

	// Vali outputs: cannot be empty
	if len(compare.Outputs) == 0 {
		return fmt.Errorf("compare must contain at least one output")
	}

	// Validate output values
	for _, output := range compare.Outputs {
		if output != OutputValues &&
			output != OutputDelta &&
			output != OutputPctChange {
			return fmt.Errorf("unsupported compare output: %s", output)
		}
	}

	return nil
}

func validateFilter(dimensions []Dimension, filter Filter) error {
	for _, dimension := range dimensions {
		if dimension.Name == filter.Field {
			if filter.Model != "" && filter.Model != dimension.Model {
				return fmt.Errorf("incorrect model for filter %s", filter.Field)
			}
			return nil
		}
	}

	return fmt.Errorf("unknown dimension: %s", filter.Field)
}

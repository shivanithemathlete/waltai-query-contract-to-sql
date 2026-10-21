package internal

import (
	"fmt"
	"slices"
)

func Validate(model SemanticModel, contract Contract) error {
	// Validate metrics
	for _, requested := range contract.Metrics {
		var metric *MetricDefinition

		for i := range model.Metrics {
			if model.Metrics[i].Name == requested.Name {
				metric = &model.Metrics[i]
				break
			}
		}

		if metric == nil {
			return fmt.Errorf("unknown metric: %s", requested.Name)
		}

		// Validate metric-level filters
		for _, requested := range contract.Metrics {
			for _, filter := range requested.Filters {
				if err := validateFilter(model.Dimensions, filter); err != nil {
					return err
				}
			}
		}
	}

	// Validate global filters
	for _, filter := range contract.Filters {
		if err := validateFilter(model.Dimensions, filter); err != nil {
			return err
		}
	}

	// Validate compare
	if contract.Compare != nil {
		if err := validateCompare(*contract.Compare, model); err != nil {
			return err
		}
	}

	// Validate group_by
	for _, field := range contract.GroupBy {
		if err := validateFilter(model.Dimensions, Filter{
			Field: field,
		}); err != nil {
			return err
		}
	}

	return nil
}

func validateCompare(compare Compare, model SemanticModel) error {
	// Validate dimension
	if err := validateFilter(model.Dimensions, Filter{
		Field: compare.Dimension,
	}); err != nil {
		return err
	}

	// Validate primary period
	if compare.Primary == "" ||
		!slices.Contains(compare.Periods, compare.Primary) {
		return fmt.Errorf("compare primary must be one of the requested periods")
	}

	// Validate outputs
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

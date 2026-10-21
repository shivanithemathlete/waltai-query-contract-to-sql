package internal

import (
	"fmt"
	"slices"
)

type Join struct {
	LeftModel   string
	LeftColumn  string
	RightModel  string
	RightColumn string
}

type Dependencies struct {
	PrimaryModel string
	Models       []string
	Joins        []Join
}

func ResolveDependencies(
	semModel SemanticModel,
	contract Contract,
) (Dependencies, error) {
	// Find the primary model from the first metric.
	var primaryModel string
	for _, metric := range semModel.Metrics {
		if metric.Name == contract.Metrics[0].Name {
			primaryModel = metric.Model
			break
		}
	}

	deps := Dependencies{
		PrimaryModel: primaryModel,
		Models:       []string{primaryModel},
	}

	deps.Models = append(deps.Models, identifyReferencedModels(semModel, contract)...)

	joins, err := identifyJoins(semModel, deps)
	if err != nil {
		return Dependencies{}, err
	}

	deps.Joins = joins

	return deps, nil
}

func identifyJoins(semModel SemanticModel, deps Dependencies) ([]Join, error) {
	primaryModel := deps.PrimaryModel
	joins := make([]Join, 0)

	// Find relationships between the primary model and required models.
	for _, requiredModel := range deps.Models {
		if requiredModel == primaryModel {
			continue
		}

		relationFound := false

		for _, relationship := range semModel.Relationships {
			if relationship.From == primaryModel &&
				relationship.To == requiredModel {
				joins = append(joins, Join{
					LeftModel:   relationship.From,
					LeftColumn:  relationship.FromColumn,
					RightModel:  relationship.To,
					RightColumn: relationship.ToColumn,
				})

				relationFound = true
				break
			}

			if relationship.To == primaryModel &&
				relationship.From == requiredModel {
				joins = append(joins, Join{
					LeftModel:   relationship.To,
					LeftColumn:  relationship.ToColumn,
					RightModel:  relationship.From,
					RightColumn: relationship.FromColumn,
				})

				relationFound = true
				break
			}
		}

		if !relationFound {
			return nil, fmt.Errorf(
				"no relationship found between %s and %s",
				primaryModel, requiredModel,
			)
		}
	}

	return joins, nil
}

func identifyReferencedModels(semModel SemanticModel, contract Contract) []string {
	models := make([]string, 0)

	// Add models referenced by metric filters.
	for _, metric := range contract.Metrics {
		for _, filter := range metric.Filters {
			addMissingModels(semModel.Dimensions, &models, filter.Field)
		}

	}

	// Add models referenced by global filters.
	for _, filter := range contract.Filters {
		addMissingModels(semModel.Dimensions, &models, filter.Field)
	}

	// Add models referenced by group by.
	for _, groupBy := range contract.GroupBy {
		addMissingModels(semModel.Dimensions, &models, groupBy)
	}

	// Add models referenced by compare.
	if contract.Compare != nil {
		addMissingModels(semModel.Dimensions, &models, contract.Compare.Dimension)
	}

	return models
}

func addMissingModels(dims []Dimension, models *[]string, field string) {
	for _, dimension := range dims {
		if dimension.Name == field {
			if !slices.Contains(*models, dimension.Model) {
				*models = append(*models, dimension.Model)
			}
			break
		}
	}
}

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	walt "waltai-query-contract-to-sql/internal"
)

func main() {
	semModel, err := loadSemanticModel("semantic_model.json")
	if err != nil {
		panic(err)
	}

	// provide contract path here
	contract, err := loadContract("contracts/contract_c.json")
	if err != nil {
		panic(err)
	}

	if err := walt.Validate(semModel, contract); err != nil {
		panic(err)
	}

	deps, err := walt.ResolveDependencies(semModel, contract)
	if err != nil {
		panic(err)
	}

	plan := walt.BuildQueryPlan(contract, deps)

	start := time.Now()
	sql, err := walt.GenerateSQL(semModel, plan, walt.DuckDB)
	elapsed := time.Since(start)

	if err != nil {
		panic(err)
	}

	fmt.Println("Generated SQL:")
	fmt.Println(sql)
	fmt.Printf("\nGeneration time: %s\n", elapsed)
}

func loadSemanticModel(path string) (walt.SemanticModel, error) {
	var model walt.SemanticModel

	data, err := os.ReadFile(path)
	if err != nil {
		return model, err
	}

	err = json.Unmarshal(data, &model)
	return model, err
}

func loadContract(path string) (walt.Contract, error) {
	var contract walt.Contract

	data, err := os.ReadFile(path)
	if err != nil {
		return contract, err
	}

	err = json.Unmarshal(data, &contract)
	return contract, err
}

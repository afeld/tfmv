package main

import (
	"fmt"
	"os"

	"github.com/hashicorp/terraform/terraform"
)

func getPlan(file string) (*terraform.Plan, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return terraform.ReadPlan(f)
}

func getChangesByType(plan *terraform.Plan) (ChangesByType, error) {
	changesByType := ChangesByType{}

	// https://github.com/palantir/tfjson/blob/master/tfjson.go
	for _, module := range plan.Diff.Modules {
		for path, resource := range module.Resources {
			addr, err := terraform.ParseResourceAddress(path)
			if err != nil {
				return changesByType, err
			}
			rType := ResourceType(addr.Type)
			changesByType.Add(rType, resource)
		}
	}

	return changesByType, nil
}

func getMoveStatements(plan *terraform.Plan) ([]string, error) {
	// TODO implement
	return []string{}, nil
}

func main() {
	// TODO parameterize
	planfile := "tfplan"
	plan, err := getPlan(planfile)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println(plan.Diff.Modules, "\n----------\n")

	changesByType, err := getChangesByType(plan)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(changesByType)

	moves, err := getMoveStatements(plan)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(moves)
}

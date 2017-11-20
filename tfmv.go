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

func getChangesByType(plan *terraform.Plan) ChangesByType {
	changesByType := ChangesByType{}

	// https://github.com/palantir/tfjson/blob/master/tfjson.go
	for _, module := range plan.Diff.Modules {
		fmt.Println(module.Path)
		for pathStr, resource := range module.Resources {
			path := ResourcePath(pathStr)
			rType := path.GetType()
			changesByType.Add(rType, resource)
		}
	}

	return changesByType
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

	changesByType := getChangesByType(plan)
	fmt.Println(changesByType)

	moves, err := getMoveStatements(plan)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(moves)
}

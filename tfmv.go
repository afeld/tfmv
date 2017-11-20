package main

import (
	"fmt"
	"log"
	"os"

	tfmt "github.com/hashicorp/terraform/command/format"
	"github.com/hashicorp/terraform/terraform"
)

func getPlan(file string) (*tfmt.Plan, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Terraform has two Plan types, for some reason. `terraform.Plan` doesn't include the module in the address, so only use it for reading from the plan file, then convert it to the other one.
	plan, err := terraform.ReadPlan(f)
	if err != nil {
		return nil, err
	}
	fmtPlan := tfmt.NewPlan(plan)
	return fmtPlan, nil
}

func getChangesByType(plan *tfmt.Plan) (ChangesByType, error) {
	changesByType := ChangesByType{}

	// inspired by
	// https://github.com/palantir/tfjson/blob/master/tfjson.go
	for _, resource := range plan.Resources {
		// TODO refactor to not pass Addr separately
		diff := ResourceDiff{Addr: *resource.Addr, Diff: *resource}
		changesByType.Add(diff)
	}

	return changesByType, nil
}

func getMoveStatements(plan *tfmt.Plan) ([]string, error) {
	// TODO implement
	return []string{}, nil
}

func main() {
	// TODO parameterize
	planfile := "tfplan"
	plan, err := getPlan(planfile)
	if err != nil {
		log.Fatalln(err)
	}

	moves, err := getMoveStatements(plan)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(moves)
}

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

func main() {
	// TODO parameterize
	planfile := "tfplan"
	plan, err := getPlan(planfile)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println(plan.Diff.Modules, "\n----------\n")

	instancesByType := map[string]*ResourceChanges{}

	// https://github.com/palantir/tfjson/blob/master/tfjson.go
	for _, module := range plan.Diff.Modules {
		fmt.Println(module.Path)
		for rType, resource := range module.Resources {
			if instancesByType[rType] == nil {
				instancesByType[rType] = &ResourceChanges{}
			}
			changes := instancesByType[rType]
			changes.Add(resource)
		}
	}

	fmt.Println(instancesByType)
}

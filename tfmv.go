package main

import (
	"fmt"
	"log"
	"os"
	"reflect"

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

func checkIfObjectsMatch(name string, creation, deletion interface{}) error {
	if reflect.DeepEqual(creation, deletion) {
		err := fmt.Errorf(name+" match, which they shouldn't:\ncreation: %+v\ndeletion:%+v\n", creation, deletion)
		return err
	}
	return nil
}

func getMoveStatements(plan *tfmt.Plan) ([]string, error) {
	moves := []string{}

	changesByType, err := getChangesByType(plan)
	if err != nil {
		return moves, err
	}

	for _, changes := range changesByType {
		for i, creation := range changes.Created {
			// stop if we're out of matches
			if i == len(changes.Destroyed) {
				break
			}
			deletion := changes.Destroyed[i]

			// sanity checks
			if err := checkIfObjectsMatch("ResourceDiffs", creation, deletion); err != nil {
				return moves, err
			}
			if err := checkIfObjectsMatch("Addrs", creation.Addr, deletion.Addr); err != nil {
				return moves, err
			}
			if err := checkIfObjectsMatch("Diffs", creation.Diff, deletion.Diff); err != nil {
				return moves, err
			}
			if err := checkIfObjectsMatch("Strings", creation.String(), deletion.String()); err != nil {
				return moves, err
			}

			moves = append(moves, "terraform state mv "+deletion.String()+" "+creation.String())
		}
	}

	return moves, nil
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
	for _, move := range moves {
		fmt.Println(move)
	}
}

package main

import (
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/hashicorp/terraform/terraform"
	"github.com/stretchr/testify/assert"
)

// https://github.com/palantir/tfjson/blob/57123411e29c8945cd8dc89b6237c8f6f31ddf6e/tfjson_test.go#L124-L132
func mustRun(t *testing.T, name string, arg ...string) {
	if _, err := exec.Command(name, arg...).Output(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			t.Fatal(string(exitError.Stderr))
		} else {
			t.Fatal(err)
		}
	}
}

func testModulePath(t *testing.T, modulePath string) string {
	module, err := filepath.Abs(modulePath)
	assert.NoError(t, err)
	return module
}

func initModule(t *testing.T, modulePath string) {
	module := testModulePath(t, modulePath)
	mustRun(t, "terraform", "init", module)
}

func generatePlan(t *testing.T, modulePath string) string {
	planFile, err := ioutil.TempFile("", "terraform-plan")
	assert.NoError(t, err)
	err = planFile.Close()
	assert.NoError(t, err)

	module := testModulePath(t, modulePath)

	planPath := planFile.Name()
	mustRun(t, "terraform", "plan", "-out="+planPath, module)
	return planPath
}

func getTestPlan(t *testing.T, modulePath string) *terraform.Plan {
	initModule(t, modulePath)
	planPath := generatePlan(t, modulePath)
	plan, err := getPlan(planPath)
	assert.NoError(t, err)
	return plan
}

func TestMissingPlan(t *testing.T) {
	_, err := getPlan("missing-file")
	assert.Error(t, err)
}

func TestEmptyPlan(t *testing.T) {
	plan := getTestPlan(t, "test/empty")
	assert.True(t, plan.Diff.Empty())
	assert.Len(t, plan.Diff.Modules, 1)
}

func TestSimplePlan(t *testing.T) {
	plan := getTestPlan(t, "test/simple")
	assert.False(t, plan.Diff.Empty())
	assert.Len(t, plan.Diff.Modules, 1)
}

func TestChangesByType_Simple(t *testing.T) {
	plan := getTestPlan(t, "test/simple")

	changesByType, err := getChangesByType(plan)
	assert.NoError(t, err)

	types := changesByType.GetTypes()
	assert.Equal(t, types, []ResourceType{"local_file"})

	changes := changesByType.Get("local_file")
	assert.Len(t, changes.Created, 1)
	assert.Len(t, changes.Destroyed, 0)
}

func TestChangesByType_Multi(t *testing.T) {
	plan := getTestPlan(t, "test/multi")

	changesByType, err := getChangesByType(plan)
	assert.NoError(t, err)

	types := changesByType.GetTypes()
	assert.Len(t, types, 2)

	changes := changesByType.Get("local_file")
	assert.Len(t, changes.Created, 1)
	assert.Len(t, changes.Destroyed, 0)

	changes = changesByType.Get("tls_private_key")
	assert.Len(t, changes.Created, 1)
	assert.Len(t, changes.Destroyed, 0)
}

func TestChangesByType_ModuleRef(t *testing.T) {
	plan := getTestPlan(t, "test/module_ref")

	changesByType, err := getChangesByType(plan)
	assert.NoError(t, err)

	types := changesByType.GetTypes()
	assert.Equal(t, types, []ResourceType{"local_file"})

	changes := changesByType.Get("local_file")
	assert.Len(t, changes.Created, 1)
	assert.Len(t, changes.Destroyed, 0)
}

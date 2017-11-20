package main

import (
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"testing"

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

func testModulePath(t *testing.T) string {
	module, err := filepath.Abs("test")
	assert.Nil(t, err)
	return module
}

func initModule(t *testing.T) {
	module := testModulePath(t)
	mustRun(t, "terraform", "init", module)
}

func plan(t *testing.T) string {
	planFile, err := ioutil.TempFile("", "terraform-plan")
	assert.Nil(t, err)
	err = planFile.Close()
	assert.Nil(t, err)

	module := testModulePath(t)

	planPath := planFile.Name()
	mustRun(t, "terraform", "plan", "-out="+planPath, module)
	return planPath
}

func TestPlan(t *testing.T) {
	initModule(t)

	planPath := plan(t)
	_, err := getPlan(planPath)
	assert.Nil(t, err)
}

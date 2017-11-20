package main

import (
	tfmt "github.com/hashicorp/terraform/command/format"
	"github.com/hashicorp/terraform/terraform"
)

type ResourceType string

type ResourceChanges struct {
	Created   []tfmt.InstanceDiff
	Destroyed []tfmt.InstanceDiff
}

func (c *ResourceChanges) Add(diff tfmt.InstanceDiff) {
	switch diff.Action {
	case terraform.DiffCreate:
		c.Created = append(c.Created, diff)
	case terraform.DiffDestroy:
		c.Destroyed = append(c.Destroyed, diff)
	}
}

type ChangesByType map[ResourceType]*ResourceChanges

func (ct ChangesByType) Add(diff tfmt.InstanceDiff) {
	rType := ResourceType(diff.Addr.Type)
	if ct[rType] == nil {
		ct[rType] = &ResourceChanges{}
	}
	changes := ct[rType]
	changes.Add(diff)
}

func (ct ChangesByType) Get(rType ResourceType) *ResourceChanges {
	return ct[rType]
}

func (ct ChangesByType) GetTypes() []ResourceType {
	// https://stackoverflow.com/a/27848197/358804
	types := make([]ResourceType, len(ct))
	i := 0
	for rType := range ct {
		types[i] = rType
		i++
	}
	return types
}

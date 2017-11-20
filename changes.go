package main

import (
	tfmt "github.com/hashicorp/terraform/command/format"
	"github.com/hashicorp/terraform/terraform"
)

type ResourceType string

type ResourceDiff struct {
	Addr terraform.ResourceAddress
	Diff tfmt.InstanceDiff
}

func (r ResourceDiff) ChangeType() terraform.DiffChangeType {
	return r.Diff.Action
}

func (r ResourceDiff) Type() ResourceType {
	return ResourceType(r.Addr.Type)
}

func (r ResourceDiff) String() string {
	return r.Addr.String()
}

type ResourceChanges struct {
	Created   []ResourceDiff
	Destroyed []ResourceDiff
}

func (c *ResourceChanges) Add(diff ResourceDiff) {
	cType := diff.ChangeType()
	if cType == terraform.DiffCreate {
		c.Created = append(c.Created, diff)
	} else if cType == terraform.DiffDestroy {
		c.Destroyed = append(c.Destroyed, diff)
	}
}

type ChangesByType map[ResourceType]*ResourceChanges

func (ct ChangesByType) Add(diff ResourceDiff) {
	rType := diff.Type()
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

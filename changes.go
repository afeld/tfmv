package main

import "github.com/hashicorp/terraform/terraform"

type ResourceType string

type ResourceChanges struct {
	Created   []*terraform.InstanceDiff
	Destroyed []*terraform.InstanceDiff
}

func (c *ResourceChanges) Add(diff *terraform.InstanceDiff) {
	cType := diff.ChangeType()
	if cType == terraform.DiffCreate {
		c.Created = append(c.Created, diff)
	} else if cType == terraform.DiffDestroy {
		c.Destroyed = append(c.Destroyed, diff)
	}
}

type ChangesByType map[ResourceType]*ResourceChanges

func (ct ChangesByType) Add(rType ResourceType, resource *terraform.InstanceDiff) {
	if ct[rType] == nil {
		ct[rType] = &ResourceChanges{}
	}
	changes := ct[rType]
	changes.Add(resource)
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

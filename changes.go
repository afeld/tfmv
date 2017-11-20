package main

import "github.com/hashicorp/terraform/terraform"

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

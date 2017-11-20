# tfmv

"Terraform Move", a tool to help with Terraform refactoring. The use case: Let's say you have a resource:

```hcl
// terraform/main.tf
resource "aws_instance" "app" {
  // ...
}
```

If you change the resource name:

```hcl
resource "aws_instance" "webapp" {
  // ...
}
```

or move it into a module:

```hcl
// terraform/main.tf --> terraform/shared/main.tf
resource "aws_instance" "app" {
  // ...
}
```

the [resource address](https://www.terraform.io/docs/internals/resource-addressing.html) is different, and Terraform treats it as the old `resource` being deleted, and the new one being created. Why not reuse them? [`terraform state mv`](https://www.terraform.io/docs/commands/state/mv.html) allows you to update the addresses of resources by hand, but this can be tedious if you are moving more than a few resources.

### Goals

In order:

1. Improve algorithm for resource matching.
    * Currently it's just matching created resources with destroyed ones of the same type, in the order it comes across them.
1. Gain enough confidence in its functionality that it can be used in deployment pipelines, where it's hard to do `state mv` by hand.
1. Propose merging into Terraform core.

## Setup

1. [Install Go 1.6+.](https://golang.org/doc/install)
1. Install the package.

    ```sh
    go get github.com/afeld/tfmv
    ```

## Usage

1. Create a plan.

    ```sh
    cd <your module>
    terraform plan -out=tfplan
    ```

1. Run the executable. It will compute `state mv` commands to efficiently reuse resources.

    ```
    $ tfmv
    terraform state mv aws_instance.my_instance module.shared.aws_instance.my_instance
    ...
    ```

1. After double-checking the output, you can run the commands to avoid deleting and recreating resources unnecessarily.

## See also

* [Google Groups discussion](https://groups.google.com/forum/#!topic/terraform-tool/CE2ScmDBTIE)
* [GitHub issue about resource equivalence maps](https://github.com/hashicorp/terraform/issues/9048)

## Development

1. [Install dep.](https://github.com/golang/dep#setup)
1. Install the dependencies.

    ```sh
    dep ensure
    ```

1. Run tests.

    ```sh
    go test
    ```

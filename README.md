# Terraform Provider for Azure Resource Manager Rest API

* [Terraform Website](https://www.terraform.io)
* [AzApi Provider Documentation](https://registry.terraform.io/providers/Azure/azapi/latest/docs)
* [AzApi Provider Usage Examples](https://github.com/Azure/terraform-provider-azapi/tree/main/examples)

## Usage Example

The following example shows how to use `azapi_resource` to manage machine learning compute resource.

```hcl
terraform {
  required_providers {
    azapi = {
      source  = "Azure/azapi"
    }
  }
}

provider "azapi" {
  # More information on the authentication methods supported by
  # the AzApi Provider can be found here:
  # https://registry.terraform.io/providers/Azure/azapi/latest/docs

  # subscription_id = "..."
  # client_id       = "..."
  # client_secret   = "..."
  # tenant_id       = "..."
}

// azurerm provider
provider "azurerm" {
  features {}
}

// create a resource group, it's recommended to use azapi provider with azurerm provider
resource "azurerm_resource_group" "example" {
  name     = "example-resources"
  location = "West Europe"
}

// create a automation account
resource "azapi_resource" "automationAccount" {
  type      = "Microsoft.Automation/automationAccounts@2021-06-22"
  name      = "myAccount"
  parent_id = azurerm_resource_group.test.id

  location = azurerm_resource_group.test.location
  body = jsonencode({
    properties = {
      disableLocalAuth    = true
      publicNetworkAccess = false
      sku = {
        name = "Basic"
      }
    }
  })
}

```


Further [usage documentation is available on the Terraform website](https://registry.terraform.io/providers/Azure/azapi/latest/docs).



## Troubleshooting

Please check the [troubleshooting doc](https://github.com/Azure/terraform-provider-azapi/blob/main/documents/troubleshooting.md).


## Developer Requirements

* [Terraform](https://www.terraform.io/downloads.html) version 0.12.x + (but 1.x is recommended)
* [Go](https://golang.org/doc/install) version 1.17.x (to build the provider plugin)

### On Windows

If you're on Windows you'll also need:
* [Git Bash for Windows](https://git-scm.com/download/win)
* [Make for Windows](http://gnuwin32.sourceforge.net/packages/make.htm)

For *GNU32 Make*, make sure its bin path is added to PATH environment variable.*

For *Git Bash for Windows*, at the step of "Adjusting your PATH environment", please choose "Use Git and optional Unix tools from Windows Command Prompt".*

Or install via [Chocolatey](https://chocolatey.org/install) (`Git Bash for Windows` must be installed per steps above)

```powershell
choco install make golang terraform -y
refreshenv
```

You must run `Developing the Provider` commands in `bash` because `sh` scrips are invoked as part of these.

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (version 1.17+ is **required**). You'll also need to correctly setup a [GOPATH](http://golang.org/doc/code.html#GOPATH), as well as adding `$GOPATH/bin` to your `$PATH`.

First clone the repository to: `$GOPATH/src/github.com/Azure/terraform-provider-azapi`

```sh
$ mkdir -p $GOPATH/src/github.com/Azure; cd $GOPATH/src/github.com/Azure
$ git clone git@github.com:Azure/terraform-provider-azapi
$ cd $GOPATH/src/github.com/Azure/terraform-provider-azapi
```

Once inside the provider directory, you can run `make tools` to install the dependent tooling required to compile the provider.

At this point you can compile the provider by running `make build`, which will build the provider and put the provider binary in the `$GOPATH/bin` directory.

```sh
$ make build
...
$ $GOPATH/bin/terraform-provider-azapi
...
```

You can also cross-compile if necessary:

```sh
GOOS=windows GOARCH=amd64 make build
```

In order to run the `Unit Tests` for the provider, you can run:

```sh
$ make test
```

The majority of tests in the provider are `Acceptance Tests` - which provisions real resources in Azure. It's possible to run the entire acceptance test suite by running `make testacc` - however it's likely you'll want to run a subset, which you can do using a prefix, by running:

```sh
make acctests TESTARGS='-run=<nameOfTheTest>' TESTTIMEOUT='60m'
```

* `<nameOfTheTest>` should be self-explanatory as it is the name of the test you want to run. An example could be `TestAccGenericResource_basic`. Since `-run` can be used with regular expressions you can use it to specify multiple tests like in `TestAccGenericResource_` to run all tests that match that expression

The following Environment Variables must be set in your shell prior to running acceptance tests:

- `ARM_CLIENT_ID`
- `ARM_CLIENT_SECRET`
- `ARM_SUBSCRIPTION_ID`
- `ARM_TENANT_ID`
- `ARM_ENVIRONMENT`
- `ARM_METADATA_HOST`
- `ARM_TEST_LOCATION`
- `ARM_TEST_LOCATION_ALT`
- `ARM_TEST_LOCATION_ALT2`

**Note:** Acceptance tests create real resources in Azure which often cost money to run.

---

## Developer: Using the locally compiled Azure Provider binary

When using Terraform 0.14 and later, after successfully compiling the Azure Provider, you must [instruct Terraform to use your locally compiled provider binary](https://www.terraform.io/docs/commands/cli-config.html#development-overrides-for-provider-developers) instead of the official binary from the Terraform Registry.

For example, add the following to `~/.terraformrc` for a provider binary located in `/home/developer/go/bin`:

```hcl
provider_installation {

  # Use /home/developer/go/bin as an overridden package directory
  # for the Azure/azapi provider. This disables the version and checksum
  # verifications for this provider and forces Terraform to look for the
  # azapi provider plugin in the given directory.
  dev_overrides {
    "Azure/azapi" = "/home/developer/go/bin"
  }

  # For all other providers, install them directly from their origin provider
  # registries as normal. If you omit this, Terraform will _only_ use
  # the dev_overrides block, and so no other providers will be available.
  direct {}
}
```

## Other docs
1. How is the resource's id built? Please check [this](https://github.com/Azure/terraform-provider-azapi/blob/main/documents/resource_id.md).

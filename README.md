# tfc - Terraform Cloud CLI Helper

[![CI](https://github.com/zkhvan/tfc/actions/workflows/ci.yaml/badge.svg)](https://github.com/zkhvan/tfc/actions/workflows/ci.yaml)
[![codecov](https://codecov.io/gh/zkhvan/tfc/graph/badge.svg?token=NM83Z3DHKO)](https://codecov.io/gh/zkhvan/tfc)

`tfc` is a command-line tool that streamlines interactions with Terraform
Enterprise and HashiCorp Cloud Platform (HCP) Terraform. It provides
convenient shortcuts and workflows for common Terraform Cloud operations.

## Features

- List workspaces and workspace variables
- List organizations

## Installation

### Using Go

```bash
go install github.com/zkhvan/tfc@latest
```

### Using Homebrew

```bash
brew install zkhvan/homebrew-tap/tfc
```

### Binary releases

Download the latest binary for your platform from the [releases
page](https://github.com/zkhvan/tfc/releases).

### From source

```bash
git clone https://github.com/zkhvan/tfc.git
cd tfc
make build
```

## Configuration

`tfc` looks for configuration in the following locations (in order of precedence):

1. Environment variables

### Environment variables

- `TFE_TOKEN`: Your Terraform Enterprise API token
- `TFE_ADDRESS`: Terraform Enterprise address (defaults to https://$TFE_HOSTNAME)
- `TFE_HOSTNAME`: Terraform Enterprise host (defaults to app.terraform.io)

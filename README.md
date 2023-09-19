# ORY Keto Terraform Provider

Terraform provider for [ORY Keto](https://github.com/ory/keto).
It allows for managing ORY Keto relationship resources using Terraform.

## Requirements
- ORY Keto 0.11.0 server or newer, versions before haven't been tested.

## Getting Started

```hcl
terraform {
  required_providers {
    oryketo = {
      source  = "76creates/oryketo"
      version = "0.0.1"
    }
  }
}

provider "oryketo" {
  write {
    url = "http://localhost:5467"
  }
  read {
    url = "http://localhost:5466"
  }
}

resource "oryketo_relationship" "this" {
  namespace  = "default"
  object     = "www"
  relation   = "read"
  subject_id = "guest"
}
```
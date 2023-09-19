# Ory Keto Provider

[Ory Keto](https://github.com/ory/keto) is authorization system that implements [Google Zanzibar](https://research.google/pubs/pub48190/).
This provider interacts with Keto API to manage relationships tuples.

## Example Usage

```hcl
provider "oryketo" {
  write {
    url = "http://localhost:4467"
  }
  read {
    url = "http://localhost:4466"
  }
}

resource "oryketo_relationship" "this" {
  namespace  = "default"
  object     = "provider"
  relation   = "user"
  subject_id = "your-name-here"
}
```

## Requirements
- Ory Keto 0.11.0 or newer, versions before haven't been tested.

## Argument Reference

* `read` (required) - Holds configuration for the read-only Keto API.
* `write` (required) - Holds configuration for the write(admin) Keto API.

The `read` block supports:

* `url` - (Required) URL for Keto read-only API. Defaults to `ORY_KETO_READ_URL` environment variable.
* `headers` - (Optional) Map of headers to add to all requests that use read API.

The `write` block supports:

* `url` - (Required) URL for Keto write(admin) API. Defaults to `ORY_KETO_WRITE_URL` environment variable.
* `headers` - (Optional) Map of headers to add to all requests that use write API.

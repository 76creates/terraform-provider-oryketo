# Data Source: oryketo_relationship_parse

Parse a Google Zanzibar relationship text notation into relationship objects, and Ory Keto JSON format.

## Example Usage

### Parse multiple text notation relationships

```hcl
data "oryketo_relationship_parse" "this" {
  from_string = <<-EOF
default:app#read@user/foo
default:app#read@guest
default:app#write@default:role/admin#member
EOF
}

locals {
  data = data.oryketo_relationship_parse.this.relation_tuple
}

resource "oryketo_relationship" "multiple" {
  count                 = length(local.data)
  namespace             = local.data[count.index].namespace
  object                = local.data[count.index].object
  relation              = local.data[count.index].relation
  subject_id            = lookup(local.data[count.index], "subject_id", null)
  subject_set_namespace = lookup(local.data[count.index], "subject_set_namespace", null)
  subject_set_object    = lookup(local.data[count.index], "subject_set_object", null)
  subject_set_relation  = lookup(local.data[count.index], "subject_set_relation", null)
}
```

## Argument Reference

* `from_string` (required) - Google Zanzibar relationship text notation to parse, can take multiple lines.

## Attributes Reference

* `relation_tuple` - List of relationship objects.
* `json` - Ory Keto schema JSON representation of the relationship objects.
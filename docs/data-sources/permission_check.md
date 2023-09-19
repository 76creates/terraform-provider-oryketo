# Data Source: oryketo_permission_check

Evaluate whether a subject has a permission to perform an action on an object.

## Example Usage

### Create simple structure and perform a check

```hcl
data "oryketo_relationship_parse" "this" {
  from_string = <<-EOF
default:app#write@default:role/admin#member
default:role/admin#member@foo
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

data "oryketo_permission_check" "should_allow" {
  depends_on = [
    oryketo_relationship.multiple
  ]
  namespace  = "default"
  object     = "app"
  relation   = "write"
  subject_id = "foo"
}

output "should_allow_result" {
  value = data.oryketo_permission_check.should_allow.allowed
}
```

## Argument Reference

* `namespace` (required) - Namespace of the relationship tuple.
* `object` (required) - Object of the relationship tuple.
* `relation` (required) - Relation of the relationship tuple.
* `subject_id` (optional) - Subject ID of the relationship tuple.
* `subject_set_namespace` (optional) - Subject Set Namespace of the relationship tuple.
* `subject_set_object` (optional) - Subject Set Object of the relationship tuple.
* `subject_set_relation` (optional) - Subject Set Relation of the relationship tuple.

~> NOTE: Either `subject_id` or `subject_set_*` group must be defined.

## Attributes Reference

* `allowed` - Boolean value indicating whether the subject has the permission to perform the action on the object.
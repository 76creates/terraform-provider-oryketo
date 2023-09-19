# Resource: oryketo_relationship

Manages relationship tuples in Ory Keto trough the [API](https://www.ory.sh/docs/keto/reference/rest-api#tag/relationship/operation/createRelationship)

## Example Usage

```hcl
resource "oryketo_relationship" "write" {
  namespace             = "default"
  object                = "app"
  relation              = "write"
  subject_set_namespace = "default"
  subject_set_object    = "role/admin"
  subject_set_relation  = "member"
}

resource "oryketo_relationship" "read" {
  namespace  = "default"
  object     = "app"
  relation   = "read"
  subject_id = "guest"
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

## Import
A Ory Keto relationship resource can be imported using its Google Zanzibar text notation, which is also used as a resource ID, e.g.
```shell
# as from the example above
$ terraform import oryketo_relationship.write 'default:app#write@default:role/admin#member'
$ terraform import oryketo_relationship.read 'default:app#read@guest'
```

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

# example modeled from https://github.com/ory/keto/tree/master/contrib/cat-videos-example
data "oryketo_relationship_parse" "cat_videos_relationships" {
  from_string = <<-EOF
videos:/cats/1.mp4#owner@videos:/cats#owner
videos:/cats/1.mp4#view@videos:/cats/1.mp4#owner
videos:/cats/1.mp4#view@*
videos:/cats/2.mp4#owner@videos:/cats#owner
videos:/cats/2.mp4#view@videos:/cats#owner
videos:/cats#owner@cat lady
videos:/cats#view@videos:/cats#owner
EOF
}

locals {
  data = data.oryketo_relationship_parse.cat_videos_relationships.relation_tuple
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
  namespace  = "videos"
  object     = "/cats/1.mp4"
  relation   = "view"
  subject_id = "*"
}

output "should_allow_result" {
  value = data.oryketo_permission_check.should_allow.allowed
}
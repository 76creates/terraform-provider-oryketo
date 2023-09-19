package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/ory/keto/ketoapi"
	hash "github.com/theTardigrade/golang-hash"
)

func dataKetoRelationshipParse() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataKetoRelationshipParseRead,
		Schema: map[string]*schema.Schema{
			"from_string": {
				Type:     schema.TypeString,
				Required: true,
			},
			"relation_tuple": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"namespace": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"object": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"relation": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"subject_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"subject_set_namespace": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"subject_set_object": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"subject_set_relation": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"json": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func dataKetoRelationshipParseRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var relationshipTuples []*ketoapi.RelationTuple
	fromString := d.Get("from_string").(string)
	for _, relString := range strings.Split(fromString, "\n") {
		cleanRelString := strings.TrimSpace(relString)
		if cleanRelString == "" {
			continue
		}
		rt, err := stringToRelationTuple(cleanRelString)
		if err != nil {
			return diag.FromErr(err)
		}
		relationshipTuples = append(relationshipTuples, rt)
	}

	jsonValue, err := flattenRelationTupleToJsonList(relationshipTuples)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("json", jsonValue); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("relation_tuple", flattenRelationTuple(relationshipTuples)); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%x", hash.UintString(fromString)))
	return nil
}

func stringToRelationTuple(s string) (*ketoapi.RelationTuple, error) {
	return (&ketoapi.RelationTuple{}).FromString(s)
}

func flattenRelationTuple(rt []*ketoapi.RelationTuple) []interface{} {
	flatten := make([]interface{}, len(rt))
	for i, rt := range rt {
		m := map[string]interface{}{
			"namespace": rt.Namespace,
			"object":    rt.Object,
			"relation":  rt.Relation,
		}
		if rt.SubjectID != nil {
			m["subject_id"] = *rt.SubjectID
		} else {
			m["subject_set_namespace"] = rt.SubjectSet.Namespace
			m["subject_set_object"] = rt.SubjectSet.Object
			m["subject_set_relation"] = rt.SubjectSet.Relation
		}

		flatten[i] = m
	}
	return flatten
}

func flattenRelationTupleToJsonList(rt []*ketoapi.RelationTuple) ([]string, error) {
	flatten := make([]string, len(rt))
	for i, rt := range rt {
		b, err := json.Marshal(rt)
		if err != nil {
			return nil, err
		}
		flatten[i] = string(b)
	}
	return flatten, nil
}

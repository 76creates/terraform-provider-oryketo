package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	hash "github.com/theTardigrade/golang-hash"
)

func dataKetoPermissionCheck() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataKetoPermissionCheckRead,
		Schema: map[string]*schema.Schema{
			"namespace": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"object": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"relation": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"subject_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"subject_set_namespace": {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
			},
			"subject_set_object": {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
			},
			"subject_set_relation": {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
			},
			"allowed": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func dataKetoPermissionCheckRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	provider := m.(*providerConfig)

	if err := validateSchemaRelationTuple(d, ""); err != nil {
		return diag.FromErr(err)
	}

	rel, err := getClientRelationship(d)
	if err != nil {
		return diag.FromErr(err)
	}

	relJson, err := json.Marshal(rel)
	if err != nil {
		return diag.FromErr(err)
	}

	request := provider.readApiClient.PermissionApi.
		CheckPermission(ctx).
		Namespace(rel.Namespace).
		Object(rel.Object).
		Relation(rel.Relation)

	if rel.SubjectId != nil {
		request = request.SubjectId(*rel.SubjectId)
	} else {
		request = request.SubjectSetNamespace(rel.SubjectSet.Namespace).
			SubjectSetObject(rel.SubjectSet.Object).
			SubjectSetRelation(rel.SubjectSet.Relation)
	}

	result, resp, err := request.Execute()
	if err != nil {
		return diag.FromErr(err)
	}
	if resp.StatusCode != 200 {
		return diag.FromErr(err)
	}

	if err := d.Set("allowed", result.GetAllowed()); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%x", hash.UintString(string(relJson))))
	return nil
}

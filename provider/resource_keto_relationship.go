package provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	ketoClient "github.com/ory/keto-client-go"
	"github.com/ory/keto/ketoapi"
)

func resourceKetoRelationship() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKetoRelationshipCreate,
		ReadContext:   resourceKetoRelationshipRead,
		DeleteContext: resourceKetoRelationshipDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceKetoRelationshipImport,
		},
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
		},
	}
}

func resourceKetoRelationshipImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	provider := m.(*providerConfig)

	id := d.Id()
	rt, err := stringToRelationTuple(id)
	if err != nil {
		return nil, fmt.Errorf("malformed id: %s", err)
	}
	rel := ketoRelationTupleToRelationship(rt)

	existingRelationships, err := getRelationshipsForTuple(ctx, provider, &rel)
	if err != nil {
		return nil, err
	}
	if len(existingRelationships) == 0 {
		d.SetId("")
		return nil, fmt.Errorf("relationship '%s' not found, data race suspected", ketoRelationshipToRelationTuple(rel).String())
	}
	relationship := existingRelationships[0]

	if err = d.Set("namespace", relationship.Namespace); err != nil {
		return nil, err
	}
	if err = d.Set("object", relationship.Object); err != nil {
		return nil, err
	}
	if err = d.Set("relation", relationship.Relation); err != nil {
		return nil, err
	}
	if relationship.SubjectId != nil {
		if err = d.Set("subject_id", *relationship.SubjectId); err != nil {
			return nil, err
		}
	} else if relationship.SubjectSet != nil {
		if err = d.Set("subject_set_namespace", relationship.SubjectSet.Namespace); err != nil {
			return nil, err
		}
		if err = d.Set("subject_set_object", relationship.SubjectSet.Object); err != nil {
			return nil, err
		}
		if err = d.Set("subject_set_relation", relationship.SubjectSet.Relation); err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New("subject_id or subject_set must be set")
	}

	return schema.ImportStatePassthroughContext(ctx, d, m)
}

func resourceKetoRelationshipCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	provider := m.(*providerConfig)

	if err := validateSchemaRelationTuple(d, ""); err != nil {
		return diag.FromErr(err)
	}

	rel, err := getClientRelationship(d)
	if err != nil {
		return diag.FromErr(err)
	}

	existingRelation, err := getRelationshipsForTuple(ctx, provider, &rel)
	if err != nil {
		return diag.FromErr(err)
	}
	if len(existingRelation) == 1 {
		return diag.FromErr(fmt.Errorf("relationship '%s' already exists", ketoRelationshipToRelationTuple(rel).String()))
	}

	body := ketoClient.CreateRelationshipBody{
		Namespace:  &rel.Namespace,
		Object:     &rel.Object,
		Relation:   &rel.Relation,
		SubjectId:  rel.SubjectId,
		SubjectSet: rel.SubjectSet,
	}

	// Todo: validate returned relationship
	_, resp, err := provider.writeApiClient.RelationshipApi.
		CreateRelationship(nil).
		CreateRelationshipBody(body).
		Execute()
	if err != nil {
		return diag.FromErr(err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			diag.FromErr(err)
		}
	}()

	if resp.StatusCode != 201 {
		return diag.Errorf("unexpected status code: %s", resp.Status)
	}
	return resourceKetoRelationshipRead(ctx, d, m)
}

func resourceKetoRelationshipDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	provider := m.(*providerConfig)

	if err := validateSchemaRelationTuple(d, ""); err != nil {
		return diag.FromErr(err)
	}

	rel, err := getClientRelationship(d)
	if err != nil {
		return diag.FromErr(err)
	}
	request := provider.writeApiClient.RelationshipApi.
		DeleteRelationships(ctx).
		Namespace(rel.Namespace).
		Object(rel.Object).
		Relation(rel.Relation)

	if rel.SubjectId != nil {
		request = request.SubjectId(*rel.SubjectId)
	} else if rel.SubjectSet != nil {
		request = request.
			SubjectSetNamespace(rel.SubjectSet.Namespace).
			SubjectSetObject(rel.SubjectSet.Object).
			SubjectSetRelation(rel.SubjectSet.Relation)
	} else {
		return diag.Errorf("subject_id or subject_set must be set")
	}

	resp, err := request.Execute()
	if err != nil {
		return diag.FromErr(err)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			diag.FromErr(err)
		}
	}()

	if resp.StatusCode != 204 {
		return diag.Errorf("unexpected status code: %s", resp.Status)
	}

	return nil
}

func resourceKetoRelationshipRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var rel ketoClient.Relationship
	var err error

	provider := m.(*providerConfig)

	if err := validateSchemaRelationTuple(d, ""); err != nil {
		return diag.FromErr(err)
	}
	rel, err = getClientRelationship(d)
	if err != nil {
		return diag.FromErr(err)
	}

	existingRelationships, err := getRelationshipsForTuple(ctx, provider, &rel)
	if err != nil {
		return diag.FromErr(err)
	}
	if len(existingRelationships) == 0 {
		d.SetId("")
		return nil
	}

	relationship := existingRelationships[0]
	setRelationshipId(d, &relationship)
	return nil
}

func setRelationshipId(d *schema.ResourceData, rel *ketoClient.Relationship) {
	d.SetId(ketoRelationshipToRelationTuple(*rel).String())
}

func getRelationshipsForTuple(ctx context.Context, provider *providerConfig, rel *ketoClient.Relationship) ([]ketoClient.Relationship, error) {
	request := provider.readApiClient.RelationshipApi.
		GetRelationships(ctx).
		Namespace(rel.Namespace).
		Object(rel.Object).
		Relation(rel.Relation)
	if rel.SubjectId != nil {
		request = request.SubjectId(*rel.SubjectId)
	} else if rel.SubjectSet != nil {
		request = request.
			SubjectSetNamespace(rel.SubjectSet.Namespace).
			SubjectSetObject(rel.SubjectSet.Object).
			SubjectSetRelation(rel.SubjectSet.Relation)
	} else {
		return nil, errors.New("subject_id or subject_set must be set")
	}

	readData, resp, err := request.
		PageSize(100). // just in case
		Execute()
	if err != nil {
		if resp != nil && resp.StatusCode == 404 {
			return nil, nil
		}
		return nil, err
	}
	tflog.Debug(ctx, fmt.Sprintf("read %d tuples", len(readData.RelationTuples)), nil)

	deduplicatedRelationships := deduplicateRelationTuple(readData.RelationTuples)
	tflog.Debug(ctx, fmt.Sprintf("deduplicated tuples %d", len(deduplicatedRelationships)), nil)

	if len(deduplicatedRelationships) > 1 {
		return nil, errors.New("multiple relationships found")
	}

	return deduplicatedRelationships, nil
}

func validateSchemaRelationTuple(d *schema.ResourceData, parentKey string) error {
	if parentKey != "" {
		if parentKey[len(parentKey)-1] != '.' {
			parentKey += "."
		}
	}
	_, namespaceOk := d.GetOk(parentKey + "namespace")
	_, objectOk := d.GetOk(parentKey + "object")
	_, relationOk := d.GetOk(parentKey + "relation")
	_, subjectIdOk := d.GetOk("subject_id")
	_, subjectSetNamespaceOk := d.GetOk("subject_set_namespace")
	_, subjectSetObjectOk := d.GetOk("subject_set_object")
	_, subjectSetRelationOk := d.GetOk("subject_set_relation")
	subjectSetOk := subjectSetNamespaceOk && subjectSetObjectOk && subjectSetRelationOk
	subjectSetOkAtLeastOne := subjectSetNamespaceOk || subjectSetObjectOk || subjectSetRelationOk
	if !namespaceOk || !objectOk || !relationOk {
		return errors.New("namespace, object and relation must be set")
	}
	if subjectIdOk && subjectSetOkAtLeastOne {
		return errors.New("only one of subject_id and subject_set group can be set")
	} else if !subjectIdOk && !subjectSetOk {
		if subjectSetOkAtLeastOne {
			return errors.New("subject_set_namespace, subject_set_object, and subject_set_relation must be defined together")
		}
		return errors.New("one of subject_id and subject_set group must be set")
	}
	return nil
}

func getClientRelationship(d *schema.ResourceData) (ketoClient.Relationship, error) {
	var relationship ketoClient.Relationship
	relationship.Namespace = d.Get("namespace").(string)
	relationship.Object = d.Get("object").(string)
	relationship.Relation = d.Get("relation").(string)
	if subjectIdRaw, ok := d.GetOk("subject_id"); ok {
		subjectId := subjectIdRaw.(string)
		relationship.SubjectId = &subjectId
	} else {
		subjectSet := ketoClient.SubjectSet{
			Namespace: d.Get("subject_set_namespace").(string),
			Object:    d.Get("subject_set_object").(string),
			Relation:  d.Get("subject_set_relation").(string),
		}
		relationship.SubjectSet = &subjectSet
	}
	return relationship, nil
}

func deduplicateRelationTuple(relationships []ketoClient.Relationship) []ketoClient.Relationship {
	keys := make(map[string]bool)
	var list []ketoClient.Relationship
	for _, entry := range relationships {
		key := ketoRelationshipToRelationTuple(entry).String()
		if entry.SubjectSet != nil {
			key += entry.SubjectSet.Namespace + entry.SubjectSet.Object + entry.SubjectSet.Relation
		}
		if _, value := keys[key]; !value {
			keys[key] = true
			list = append(list, entry)
		}
	}
	return list
}

func ketoRelationshipToRelationTuple(d ketoClient.Relationship) *ketoapi.RelationTuple {
	relationTuple := ketoapi.RelationTuple{
		Namespace:  d.Namespace,
		Object:     d.Object,
		Relation:   d.Relation,
		SubjectID:  d.SubjectId,
		SubjectSet: nil,
	}
	if d.SubjectSet != nil {
		relationTuple.SubjectSet = &ketoapi.SubjectSet{
			Namespace: d.SubjectSet.Namespace,
			Object:    d.SubjectSet.Object,
			Relation:  d.SubjectSet.Relation,
		}
	}
	return &relationTuple
}

func ketoRelationTupleToRelationship(d *ketoapi.RelationTuple) ketoClient.Relationship {
	relationship := ketoClient.Relationship{
		Namespace: d.Namespace,
		Object:    d.Object,
		Relation:  d.Relation,
		SubjectId: d.SubjectID,
	}
	if d.SubjectSet != nil {
		relationship.SubjectSet = &ketoClient.SubjectSet{
			Namespace: d.SubjectSet.Namespace,
			Object:    d.SubjectSet.Object,
			Relation:  d.SubjectSet.Relation,
		}
	}
	return relationship
}

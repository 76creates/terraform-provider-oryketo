package provider

import (
	"context"
	"net/url"

	"github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	keto "github.com/ory/keto-client-go"
)

type providerConfig struct {
	readApiClient  *keto.APIClient
	writeApiClient *keto.APIClient
}

func Provider(ctx context.Context) *schema.Provider {
	provider := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"read": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"url": {
							Type:        schema.TypeString,
							Required:    true,
							DefaultFunc: schema.EnvDefaultFunc("ORY_KETO_READ_URL", nil),
						},
						"headers": {
							Type:     schema.TypeMap,
							Optional: true,
						},
					},
				},
			},
			"write": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"url": {
							Type:        schema.TypeString,
							Required:    true,
							DefaultFunc: schema.EnvDefaultFunc("ORY_KETO_WRITE_URL", nil),
						},
						"headers": {
							Type:     schema.TypeMap,
							Optional: true,
						},
					},
				},
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"oryketo_relationship": resourceKetoRelationship(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"oryketo_relationship_parse": dataKetoRelationshipParse(),
			"oryketo_permission_check":   dataKetoPermissionCheck(),
		},
		ConfigureContextFunc: configureProvider,
	}
	return provider
}

func configureProvider(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	readObject := d.Get("read").([]interface{})[0].(map[string]interface{})
	readUrl, err := url.Parse(readObject["url"].(string))
	if err != nil {
		return nil, diag.Errorf("parse keto url: %v", err)
	}

	confReadHeaders := readObject["headers"].(map[string]interface{})
	readHeaders := make(map[string]string)
	for k, v := range confReadHeaders {
		readHeaders[k] = v.(string)
	}

	httpReadClient := cleanhttp.DefaultClient()
	readClientConfig := keto.NewConfiguration()
	readClientConfig.Host = readUrl.Host
	readClientConfig.Scheme = readUrl.Scheme
	readClientConfig.DefaultHeader = readHeaders
	readClientConfig.UserAgent = "terraform/ory-keto-provider"
	readClientConfig.Debug = false
	readClientConfig.HTTPClient = httpReadClient
	readApiClient := keto.NewAPIClient(readClientConfig)

	writeObject := d.Get("write").([]interface{})[0].(map[string]interface{})
	writeUrl, err := url.Parse(writeObject["url"].(string))
	if err != nil {
		return nil, diag.Errorf("parse keto url: %v", err)
	}

	confWriteHeaders := writeObject["headers"].(map[string]interface{})
	writeHeaders := make(map[string]string)
	for k, v := range confWriteHeaders {
		writeHeaders[k] = v.(string)
	}

	httpWriteClient := cleanhttp.DefaultClient()
	writeClientConfig := keto.NewConfiguration()
	writeClientConfig.Host = writeUrl.Host
	writeClientConfig.Scheme = writeUrl.Scheme
	writeClientConfig.DefaultHeader = writeHeaders
	writeClientConfig.UserAgent = "terraform/ory-keto-provider"
	writeClientConfig.Debug = false
	writeClientConfig.HTTPClient = httpWriteClient
	writeApiClient := keto.NewAPIClient(writeClientConfig)

	return &providerConfig{
		readApiClient:  readApiClient,
		writeApiClient: writeApiClient,
	}, diag.Diagnostics{}
}

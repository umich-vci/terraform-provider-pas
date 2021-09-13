package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/umich-vci/gopas"
)

func init() {
	// Set descriptions to support markdown syntax, this will be used in document generation
	// and the language server.
	schema.DescriptionKind = schema.StringMarkdown

	// Customize the content of descriptions when output. For example you can add defaults on
	// to the exported descriptions if present.
	schema.SchemaDescriptionBuilder = func(s *schema.Schema) string {
		desc := s.Description
		if s.Default != nil {
			desc += fmt.Sprintf(" Defaults to `%v`.", s.Default)
		}
		return strings.TrimSpace(desc)
	}
}

func New(version string) func() *schema.Provider {
	return func() *schema.Provider {
		p := &schema.Provider{
			Schema: map[string]*schema.Schema{
				"username": {
					Type:        schema.TypeString,
					Required:    true,
					DefaultFunc: schema.EnvDefaultFunc("PAS_USERNAME", nil),
					Description: "This is the username to use to access the Red Hat Satellite server. This must be provided in the config or in the environment variable `PAS_USERNAME`.",
				},
				"password": {
					Type:        schema.TypeString,
					Required:    true,
					Sensitive:   true,
					DefaultFunc: schema.EnvDefaultFunc("PAS_PASSWORD", nil),
					Description: "This is the password to use to access the Red Hat Satellite server. This must be provided in the config or in the environment variable `PAS_PASSWORD`.",
				},
				"pas_host": {
					Type:        schema.TypeString,
					Required:    true,
					DefaultFunc: schema.EnvDefaultFunc("PAS_HOST", nil),
					Description: "This is the hostname or IP address of the CyberArk PAS server. This must be provided in the config or in the environment variable `PAS_HOST`.",
				},
				"auth_type": {
					Type:        schema.TypeString,
					Required:    true,
					DefaultFunc: schema.EnvDefaultFunc("PAS_AUTH_TYPE", nil),
					Description: "This is the authentication type to use with the CyberArk PAS server. This must be provided in the config or in the environment variable `PAS_AUTH_TYPE`.",
				},
			},
			DataSourcesMap: map[string]*schema.Resource{
				"scaffolding_data_source": dataSourceScaffolding(),
			},
			ResourcesMap: map[string]*schema.Resource{
				"scaffolding_resource": resourceScaffolding(),
			},
		}

		p.ConfigureContextFunc = configure(version, p)

		return p
	}
}

type apiClient struct {
	Client gopas.APIClient
}

func configure(version string, p *schema.Provider) func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) {
	return func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		userAgent := p.UserAgent("terraform-provider-pas", version)

		username := d.Get("username").(string)
		password := d.Get("password").(string)
		authType := d.Get("auth_type").(string)
		host := d.Get("pas_host").(string)
		concurrent := true

		config := gopas.NewConfiguration()
		config.UserAgent = userAgent
		config.Host = host

		data := *gopas.NewLogonData()
		data.UserName = &username
		data.Password = &password
		data.ConcurrentSession = &concurrent

		client := gopas.NewAPIClient(config)
		client.AuthApi.AuthLogon(ctx, authType).Data(data).Execute()

		return &apiClient{Client: *client}, nil
	}
}

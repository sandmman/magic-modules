package google

import (
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceGkeonpremBareMetalVersions() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceGkeonpremBareMetalVersionsRead,

		Schema: map[string]*schema.Schema{
			"create_config": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"admin_cluster_membership": {
							Optional:     true,
							Type:         schema.TypeString,
							ExactlyOneOf: []string{"create_config.0.admin_cluster_membership", "create_config.0.admin_cluster_name"},
						},
						"admin_cluster_name": {
							Optional:     true,
							Type:         schema.TypeString,
							ExactlyOneOf: []string{"create_config.0.admin_cluster_membership", "create_config.0.admin_cluster_name"},
						},
					},
				},
			},
			"valid_versions": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceGkeonpremBareMetalVersionsRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	userAgent, err := generateUserAgentString(d, config.userAgent)
	if err != nil {
		return err
	}

	project, err := getProject(d, config)
	if err != nil {
		return err
	}

	location, err := getLocation(d, config)
	if err != nil {
		return err
	}
	if len(location) == 0 {
		return fmt.Errorf("Cannot determine location: set location in this data source or at provider-level")
	}

	obj := make(map[string]interface{})
	if v, ok := d.GetOkExists("parent"); ok {
		obj["parent"] = v
	}

	if v, ok := d.GetOkExists("create_config"); ok {
		l := v.([]interface{})
		raw := l[0]
		config := raw.(map[string]interface{})

		createConfig := make(map[string]interface{})

		if v, ok := config["admin_cluster_membership"]; ok {
			createConfig["admin_cluster_membership"] = v
		}

		if v, ok := config["admin_cluster_name"]; ok {
			createConfig["admin_cluster_name"] = v
		}

		obj["create_config"] = createConfig
	}

	url, err := replaceVars(d, config, "{{GkeonpremBasePath}}projects/{{project}}/locations/{{location}}/bareMetalClusters:queryVersionConfig")
	if err != nil {
		return err
	}

	res, err := sendRequest(config, "POST", project, url, userAgent, obj)
	if err != nil {
		return err
	}

	if err := d.Set("versions", res["versions"]); err != nil {
		return err
	}

	var validVersions []string
	for _, v := range res["validVersions"].([]interface{}) {
		vm := v.(map[string]interface{})
		validVersions = append(validVersions, vm["version"].(string))
	}
	if err := d.Set("valid_versions", validVersions); err != nil {
		return err
	}

	d.SetId(time.Now().UTC().String())

	return nil
}

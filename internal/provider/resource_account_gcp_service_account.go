package provider

import (
	"context"
	"io"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/umich-vci/gopas"
)

func resourceAccountGCPServiceAccount() *schema.Resource {
	return &schema.Resource{
		Description: "Resource to manage a GCP Service Account in CyberArk PAS",

		CreateContext: resourceAccountGCPServiceAccountCreate,
		ReadContext:   resourceAccountGCPServiceAccountRead,
		UpdateContext: resourceAccountGCPServiceAccountUpdate,
		DeleteContext: resourceAccountGCPServiceAccountDelete,

		Schema: map[string]*schema.Schema{
			"safe_name": {
				Description: "The name of the safe to create the GCP service account in.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"username": {
				Description: "The e-mail of the GCP service account.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"impersonate_user": {
				Description: "The name of the user with user management permissions that the plugin uses for connecting and managing account passwords for the GCP Account Management plugin.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"name": {
				Description: "The name of the account.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"populate_key": {
				Description: "Indicates whether to populate the key if it doesn't exist on reconcile.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"category_modification_time": {
				Description: "",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"created_time": {
				Description: "",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"key_id": {
				Description: "The ID of the GCP key",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func resourceAccountGCPServiceAccountCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*apiClient).Client

	platformID := "GCPServiceAccount"
	safeName := d.Get("safe_name").(string)

	account := *gopas.NewAccountModel(platformID, safeName)

	username := d.Get("username").(string)
	account.UserName = &username

	if n, ok := d.GetOk("name"); ok {
		name := n.(string)
		account.Name = &name
	}

	prop := make(map[string]string)
	prop["KeyID"] = "TerraformCreatedAccount"

	if i, ok := d.GetOk("impersonate_user"); ok {
		prop["ImpersonateUser"] = i.(string)
	}

	if k, ok := d.GetOk("key"); ok {
		secret := k.(string)
		account.Secret = &secret
	}

	if p, ok := d.GetOk("populate_key"); ok {
		populate := "No"
		if p.(bool) {
			populate = "Yes"
		}
		prop["PopulateKey"] = populate
	}

	account.PlatformAccountProperties = &prop

	act, resp, err := client.AccountsApi.AccountsAddAccount(ctx).Account(account).Execute()
	if err != nil {
		var diags diag.Diagnostics
		diags = append(diags, diag.FromErr(err)...)

		b, err := io.ReadAll(resp.Body)
		if err != nil {
			diags = append(diags, diag.FromErr(err)...)
		}

		diags = append(diags, diag.Errorf(string(b))...)
		return diags
	}

	d.SetId(act["id"].(string))

	return resourceAccountGCPServiceAccountRead(ctx, d, meta)
}

func resourceAccountGCPServiceAccountRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*apiClient).Client

	id := d.Id()

	account, resp, err := client.AccountsApi.AccountsGetAccount(ctx, id).Execute()
	if err != nil {
		if resp.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.Set("key_id", (*account.PlatformAccountProperties)["KeyID"])
	d.Set("populate_key", (*account.PlatformAccountProperties)["PopulateKey"])
	d.Set("impersonate_user", (*account.PlatformAccountProperties)["ImpersonateUser"])
	d.Set("safe_name", account.SafeName)
	d.Set("username", account.UserName)
	d.Set("password", account.Secret)
	d.Set("name", account.Name)
	d.Set("category_modification_time", account.CategoryModificationTime)
	d.Set("created_time", account.CreatedTime)

	return nil
}

func resourceAccountGCPServiceAccountUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*apiClient).Client

	id := d.Id()

	accountPatch := *gopas.NewJsonPatchDocumentAccountModel()
	operations := []gopas.OperationAccountModel{}
	op := "replace"

	rootProp := make(map[string]interface{})

	if d.HasChange("safe_name") {
		rootProp["safeName"] = d.Get("safe_name")
	}

	if d.HasChange("name") {
		rootProp["name"] = d.Get("name")
	}

	if d.HasChange("username") {
		rootProp["userName"] = d.Get("name")
	}

	if d.HasChange("key") {
		rootProp["secret"] = d.Get("key")
	}

	if len(rootProp) > 0 {
		path := "/"
		operation := *gopas.NewOperationAccountModel()
		operation.Op = &op
		operation.Path = &path
		operation.Value = &rootProp
		operations = append(operations, operation)
	}

	prop := make(map[string]interface{})

	if d.HasChange("key_id") {
		prop["KeyID"] = d.Get("key_id")
	}

	if d.HasChange("impersonate_user") {
		prop["ImpersonateUser"] = d.Get("impersonate_user")
	}

	if d.HasChange("populate_key") {
		populate := "No"
		if d.Get("populate_key").(bool) {
			populate = "Yes"
		}
		prop["PopulateKey"] = populate
	}

	if len(prop) > 0 {
		path := "/platformAccountProperties"
		operation := *gopas.NewOperationAccountModel()
		operation.Op = &op
		operation.Path = &path
		operation.Value = &prop
		operations = append(operations, operation)
	}

	accountPatch.Operations = &operations

	_, _, err := client.AccountsApi.AccountsUpdateAccount(ctx, id).AccountPatch(accountPatch).Execute()
	if err != nil {
		diag.FromErr(err)
	}

	return resourceAccountGCPServiceAccountRead(ctx, d, meta)
}

func resourceAccountGCPServiceAccountDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*apiClient).Client

	id := d.Id()

	_, _, err := client.AccountsApi.AccountsDeleteAccount(ctx, id).Execute()
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return nil
}

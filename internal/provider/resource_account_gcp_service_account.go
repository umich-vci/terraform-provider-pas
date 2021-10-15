package provider

import (
	"context"
	"io"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/umich-vci/gopas"
)

func resourceAccountGCPServiceAccount() *schema.Resource {
	return &schema.Resource{
		Description: "Resource to manage a GCP Service Account in CyberArk PAS",

		CreateContext: resourceAccountGCPServiceAccountCreate,
		ReadContext:   resourceAccountGCPServiceAccountRead,
		UpdateContext: resourceAccountGCPServiceAccountUpdate,
		DeleteContext: resourceAccountGCPServiceAccountDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"safe_name": {
				Description:  "The name of the safe to create the GCP service account in.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"username": {
				Description:  "The e-mail of the GCP service account.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"change_account": {
				Description: "The account to use as the change account.",
				Type:        schema.TypeSet,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"safe_name": {
							Description:  "The name of the safe that contains the account to use as the change account.",
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},
						"name": {
							Description:  "The name of the account to use as the change account.",
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},
						"folder": {
							Description:  "The folder the change account is located in.",
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "Root",
							ValidateFunc: validation.StringIsNotEmpty,
						},
					},
				},
				Optional: true,
				MaxItems: 1,
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
			"platform_id": {
				Description: "The Platform ID to use for the GCP Service Account.",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "GCPServiceAccount",
			},
			"populate_key": {
				Description: "Indicates whether to populate the key if it doesn't exist on reconcile.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"reconcile_account": {
				Description: "The account to use as the reconcile account.",
				Type:        schema.TypeSet,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"safe_name": {
							Description:  "The name of the safe that contains the account to use as the change account.",
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},
						"name": {
							Description:  "The name of the account to use as the change account.",
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},
						"folder": {
							Description:  "The folder the change account is located in.",
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "Root",
							ValidateFunc: validation.StringIsNotEmpty,
						},
					},
				},
				Optional: true,
				MaxItems: 1,
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

	platformID := d.Get("platform_id").(string)
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

	id := act["id"].(string)
	d.SetId(id)

	if c, ok := d.GetOk("change_account"); ok {
		changeAccount := c.(*schema.Set).List()[0].(map[string]interface{})
		safeName := changeAccount["safe_name"].(string)
		name := changeAccount["name"].(string)
		folder := changeAccount["folder"].(string)

		linkAccount := *gopas.NewLinkAccountData(name, safeName, folder, 2)

		_, err := client.AccountsApi.AccountsLinkAccount(ctx, id).LinkAccount(linkAccount).Execute()
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if c, ok := d.GetOk("reconcile_account"); ok {
		reconcileAccount := c.(*schema.Set).List()[0].(map[string]interface{})
		safeName := reconcileAccount["safe_name"].(string)
		name := reconcileAccount["name"].(string)
		folder := reconcileAccount["folder"].(string)

		linkAccount := *gopas.NewLinkAccountData(name, safeName, folder, 3)

		_, err := client.AccountsApi.AccountsLinkAccount(ctx, id).LinkAccount(linkAccount).Execute()
		if err != nil {
			return diag.FromErr(err)
		}
	}

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

	accountv1, _, err := client.AccountsApi.AccountsGetAccountLegacy(ctx).Keywords(*account.UserName).Safe(account.SafeName).Execute()
	if err != nil {
		var diags diag.Diagnostics
		diags = append(diags, diag.Errorf("Error finding account name %s with legacy API", *account.Name)...)
		diags = append(diags, diag.FromErr(err)...)
		return diags
	}

	if *accountv1.Count == 0 {
		return diag.Errorf("No accounts found with legacy API")
	}

	changeAccount := make(map[string]interface{})
	reconcileAccount := make(map[string]interface{})

	for _, kv := range *(*accountv1.Accounts)[0].InternalProperties {
		switch *kv.Key {
		case "ExtraPass2Name":
			changeAccount["name"] = *kv.Value
		case "ExtraPass2Folder":
			changeAccount["folder"] = *kv.Value
		case "ExtraPass2Safe":
			changeAccount["safe_name"] = *kv.Value
		case "ExtraPass3Name":
			reconcileAccount["name"] = *kv.Value
		case "ExtraPass3Folder":
			reconcileAccount["folder"] = *kv.Value
		case "ExtraPass3Safe":
			reconcileAccount["safe_name"] = *kv.Value
		}
	}

	if len(changeAccount) > 0 {
		d.Set("change_account", []map[string]interface{}{changeAccount})
	} else {
		d.Set("change_account", nil)
	}

	if len(reconcileAccount) > 0 {
		d.Set("reconcile_account", []map[string]interface{}{reconcileAccount})
	} else {
		d.Set("reconcile_account", nil)
	}

	if (*account.PlatformAccountProperties)["PopulateKey"] == "Yes" {
		d.Set("populate_key", true)
	} else if (*account.PlatformAccountProperties)["PopulateKey"] == "No" {
		d.Set("populate_key", false)
	} else {
		d.Set("populate_key", nil)
	}

	d.Set("key_id", (*account.PlatformAccountProperties)["KeyID"])
	d.Set("impersonate_user", (*account.PlatformAccountProperties)["ImpersonateUser"])
	d.Set("safe_name", account.SafeName)
	d.Set("username", account.UserName)
	d.Set("name", account.Name)
	d.Set("category_modification_time", account.CategoryModificationTime)
	d.Set("created_time", account.CreatedTime)
	d.Set("platform_id", account.PlatformId)

	return nil
}

func resourceAccountGCPServiceAccountUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*apiClient).Client

	id := d.Id()

	operations := []gopas.OperationAccountModel{}
	opReplace := "replace"
	opRemove := "remove"
	opAdd := "add"

	rootProp := make(map[string]interface{})

	if d.HasChange("safe_name") {
		_, n := d.GetChange("safe_name")
		rootProp["safeName"] = n
	}

	if d.HasChange("name") {
		_, n := d.GetChange("name")
		rootProp["name"] = n
	}

	if d.HasChange("username") {
		_, n := d.GetChange("username")
		rootProp["userName"] = n
	}

	if len(rootProp) > 0 {
		path := "/"
		operation := *gopas.NewOperationAccountModel()
		operation.Op = &opReplace
		operation.Path = &path
		operation.Value = &rootProp
		operations = append(operations, operation)
	}

	prop := make(map[string]interface{})
	path := "/platformAccountProperties"

	if d.HasChange("impersonate_user") {
		o, n := d.GetChange("impersonate_user")
		if o == nil {
			operation := gopas.OperationAccountModel{
				Op:    &opAdd,
				Path:  &path,
				Value: &map[string]interface{}{"ImpersonateUser": n},
			}
			operations = append(operations, operation)
		} else if n == nil {
			operation := gopas.OperationAccountModel{
				Op:    &opRemove,
				Path:  &path,
				Value: &map[string]interface{}{"ImpersonateUser": ""},
			}
			operations = append(operations, operation)
		} else {
			prop["ImpersonateUser"] = n
		}

	}

	if d.HasChange("populate_key") {
		_, n := d.GetChange("populate_key")
		populate := "No"
		if n == nil {
			populate = ""
		} else if n.(bool) {
			populate = "Yes"
		}
		prop["PopulateKey"] = populate
	}

	if len(prop) > 0 {
		path := "/platformAccountProperties"
		operation := *gopas.NewOperationAccountModel()
		operation.Op = &opReplace
		operation.Path = &path
		operation.Value = &prop
		operations = append(operations, operation)
	}

	if len(operations) > 0 {
		patch := *gopas.NewJsonPatchDocumentAccountModel()
		patch.SetOperations(operations)

		_, resp, err := client.AccountsApi.AccountsUpdateAccount(ctx, id).AccountPatch(patch).Execute()
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
	}

	if d.HasChange("change_account") {
		_, n := d.GetChange("change_account")
		if n == nil {
			// delete not implemented yet
		} else {
			changeAccount := n.(*schema.Set).List()[0].(map[string]interface{})
			safeName := changeAccount["safe_name"].(string)
			name := changeAccount["name"].(string)
			folder := changeAccount["folder"].(string)
			linkAccount := *gopas.NewLinkAccountData(name, safeName, folder, 2)

			resp, err := client.AccountsApi.AccountsLinkAccount(ctx, id).LinkAccount(linkAccount).Execute()
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
		}
	}

	if d.HasChange("reconcile_account") {
		_, n := d.GetChange("reconcile_account")
		if n == nil {
			// delete not implemented yet
		} else {
			reconcileAccount := n.(*schema.Set).List()[0].(map[string]interface{})
			safeName := reconcileAccount["safe_name"].(string)
			name := reconcileAccount["name"].(string)
			folder := reconcileAccount["folder"].(string)
			linkAccount := *gopas.NewLinkAccountData(name, safeName, folder, 3)

			resp, err := client.AccountsApi.AccountsLinkAccount(ctx, id).LinkAccount(linkAccount).Execute()
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
		}
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

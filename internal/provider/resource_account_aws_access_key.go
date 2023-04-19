package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/umich-vci/gopas"
)

func resourceAccountAWSAccessKey() *schema.Resource {
	return &schema.Resource{
		Description: "Resource to manage AWS IAM User credentials in CyberArk PAS",

		CreateContext: resourceAccountAWSAccessKeyCreate,
		ReadContext:   resourceAccountAWSAccessKeyRead,
		UpdateContext: resourceAccountAWSAccessKeyUpdate,
		DeleteContext: resourceAccountAWSAccessKeyDelete,

		Schema: map[string]*schema.Schema{
			"aws_account_id": {
				Description: "The account ID on the AWS console. This is a 12-digit number such as 123456789012.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"safe_name": {
				Description: "The name of the safe to create the AWS IAM account in.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"aws_access_key_id": {
				Description: "The unique ID of the Amazon Web Services (AWS) access key that is used by APIs to access the AWS console.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"password": {
				Description: "The password of the IAM user.",
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
			},
			"aws_arn_role": {
				Description: "The role that can securely access the AWS console.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"aws_policy": {
				Description: "The policy that enables access to the AWS console for the specified user.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"aws_account_alias_name": {
				Description: "A friendly identifier of your AWS account ID that can be used for your sign-in page to contain your company name, instead of your AWS account ID.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"name": {
				Description: "The name of the account. If not specified, one is generated.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"category_modification_time": {
				Description: "TODO",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"created_time": {
				Description: "When the account was created",
				Type:        schema.TypeInt,
				Computed:    true,
			},
		},
	}
}

func resourceAccountAWSAccessKeyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*apiClient).Client

	platformID := "AWS"
	safeName := d.Get("safe_name").(string)

	account := *gopas.NewAccountModel(platformID, safeName)

	username := d.Get("username").(string)
	account.UserName = &username

	address := d.Get("address").(string)
	account.Address = &address

	if n, ok := d.GetOk("name"); ok {
		name := n.(string)
		account.Name = &name
	}

	if p, ok := d.GetOk("password"); ok {
		password := p.(string)
		account.Secret = &password
	}

	prop := make(map[string]string)
	prop["AWSAccountID"] = d.Get("aws_account_id").(string)

	if arn, ok := d.GetOk("aws_arn_role"); ok {
		prop["AWSARNRole"] = arn.(string)
	}
	if policy, ok := d.GetOk("aws_policy"); ok {
		prop["AWSPolicy"] = policy.(string)
	}
	if alias, ok := d.GetOk("aws_account_alias_name"); ok {
		prop["AWSAccountAliasName"] = alias.(string)
	}
	account.PlatformAccountProperties = &prop

	act, resp, err := client.AccountsApi.AccountsAddAccount(ctx).Account(account).Execute()
	if err != nil {
		return returnResponseErr(resp, err)
	}

	d.SetId(act["id"].(string))

	return resourceAccountAWSAccessKeyRead(ctx, d, meta)
}

func resourceAccountAWSAccessKeyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*apiClient).Client

	id := d.Id()

	account, resp, err := client.AccountsApi.AccountsGetAccount(ctx, id).Execute()
	if err != nil {
		if resp.StatusCode == 404 {
			d.SetId("")
			return nil
		}

		return returnResponseErr(resp, err)
	}

	d.Set("address", account.Address)
	d.Set("aws_account_alias_name", (*account.PlatformAccountProperties)["AWSAccountAliasName"])
	d.Set("aws_account_id", (*account.PlatformAccountProperties)["AWSAccountID"])
	d.Set("aws_arn_role", (*account.PlatformAccountProperties)["AWSARNRole"])
	d.Set("aws_policy", (*account.PlatformAccountProperties)["AWSPolicy"])
	d.Set("category_modification_time", account.CategoryModificationTime)
	d.Set("created_time", account.CreatedTime)
	d.Set("name", account.Name)
	d.Set("password", account.Secret)
	d.Set("safe_name", account.SafeName)
	d.Set("username", account.UserName)

	return nil
}

func resourceAccountAWSAccessKeyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*apiClient).Client

	id := d.Id()

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

	if d.HasChange("password") {
		rootProp["secret"] = d.Get("password")
	}

	if d.HasChange("address") {
		rootProp["address"] = d.Get("address")
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

	if d.HasChange("aws_account_id") {
		prop["AWSAccountID"] = d.Get("aws_account_id")
	}

	if d.HasChange("aws_arn_role") {
		prop["AWSARNRole"] = d.Get("aws_arn_role")
	}

	if d.HasChange("aws_policy") {
		prop["AWSPolicy"] = d.Get("aws_policy")
	}

	if d.HasChange("aws_account_alias_name") {
		prop["AWSAccountAliasName"] = d.Get("aws_account_alias_name")
	}

	if len(prop) > 0 {
		path := "/platformAccountProperties"
		operation := *gopas.NewOperationAccountModel()
		operation.Op = &op
		operation.Path = &path
		operation.Value = &prop
		operations = append(operations, operation)
	}

	patch := *gopas.NewJsonPatchDocumentAccountModel()
	patch.SetOperations(operations)

	_, resp, err := client.AccountsApi.AccountsUpdateAccount(ctx, id).AccountPatch(patch).Execute()
	if err != nil {
		return returnResponseErr(resp, err)
	}

	return resourceAccountAWSAccessKeyRead(ctx, d, meta)
}

func resourceAccountAWSAccessKeyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*apiClient).Client

	id := d.Id()

	_, resp, err := client.AccountsApi.AccountsDeleteAccount(ctx, id).Execute()
	if err != nil {
		return returnResponseErr(resp, err)
	}

	d.SetId("")

	return nil
}

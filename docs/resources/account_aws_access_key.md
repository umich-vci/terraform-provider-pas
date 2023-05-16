---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "pas_account_aws_access_key Resource - terraform-provider-pas"
subcategory: ""
description: |-
  Resource to manage AWS IAM User credentials in CyberArk PAS
---

# pas_account_aws_access_key (Resource)

Resource to manage AWS IAM User credentials in CyberArk PAS



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `aws_access_key_id` (String) The unique ID of the Amazon Web Services (AWS) access key that is used by APIs to access the AWS console.
- `aws_account_id` (String) The account ID on the AWS console. This is a 12-digit number such as 123456789012.
- `safe_name` (String) The name of the safe to create the AWS IAM account in.

### Optional

- `aws_account_alias_name` (String) A friendly identifier of your AWS account ID that can be used for your sign-in page to contain your company name, instead of your AWS account ID.
- `aws_arn_role` (String) The role that can securely access the AWS console.
- `aws_policy` (String) The policy that enables access to the AWS console for the specified user.
- `name` (String) The name of the account. If not specified, one is generated.
- `password` (String, Sensitive) The password of the IAM user.

### Read-Only

- `category_modification_time` (Number) TODO
- `created_time` (Number) When the account was created
- `id` (String) The ID of this resource.


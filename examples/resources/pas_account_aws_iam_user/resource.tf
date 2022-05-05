resource "pas_account_aws_iam_user" "aws" {
  safe_name      = "MySafe"
  username       = "iamuser"
  aws_account_id = "123456789012"
}

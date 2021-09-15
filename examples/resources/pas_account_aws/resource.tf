resource "pas_account_aws" "aws" {
  safe_name      = "MySafe"
  username       = "iamuser"
  aws_account_id = "123456789012"
}

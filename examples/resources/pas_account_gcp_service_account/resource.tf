resource "pas_account_gcp_service_account" "gcp_sa" {
  safe_name = "MySafe"
  username  = "service-account@project-id.iam.gserviceaccount.com"
  key_id    = "fOaULWiinTfatlEMeUPScRqi9n0oQmBywvtN3jc7"
}

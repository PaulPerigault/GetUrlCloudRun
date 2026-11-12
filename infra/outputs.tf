output "service_url" {
  description = "URL of the Cloud Run service. Calls need an identity token unless allow_unauthenticated is set."
  value       = google_cloud_run_v2_service.this.uri
}

output "service_account_email" {
  description = "Runtime service account of the service."
  value       = google_service_account.runtime.email
}

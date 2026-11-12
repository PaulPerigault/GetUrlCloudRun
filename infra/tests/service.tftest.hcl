# Offline tests: the Google provider is mocked, no credential and no network.
# Run with `task infra:test`.

mock_provider "google" {}

variables {
  project_id = "demo-project-1"
  image      = "us-central1-docker.pkg.dev/demo-project-1/apps/geturlcloudrun:v0.1.0"
}

run "private_by_default" {
  command = plan

  assert {
    condition     = length(google_cloud_run_v2_service_iam_member.invoker) == 0
    error_message = "no invoker binding may exist unless one is requested"
  }
}

run "public_only_when_asked" {
  command = plan

  variables {
    allow_unauthenticated = true
  }

  assert {
    condition     = contains(keys(google_cloud_run_v2_service_iam_member.invoker), "allUsers")
    error_message = "allow_unauthenticated must bind allUsers"
  }
}

run "named_invokers" {
  command = plan

  variables {
    invoker_members = ["serviceAccount:caller@demo-project-1.iam.gserviceaccount.com"]
  }

  assert {
    condition     = length(google_cloud_run_v2_service_iam_member.invoker) == 1 && !contains(keys(google_cloud_run_v2_service_iam_member.invoker), "allUsers")
    error_message = "one named invoker, and nobody else"
  }
}

run "runtime_configuration" {
  command = plan

  variables {
    allowed_origins = ["https://app.example.com", "http://localhost:4200"]
  }

  assert {
    condition     = one([for e in google_cloud_run_v2_service.this.template[0].containers[0].env : e.value if e.name == "PROVIDER"]) == "gcp"
    error_message = "the deployed service must use the gcp provider, never the fake one"
  }

  assert {
    condition     = one([for e in google_cloud_run_v2_service.this.template[0].containers[0].env : e.value if e.name == "ALLOWED_ORIGINS"]) == "https://app.example.com,http://localhost:4200"
    error_message = "allowed origins must be passed as a comma-separated list"
  }

  assert {
    condition     = one([for e in google_cloud_run_v2_service.this.template[0].containers[0].env : e.value if e.name == "GCP_PROJECT_ID"]) == "demo-project-1"
    error_message = "the project ID must reach the container"
  }
}

run "protected_and_least_privilege" {
  command = plan

  assert {
    condition     = google_cloud_run_v2_service.this.deletion_protection == true
    error_message = "deletion protection must be on by default"
  }

  assert {
    condition     = google_project_iam_member.runtime_run_viewer.role == "roles/run.viewer"
    error_message = "the runtime service account only needs the read-only viewer role"
  }
}

run "rejects_latest_tag" {
  command = plan

  variables {
    image = "us-central1-docker.pkg.dev/demo-project-1/apps/geturlcloudrun:latest"
  }

  expect_failures = [var.image]
}

run "rejects_untagged_image" {
  command = plan

  variables {
    image = "us-central1-docker.pkg.dev/demo-project-1/apps/geturlcloudrun"
  }

  expect_failures = [var.image]
}

run "accepts_digest" {
  command = plan

  variables {
    image = "us-central1-docker.pkg.dev/demo-project-1/apps/geturlcloudrun@sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
  }
}

run "rejects_inverted_scaling" {
  command = plan

  variables {
    min_instances = 5
    max_instances = 2
  }

  expect_failures = [var.max_instances]
}

run "rejects_wildcard_member" {
  command = plan

  variables {
    invoker_members = ["allUsers"]
  }

  expect_failures = [var.invoker_members]
}

run "rejects_bad_project_id" {
  command = plan

  variables {
    project_id = "Bad_Project"
  }

  expect_failures = [var.project_id]
}

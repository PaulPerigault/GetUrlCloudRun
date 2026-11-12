terraform {
  # The exact patch is pinned in .opentofu-version and checked by `task infra:tool`.
  required_version = "~> 1.13.9"

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "8.11.0"
    }
  }

  # Partial configuration: the state lives in a bucket outside this repository.
  # Pass the bucket at init time, see infra/README.md and backend.hcl.example.
  backend "gcs" {}
}

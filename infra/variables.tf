variable "project_id" {
  description = "ID of the GCP project that hosts the service."
  type        = string
  nullable    = false

  validation {
    condition     = can(regex("^[a-z][-a-z0-9]{4,28}[a-z0-9]$", var.project_id))
    error_message = "project_id must be a valid GCP project ID (6 to 30 characters)."
  }
}

variable "region" {
  description = "Cloud Run region. The service lists the services of this region only."
  type        = string
  default     = "us-central1"
  nullable    = false

  validation {
    condition     = can(regex("^[a-z]+-[a-z]+[0-9]+$", var.region))
    error_message = "region must look like europe-west1."
  }
}

variable "service_name" {
  description = "Name of the Cloud Run service."
  type        = string
  default     = "geturlcloudrun"
  nullable    = false

  validation {
    condition     = can(regex("^[a-z]([-a-z0-9]{0,47}[a-z0-9])?$", var.service_name))
    error_message = "service_name must be a valid Cloud Run service name (1 to 49 characters)."
  }
}

variable "service_account_id" {
  description = "Account ID of the runtime service account (6 to 30 characters)."
  type        = string
  default     = "geturlcloudrun-sa"
  nullable    = false

  validation {
    condition     = can(regex("^[a-z][-a-z0-9]{4,28}[a-z0-9]$", var.service_account_id))
    error_message = "service_account_id must be 6 to 30 characters: lower-case letters, digits, dashes."
  }
}

variable "image" {
  description = "Container image, with an explicit tag or a sha256 digest. The tag latest is refused."
  type        = string
  nullable    = false

  validation {
    condition = (
      can(regex("^[a-z0-9.-]+(:[0-9]+)?(/[A-Za-z0-9._-]+)+(:[A-Za-z0-9._-]+|@sha256:[a-f0-9]{64})$", var.image))
      && !endswith(var.image, ":latest")
    )
    error_message = "image must be registry/path:tag or registry/path@sha256:digest, and never :latest."
  }
}

variable "invoker_members" {
  description = "IAM members allowed to call the service (roles/run.invoker), for example serviceAccount:caller@project.iam.gserviceaccount.com. Authentication is delegated to Cloud Run IAM."
  type        = list(string)
  default     = []
  nullable    = false

  validation {
    condition     = alltrue([for m in var.invoker_members : can(regex("^(user|group|serviceAccount|domain):[^ ]+$", m))])
    error_message = "Each member must start with user:, group:, serviceAccount: or domain:. Use allow_unauthenticated for public access."
  }
}

variable "allow_unauthenticated" {
  description = "Make the service public (grants roles/run.invoker to allUsers). Off by default: the service is private."
  type        = bool
  default     = false
  nullable    = false
}

variable "allowed_origins" {
  description = "Browser origins allowed by CORS (exact, for example https://app.example.com). Empty means CORS is refused."
  type        = list(string)
  default     = []
  nullable    = false
}

variable "min_instances" {
  description = "Minimum number of instances. 0 lets the service scale to zero."
  type        = number
  default     = 0
  nullable    = false

  validation {
    condition     = var.min_instances >= 0 && var.min_instances <= 10
    error_message = "min_instances must be between 0 and 10."
  }
}

variable "max_instances" {
  description = "Maximum number of instances. Bounds the cost and the load on the Cloud Run Admin API."
  type        = number
  default     = 3
  nullable    = false

  validation {
    condition     = var.max_instances >= 1 && var.max_instances <= 100 && var.max_instances >= var.min_instances
    error_message = "max_instances must be between 1 and 100 and at least min_instances."
  }
}

variable "deletion_protection" {
  description = "Refuse to delete the service. Set to false in a separate, announced change before any destroy."
  type        = bool
  default     = true
  nullable    = false
}

# Only the ruleset bundled in the pinned tflint binary is used: the Google
# ruleset would be a plugin downloaded at run time, which the repository
# forbids without a pinned fingerprint.
config {
  call_module_type = "none"
}

plugin "terraform" {
  enabled = true
  preset  = "recommended"
}

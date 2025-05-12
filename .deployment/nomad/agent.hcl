acl {
  enabled    = true
  token_ttl  = "30s"
  policy_ttl = "60s"
  role_ttl   = "60s"
}

plugin "docker" {
    config {
        volumes {
            enabled      = true
            selinuxlabel = "z"
        }
    }
}
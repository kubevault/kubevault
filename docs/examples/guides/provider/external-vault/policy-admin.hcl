path "sys/policy/*" {
  capabilities = ["create", "update", "read", "delete", "list"]
}

path "sys/policy" {
  capabilities = ["read", "list"]
}

path "auth/kubernetes/role" {
  capabilities = ["read", "list"]
}

path "auth/kubernetes/role/*" {
  capabilities = ["create", "update", "read", "delete", "list"]
}
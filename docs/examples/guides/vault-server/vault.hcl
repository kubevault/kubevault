path "sys/mounts" {
  capabilities = ["read", "list"]
}
path "sys/mounts/*" {
  capabilities = ["create", "read", "update", "delete"]
}
path "sys/leases/revoke/*" {
    capabilities = ["update"]
}
path "sys/policy/*" {
	capabilities = ["create", "update", "read", "delete", "list"]
}
path "sys/policy" {
	capabilities = ["read", "list"]
}
path "sys/policies" {
	capabilities = ["read", "list"]
}
path "sys/policies/*" {
	capabilities = ["create", "update", "read", "delete", "list"]
}
path "auth/kubernetes/role" {
	capabilities = ["read", "list"]
}
path "auth/kubernetes/role/*" {
	capabilities = ["create", "update", "read", "delete", "list"]
}
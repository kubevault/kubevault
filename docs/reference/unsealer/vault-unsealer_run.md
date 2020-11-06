---
title: Vault-Unsealer Run
menu:
  docs_{{ .version }}:
    identifier: vault-unsealer-run
    name: Vault-Unsealer Run
    parent: reference-unsealer
menu_name: docs_{{ .version }}
section_menu_id: reference
---
## vault-unsealer run

Launch Vault unsealer

```
vault-unsealer run [flags]
```

### Options

```
      --auth.k8s-ca-cert string                           PEM encoded CA cert for use by the TLS client used to talk with the Kubernetes API
      --auth.k8s-host string                              Host must be a host string, a host:port pair, or a URL to the base of the Kubernetes API server
      --auth.k8s-token-reviewer-jwt string                A service account JWT used to access the TokenReview API to validate other JWTs during login. If this flag is not provided, then the value from K8S_TOKEN_REVIEWER_JWT environment variable will be used
      --aws.kms-key-id string                             The ID or ARN of the AWS KMS key to encrypt values
      --aws.ssm-key-prefix string                         The Key Prefix for SSM Parameter store
      --aws.use-secure-string                             Use secure string parameter, for more info https://docs.aws.amazon.com/systems-manager/latest/userguide/sysman-paramstore-about.html#sysman-paramstore-securestring
      --azure.client-cert-password string                 The password of the client certificate for an AAD application
      --azure.client-cert-path string                     The path of a client certificate for an AAD application
      --azure.client-id string                            The ClientID for an AAD application.
      --azure.client-secret string                        The ClientSecret for an AAD application
      --azure.cloud string                                The cloud environment identifier (default "AZUREPUBLICCLOUD")
      --azure.secret-prefix string                        Prefix to use in secret name for azure key vault
      --azure.tenant-id string                            The AAD Tenant ID
      --azure.use-managed-identity                        Use managed service identity for the virtual machine
      --azure.vault-base-url string                       Azure key vault url, for example https://myvault.vault.azure.net
      --google.kms-crypto-key string                      The name of the Google Cloud KMS crypto key to use
      --google.kms-key-ring string                        The name of the Google Cloud KMS key ring to use
      --google.kms-location string                        The Google Cloud KMS location to use (eg. 'global', 'europe-west1')
      --google.kms-project string                         The Google Cloud KMS project to use
      --google.storage-bucket string                      The name of the Google Cloud Storage bucket to store values in
      --google.storage-prefix string                      The prefix to use for values store in Google Cloud Storage
  -h, --help                                              help for run
      --k8s.secret-name string                            Secret name to use when creating secret containing root token and shared keys
      --mode string                                       Select the mode to use 'google-cloud-kms-gcs' => Google Cloud Storage with encryption using Google KMS; 'aws-kms-ssm' => AWS SSM parameter store using AWS KMS; 'azure-key-vault' => Azure Key Vault Secret store; 'kubernetes-secret' => Kubernetes secret to store unseal keys
      --overwrite-existing                                overwrite existing unseal keys and root tokens, possibly dangerous!
      --policy-manager.name string                        Name of the policy. A policy and a  vault kubernetes auth role will be created using this name
      --policy-manager.service-account-name string        Name of the service account
      --policy-manager.service-account-namespace string   Namespace of the service account
      --retry-period duration                             How often to attempt to unseal the vault instance (default 10s)
      --secret-shares int                                 Total count of secret shares that exist (default 5)
      --secret-threshold int                              Minimum required secret shares to unseal (default 3)
      --store-root-token                                  should the root token be stored in the key store (default true)
      --vault.address string                              Specifies the vault address. Address form : scheme://host:port (default "https://127.0.0.1:8200")
      --vault.ca-cert string                              Specifies the CA cert that will be used to verify self signed vault server certificate
      --vault.insecure-skip-tls-verify                    To skip tls verification when communicating with vault server
```

### Options inherited from parent commands

```
      --alsologtostderr                  log to standard error as well as files
      --enable-analytics                 Send analytical events to Google Analytics (default true)
      --log_backtrace_at traceLocation   when logging hits line file:N, emit a stack trace (default :0)
      --log_dir string                   If non-empty, write log files in this directory
      --logtostderr                      log to standard error instead of files
      --stderrthreshold severity         logs at or above this threshold go to stderr
      --use-kubeapiserver-fqdn-for-aks   if true, uses kube-apiserver FQDN for AKS cluster to workaround https://github.com/Azure/AKS/issues/522 (default true)
  -v, --v Level                          log level for V logs
      --vmodule moduleSpec               comma-separated list of pattern=N settings for file-filtered logging
```

### SEE ALSO

* [vault-unsealer](/docs/reference/unsealer/vault-unsealer.md)	 - Automates initialisation and unsealing of Hashicorp Vault


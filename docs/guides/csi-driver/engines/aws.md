
## Setup

1. Enable the AWS secret engine:

```bash
$ vault secrets enable aws
Success! Enabled the aws secrets engine at: aws/
```

2. Configure the credentials that Vault uses to communicate with AWS to generate the IAM credentials

```bash
$ vault write aws/config/root \
    access_key=AKIAJWVN5Z4FOFT7NLNA \
    secret_key=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 \
    region=us-east-1
```

3. Create IAM policy on aws with following and copy the value of policy ARN:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "iam:AttachUserPolicy",
        "iam:CreateAccessKey",
        "iam:CreateUser",
        "iam:DeleteAccessKey",
        "iam:DeleteUser",
        "iam:DeleteUserPolicy",
        "iam:DetachUserPolicy",
        "iam:ListAccessKeys",
        "iam:ListAttachedUserPolicies",
        "iam:ListGroupsForUser",
        "iam:ListUserPolicies",
        "iam:PutUserPolicy",
        "iam:RemoveUserFromGroup"
      ],
      "Resource": [
        "arn:aws:iam::ACCOUNT-ID-WITHOUT-HYPHENS:user/vault-*"
      ]
    }
  ]
}
```

4. Configure a vault role that maps to a set of permissions in AWS and an AWS credential type. When users generate credentials, they are generated against this role,

```bash
$ vault write aws/roles/my-aws-role \
    arn=arn:aws:iam::452618475015:policy/vaultiampolicy \
    credential_type=iam_user \
    policy_document=-<<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": "ec2:*",
      "Resource": "*"
    }
  ]
}
EOF

```

Here, my-aws-role was treated as secret name on storage class.

### Create Kubernetes role

To create role on kubernetes cluster run

```bash
$ vault write auth/kubernetes/role/aws-cred-role bound_service_account_names=aws-vault bound_service_account_namespaces=default policies=test-policy ttl=24h

```
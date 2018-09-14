// +build !ignore_autogenerated

/*
Copyright 2018 The Kube Vault Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by deepcopy-gen. DO NOT EDIT.

package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AwsKmsSsmSpec) DeepCopyInto(out *AwsKmsSsmSpec) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AwsKmsSsmSpec.
func (in *AwsKmsSsmSpec) DeepCopy() *AwsKmsSsmSpec {
	if in == nil {
		return nil
	}
	out := new(AwsKmsSsmSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AzureKeyVault) DeepCopyInto(out *AzureKeyVault) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AzureKeyVault.
func (in *AzureKeyVault) DeepCopy() *AzureKeyVault {
	if in == nil {
		return nil
	}
	out := new(AzureKeyVault)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AzureSpec) DeepCopyInto(out *AzureSpec) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AzureSpec.
func (in *AzureSpec) DeepCopy() *AzureSpec {
	if in == nil {
		return nil
	}
	out := new(AzureSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *BackendStorageSpec) DeepCopyInto(out *BackendStorageSpec) {
	*out = *in
	if in.Etcd != nil {
		in, out := &in.Etcd, &out.Etcd
		if *in == nil {
			*out = nil
		} else {
			*out = new(EtcdSpec)
			**out = **in
		}
	}
	if in.Gcs != nil {
		in, out := &in.Gcs, &out.Gcs
		if *in == nil {
			*out = nil
		} else {
			*out = new(GcsSpec)
			**out = **in
		}
	}
	if in.S3 != nil {
		in, out := &in.S3, &out.S3
		if *in == nil {
			*out = nil
		} else {
			*out = new(S3Spec)
			**out = **in
		}
	}
	if in.Azure != nil {
		in, out := &in.Azure, &out.Azure
		if *in == nil {
			*out = nil
		} else {
			*out = new(AzureSpec)
			**out = **in
		}
	}
	if in.PostgreSQL != nil {
		in, out := &in.PostgreSQL, &out.PostgreSQL
		if *in == nil {
			*out = nil
		} else {
			*out = new(PostgreSQLSpec)
			**out = **in
		}
	}
	if in.MySQL != nil {
		in, out := &in.MySQL, &out.MySQL
		if *in == nil {
			*out = nil
		} else {
			*out = new(MySQLSpec)
			**out = **in
		}
	}
	if in.File != nil {
		in, out := &in.File, &out.File
		if *in == nil {
			*out = nil
		} else {
			*out = new(FileSpec)
			**out = **in
		}
	}
	if in.DynamoDB != nil {
		in, out := &in.DynamoDB, &out.DynamoDB
		if *in == nil {
			*out = nil
		} else {
			*out = new(DynamoDBSpec)
			**out = **in
		}
	}
	if in.Swift != nil {
		in, out := &in.Swift, &out.Swift
		if *in == nil {
			*out = nil
		} else {
			*out = new(SwiftSpec)
			**out = **in
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new BackendStorageSpec.
func (in *BackendStorageSpec) DeepCopy() *BackendStorageSpec {
	if in == nil {
		return nil
	}
	out := new(BackendStorageSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DynamoDBSpec) DeepCopyInto(out *DynamoDBSpec) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DynamoDBSpec.
func (in *DynamoDBSpec) DeepCopy() *DynamoDBSpec {
	if in == nil {
		return nil
	}
	out := new(DynamoDBSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EtcdSpec) DeepCopyInto(out *EtcdSpec) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EtcdSpec.
func (in *EtcdSpec) DeepCopy() *EtcdSpec {
	if in == nil {
		return nil
	}
	out := new(EtcdSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FileSpec) DeepCopyInto(out *FileSpec) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FileSpec.
func (in *FileSpec) DeepCopy() *FileSpec {
	if in == nil {
		return nil
	}
	out := new(FileSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GcsSpec) DeepCopyInto(out *GcsSpec) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GcsSpec.
func (in *GcsSpec) DeepCopy() *GcsSpec {
	if in == nil {
		return nil
	}
	out := new(GcsSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GoogleKmsGcsSpec) DeepCopyInto(out *GoogleKmsGcsSpec) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GoogleKmsGcsSpec.
func (in *GoogleKmsGcsSpec) DeepCopy() *GoogleKmsGcsSpec {
	if in == nil {
		return nil
	}
	out := new(GoogleKmsGcsSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KubernetesSecretSpec) DeepCopyInto(out *KubernetesSecretSpec) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KubernetesSecretSpec.
func (in *KubernetesSecretSpec) DeepCopy() *KubernetesSecretSpec {
	if in == nil {
		return nil
	}
	out := new(KubernetesSecretSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ModeSpec) DeepCopyInto(out *ModeSpec) {
	*out = *in
	if in.KubernetesSecret != nil {
		in, out := &in.KubernetesSecret, &out.KubernetesSecret
		if *in == nil {
			*out = nil
		} else {
			*out = new(KubernetesSecretSpec)
			**out = **in
		}
	}
	if in.GoogleKmsGcs != nil {
		in, out := &in.GoogleKmsGcs, &out.GoogleKmsGcs
		if *in == nil {
			*out = nil
		} else {
			*out = new(GoogleKmsGcsSpec)
			**out = **in
		}
	}
	if in.AwsKmsSsm != nil {
		in, out := &in.AwsKmsSsm, &out.AwsKmsSsm
		if *in == nil {
			*out = nil
		} else {
			*out = new(AwsKmsSsmSpec)
			**out = **in
		}
	}
	if in.AzureKeyVault != nil {
		in, out := &in.AzureKeyVault, &out.AzureKeyVault
		if *in == nil {
			*out = nil
		} else {
			*out = new(AzureKeyVault)
			**out = **in
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ModeSpec.
func (in *ModeSpec) DeepCopy() *ModeSpec {
	if in == nil {
		return nil
	}
	out := new(ModeSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MySQLSpec) DeepCopyInto(out *MySQLSpec) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MySQLSpec.
func (in *MySQLSpec) DeepCopy() *MySQLSpec {
	if in == nil {
		return nil
	}
	out := new(MySQLSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PostgreSQLSpec) DeepCopyInto(out *PostgreSQLSpec) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PostgreSQLSpec.
func (in *PostgreSQLSpec) DeepCopy() *PostgreSQLSpec {
	if in == nil {
		return nil
	}
	out := new(PostgreSQLSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *S3Spec) DeepCopyInto(out *S3Spec) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new S3Spec.
func (in *S3Spec) DeepCopy() *S3Spec {
	if in == nil {
		return nil
	}
	out := new(S3Spec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SwiftSpec) DeepCopyInto(out *SwiftSpec) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SwiftSpec.
func (in *SwiftSpec) DeepCopy() *SwiftSpec {
	if in == nil {
		return nil
	}
	out := new(SwiftSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TLSPolicy) DeepCopyInto(out *TLSPolicy) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TLSPolicy.
func (in *TLSPolicy) DeepCopy() *TLSPolicy {
	if in == nil {
		return nil
	}
	out := new(TLSPolicy)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *UnsealerSpec) DeepCopyInto(out *UnsealerSpec) {
	*out = *in
	in.Mode.DeepCopyInto(&out.Mode)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new UnsealerSpec.
func (in *UnsealerSpec) DeepCopy() *UnsealerSpec {
	if in == nil {
		return nil
	}
	out := new(UnsealerSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VaultServer) DeepCopyInto(out *VaultServer) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VaultServer.
func (in *VaultServer) DeepCopy() *VaultServer {
	if in == nil {
		return nil
	}
	out := new(VaultServer)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *VaultServer) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VaultServerCondition) DeepCopyInto(out *VaultServerCondition) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VaultServerCondition.
func (in *VaultServerCondition) DeepCopy() *VaultServerCondition {
	if in == nil {
		return nil
	}
	out := new(VaultServerCondition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VaultServerList) DeepCopyInto(out *VaultServerList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]VaultServer, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VaultServerList.
func (in *VaultServerList) DeepCopy() *VaultServerList {
	if in == nil {
		return nil
	}
	out := new(VaultServerList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *VaultServerList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VaultServerSpec) DeepCopyInto(out *VaultServerSpec) {
	*out = *in
	if in.ConfigSource != nil {
		in, out := &in.ConfigSource, &out.ConfigSource
		if *in == nil {
			*out = nil
		} else {
			*out = new(v1.VolumeSource)
			(*in).DeepCopyInto(*out)
		}
	}
	if in.TLS != nil {
		in, out := &in.TLS, &out.TLS
		if *in == nil {
			*out = nil
		} else {
			*out = new(TLSPolicy)
			**out = **in
		}
	}
	in.Backend.DeepCopyInto(&out.Backend)
	if in.Unsealer != nil {
		in, out := &in.Unsealer, &out.Unsealer
		if *in == nil {
			*out = nil
		} else {
			*out = new(UnsealerSpec)
			(*in).DeepCopyInto(*out)
		}
	}
	in.PodTemplate.DeepCopyInto(&out.PodTemplate)
	in.ServiceTemplate.DeepCopyInto(&out.ServiceTemplate)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VaultServerSpec.
func (in *VaultServerSpec) DeepCopy() *VaultServerSpec {
	if in == nil {
		return nil
	}
	out := new(VaultServerSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VaultServerStatus) DeepCopyInto(out *VaultServerStatus) {
	*out = *in
	if in.ObservedGeneration != nil {
		in, out := &in.ObservedGeneration, &out.ObservedGeneration
		if *in == nil {
			*out = nil
		} else {
			*out = (*in).DeepCopy()
		}
	}
	in.VaultStatus.DeepCopyInto(&out.VaultStatus)
	if in.UpdatedNodes != nil {
		in, out := &in.UpdatedNodes, &out.UpdatedNodes
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]VaultServerCondition, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VaultServerStatus.
func (in *VaultServerStatus) DeepCopy() *VaultServerStatus {
	if in == nil {
		return nil
	}
	out := new(VaultServerStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VaultStatus) DeepCopyInto(out *VaultStatus) {
	*out = *in
	if in.Standby != nil {
		in, out := &in.Standby, &out.Standby
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Sealed != nil {
		in, out := &in.Sealed, &out.Sealed
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Unsealed != nil {
		in, out := &in.Unsealed, &out.Unsealed
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VaultStatus.
func (in *VaultStatus) DeepCopy() *VaultStatus {
	if in == nil {
		return nil
	}
	out := new(VaultStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VaultserverVersion) DeepCopyInto(out *VaultserverVersion) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VaultserverVersion.
func (in *VaultserverVersion) DeepCopy() *VaultserverVersion {
	if in == nil {
		return nil
	}
	out := new(VaultserverVersion)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *VaultserverVersion) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VaultserverVersionList) DeepCopyInto(out *VaultserverVersionList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]VaultserverVersion, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VaultserverVersionList.
func (in *VaultserverVersionList) DeepCopy() *VaultserverVersionList {
	if in == nil {
		return nil
	}
	out := new(VaultserverVersionList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *VaultserverVersionList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VaultserverVersionSpec) DeepCopyInto(out *VaultserverVersionSpec) {
	*out = *in
	out.Vault = in.Vault
	out.Unsealer = in.Unsealer
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VaultserverVersionSpec.
func (in *VaultserverVersionSpec) DeepCopy() *VaultserverVersionSpec {
	if in == nil {
		return nil
	}
	out := new(VaultserverVersionSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VaultserverVersionUnsealer) DeepCopyInto(out *VaultserverVersionUnsealer) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VaultserverVersionUnsealer.
func (in *VaultserverVersionUnsealer) DeepCopy() *VaultserverVersionUnsealer {
	if in == nil {
		return nil
	}
	out := new(VaultserverVersionUnsealer)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VaultserverVersionVault) DeepCopyInto(out *VaultserverVersionVault) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VaultserverVersionVault.
func (in *VaultserverVersionVault) DeepCopy() *VaultserverVersionVault {
	if in == nil {
		return nil
	}
	out := new(VaultserverVersionVault)
	in.DeepCopyInto(out)
	return out
}

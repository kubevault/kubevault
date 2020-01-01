package types

import (
	"encoding/json"
	"reflect"
	"testing"

	config "kubevault.dev/operator/apis/config/v1alpha1"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	kfake "k8s.io/client-go/kubernetes/fake"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
)

func TestGetAuthInfoFromAppBinding(t *testing.T) {
	kc := kfake.NewSimpleClientset()
	vApp := &appcat.AppBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "vault-app",
			Namespace: "demo",
		},
		Spec: appcat.AppBindingSpec{
			ClientConfig: appcat.ClientConfig{
				Service: &appcat.ServiceReference{
					Scheme: "HTTPS",
					Name:   "vault",
					Port:   8200,
				},
				InsecureSkipTLSVerify: false,
				CABundle:              []byte("LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUN1RENDQWFDZ0F3SUJBZ0lCQURBTkJna3Foa2lHOXcwQkFRc0ZBREFOTVFzd0NRWURWUVFERXdKallUQWUKRncweE9ERXlNamN3TkRVNU1qVmFGdzB5T0RFeU1qUXdORFU1TWpWYU1BMHhDekFKQmdOVkJBTVRBbU5oTUlJQgpJakFOQmdrcWhraUc5dzBCQVFFRkFBT0NBUThBTUlJQkNnS0NBUUVBMVhid2wyQ1NNc2VQTU5RRzhMd3dUVWVOCkI1T05oSTlDNzFtdUoyZEZjTTlUc1VDQnlRRk1weUc5dWFvV3J1ZDhtSWpwMVl3MmVIUW5udmoybXRmWGcrWFcKSThCYkJUaUFKMWxMMFE5MlV0a1BLczlXWEt6dTN0SjJUR1hRRDhhbHZhZ0JrR1ViOFJYaUNqK2pnc1p6TDRvQQpNRWszSU9jS0xnMm9ldFZNQ0hwNktpWTBnQkZiUWdJZ1A1TnFwbksrbU02ZTc1ZW5hWEdBK2V1d09FT0YwV0Z2CmxGQmgzSEY5QlBGdTJKbkZQUlpHVDJKajBRR1FNeUxodEY5Tk1pZTdkQnhiTWhRVitvUXp2d1EvaXk1Q2pndXQKeDc3d29HQ2JtM0o4cXRybUg2Tjl6Tlc3WlR0YTdLd05PTmFoSUFEMSsrQm5rc3JvYi9BYWRKT0tMN2dLYndJRApBUUFCb3lNd0lUQU9CZ05WSFE4QkFmOEVCQU1DQXFRd0R3WURWUjBUQVFIL0JBVXdBd0VCL3pBTkJna3Foa2lHCjl3MEJBUXNGQUFPQ0FRRUFXeWFsdUt3Wk1COWtZOEU5WkdJcHJkZFQyZnFTd0lEOUQzVjN5anBlaDVCOUZHN1UKSS8wNmpuRVcyaWpESXNHNkFDZzJKOXdyaSttZ2VIa2Y2WFFNWjFwZHRWeDZLVWplWTVnZStzcGdCRTEyR2NPdwpxMUhJb0NrekVBMk5HOGRNRGM4dkQ5WHBQWGwxdW5veWN4Y0VMeFVRSC9PRlc4eHJxNU9vcXVYUkxMMnlKcXNGCmlvM2lJV3EvU09Yajc4MVp6MW5BV1JSNCtSYW1KWjlOcUNjb1Z3b3R6VzI1UWJKWWJ3QzJOSkNENEFwOUtXUjUKU2w2blk3NVMybEdSRENsQkNnN2VRdzcwU25seW5mb3RaTUpKdmFzbStrOWR3U0xtSDh2RDNMMGNGOW5SOENTSgpiTjBiZzczeVlWRHgyY3JRYk0zcko4dUJnY3BsWlRpUy91SXJ2QT09Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K"),
			},
			Parameters: &runtime.RawExtension{},
		},
	}

	tests := []struct {
		name    string
		kClient kubernetes.Interface
		secret  *corev1.Secret
		vApp    *appcat.AppBinding
		vConfig *config.VaultServerConfiguration
		want    *AuthInfo
		wantErr bool
	}{
		{
			name:    "Should be failed; Empty kubernetes client",
			kClient: nil,
			secret:  nil,
			vApp:    vApp,
			vConfig: nil,
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Should be failed; Empty AppBinding",
			kClient: kc,
			secret:  nil,
			vApp:    nil,
			vConfig: nil,
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Should be passed; Correct inputs",
			kClient: kc,
			secret:  nil,
			vApp:    vApp,
			vConfig: &config.VaultServerConfiguration{
				Path:      "kubernetes",
				VaultRole: "policy-controller-role",
				Kubernetes: &config.KubernetesAuthConfig{
					ServiceAccountName: "vault",
				},
			},
			want: &AuthInfo{
				VaultApp: vApp,
				ServiceAccountRef: &corev1.ObjectReference{
					Namespace: "demo",
					Name:      "vault",
				},
				Secret: nil,
				ExtraInfo: &AuthExtraInfo{
					Kubernetes: &config.KubernetesAuthConfig{
						ServiceAccountName: "vault",
					},
				},
				VaultRole: "policy-controller-role",
				Path:      "kubernetes",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.vConfig != nil {
				data, err := json.Marshal(tt.vConfig)
				assert.Nil(t, err)
				if tt.vApp != nil {
					tt.vApp.Spec.Parameters.Raw = data
				}
			}
			got, err := GetAuthInfoFromAppBinding(tt.kClient, tt.vApp)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAuthInfoFromAppBinding() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAuthInfoFromAppBinding() got = %v, want %v", got, tt.want)
			}
		})
	}
}

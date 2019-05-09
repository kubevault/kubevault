module github.com/kubevault/operator

go 1.12

require (
	contrib.go.opencensus.io/exporter/ocagent v0.4.12 // indirect
	github.com/Azure/go-autorest v12.0.0+incompatible // indirect
	github.com/SermoDigital/jose v0.0.0-20180104203859-803625baeddc // indirect
	github.com/appscode/go v0.0.0-20190424183524-60025f1135c9
	github.com/appscode/pat v0.0.0-20170521084856-48ff78925b79
	github.com/armon/go-radix v1.0.0 // indirect
	github.com/aws/aws-sdk-go v1.19.27
	github.com/codeskyblue/go-sh v0.0.0-20190412065543-76bd3d59ff27
	github.com/coreos/prometheus-operator v0.29.0
	github.com/cpuguy83/go-md2man v1.0.10 // indirect
	github.com/evanphx/json-patch v4.2.0+incompatible
	github.com/go-openapi/spec v0.19.0
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/golang/snappy v0.0.1 // indirect
	github.com/gophercloud/gophercloud v0.0.0-20190509013533-844afee4f565 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.1
	github.com/hashicorp/go-hclog v0.9.0 // indirect
	github.com/hashicorp/go-immutable-radix v1.0.0 // indirect
	github.com/hashicorp/go-plugin v1.0.0 // indirect
	github.com/hashicorp/go-retryablehttp v0.5.3 // indirect
	github.com/hashicorp/go-rootcerts v1.0.0 // indirect
	github.com/hashicorp/go-sockaddr v1.0.2 // indirect
	github.com/hashicorp/go-uuid v1.0.1 // indirect
	github.com/hashicorp/go-version v1.1.0 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/hashicorp/vault v1.0.1
	github.com/hashicorp/yamux v0.0.0-20181012175058-2f1d1f20f75d // indirect
	github.com/kubedb/apimachinery v0.0.0-20190506191700-871d6b5d30ee
	github.com/lib/pq v0.0.0-20180201184707-88edab080323
	github.com/mitchellh/go-testing-interface v1.0.0 // indirect
	github.com/ncw/swift v1.0.47
	github.com/onsi/ginkgo v1.8.0
	github.com/onsi/gomega v1.5.0
	github.com/pierrec/lz4 v2.0.5+incompatible // indirect
	github.com/pkg/errors v0.8.1
	github.com/prometheus/common v0.4.0 // indirect
	github.com/ryanuber/go-glob v1.0.0 // indirect
	github.com/spf13/afero v1.2.2
	github.com/spf13/cobra v0.0.3
	github.com/spf13/pflag v1.0.3
	github.com/stretchr/testify v1.3.0
	golang.org/x/oauth2 v0.0.0-20190402181905-9f3314589c9a
	golang.org/x/sync v0.0.0-20190423024810-112230192c58 // indirect
	golang.org/x/sys v0.0.0-20190508220229-2d0786266e9c // indirect
	gomodules.xyz/cert v1.0.0
	google.golang.org/api v0.4.0
	google.golang.org/genproto v0.0.0-20190508193815-b515fa19cec8 // indirect
	google.golang.org/grpc v1.20.1 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.0.0 // indirect
	k8s.io/api v0.0.0-20190503110853-61630f889b3c
	k8s.io/apiextensions-apiserver v0.0.0-20190508224317-421cff06bf05
	k8s.io/apimachinery v0.0.0-20190508063446-a3da69d3723c
	k8s.io/apiserver v0.0.0-20190508223931-4756b09d7af2
	k8s.io/cli-runtime v0.0.0-20190508184404-b26560c459bd // indirect
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/component-base v0.0.0-20190508223741-40efa6d42997 // indirect
	k8s.io/kube-aggregator v0.0.0-20190508224022-f9852b6d3a84
	k8s.io/kube-openapi v0.0.0-20190502190224-411b2483e503
	k8s.io/kubernetes v1.14.1
	kmodules.xyz/client-go v0.0.0-20190508091620-0d215c04352f
	kmodules.xyz/custom-resources v0.0.0-20190225012057-ed1c15a0bbda
	kmodules.xyz/monitoring-agent-api v0.0.0-20190508125842-489150794b9b
	kmodules.xyz/objectstore-api v0.0.0-20190506085934-94c81c8acca9 // indirect
	kmodules.xyz/offshoot-api v0.0.0-20190508142450-1c69d50f3c1c
	kmodules.xyz/webhook-runtime v0.0.0-20190508093950-b721b4eba5e5
)

replace (
	github.com/graymeta/stow => github.com/appscode/stow v0.0.0-20190506085026-ca5baa008ea3
	gopkg.in/robfig/cron.v2 => github.com/appscode/cron v0.0.0-20170717094345-ca60c6d796d4
	k8s.io/api => k8s.io/api v0.0.0-20190313235455-40a48860b5ab
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20190315093550-53c4693659ed
	k8s.io/apimachinery => github.com/kmodules/apimachinery v0.0.0-20190508045248-a52a97a7a2bf
	k8s.io/apiserver => github.com/kmodules/apiserver v0.0.0-20190508082252-8397d761d4b5
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.0.0-20190314001948-2899ed30580f
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.0.0-20190314002645-c892ea32361a
	k8s.io/component-base => k8s.io/component-base v0.0.0-20190314000054-4a91899592f4
	k8s.io/klog => k8s.io/klog v0.3.0
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.0.0-20190314000639-da8327669ac5
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20190228160746-b3a7cee44a30
	k8s.io/metrics => k8s.io/metrics v0.0.0-20190314001731-1bd6a4002213
	k8s.io/utils => k8s.io/utils v0.0.0-20190221042446-c2654d5206da
)

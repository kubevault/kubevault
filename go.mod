module kubevault.dev/operator

go 1.12

require (
	cloud.google.com/go v0.39.0 // indirect
	github.com/PuerkitoBio/purell v1.1.1 // indirect
	github.com/SermoDigital/jose v0.9.1 // indirect
	github.com/appscode/go v0.0.0-20191006073906-e3d193d493fc
	github.com/armon/go-radix v1.0.0 // indirect
	github.com/aws/aws-sdk-go v1.20.20
	github.com/census-instrumentation/opencensus-proto v0.2.0 // indirect
	github.com/codeskyblue/go-sh v0.0.0-20190412065543-76bd3d59ff27
	github.com/coreos/prometheus-operator v0.31.1
	github.com/evanphx/json-patch v4.5.0+incompatible
	github.com/frankban/quicktest v1.4.0 // indirect
	github.com/go-openapi/spec v0.19.0
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/gorilla/mux v1.7.3
	github.com/hashicorp/go-cleanhttp v0.5.1
	github.com/hashicorp/go-immutable-radix v1.1.0 // indirect
	github.com/hashicorp/go-plugin v1.0.1 // indirect
	github.com/hashicorp/go-retryablehttp v0.6.2 // indirect
	github.com/hashicorp/vault v1.0.1
	github.com/hashicorp/yamux v0.0.0-20181012175058-2f1d1f20f75d // indirect
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/lib/pq v1.1.0
	github.com/ncw/swift v1.0.47
	github.com/onsi/ginkgo v1.8.0
	github.com/onsi/gomega v1.5.0
	github.com/pierrec/lz4 v2.2.5+incompatible // indirect
	github.com/pkg/errors v0.8.1
	github.com/prometheus/client_golang v0.9.4 // indirect
	github.com/ryanuber/go-glob v1.0.0 // indirect
	github.com/smartystreets/assertions v0.0.0-20190116191733-b6c0e53d7304 // indirect
	github.com/spf13/afero v1.2.2
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.3
	github.com/stretchr/testify v1.3.0
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45
	gomodules.xyz/cert v1.0.0
	google.golang.org/api v0.7.0
	k8s.io/api v0.0.0-20190503110853-61630f889b3c
	k8s.io/apiextensions-apiserver v0.0.0-20190516231611-bf6753f2aa24
	k8s.io/apimachinery v0.0.0-20190508063446-a3da69d3723c
	k8s.io/apiserver v0.0.0-20190516230822-f89599b3f645
	k8s.io/cli-runtime v0.0.0-20190516231937-17bc0b7fcef5 // indirect
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/component-base v0.0.0-20190508223741-40efa6d42997 // indirect
	k8s.io/kube-aggregator v0.0.0-20190508224022-f9852b6d3a84
	k8s.io/kube-openapi v0.0.0-20190502190224-411b2483e503
	k8s.io/kubernetes v1.14.0
	kmodules.xyz/client-go v0.0.0-20191006173540-91f8ee6b6b4b
	kmodules.xyz/custom-resources v0.0.0-20190927035424-65fe358bb045
	kmodules.xyz/monitoring-agent-api v0.0.0-20190808150221-601a4005b7f7
	kmodules.xyz/objectstore-api v0.0.0-20191006080053-fc8b57fadcf0
	kmodules.xyz/offshoot-api v0.0.0-20190901210649-de049192326c
	kmodules.xyz/webhook-runtime v0.0.0-20190808145328-4186c470d56b
)

replace (
	contrib.go.opencensus.io/exporter/ocagent => contrib.go.opencensus.io/exporter/ocagent v0.3.0
	git.apache.org/thrift.git => github.com/apache/thrift v0.12.0
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v12.3.0+incompatible
	github.com/census-instrumentation/opencensus-proto => github.com/census-instrumentation/opencensus-proto v0.1.0
	github.com/golang/protobuf => github.com/golang/protobuf v1.2.0
	go.opencensus.io => go.opencensus.io v0.21.0
	k8s.io/api => k8s.io/api v0.0.0-20190313235455-40a48860b5ab
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20190315093550-53c4693659ed
	k8s.io/apimachinery => github.com/kmodules/apimachinery v0.0.0-20190508045248-a52a97a7a2bf
	k8s.io/apiserver => github.com/kmodules/apiserver v0.0.0-20190811223248-5a95b2df4348
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.0.0-20190314001948-2899ed30580f
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.0.0-20190314002645-c892ea32361a
	k8s.io/component-base => k8s.io/component-base v0.0.0-20190314000054-4a91899592f4
	k8s.io/klog => k8s.io/klog v0.3.0
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.0.0-20190314000639-da8327669ac5
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20190228160746-b3a7cee44a30
	k8s.io/metrics => k8s.io/metrics v0.0.0-20190314001731-1bd6a4002213
	k8s.io/utils => k8s.io/utils v0.0.0-20190221042446-c2654d5206da
	kmodules.xyz/custom-resources => github.com/kamolhasan/custom-resources v0.0.0-20190826094133-83f06ebdbc32
	sigs.k8s.io/structured-merge-diff => sigs.k8s.io/structured-merge-diff v0.0.0-20190302045857-e85c7b244fd2
)

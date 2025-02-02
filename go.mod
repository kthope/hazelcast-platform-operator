module github.com/hazelcast/hazelcast-platform-operator

go 1.16

require (
	cloud.google.com/go/bigquery v1.4.0
	github.com/go-logr/logr v0.3.0
	github.com/hazelcast/hazelcast-go-client v1.2.0
	github.com/onsi/ginkgo/v2 v2.1.3
	github.com/onsi/gomega v1.18.1
	github.com/robfig/cron/v3 v3.0.0
	golang.org/x/tools v0.1.7 // indirect
	google.golang.org/api v0.20.0
	gopkg.in/yaml.v3 v3.0.1
	k8s.io/api v0.20.2
	k8s.io/apimachinery v0.20.2
	k8s.io/client-go v0.20.2
	k8s.io/utils v0.0.0-20210111153108-fddb29f9d009
	sigs.k8s.io/controller-runtime v0.8.3
)

// to fix vulnerability: CVE-2021-3121 in github.com/gogo/protobuf < v1.3.2
replace github.com/gogo/protobuf => github.com/gogo/protobuf v1.3.2

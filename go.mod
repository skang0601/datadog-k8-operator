module github.com/skang0601/datadog-k8s-operator

go 1.13

require (
	github.com/cenkalti/backoff v2.2.1+incompatible // indirect
	github.com/go-logr/logr v0.1.0
	github.com/onsi/ginkgo v1.11.0
	github.com/onsi/gomega v1.8.1
	github.com/zorkian/go-datadog-api v2.29.0+incompatible
	k8s.io/apimachinery v0.17.2
	k8s.io/client-go v0.17.2
	sigs.k8s.io/controller-runtime v0.5.0
)

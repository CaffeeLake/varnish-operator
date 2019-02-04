// Generate some defaults for VarnishService
//go:generate go run ../../../../vendor/k8s.io/code-generator/cmd/defaulter-gen/main.go -O zz_generated.defaults -i . -h ../../../../hack/boilerplate.go.txt

package v1alpha1

import (
	"fmt"
	"icm-varnish-k8s-operator/pkg/varnishservice/config"
	"math"
	"regexp"
	"sort"

	"k8s.io/apimachinery/pkg/util/intstr"

	"k8s.io/api/core/v1"
	kv1 "k8s.io/kubernetes/pkg/apis/core/v1"
)

var operatorCfg *config.Config

var varnishArgsKeyRegexp = regexp.MustCompile("-\\w")

// Init is used to inject config to the package. Autogeneration requires to have functions with only one argument
// of specific type so that makes not possible to have structs where you can inject configs or other dependencies.
// Needs to be called before package usage.
func Init(cfg *config.Config) {
	operatorCfg = cfg
}

// SetDefaults_VarnishService sets defaults for everything inside VarnishService.
// Normally, this would be handled by the generator code, but in this case, the order that things are defaulted matters,
// since some defaults depend on other values already being defaulted.
// Therefore, add the `defaulter-gen=covers` flag which tells the generator to ignore any recursive `SetDefaults` functions,
// meaning they must be included here instead. With this, the order of defaults can be controlled
// +k8s:defaulter-gen=covers
func SetDefaults_VarnishService(in *VarnishService) {
	SetDefaults_VarnishServiceSpec(&in.Spec)
	SetDefaults_VarnishVCLConfigMap(&in.Spec.VCLConfigMap)
	SetDefaults_VarnishDeployment(&in.Spec.Deployment)
	SetDefaults_VarnishContainer(&in.Spec.Deployment.Container)
	SetDefaults_VarnishServiceService(&in.Spec.Service)

	// setVarnishArgs must go last, since it depends on defaults set above
	// this should actually go into the mutating webhook, but it cannot go there due to a bug that was fixed in k8s 1.12+. see /pkg/varnishservice/webhooks/webhooks.go
	setVarnishArgs(&in.Spec)
}

func SetDefaults_VarnishServiceSpec(in *VarnishServiceSpec) {
	if in.LogLevel == "" {
		in.LogLevel = operatorCfg.OperatorLogLevel.String()
	}
	if in.LogFormat == "" {
		in.LogFormat = operatorCfg.OperatorLogFormat
	}
}

func SetDefaults_VarnishVCLConfigMap(in *VarnishVCLConfigMap) {
	if in.BackendsFile == "" {
		in.BackendsFile = operatorCfg.VclConfigMapBackendsFile
	}
	if in.DefaultFile == "" {
		in.DefaultFile = operatorCfg.VclConfigMapDefaultFile
	}
	in.BackendsTmplFile = in.BackendsFile + ".tmpl"
}

func SetDefaults_VarnishDeployment(in *VarnishDeployment) {
	if in.Replicas == nil {
		in.Replicas = &operatorCfg.DeploymentReplicas
	}
}

func SetDefaults_VarnishContainer(in *VarnishContainer) {
	if in.Image == "" {
		in.Image = operatorCfg.DeploymentContainerImage
	}
	if in.ImagePullPolicy == nil {
		in.ImagePullPolicy = &operatorCfg.DeploymentContainerImagePullPolicy
	}
	if in.RestartPolicy == "" {
		in.RestartPolicy = operatorCfg.DeploymentContainerRestartPolicy
	}
	if in.Resources == nil {
		operatorCfg.DeploymentContainerResources.DeepCopyInto(in.Resources)
	}
	if in.ImagePullSecret == nil {
		in.ImagePullSecret = &operatorCfg.DeploymentContainerImagePullSecret
	}
	if in.LivenessProbe == nil && operatorCfg.DeploymentContainerLivenessProbe != nil {
		operatorCfg.DeploymentContainerLivenessProbe.DeepCopyInto(in.LivenessProbe)
	}
	if in.ReadinessProbe == nil && operatorCfg.DeploymentContainerReadinessProbe != nil {
		operatorCfg.DeploymentContainerReadinessProbe.DeepCopyInto(in.ReadinessProbe)
	}
}

func SetDefaults_VarnishServiceService(in *VarnishServiceService) {
	s := &v1.Service{Spec: in.ServiceSpec}
	kv1.SetObjectDefaults_Service(s)
	s.Spec.DeepCopyInto(&in.ServiceSpec)

	if in.VarnishPort.Name == "" {
		in.VarnishPort.Name = "varnish"
	}
	if in.VarnishPort.TargetPort == (intstr.IntOrString{}) {
		in.VarnishPort.TargetPort = intstr.IntOrString{
			Type:   intstr.Int,
			IntVal: in.VarnishPort.Port,
		}
	}
	if in.VarnishExporterPort.Name == "" {
		in.VarnishExporterPort.Name = "varnishexporter"
	}
	if in.VarnishExporterPort.TargetPort == (intstr.IntOrString{}) {
		in.VarnishExporterPort.TargetPort = intstr.IntOrString{
			Type:   intstr.Int,
			IntVal: in.VarnishExporterPort.Port,
		}
	}
	// if the user tried to stick the named ports in the .ports array, remove them
	for idx, port := range in.Ports {
		if port.Name == in.VarnishPort.Name ||
			port.Port == in.VarnishPort.Port ||
			port.TargetPort == in.VarnishPort.TargetPort ||
			port.Name == in.VarnishExporterPort.Name ||
			port.Port == in.VarnishExporterPort.Port ||
			port.TargetPort == in.VarnishExporterPort.TargetPort {

			in.Ports[idx] = in.Ports[len(in.Ports)-1]
			in.Ports = in.Ports[:len(in.Ports)-1]
		}
	}

	in.Ports = append(in.Ports, *in.VarnishPort, *in.VarnishExporterPort)

	if in.PrometheusAnnotations == nil {
		in.PrometheusAnnotations = &operatorCfg.ServicePrometheusAnnotations
	}
}

// due to validating webhook, we can assume args are properly formed (in key/value pairs, with optional value) and there are no override args present in the list
func setVarnishArgs(in *VarnishServiceSpec) {
	varnishArgsOverrides := []string{
		"-F",
		"-a", fmt.Sprintf("0.0.0.0:%d", in.Service.VarnishPort.Port),
		"-S", "/etc/varnish/secret",
		"-f", "/etc/varnish/" + in.VCLConfigMap.DefaultFile,
	}
	varnishArgsDefaults := map[string][]string{
		"-s": {"-s", fmt.Sprintf("malloc,%dM", (int64)((float64)(in.Deployment.Container.Resources.Limits.Memory().Value())*.9/math.Pow(2, 20)))},
		"-p": {"-p", "default_ttl=3600", "-p", "default_grace=3600"},
		"-T": {"-T", "127.0.0.1:6082"},
	}
	sortedDefaults := func() []string {
		var unsorted [][]string
		for _, args := range varnishArgsDefaults {
			unsorted = append(unsorted, args)
		}
		sort.SliceStable(unsorted, func(i, j int) bool {
			return unsorted[i][0] < unsorted[j][0]
		})
		var out []string
		for _, arg := range unsorted {
			out = append(out, arg...)
		}
		return out
	}

	rawArgs := in.Deployment.Container.VarnishArgs
	var parsedArgs [][]string

	for i := 0; i < len(rawArgs); {
		var nextArg []string
		// if arg has a default, remove the default
		delete(varnishArgsDefaults, rawArgs[i])

		// add arg key to output
		nextArg = append(nextArg, rawArgs[i])
		i++
		// if there is an arg value (as defined by NOT being a key), add it to output
		if i < len(rawArgs) && !varnishArgsKeyRegexp.MatchString(rawArgs[i]) {
			nextArg = append(nextArg, rawArgs[i])
			i++
		}
		parsedArgs = append(parsedArgs, nextArg)
	}

	sort.SliceStable(parsedArgs, func(i, j int) bool {
		return parsedArgs[i][0] < parsedArgs[j][0]
	})

	var out []string

	// append sorted args
	for _, parsed := range parsedArgs {
		out = append(out, parsed...)
	}

	// append any remaining defaults, in sorted order
	out = append(out, sortedDefaults()...)

	// append all overrides
	out = append(out, varnishArgsOverrides...)

	in.Deployment.Container.VarnishArgs = out
}

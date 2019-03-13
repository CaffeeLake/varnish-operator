package webhooks

import (
	"icm-varnish-k8s-operator/pkg/apis/icm/v1alpha1"
	"icm-varnish-k8s-operator/pkg/logger"
	"icm-varnish-k8s-operator/pkg/varnishservice/config"

	"github.com/pkg/errors"

	admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission/builder"
)

// InstallWebhooks creates webhooks to be used with the VarnishService CRD
// +kubebuilder:rbac:groups=admissionregistration.k8s.io,resources=mutatingwebhookconfigurations;validatingwebhookconfigurations,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
func InstallWebhooks(mgr manager.Manager, cfg *config.Config, logr *logger.Logger) error {
	validatingWebhook, err := builder.NewWebhookBuilder().
		Name("validating-webhook.varnish-operator.icm.ibm.com").
		Validating().
		Operations(admissionregistrationv1beta1.Create, admissionregistrationv1beta1.Update).
		WithManager(mgr).
		ForType(&v1alpha1.VarnishService{}).
		Handlers(&validationWebhook{logger: logr}).
		FailurePolicy(admissionregistrationv1beta1.Ignore). //change to Fail for debugging
		Build()

	if err != nil {
		return errors.Wrap(err, "Can't create validating webhook")
	}

	err = validatingWebhook.Validate()
	if err != nil {
		return errors.Wrap(err, "Invalid validating webhook")
	}

	mutatingWebhook, err := builder.NewWebhookBuilder().
		Name("mutating-webhook.varnish-operator.icm.ibm.com").
		Mutating().
		Operations(admissionregistrationv1beta1.Create, admissionregistrationv1beta1.Update).
		WithManager(mgr).
		ForType(&v1alpha1.VarnishService{}).
		Handlers(&mutatingWebhook{scheme: mgr.GetScheme(), logger: logr}).
		FailurePolicy(admissionregistrationv1beta1.Ignore). //change to Fail for debugging
		Build()

	if err != nil {
		return errors.Wrap(err, "Can't create mutating webhook")
	}

	err = mutatingWebhook.Validate()
	if err != nil {
		return errors.Wrap(err, "Invalid mutating webhook")
	}

	srv, err := webhook.NewServer("varnish-operator-webhook-server", mgr, webhook.ServerOptions{
		Port:    cfg.ContainerWebhookPort,
		CertDir: "/tmp/varnish-operator/webhook/certs",
		BootstrapOptions: &webhook.BootstrapOptions{
			ValidatingWebhookConfigName: "varnish-operator-validating-webhook-config",
			//MutatingWebhookConfigName:   "varnish-operator-mutating-webhook-config",
			Service: &webhook.Service{
				Namespace: cfg.Namespace,
				Name:      "varnish-operator-webhook-service",
				// Selectors should select the pods that runs this webhook server.
				Selectors: map[string]string{
					"admission-controller": "varnish-service-admission-controller",
				},
			},
		},
	})

	if err != nil {
		return errors.Wrap(err, "Can't create validating webhook server")
	}

	// The mutating webhook is disabled due to a bug in kubernetes 1.11.
	// It leaded to errors like this (shortened):
	// Internal error occurred: jsonpatch replace operation does not apply: doc is missing key: /spec/service/ports/0/targetPort
	// It was caused by the mutating webhook that was setting default values for the service.
	// For now the defaults setting is happening in the reconcile loop until we decide to drop Kubernetes 1.11 support.
	// You can use mutating webhooks for different logic and also safely use the validating webhook functions if you need.
	// To do so, just uncomment the webhooks registering below and make sure you run the server not in Dryrun mode.
	//err = srv.Register(validatingWebhook, mutatingWebhook)
	err = srv.Register(validatingWebhook)
	if err != nil {
		return errors.Wrap(err, "Can't register validating webhook in the admission server")
	}

	return nil
}

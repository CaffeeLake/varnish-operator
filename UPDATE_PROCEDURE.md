# Update Procedure

Steps needed to update varnish operator:

* change version.txt
* change varnish-operator/values.yaml#container.image
* change varnish-operator/Chart.yaml#appVersion
* change varnish-operator/Chart.yaml#version
* change config/samples/icm_v1alpha1_varnishcluster.yaml#spec.varnish.image

Sure wish all of this was automated...

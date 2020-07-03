## Datdog-K8s-Operator

### Monitoring
This operator allows for the creation of specific DataDog monitors through a K8s Operator.

The creation/deletion of the monitors are tracked completely by the CRDs.

Specific monitors

```yaml
apiVersion: v1alpha1
kind: DatadogMonitor
metadata:
  name: datadogmonitor-sample
spec:
  # Add fields here
  name: "This operator is doing things"
  query: "avg(last_5m):avg:kubernetes.pods.running{env:test} > 1"
  message: "This was created by an operator"
  type: "metric alert"
  tags:
    - test
  options:
    locked: false
    new_host_delay: 300
    require_full_window: true
    thresholds:
      critical: "1"
```

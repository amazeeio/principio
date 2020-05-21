# Principio

When a namespace is created, it will serch the cluster for `InitConfig` kinds within the cluster.

It will apply the items in `spec.initItems` to the namespace if the labels in `spec.initLabels` match or not.

# Spec
```
kind: InitConfig
apiVersion: init.amazee.io/v1alpha1
metadata:
  name: my-custom-service
  namespace: principio-controller
spec:
  initLabels:
    - key: principio.amazee.io/no-custom-service
      operator: DoesNotExist
  initItems:
    - apiVersion: v1
      kind: Service
      metadata:
        name: custom-service
      spec:
        clusterIP: None
        ports:
          - port: 1
            protocol: TCP
            targetPort: 1
        type: ClusterIP
    - apiVersion: v1
      kind: Endpoints
      metadata:
        name: custom-service
      subsets:
        - addresses:
            - ip: 10.1.2.3
            - ip: 10.1.2.4
            - ip: 10.1.2.5
          ports:
            - port: 1
              protocol: TCP

```

# Notes

This will only work if a namespace is created with the labels as this controller only watches for the initial creation of a namespace.

```
apiVersion: v1
kind: Namespace
metadata:
  name: my-custom-namespace
  labels:
    principio.amazee.io/no-custom-service: 'true'
```
# Principio

When a namespace is created, it will serch the cluster for `InitConfig` kinds within the cluster.

It will apply the items in `spec.initItems` to the namespace if the labels in `spec.initLabels` match or not.

# Install

Set your kube context to the cluster you want to install on, then run the following
```
make IMG=amazeeio/principio:v0.0.1 deploy
make install
```
This will create `principio-system` namespace in the cluster and set up the controller.

## Notes
> Initially it will reconcile all namespaces, but with no config it will not act on them.
> Once config is added, it will only apply to newly created namespaces.
> If the controller re-starts, it will reconcile all namespaces again, but this time it will act on them if it needs to.

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

Based on the example `Spec` above, the following namespace **WILL NOT** get the `custom-service`
```
apiVersion: v1
kind: Namespace
metadata:
  name: my-custom-namespace
  labels:
    principio.amazee.io/no-custom-service: 'true'
```

Based on the example `Spec` above, the following namespace **WILL** get the `custom-service`
```
apiVersion: v1
kind: Namespace
metadata:
  name: my-other-custom-namespace
```
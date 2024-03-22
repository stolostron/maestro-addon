# maestro-addon

The maestro addon is used to enable the maestro in the ACM

## Build

```
make build
```

## Image

```
make image
```

## Deploy

1. deploy maestro-addon manager on the hub
```
helm install maestro-addon ./charts/maestro-addon
```

2. create a `ManagedClusterAddOn` for one managed cluster on the hub

```
cat << EOF | kubectl -n <cluster-name> apply -f
apiVersion: addon.open-cluster-management.io/v1alpha1
kind: ManagedClusterAddOn
metadata:
  name: maestro-addon
spec: {}
EOF
```

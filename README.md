# maestro-addon

The maestro addon is used to enable the maestro in the ACM

## Build

### Build binary

```
make build
```

### Build Image

```
make image
```

## Deploy

1. deploy maestro-addon manager on the hub

  ```
  helm install maestro-addon ./charts/maestro-addon
  ```

  - To use the customized images:
    ```
    --set global.imageOverrides.maestroImage=<your customized image>
    --set global.imageOverrides.maestroAddOnImage=<your customized image>
    ```

  - To specify the AMQ Streams Kafka cluster name and namespace:
    ```
    --set messageQueue.amqStreams.name=<your AMQ Streams Kafka CR name>
    --set messageQueue.amqStreams.namespace=<your AMQ Streams Kafka CR namespace>
    ```

  More available config values can be found from [here](charts/maestro-addon/values.yaml).

2. create a `ManagedClusterAddOn` in one managed cluster namespace on the hub to install the maestro-addon agent on this managed cluster

  ```
  cat << EOF | kubectl -n <cluster-name> apply -f -
  apiVersion: addon.open-cluster-management.io/v1alpha1
  kind: ManagedClusterAddOn
  metadata:
    name: maestro-addon
  spec: {}
  EOF
  ```

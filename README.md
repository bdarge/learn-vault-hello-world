# hello-vault

A sample app to test Vault KV engine with root token and k8s authentication

## Local Development

This application is built to run in a cluster and not locally. It would take
some additional changes to have it work locally.

### Run a local vault server
```shell
$ vault server -dev -dev-root-token-id=root
$ vault kv put secret/webapp/config username="static-user" password="static-password"
```

### Run the local app server
```shell
$ export VAULT_ADDR="http://localhost:8200"
$ go run *.go
```

## Docker Image

Build and push the Docker image. (multi-arch)

```shell
$ docker buildx build --push --platform linux/amd64,linux/arm64,linux/arm/v7 -t USERNAME/hello-vault:k8s .
```

## Load it into Kubernetes

The assumption is Kubernetes & Vault are configured correctly. Update the configuration file to use your Docker image.

Apply the configuration that describes the hello-vault pod.

```shell
$ kubectl apply -f hello-vault.yaml
```

Check the logs of the server.

```shell
$ kubectl logs hello-vault-e54r445b4c-psdlk
```

Login to the instance.

```shell
$ kubectl exec -it hello-vault-e54r445b4c-psdlk /bin/bash
```
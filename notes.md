# Notes

```sh

az connectedk8s show --resource-group $RESOURCE_GROUP --name $CLUSTER_NAME  --query "{oidcIssuerEnabled:oidcIssuerProfile.enabled, workloadIdentityEnabled: securityProfile.workloadIdentity.enabled}"

az connectedk8s update --resource-group $RESOURCE_GROUP --name $CLUSTER_NAME --enable-oidc-issuer --enable-workload-identity 


sudo nano /etc/rancher/k3s/config.yaml


```

file contenst:

```yaml
kube-apiserver-arg:
 - service-account-issuer=https://oidcdiscovery-europe-endpoint-bygcg7c6dafqeqca.z01.azurefd.net/88ad2f8c-7074-4c35-8c14-e263f24401b3/
 - service-account-max-token-expiration=24h 
```


k3d
```sh
k3d config init 
# change file
k3d config migrate k3d-default.yaml 
```


k3d file contents:
```yaml
---
apiVersion: k3d.io/v1alpha5
kind: Simple
metadata:
  name: k3s-default
servers: 1
agents: 0
image: docker.io/rancher/k3s:v1.30.4-k3s1
          - http://my.company.registry:5000
options:
  k3s: # options passed on to K3s itself
    extraArgs: # additional arguments passed to the `k3s server|agent` command; same as `--k3s-arg`
      - arg: --kube-apiserver-arg=service-account-issuer=https://oidcdiscovery-europe-endpoint-bygcg7c6dafqeqca.z01.azurefd.net/88ad2f8c-7074-4c35-8c14-e263f24401b3/
      - arg: --kube-apiserver-arg=service-account-max-token-expiration=24h
```




create secrets
```sh
KUBERNETES_NAMESPACE="workloads"
SERVICE_ACCOUNT_NAME=aio-ssc-sa # created by az iot ops secretsync enable

kubectl create ns ${KUBERNETES_NAMESPACE}
cat <<EOF | kubectl apply -f -
  apiVersion: v1
  kind: ServiceAccount
  metadata:
    name: ${SERVICE_ACCOUNT_NAME}
    namespace: ${KUBERNETES_NAMESPACE}
EOF

FEDERATED_IDENTITY_CREDENTIAL_NAME="workloadFedIdentity"
az identity federated-credential create --name ${FEDERATED_IDENTITY_CREDENTIAL_NAME} --identity-name ${USER_ASSIGNED_MI_NAME} --resource-group ${RESOURCE_GROUP} --issuer ${SERVICE_ACCOUNT_ISSUER} --subject system:serviceaccount:${KUBERNETES_NAMESPACE}:${SERVICE_ACCOUNT_NAME}





export USER_ASSIGNED_CLIENT_ID="$(az identity show --resource-group "${RESOURCE_GROUP}" --name "${USER_ASSIGNED_MI_NAME}" --query 'clientId' -otsv)"
export AZURE_TENANT_ID="$(az account show | jq -r '.tenantId')"
```

get secret in aio namespace (much fucking easier):

```sh
export KUBERNETES_NAMESPACE="azure-iot-operations"
export USER_ASSIGNED_CLIENT_ID="$(az identity show --resource-group "${RESOURCE_GROUP}" --name "${USER_ASSIGNED_MI_NAME}" --query 'clientId' -otsv)"
export AZURE_TENANT_ID="$(az account show | jq -r '.tenantId')"
export KEYVAULT_SECRET_NAME="my-test-secret"


cat <<EOF > spc.yaml
apiVersion: secrets-store.csi.x-k8s.io/v1
kind: SecretProviderClass
metadata:
  name: secret-provider-class-name                      # Name of the class; must be unique per Kubernetes namespace
  namespace: ${KUBERNETES_NAMESPACE}                    # Kubernetes namespace to make the secrets accessible in
spec:
  provider: azure
  parameters:
    clientID: "${USER_ASSIGNED_CLIENT_ID}"               # Managed Identity Client ID for accessing the Azure Key Vault with.
    keyvaultName: ${KEYVAULT_NAME}                       # The name of the Azure Key Vault to synchronize secrets from.
    objects: |
      array:
        - |
          objectName: ${KEYVAULT_SECRET_NAME}            # The name of the secret to sychronize.
          objectType: secret
          objectVersionHistory: 2                       # [optional] The number of versions to synchronize, starting from latest.      
    tenantID: "${AZURE_TENANT_ID}"                       # The tenant ID of the Key Vault 
EOF

SECRET_PROVIDER_CLASS_NAME=secret-provider-class-name

SERVICE_ACCOUNT_NAME=aio-ssc-sa # created by az iot ops secretsync enable


cat <<EOF > ss.yaml
apiVersion: secret-sync.x-k8s.io/v1alpha1
kind: SecretSync
metadata:
  name: secret-sync-name                                  # Name of the object; must be unique per Kubernetes namespace
  namespace: ${KUBERNETES_NAMESPACE}                      # Kubernetes namespace
spec:
  serviceAccountName: ${SERVICE_ACCOUNT_NAME}             # The Kubernetes service account to be given permissions to access the secret.
  secretProviderClassName: ${SECRET_PROVIDER_CLASS_NAME}  # The name of the matching SecretProviderClass with the configuration to access the AKV storing this secret
  secretObject:
    type: Opaque
    data:
    - sourcePath: ${KEYVAULT_SECRET_NAME}/0                # Name of the secret in Azure Key Vault with an optional version number (defaults to latest)
      targetKey: ${KEYVAULT_SECRET_NAME}-data-key0         # Target name of the secret in the Kubernetes Secret Store (must be unique)
    - sourcePath: ${KEYVAULT_SECRET_NAME}/1                # [optional] Next version of the AKV secret. Note that versions of the secret must match the configured objectVersionHistory in the secrets provider class 
      targetKey: ${KEYVAULT_SECRET_NAME}-data-key1         # [optional] Next target name of the secret in the K8s Secret Store
EOF

kubectl apply -f ./spc.yaml
kubectl apply -f ./ss.yaml
```
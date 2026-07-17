param aksName string = 'cas-test'
param acrName string = 'castestacr'
param location string = resourceGroup().location
param dnsPrefix string = aksName
param vmSize string = 'Standard_DS2_v2'
param casUserAssignedIdentityName string = 'cas-msi'
param casFederatedCredenatialName string = 'cas-federated-credential'
param casNamespace string = 'kube-system'

resource aks 'Microsoft.ContainerService/managedClusters@2023-11-01' = {
  location: location
  name: aksName
  identity: { type: 'SystemAssigned' }  // --enable-managed-identity

  properties: {
    dnsPrefix: dnsPrefix
    oidcIssuerProfile: { enabled: true }  // --enable-oidc-issuer
    securityProfile: {
       workloadIdentity: { enabled: true } // --enable-workload-identity 
    } 
    agentPoolProfiles: [
      {
        count: 1
        mode: 'System'
        name: 'nodepool1'
        type: 'VirtualMachineScaleSets'
        vmSize: vmSize
      }
      {
        count: 3
        mode: 'User'
        name: 'nodepool2'
        type: 'VirtualMachineScaleSets'
        vmSize: vmSize
      }
    ]
    networkProfile: {
      networkPlugin: 'azure'
      networkPluginMode: 'overlay'
    } 
  }
}

resource acr 'Microsoft.ContainerRegistry/registries@2023-07-01' = {
  location: location
  name: acrName
  sku: { name: 'Basic' }
}

var AcrPull = subscriptionResourceId('Microsoft.Authorization/roleDefinition', '7f951dda-4ed3-4680-a7ca-43fe172d538d')

@description('AKS can pull images from ACR')
resource aksAcrPull 'Microsoft.Authorization/roleAssignments@2022-04-01' = {
  name: guid(resourceGroup().id, acr.name, aks.name, AcrPull)
  scope: acr
  properties: {
    principalId: aks.properties.identityProfile.kubeletidentity.objectId
    principalType: 'ServicePrincipal'
    roleDefinitionId: AcrPull
  }
}

@description('CAS user assigned identity')
resource casUserAssignedIdentity 'Microsoft.ManagedIdentity/userAssignedIdentities@2023-01-31' = {
  location: location
  name: casUserAssignedIdentityName
  resource caseFederatedCredential 'federatedIdentityCredentials' = {
    name: casFederatedCredenatialName
    properties: {
      issuer: aks.properties.oidcIssuerProfile.issuerURL
      subject: 'system:serviceaccount:${casNamespace}:cluster-autoscaler' // TODO: parameterize namespace
      audiences: ['api://AzureADTokenExchange']
    }
  }
}

output acrName string = acr.name
output aksName string = aks.name
output nodeResourceGroup string = aks.properties.nodeResourceGroup
output userNodePoolName string = aks.properties.agentPoolProfiles[1].name
output casUserAssignedIdentityPrincipal string = casUserAssignedIdentity.properties.principalId
output casUserAssignedIdentityClientId string = casUserAssignedIdentity.properties.clientId

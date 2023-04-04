# gridscale Go Client Library

[![GitHub tag (latest SemVer)](https://img.shields.io/github/v/tag/gridscale/gsclient-go?label=release)](https://github.com/gridscale/gsclient-go/releases) [![PkgGoDev](https://pkg.go.dev/badge/github.com/gridscale/gsclient-go/v3)](https://pkg.go.dev/github.com/gridscale/gsclient-go/v3)
[![gsclient-go status](https://github.com/gridscale/gsclient-go/workflows/Go/badge.svg)](https://github.com/gridscale/gsclient-go/actions)

This is a client for the gridscale API. It can be used to make an application interact with the gridscale cloud platform to create and manage resources.

## Prerequisites

To be able to use this client a number of steps need to be taken. First a gridscale account will be required, which can be created [here](https://my.gridscale.io/signup/). Then an API-token [should be created](https://my.gridscale.io/APIs/).

## Installation

First the Go programming language will need to be installed. This can be done by using [the official Go installation guide](https://golang.org/doc/install) or by using the packages provided by your distribution.

Downloading the gridscale Go client can be done with the following go command:

    $ go get github.com/gridscale/gsclient-go/v3

## Using the Client

To be able to use the gridscale Go client in an application it can be imported in a go file. This can be done with the following code:

```go
import "github.com/gridscale/gsclient-go/v3"
```

To get access to the functions of the Go client, a Client type needs to be created. This requires a Config type. Both of these can be created with the following code:

```go
//Using default config
config := gsclient.DefaultConfiguration("User-UUID", "API-token")

//OR Custom config
config := gsclient.NewConfiguration(
            "API-URL",
            "User-UUID",
            "API-token",
            false, //Set debug mode
            true, //Set sync mode
            500, //Delay (in milliseconds) between requests (or retry 503 error code)
            100, //Maximum number of retries when server returns 503 error code
        )
client := gsclient.NewClient(config)
```

To trace the duration of individual client calls, set logger to `Trace` level via `gsclient.SetLogLevel()` function. Other log levels: https://github.com/sirupsen/logrus#level-logging
```go
gsclient.SetLogLevel(logrus.TraceLevel)
```

`Trace` message looks like following:
```
TRAC[2021-03-12T10:32:43+01:00] Successful method="github.com/gridscale/gsclient-go/v3.(*Client).GetServer" requestUUID=035fc625-199d-41da-93c4-f32502d101c1 timeMs=350
```

Make sure to replace the user-UUID and API-token strings with valid credentials or variables containing valid credentials. It is recommended to use environment variables for them.

## Using API endpoints

***Note: `context` has to be passed to all APIs of `gsclient-go` as the first parameter. In case you want to set timeout for a specific operation, you can pass a context with timeout (via `context.WithTimeout` or `context.WithDeadline`)

After having created a Client type, as shown above, it will be possible to interact with the API. An example would be the [Servers Get endpoint](https://gridscale.io/en/api-documentation/index.html#servers-get):

```go
ctx := context.Background()
servers := client.GetServerList(ctx)
```

For creating and updating/patching objects in gridscale, it will be required to use the respective CreateRequest and UpdateRequest types. For creating an IP that would be IPCreateRequest and IPUpdateRequest. Here an example:

```go
ctx := context.Background()
requestBody := gsclient.IPCreateRequest{
    Name:       "IPTest",
    Family:     gsclient.IPv6Type,
    Failover:   false,
    ReverseDNS: "my-reverse-dns-entry.tld",
    Labels:     []string{"MyLabel"},
}

client.CreateIP(ctx, requestBody)
```

For updating/scaling server resources you could use:

```go
myServerUuid := "[Server UUID]"
backgroundContext := context.Background()

// No hotplug available for scaling resources down, shutdown server first via ACPI
shutdownErr := client.ShutdownServer(backgroundContext, myServerUuid)
if shutdownErr != nil{
    log.Error("Shutdown server failed", shutdownErr)
    return
}

// Update servers resources
requestBody := gsclient.ServerUpdateRequest{
    Memory:          12,
    Cores:           4,
}

updateErr := client.UpdateServer(backgroundContext, myServerUuid, requestBody)
if updateErr != nil{
    log.Error("Serverupdate failed", updateErr)
    return
}

// Start server again
poweronErr := client.StartServer(backgroundContext, myServerUuid)
if poweronErr != nil{
    log.Error("Start server failed", poweronErr)
    return
}
```

What options are available for each create and update request can be found in the source code. After installing it should be located in `$GOPATH/src/github.com/gridscale/gsclient-go`.

## Examples

Examples on how to use each resource can be found in the examples folder:

* Firewall (firewall.go)
* IP (ip.go)
* ISO-image (isoimage.go)
* Loadbalancer (loadbalancer.go)
* Network (network.go)
* Object Storage (objectstorage.go)
* PaaS service (paas.go)
* Server (server.go)
* Storage (storage.go)
* Storage snapshot (snapshot.go)
* Storage snapshot schedule (snapshotschedule.go)
* SSH-key (sshkey.go)
* Template (template.go)

## Implemented API Endpoints

Not all endpoints have been implemented in this client, but new ones will be added in the future. Here is the current list of implemented endpoints and their respective function written like endpoint (function):

* Servers
  * Servers Get (GetServerList)
  * Server Get (GetServer)
  * Server Create (CreateServer)
  * Server Patch (UpdateServer)
  * Server Delete (DeleteServer)
  * Server Events Get (GetServerEventList)
  * Server Metrics Get (GetServerMetricList)
  * ACPI Shutdown (ShutdownServer) *NOTE: ShutdownServer() will not run StopServer() when it fails to shutdown a server*
  * Server On/Off (StartServer, StopServer)
  * Server's Storages Get (GetServerStorageList)
  * Server's Storage Get (GetServerStorage)
  * Server's Storage Create (CreateServerStorage)
  * Server's Storage Update (UpdateServerStorage)
  * Server's Storage Delete (DeleteServerStorage)
  * Link Storage (LinkStorage)
  * Unlink Storage (UnlinkStorage)
  * Server's Networks Get (GetServerNetworkList)
  * Server's Network Get (GetServerNetwork)
  * Server's Network Create (CreateServerNetwork)
  * Server's Network Update (UpdateServerNetwork)
  * Server's Network Delete (DeleteServerNetwork)
  * Link Network (LinkNetwork)
  * Unlink Network (UnlinkNetwork)
  * Server's IPs Get (GetServerNetworkList)
  * Server's IP Get (GetServerNetwork)
  * Server's IP Create (CreateServerNetwork)
  * Server's IP Update (UpdateServerNetwork)
  * Server's IP Delete (DeleteServerNetwork)
  * Link IP (LinkIP)
  * Unlink IP (UnlinkIP)
  * Server's ISO images Get (GetServerIsoImageList)
  * Server's ISO image Get (GetServerIsoImage)
  * Server's ISO image Create (CreateServerIsoImage)
  * Server's ISO image Update (UpdateServerIsoImage)
  * Server's ISO image Delete (DeleteServerIsoImage)
  * Link ISO image (LinkIsoimage)
  * Unlink ISO image (UnlinkIsoimage)
* Storages
  * Storages Get (GetStorageList)
  * Storage Get (GetStorage)
  * Storage Create (CreateStorage)
  * Storage Create From A Backup (CreateStorageFromBackup)
  * Storage Clone (CloneStorage)
  * Storage Patch (UpdateStorage)
  * Storage Delete (DeleteStorage)
  * Storage's events Get (GetStorageEventList)
* Networks
  * Networks Get (GetNetworkList)
  * Network Get (GetNetwork)
  * Network Create (CreateNetwork)
  * Network Patch (UpdateNetwork)
  * Network Delete (DeleteNetwork)
  * Network Events Get (GetNetworkEventList)
  * (GetNetworkPublic) No official endpoint, but gives the Public Network
* Load balancers
  * LoadBalancers Get (GetLoadBalancerList)
  * LoadBalancer Get (GetLoadBalancer)
  * LoadBalancer Create (CreateLoadBalancer)
  * LoadBalancer Patch (UpdateLoadBalancer)
  * LoadBalancer Delete (DeleteLoadBalancer)
  * LoadBalancerEvents Get (GetLoadBalancerEventList)
* IPs
  * IPs Get (GetIPList)
  * IP Get (GetIP)
  * IP Create (CreateIP)
  * IP Patch (UpdateIP)
  * IP Delete (DeleteIP)
  * IP Events Get (GetIPEventList)
  * IP Version Get (GetIPVersion)
* SSH-Keys
  * SSH-Keys Get (GetSshkeyList)
  * SSH-Key Get (GetSshkey)
  * SSH-Key Create (CreateSshkey)
  * SSH-Key Patch (UpdateSshkey)
  * SSH-Key Delete (DeleteSshkey)
  * SSH-Key's events Get (GetSshkeyEventList)
* Template
  * Templates Get (GetTemplateList)
  * Template Get (GetTemplate)
  * (GetTemplateByName) No official endpoint, but gives a template which matches the exact name given.
  * Template Create (CreateTemplate)
  * Template Update (UpdateTemplate)
  * Template Delete (DeleteTemplate)
  * Template's events Get (GetTemplateEventList)
* PaaS
  * PaaS services Get (GetPaaSServiceList)
  * PaaS service Get (GetPaaSService)
  * PaaS service Create (CreatePaaSService)
  * PaaS service Update (UpdatePaaSService)
  * PaaS service Delete (DeletePaaSService)
  * PaaS service metrics Get (GetPaaSServiceMetrics)
  * PaaS service templates Get (GetPaaSTemplateList)
  * PaaS service security zones Get (GetPaaSSecurityZoneList)
  * Paas service security zone Get (GetPaaSSecurityZone)
  * PaaS service security zone Create (CreatePaaSSecurityZone)
  * PaaS service security zone Update (UpdatePaaSSecurityZone)
  * PaaS service security zone Delete (DeletePaaSSecurityZone)
* ISO Image
  * ISO Images Get (GetISOImageList)
  * ISO Image Get (GetISOImage)
  * ISO Image Create (CreateISOImage)
  * ISO Image Update (UpdateISOImage)
  * ISO Image Delete (DeleteISOImage)
  * ISO Image Events Get (GetISOImageEventList)
* Object Storage
  * Object Storage's Access Keys Get (GetObjectStorageAccessKeyList)
  * Object Storage's Access Key Get (GetObjectStorageAccessKey)
  * Object Storage's Access Key Create (CreateObjectStorageAccessKey)
  * Object Storage's Access Key Delete (DeleteObjectStorageAccessKey)
  * Object Storage's Buckets Get (GetObjectStorageBucketList)
* Storage Snapshot Scheduler
  * Storage Snapshot Schedules Get (GetStorageSnapshotScheduleList)
  * Storage Snapshot Schedule Get (GetStorageSnapshotSchedule)
  * Storage Snapshot Schedule Create (CreateStorageSnapshotSchedule)
  * Storage Snapshot Schedule Update (UpdateStorageSnapshotSchedule)
  * Storage Snapshot Schedule Delete (DeleteStorageSnapshotSchedule)
* Storage Snapshot
  * Storage Snapshots Get (GetStorageSnapshotList)
  * Storage Snapshot Get (GetStorageSnapshot)
  * Storage Snapshot Create (CreateStorageSnapshot)
  * Storage Snapshot Update (UpdateStorageSnapshot)
  * Storage Snapshot Delete (DeleteStorageSnapshot)
  * Storage Rollback (RollbackStorage)
  * Storage Snapshot Export to S3 (ExportStorageSnapshotToS3)
* Storage Backup
  * Storage Backups Get (GetStorageBackupList)
  * Storage Backup Delete (DeleteStorageBackup)
  * Storage Backup Rollback (RollbackStorageBackup)
* Storage Backup Schedule
  * Storage Backup Schedules Get (GetStorageBackupScheduleList)
  * Storage Backup Schedule Get (GetStorageBackupSchedule)
  * Storage Backup Schedule Create (CreateStorageBackupSchedule)
  * Storage Backup Schedule Update (UpdateStorageBackupSchedule)
  * Storage Backup Schedule Delete (DeleteStorageBackupSchedule)
* Firewall
  * Firewalls Get (GetFirewallList)
  * Firewall Get (GetFirewall)
  * Firewall Create (CreateFirewall)
  * Firewall Update (UpdateFirewall)
  * Firewall Delete (DeleteFirewall)
  * Firewall Events Get (GetFirewallEventList)
* Marketplace Application
  * Marketplace Applications Get (GetMarketplaceApplicationList)
  * Marketplace Application Get (GetMarketplaceApplication)
  * Marketplace Application Create (CreateMarketplaceApplication)
  * Marketplace Application Import (ImportMarketplaceApplication)
  * Marketplace Application Update (UpdateMarketplaceApplication)
  * Marketplace Application Delete (DeleteMarketplaceApplication)
  * Marketplace Application Events Get (GetMarketplaceApplicationEventList)
* Event
  * Events Get (GetEventList)
* Label
  * Labels Get (GetLabelList)
* Location
  * Locations Get (GetLocationList)
  * Location Get (GetLocation)
  * Location IPs Get (GetIPsByLocation)
  * Location ISO Images Get (GetISOImagesByLocation)
  * Location Networks Get (GetNetworksByLocation)
  * Location Servers Get (GetServersByLocation)
  * Location Snapshots Get (GetSnapshotsByLocation)
  * Location Storages Get (GetStoragesByLocation)
  * Location Templates Get (GetTemplatesByLocation)
* Deleted
  * Deleted IPs Get (GetDeletedIPs)
  * Deleted ISO Images Get (GetDeletedISOImages)
  * Deleted Networks Get (GetDeletedNetworks)
  * Deleted Servers Get (GetDeletedServers)
  * Deleted Snapshots Get (GetDeletedSnapshots)
  * Deleted Storages Get (GetDeletedStorages)
  * Deleted Templates Get (GetDeletedTemplates)
  * Deleted PaaS Services Get (GetDeletedPaaSServices)
* SSL certificate
  * SSL certificates Get (GetSSLCertificateList)
  * SSL certificate Get (GetSSLCertificate)
  * SSL certificate Create (CreateSSLCertificate)
  * SSL certificate Delete (DeleteSSLCertificate)

Note: The functions in this list can be called with a Client type.

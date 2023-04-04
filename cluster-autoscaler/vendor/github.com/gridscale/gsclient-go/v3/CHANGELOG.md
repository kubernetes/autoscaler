# Changelog

## 3.11.1 (Jan 04, 2023)

BUG FIXES:

* (Hot fix) Omit `HardwareProfileConfig` when it is not set during server creation/update.

## 3.11.0 (Dec 21, 2022)

IMPROVEMENTS:
- Allow to custom hardware profile configuration [#223](https://github.com/gridscale/gsclient-go/pull/223)

## 3.10.1 (Oct 25, 2022)

IMPROVEMENTS:
- Add retry when API backend returns error code 409 [#222](https://github.com/gridscale/gsclient-go/pull/222)

## 3.10.0 (Apr 28, 2022)

IMPROVEMENTS:
- Add `StorageType` to PaaS template properties [#219](https://github.com/gridscale/gsclient-go/pull/219)
- Add retry when API backend returns error code 424, 500 [#218](https://github.com/gridscale/gsclient-go/pull/218)

## 3.9.1 (Apr 6, 2022)

BUG FIXES:

* (Hot fix) Add missing omitempty to dhcp update fields [#217](https://github.com/gridscale/gsclient-go/pull/217).

## 3.9.0 (Apr 5, 2022)

FEATURES:
- Allow to attach PaaS service to a private network [#216](https://github.com/gridscale/gsclient-go/pull/216).

## 3.8.1 (Jan 26, 2022)

IMPROVEMENTS:
- Allow update server's hw profile, according to the official API docs [#215](https://github.com/gridscale/gsclient-go/pull/215).

BUG FIXES:
- Allow to remove DHCP options (except DHCPRange) [#214](https://github.com/gridscale/gsclient-go/pull/214).

## 3.8.0 (Dec 16, 2021)

FEATURES:

* Add network's `pinned_server` API endpoints. Servers in a DHCP-enabled network can be assigned to a specific DHCP IP address ([#207](https://github.com/gridscale/gsclient-go/issues/207)).
* Add missing location API endpoints ([#208](https://github.com/gridscale/gsclient-go/issues/208)).
* Add `usage` API endpoints ([#208](https://github.com/gridscale/gsclient-go/issues/208)).
* Add storage backup location. Backup schedule can be set to run in a specific location ([#210](https://github.com/gridscale/gsclient-go/issues/210)).

## 3.7.0 (Aug 30, 2021)

FEATURES:

* Add DHCP options to network create/update operations ([#206](https://github.com/gridscale/gsclient-go/issues/206)).
* Add DHCP-relating properties to network's properties ([#206](https://github.com/gridscale/gsclient-go/issues/206)).

## 3.6.3 (May 11, 2021)

BUG FIXES:
- Fix an error when requests are read to the end during re-try ([#201](https://github.com/gridscale/gsclient-go/issues/201)).
- Fix handling of retry-able requests 429 and 503 ([#200](https://github.com/gridscale/gsclient-go/issues/200)).

## 3.6.2 (April 20, 2021)

BUG FIXES:
- Fix `PROTOCOL_ERROR` when using moderate amount of concurrency on `gsclient-go`. The issue causes the body of the HTTP2 upgraded request loss during retrial. ([#199](https://github.com/gridscale/gsclient-go/pull/199)).

## 3.6.1 (April 13, 2021)

BUG FIXES:
- Correct `service_template_uuid` field in PaaS update request ([#197](https://github.com/gridscale/gsclient-go/issues/197))

## 3.6.0 (March 30, 2021)

FEATURES:

* Add SSL certificate operations ([#194](https://github.com/gridscale/gsclient-go/issues/194)).

IMPROVEMENTS:

* Add `AutoscalingProperties` field to `AutoscalingProperties` ([#192](https://github.com/gridscale/gsclient-go/issues/192)).

* Add `CertificateUUID` field in loadbalancer, allow to use a custom certificate ([#194](https://github.com/gridscale/gsclient-go/issues/194)).

BUG FIXES:

* Change `LetsencryptSSL` field in loadbalancer from `interface{}` to `*string` ([#194](https://github.com/gridscale/gsclient-go/issues/194)).

* Add DEPRECATED notice to `CurrentPrice` field docs.

## 3.5.0 (March 12, 2021)

This release focuses on performance and documentation fixes.

FEATURES:

* Storage variant now can be set ("distributed", or "local") via method `CreateStorage`.(See [PR #188](https://github.com/gridscale/gsclient-go/pull/188) for details.)

IMPROVEMENTS:

* gsclient-go now allows tracing the duration of individual client calls (See [#186](https://github.com/gridscale/gsclient-go/issues/186)).
* Performance got improved a lot in most situations since gsclient-go won't needlessly delay sending the actual HTTP request anymore. More [here](https://github.com/gridscale/gsclient-go/issues/186#issuecomment-795466954).
* Documentation has just got better (See [#184](https://github.com/gridscale/gsclient-go/issues/184)).

## 3.4.0 (February 8, 2021)

This release catches up on some gridscale.io API additions.

IMPROVEMENTS:

* Add `Initiator` field to `EventProperties` ([#175](https://github.com/gridscale/gsclient-go/issues/175)).
* `PaaSTemplateProperties` also gained new fields, most notably a new `Release` and `Version` field got added ([#174](https://github.com/gridscale/gsclient-go/issues/174)).

## 3.3.2 (January 12, 2021)

IMPROVEMENTS:

* `PaaSServiceTemplateUUID` in PaaS service can be updated ([PR #172](https://github.com/gridscale/gsclient-go/pull/172)).

## 3.3.1 (January 05, 2021)

IMPROVEMENTS:

* Rename function `retryWithLimitedNumOfRetries` to `retryNTimes`.

BUG FIXES:

* Convert type of err in `retryNTimes` to `RequestError` type ([PR #170](https://github.com/gridscale/gsclient-go/pull/170)).

## 3.3.0 (December 17, 2020)

FEATURES:

* Add CreateStorageFromBackup function ([PR #167](https://github.com/gridscale/gsclient-go/pull/167)).

IMPROVEMENTS:

* Remove L3 Security issue from known issues ([Issue #166](https://github.com/gridscale/gsclient-go/issues/166)).
* Various documentation fixes ([PR #158](https://github.com/gridscale/gsclient-go/pull/158)).
* Completely get rid of Travis CI ([PR #163](https://github.com/gridscale/gsclient-go/pull/163)).

BUG FIXES:

* Remove incorrect IP relations ([PR #162](https://github.com/gridscale/gsclient-go/pull/162)).

## 3.2.2 (October 27, 2020)

IMPROVEMENTS:

* Omit `X-Auth-UserID` and `X-Auth-Token` HTTP headers when they are empty.
* Documentation fixes
* StorageBackup and MarketPlaceApplication related schema fixes

## 3.2.1 (September 4, 2020)

BUG FIXES:

* Fixed issue making it unable to get related backups of a backup schedule.

## 3.2.0 (September 2, 2020)

FEATURES:

* Add storage backup functionality.
* Add storage backup schedule's functionality.
* Add marketplace application's functionality.

IMPROVEMENTS:

* Add more examples.
* Added APIs for easy development (can be used for mocking gsclient-go).

BUG FIXES:

* Fixed failed unit tests.
* Added missing fields in some structs.
* Fixed an error occurring when parsing an empty string to a GSTime variable.

## 3.1.0 (June 30, 2020)

FEATURES:

* Storage can be increased its speed via UpdateStorage.
* Support custom HTTP headers.
* Be able to deal with gridscale API rate limiting.

IMPROVEMENTS:

* Refactor `execute()` function.
* Global Logger.

## 3.0.1 (May 19, 2020)

FEATURES:

* Add Kubernetes support (get k8s config, renew k8s cluster's credentials).

BUG FIXES:

* Fixed wrong import path in examples.
* Fixed wrong import path in README.

## 3.0.0 (May 6, 2020)

FEATURES:

* Add storage cloning.

DEPRECATED FEATURES:

* `requestCheckTimeoutSecs` is removed from `NewConfiguration` function.

IMPROVEMENTS:

* Every function (mostly) can now be controlled through context.
* `ShutdownServer` does not run powering off when the server cannot be shut down gracefully. To ungracefully power off a server, the `StopServer` function should be used when `ShutdownServer` fails.
* Reduce size of `vendor` directory by removing unnecessary packages.

## 2.2.2 (April 8, 2020)

DEPRECATED FEATURES:

* Deprecated and removed labels create/delete options.

BUG FIXES:

* Fixed "context is expired but still retrying".
* Fixed some typos.

## 2.2.1 (January 21, 2020)

BUG FIXES:

* (Hot fix) Fix incompatible major version when using gomod due to missing `/v2` suffix in module path of `go.mod` file

## 2.2.0 (January 21, 2020)

IMPROVEMENTS:

* Retry requests in case of network issues (timeouts, connection resets, connection refused, etc)
* Simple requests back-off in case of retrying errors
* Increase defaultDelayIntervalMilliSecs to 1000 to reduce stress on API
* Better variables/functions' names
* Remove `LocationUUID` as objects' location depends on Project's location
* Add gomod

BUG FIXES:

* Fix "cannot remove all labels of an object"
* Fix resource leak due to not closing response's body

## 2.1.0 (November 05, 2019)

IMPROVEMENTS:

* Errors that are from http requests now include request UUIDs
* No need to create structs when exporting snapshots to S3
* Waiting for asynchronous requests is now faster and more memory-friendly

BUG FIXES:

* Fix README
* Fix missing JSON properties

## 2.0.0 (October 07, 2019)

FEATURES:

* Add `sync` mode
* The library now can be fully controlled through `context`
* Auto retry when server returns 5xx and 424 http codes
* Add a default configuration for the client

IMPROVEMENTS:

* Server type is now limited to pre-defined values
* Storage type is now limited to pre-defined values
* IP address family is now limited to pre-defined values
* Load balancer algorithm is now limited to pre-defined values
* All time-related properties are now type of GSTime (a custom type of time.Time)
* Friendly godoc.

BUG FIXES:

* Fixed bugs when un-marshaling JSON to concrete type (mismatched type)

## 1.0.0 (September 05, 2019)

FEATURES:

* Add support for Locations
* Add support for Events
* Add support for Labels
* Add support for Deletes

IMPROVEMENTS:

* Heavily code refactoring to improve code quality
* Achieve 95% test coverage
* Achieve 100% compliant golang code conventions based on goreportcard
* Power-off server if graceful shutdown fails
* Backward compatibility for server creation API

## 0.2.0 (August 23, 2019)

FEATURES:

* Add support for LBaaS (GH-2)
* Add support for PaaS (GH-6)
* Add support for ISO Image Handling (GH-8)
* Add support for Object Storage (GH-11)
* Add support for Snapshots (GH-12) and Snapshot Scheduler (GH-13)
* Add support for Firewall Handling (GH-14)

IMPROVEMENTS:

* Unit Tests for all functionality
* Logging support
* Many examples have been added
* Consistency as well as language styles improved

## 0.1.0 (January 2, 2019)

* Initial release of gsclient-go.

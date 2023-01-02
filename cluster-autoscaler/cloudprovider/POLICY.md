# Cloudprovider policy

As of the moment this policy is written (September 2022) Cluster Autoscaler has
integrations with almost 30 different cloudproviders. At the same time there
are only a handful of core CA maintainers. The maintainers don't have the
capacity to build new integrations or maintain existing ones. In most cases they
also have no experience with particular clouds and no access to a test
environment.

Due to above reasons each integration is required to have a set of OWNERS who
are responsible for development and maintenance of the integration. This
document describes the role and responsibilities of core maintainers and
integration owners. A lot of what is described below has been unofficial
practice for multiple years now, but this policy also introduces some new
requirements for cloudprovider maintenance.

## Responsbilities

Cloudprovider owners are responsible for:

  * Maintaining their integrations.
  * Testing their integrations. Currently any new CA release is tested e2e on
    GCE, testing on other platforms is the responsibility of cloudprovider
    maintainers (note: there is an effort to make automated e2e tests possible
    to run on other providers, so this may improve in the future).
  * Addressing any issues raised in autoscaler github repository related to a
    given provider.
  * Reviewing any pull requests to their cloudprovider.
    * Pull requests that only change cloudprovider code do not require any
      review or approval from core maintainers.
    * Pull requests that change cloudprovider and core code require approval
      from both the cloudprovider owner and core maintainer.

The core maintainers will generally not interfere with cloudprovider
development, but they may take the following actions without seeking approval
from cloudprovider owners:

  * Make trivial changes to cloudproviders when needed to implement changes in
    CA core (ex. updating function signatures when a go interface
    changes).
  * Revert any pull requests that break tests, prevent CA from compiling, etc.
    This includes pull requests adding new providers if they cause the tests to
    start failing or break the rules defined below.

## Adding new cloud provider integration

### External provider

One way to integrate CA with a cloudprovider is to use existing
[External
gRPC](https://github.com/kubernetes/autoscaler/tree/master/cluster-autoscaler/cloudprovider/externalgrpc)
provider. Integrating with gRPC interface may be easier than implementing an
in-tree cloudprovider and the gRPC provider comes with some essential caching
built in.

An external cloudprovider implementation doesn't live in this repository and is
not a part of CA image. As such it is also not a subject to this policy.

### In-tree provider

An alternative to External gRPC provider is an in-tree cloudprovider
integration. An in-tree provider allows more customization (ex. by implementing
[custom processors](https://github.com/kubernetes/autoscaler/tree/master/cluster-autoscaler/processors)
that integrate with a specific provider), but it requires significantly more effort to
implement and maintain.

In order to add new in-tree integration you need to open a pull request implementing
the interfaces defined in cloud\_provider.go. This policy requires that any new
in-tree cloudprovider follows the following rules:

  * Cloudprovider needs to have an OWNERS file that lists its maintainers.
    Kubernetes policy requires that code OWNERS are members of the Kubernetes
    organization.
    * It is required that both reviewers and approvers sections of OWNERS file
      are non-empty.
    * This can create a chicken and egg problem, where adding a cloudprovider
      requires being a member of Kubernetes org and becoming a member of the
      organization requires a history of code contributions. For this reason it
      is allowed for the OWNERS file to temporarily contain commented out github
      handles. There is an expectation that at least some of the owners will
      join Kubernetes organization (by following the
      [process](https://github.com/kubernetes/community/blob/master/community-membership.md))
      within one release cycly, so that they can approve PRs to their
      cloudprovider.
  * Cloudprovider shouldn't introduce new dependencies (such as clients/SDKs)
    to top-level go.mod vendor, unless those dependencies are already imported
    by kubernetes/kubernetes repository and the same version of the library is
    used by CA and Kubernetes. This requirement is mainly driven by
    the problems with version conflicts in transitive dependencies we've
    experienced in the past.
    * Cloudproviders are welcome to carry their dependencies inside their
      directories as needed.

Note: Any functions in cloud\_provider.go marked as 'Implementation optional'
may be left unimplemented. Those functions provide additional functionality, but
are not critical. To leave a function unimplemented just have it return
cloudprovider.ErrNotImplemented.

## Cloudprovider maintenance requirements

In order to allow code changes to Cluster Autoscaler that would require
non-trivial changes in cloudproviders this policy introduces _Cloudprovider
maintenance request_ (CMR) mechanism.

 * CMR will be issued via a github issue tagging all
   cloudprovider owners and describing the problem being solved and the changes
   requested.
 * CMR will clearly state the minor version in which the changes are expected
   (ex. 1.26).
 * CMR will need to be discussed on sig-autoscaling meeting and approved by
   sig leads before being issued. It will also be announced on sig-autoscaling
   slack channel and highlited in sig-autoscaling meeting notes.
 * A CMR may be issued no later then [enhancements
   freeze](https://github.com/kubernetes/sig-release/blob/master/releases/release_phases.md#enhancements-freeze)
   of a given Kubernetes minor version.
 * If a given cloud provider was added more than one release cycle ago and there
   are no valid OWNERS, CMR should request OWNERS file update.

Cloudprovider owners will be required to address CMR or request an exception via
the CMR github issue. A failure to take any action will result in cloudprovider
being considered abandoned and marking it as deprecated as described below.

### Empty maintenance request

If no CMRs are issued in a given minor release, core maintainers will issue an
_empty CMR_. The purpose of an empty CMR is to verify that cloudprovider owners
are still actively maintaining their integration. The only action required for
an empty CMR is replying on the github issue. Only one owner from each
cloudprovider needs to reply on the issue.

Empty CMR follows the same rules as any other CMR. In particular it needs to be
issued by enhancements freeze.

### Cloudprovider deprecation and deletion

If cloudprovider owners fail to take actions described above, the particular
integration will be marked as deprecated in the next CA minor release. A
deprecated cloudprovider will be completely removed after 1 year as per
[Kubernetes deprecation
policy](https://kubernetes.io/docs/reference/using-api/deprecation-policy/#deprecating-a-feature-or-behavior).

A deprecated cloudprovider may become maintained again if the owners become
active again or new owners step up. In order to regain maintained status any
outstanding CMRs will need to be addressed.

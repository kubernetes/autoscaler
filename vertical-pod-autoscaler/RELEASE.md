# VPA Release Instructions

These are instructions for releasing VPA. We aim to release a new VPA minor version after each minor Kubernetes release.
We release patch versions as needed.

Before doing the release for the first time check if you have all the necessary permissions (see
[Permissions](#permissions) section below).

There are the following steps of the release process:

1. [ ] Open issue to track the release.
2. [ ] Update VPA version const.
3. [ ] Build and stage images.
4. [ ] Test the release.
5. [ ] Promote image.
6. [ ] Finalize release.

## Open issue to track the release

Open a new issue to track the release, use the [vpa_release](https://github.com/kubernetes/autoscaler/issues/new?&template=vpa_release.md) template.
We use the issue to communicate what is state of the release.

## Update VPA version const

1. [ ] Wait for all VPA changes that will be in the release to merge.
2. [ ] Wait for [the end to end tests](https://testgrid.k8s.io/sig-autoscaling-vpa) to run with all VPA changes
   included.
   To see what code was actually tested, look for `===== last commit =====`
   entries in the full `build-log.txt` of a given test run.
3. [ ] Make sure the end to end VPA tests are green.

### New minor release

1. [ ] Change the version in
    [common/version-go](https://github.com/kubernetes/autoscaler/blob/master/vertical-pod-autoscaler/common/version.go)
    to `1.${next-minor}.0`,
2. [ ] Commit and merge the change,
3. [ ] Go to the merged change,
4. [ ] [Create a new branch](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/proposing-changes-to-your-work-with-pull-requests/creating-and-deleting-branches-within-your-repository) named `vpa-release-1.${next-minor}` from the
    merged change.

### New patch release

1.  [ ] Bump the patch version number in VerticalPodAutoscalerVersion constant in
    [common/version.go](https://github.com/kubernetes/autoscaler/blob/master/vertical-pod-autoscaler/common/version.go).
    Create a commit and merge by making a PR to the `vpa-release-1.${minor}` branch.

## Build and stage images

Create a fresh clone of the repo and switch to the `vpa-release-1.${minor}`
branch. This makes sure you have no local changes while building the images.

For example:
```sh
git clone git@github.com:kubernetes/autoscaler.git
git switch vpa-release-1.0
```

Once in the freshly cloned repo, build and stage the images.

```sh
cd vertical-pod-autoscaler/
for component in recommender updater admission-controller ; do TAG=`grep 'const VerticalPodAutoscalerVersion = ' common/version.go | cut -d '"' -f 2` REGISTRY=gcr.io/k8s-staging-autoscaling make release --directory=pkg/${component}; done
```

## Test the release

1.  [ ] Create a Kubernetes cluster. If you're using GKE you can use the following command:

    ```shell
    gcloud container clusters create e2e-test --machine-type=n1-standard-2 --image-type=COS_CONTAINERD --num-nodes=3
    ```

1. [ ]  Create clusterrole. If you're using GKE you can use the following command:

    ```shell
    kubectl create clusterrolebinding my-cluster-admin-binding --clusterrole=cluster-admin --user=`gcloud config get-value account`
    ```

1.  [ ] Deploy VPA:
    ```shell
    REGISTRY=gcr.io/k8s-staging-autoscaling TAG=`grep 'const VerticalPodAutoscalerVersion = ' common/version.go | cut -d '"' -f 2` ./hack/vpa-up.sh
    ```

1.  [ ] [Run](https://github.com/kubernetes/autoscaler/blob/master/vertical-pod-autoscaler/hack/run-e2e-tests.sh)
    the `full-vpa` test suite:

    ```shell
    ./hack/run-e2e-tests.sh full-vpa
    ```

## Promote image

To promote image from staging repo send out PR updating
[autoscaling images mapping](https://github.com/kubernetes/k8s.io/blob/master/registry.k8s.io/images/k8s-staging-autoscaling/images.yaml)
([example](https://github.com/kubernetes/k8s.io/pull/1318)).

NOTE: Please use the [add-version.sh
script](https://github.com/kubernetes/k8s.io/blob/main/registry.k8s.io/images/k8s-staging-autoscaling/add-version.sh)
to prepare the changes automatically.

When PR merges the promoter will run automatically and upload the image from
staging repo to final repo. The post submit job status can be tracked on
[testgrid](https://testgrid.k8s.io/sig-k8s-infra-k8sio#post-k8sio-image-promo).
To verify if the promoter finished its job one can use gcloud. E.g.:

```sh
gcloud container images describe registry.k8s.io/autoscaling/vpa-recommender:[*vpa-version*]
```

## Finalize release

NOTE: We currently use two tags for releases:
`vertical-pod-autoscaler-[*vpa-version*]` and
`vertical-pod-autoscaler/v[*vpa-version*]`. We need
`vertical-pod-autoscaler/v[*vpa-version*]` for `go get
k8s.io/autoscaler/vertical-pod-autoscaler@v[*vpa-version*]` to work. We can
consider stopping using `vertical-pod-autoscaler-[*vpa-version*]` tags but
we've been using them since `vertical-pod-autoscaler-0.1` and tags with the
other pattern start only with `vertical-pod-autoscaler/v0.9.0` so we should make
sure nothing we care about will break if we do.

1.  [ ] Update information about newest version and K8s compatibility in
    [the installation section of README](https://github.com/kubernetes/autoscaler/blob/master/vertical-pod-autoscaler/README.md#installation).

1.  [ ] Update the yaml files:

    ```sh
    sed -i -s "s|[0-9]\+\.[0-9]\+\.[0-9]\+|[*vpa-version*]|" ./deploy/*-deployment*.yaml ./hack/vpa-process-yaml.sh
    ```
1.  [ ] Update the default tag in  [vpa-up.sh](https://github.com/kubernetes/autoscaler/blob/master/vertical-pod-autoscaler/hack/vpa-up.sh).

1.  [ ] Merge these changes into branch vpa-release-1.{$minor} and optionally into master if 1.{$minor} is the latest minor release
    (example PR: [#5460](https://github.com/kubernetes/autoscaler/pull/5460)).

1.  [ ] Tag the commit with version const change

    ```sh
    git tag -a vertical-pod-autoscaler-[*vpa-version*] -m "Vertical Pod Autoscaler release [*vpa-version*]"
    git tag -a vertical-pod-autoscaler/v[*vpa-version*] -m "Vertical Pod Autoscaler release [*vpa-version*]"
    ```

1.  [ ] Push tag

    ```sh
    git push git@github.com:kubernetes/autoscaler.git vertical-pod-autoscaler-[*vpa-version*]
    git push git@github.com:kubernetes/autoscaler.git vertical-pod-autoscaler/v[*vpa-version*]
    ```

1.  [ ] To create and publish a github release from pushed tag go to
    https://github.com/kubernetes/autoscaler/releases/tag/vertical-pod-autoscaler-[*vpa-version*],
    press `Create release from tag`, complete release title and release notes, tick the
    `This is a pre-release` box and press `Publish release`.

## Permissions

* Permissions to access `gcr.io/k8s-staging-autoscaling` are governed by list
    of people in
    [groups.yaml](https://github.com/kubernetes/k8s.io/blob/master/groups/sig-autoscaling/groups.yaml)
    under k8s-infra-staging-autoscaling.
* Permissions to add images to
    [`k8s.io/registry.k8s.io/images/k8s-staging-autoscaling/images.yaml`](https://github.com/kubernetes/k8s.io/blob/main/registry.k8s.io/images/k8s-staging-autoscaling/images.yaml) are governed by
    [OWNERS file](https://github.com/kubernetes/k8s.io/blob/main/registry.k8s.io/images/k8s-staging-autoscaling/OWNERS).
* Permissions to add tags to
    [kubernetes/autoscaler](https://github.com/kubernetes/autoscaler) and create
    releases in the repo you must be:

    *   a collaborator on the
        [kubernetes/autoscaler](https://github.com/kubernetes/autoscaler) repo
        or
    *   a member of the
        [autoscaler-maintainers](https://github.com/orgs/kubernetes/teams/autoscaler-maintainers/members)
        team.

    A member of the
    [autoscaler-admins](https://github.com/orgs/kubernetes/teams/autoscaler-admins)
    can add you to add you as a collaborator.

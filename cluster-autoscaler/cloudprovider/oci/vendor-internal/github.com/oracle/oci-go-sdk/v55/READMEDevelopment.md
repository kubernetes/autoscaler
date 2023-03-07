# Running the Code Generator
## Pre-requisites
- Maven
- Python (and tools, see below)
- Go(make sure you set the GOPATH enviroment correctly)
- Go tools see: https://github.com/oracle/oci-go-sdk#building-and-testing
- oci-go-sdk commons go package
  - You can install the lastest version by installing pulling the current version of the go sdk
- Make

## GO Module
We have introduced Go Module support and there are some tricky part in development. For developers who want to use an internal Go SDK version (which is hosted in bitbucket instead of public github), please follow these steps to setup your environment
1. run 
   ```
    git config --global url."ssh://git@bitbucket.oci.oraclecorp.com:7999".insteadOf "https://bitbucket.oci.oraclecorp.com/scm"
   ```
2. disable GOSUMDB, `export GOSUMDB=off`
3. import oci-go-sdk as usual in `go.mod` file
4. add "replace" statement at the end of `go.mod` file, for example: 
    ```
    replace github.com/oracle/oci-go-sdk/v35 => bitbucket.oci.oraclecorp.com/sdk/oci-go-sdk/v35 v35.0.0-20210209162633-71904f09d1f4
    ```
5. replace this complex version with the commit id you want, then run `go mod tidy`/`go mod vendor`/`go build`, etc
    ```
    replace github.com/oracle/oci-go-sdk/v36 v36.0.0 => bitbucket.oci.oraclecorp.com/sdk/oci-go-sdk/v36 15b567ee1ef
    ```
6. after `v36.1.0`, replacing tag for Bitbucket preview branch is also supported, then run `go mod tidy`/`go mod vendor`/`go build`, etc
    ```
    replace github.com/oracle/oci-go-sdk/v36 v36.1.0 => bitbucket.oci.oraclecorp.com/sdk/oci-go-sdk/v36 v36.1.0-p
    ```

## Python setup
Use ``virtualenv`` to create a virtual environment that runs Python 2.x (of course, you have to have Python 2.x installed somewhere):

    # Install the virtual environment in the current directory
    virtualenv --python=<path to Python 2.x executable> temp/python2
    # Activate virtual environment
    source temp/python2/bin/activate
    # Install packages
    pip install PyYAML
    pip install six



## Start here!
The build functionality is driven by 2 make files

- Makefile: Is public and exposed, builds the sdk and runs unittest
- MakefileDevelopment: Private. builds, generates new sdk, runs private integtests and autotest. Most of the time this the one you want to work with


## Layout
The  organiztion of the public parts of  sdk is described here:  https://github.com/oracle/oci-go-sdk#organization-of-the-sdk .
In addition in order to support generation, the following files are present:

- featureId.yaml: legacy file where the different conditinal directives were getting saved, used by the generator to turn spec features on/off
- codegenConfig: current directory where the feature flags are stored and read from
- github.whitelist: our release process uses this file to copy/push all the files that match one or more of the regexes in this file
- pom.xml: Controls the version of the generator as well as locations of the specs used to generate the different packages. For go in particular, this file contains configurations for a successful generation and it usually looks like this.

            <configuration>
              <language>oracle-go-sdk</language>
              <specPath>${preprocessed-temp-dir}/announcements-service-spec/${announcements-service-spec-spec-file}</specPath>
              <outputDir>${env.GOPATH}/src/${fullyQualifiedProjectName}</outputDir>  <-- The root of the sdk, notice it reads the env var $GOPATH
              <basePackage>announcementsservice</basePackage>  <--- The name of the package, it will create a directory with that name and put all generated files in there
              <specGenerationType>${generationType}</specGenerationType>
              <additionalProperties>
                <specName>announcementsservice</specName>
                <fqProjectName>${fullyQualifiedProjectName}</fqProjectName>  <-- The name of the root packate usually github.com/oracle/oci-go-sdk
                <serviceHostName>announcements</serviceHostName>
              </additionalProperties>
              <featureIdConfigFile>${feature-id-file}</featureIdConfigFile>
              <featureIdConfigDir>${feature-id-dir}</featureIdConfigDir>
            </configuration>


## Help
The `MakefileDevelopment.mk` file contains a help command to help you nagivate its options. To bring out the help execute:

    make -f MakefileDevelopment.mk help

## Building and Generating
The generation makefile is: ***MakefileDevelopment.mk***
You run the code generator by executing. This will generate the code as well as build it

    make -f MakefileDevelopment.mk build

After executing this command the source code will be placed under the canonical repository `$GOPATH/src/$PROJECT_NAME` where $PROJECT_NAME is the fully qualified project name: `github.com/oracle/oci-go-sdk`.

The above command executes the  `generation` and `build` target which generates the sdk. If you want to just build the sdk, issue:

    make -f MakefileDevelopment.mk build-sdk

You can also build packages individually by issuing:

    make build-[package_name]

## Testing
The go sdk has 2 types of testing and they can all be executed through the make files

### Unittest
The unitest of the go-sdk cover functionality used internally by all the sdk packages, mostly present in the `common` and `common/auth` package. It can be executed by running

    make test

### Autotests
Autotests have the highest coverage for the sdk api and are genereated automatically. These test work in conjunction with the testing service, so when running them, you need to have an instance of the testing service running. You can execute these tests by

    make -f MakefileDevelopment.mk autotest-identity ## will execute the autotest for the identity service

    make -f MakefileDevelopment.mk autotest-all ## will execute all autotest, be careful, this can take a long time

### Running a single test
A single test can be run with the following command. Notices `TEST_NAME` is the name of the test that you wish to run and can be a string or a regex, eg:

    make -f MakefileDevelopment.mk autotest TEST_NAME=TestIdentityClientListUsers ## will execute *just* the TestIdentityClientListUsers test

    make -f MakefileDevelopment.mk autotest TEST_NAME=^TestIdentity ## will execute all tests that start with "TestIdentity"


## Release
Instead of the `build` target. Execute the release target like so:

    make -f MakefileDevelopment.mk release

Do not forget to setup major, minor versions by updating the variables in: `MakefileDevelopment.mk`

    VER_MAJOR = x
    VER_MINOR = y


## Some tips
Often when working on new feature of the sdk, you'll need to generate and build, most of the time you don't need to generate the whole sdk but only a subset. This is a bit challening since maven thougt it allows you to target a specific step, the names of the stepts in the maven file are not very intuitive. I find the following commands super helpful

- Targeting a specific package for generation, where `$1` is the execution `<id>` of the package you want to generate

        PROJECT_NAME=github.com/oracle/oci-go-sdk mvn bmc-sdk-swagger:generate@$1
        PROJECT_NAME=github.com/oracle/oci-go-sdk mvn bmc-sdk-swagger:generate@go-public-sdk-maestro-spec

- Linting and rebuilding sdk and tests for a specific package

        make lint-$1 build-$1 pre-doc && make -f MakefileDevelopment.mk build-autotest
        make lint-resourcemanager build-resourcemanager pre-doc && make -f MakefileDevelopment.mk build-autotest

- Often you have to rebuild the generator and then execute this steps, so the whole command line ends up looking like this:

        (cd /Users/eginez/repos/bmc-sdk-swagger && mvn install -D=skipTests) && mvngen go-public-sdk-maestro-spec && make lint-resourcemanager build-resourcemanager pre-doc && make -f MakefileDevelopment.mk build-autotest


## Self-service for adding features and services

[Requesting a preview SDK](<https://confluence.oci.oraclecorp.com/display/DEX/Requesting+a+preview+SDK+CLI>)

[Requesting a public SDK](<https://confluence.oci.oraclecorp.com/pages/viewpage.action?pageId=43683000>)

[Self-Service Testing and Development](<https://confluence.oci.oraclecorp.com/pages/viewpage.action?spaceKey=DEX&title=Self-Service+Testing+and+Development>)

[SDK Testing with OCI Testing Service Overview](<https://confluence.oci.oraclecorp.com/display/DEX/SDK+Testing+with+OCI+Testing+Service+Overview>)

[SDK / CLI Sample Requirements](<https://confluence.oci.oraclecorp.com/pages/viewpage.action?pageId=43687174>)

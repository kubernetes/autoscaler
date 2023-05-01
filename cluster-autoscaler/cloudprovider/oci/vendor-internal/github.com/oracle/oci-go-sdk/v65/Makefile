DOC_SERVER_URL=https:\/\/docs.cloud.oracle.com

GEN_TARGETS = identity core objectstorage loadbalancer database audit dns filestorage email containerengine resourcesearch keymanagement announcementsservice healthchecks waas autoscaling streaming ons monitoring resourcemanager budget workrequests functions limits events dts oce oda analytics integration osmanagement marketplace apigateway applicationmigration datacatalog dataflow datascience nosql secrets vault bds cims datasafe mysql dataintegration ocvp usageapi blockchain loggingingestion logging loganalytics managementdashboard sch loggingsearch managementagent cloudguard opsi computeinstanceagent optimizer tenantmanagercontrolplane rover databasemanagement artifacts apmsynthetics goldengate apmcontrolplane apmtraces networkloadbalancer vulnerabilityscanning databasemigration servicecatalog ailanguage operatoraccesscontrol bastion genericartifactscontent jms devops aianomalydetection datalabelingservice datalabelingservicedataplane apmconfig waf certificates certificatesmanagement usage databasetools servicemanagerproxy appmgmtcontrol ospgateway identitydataplane visualbuilder osubusage osubsubscription osuborganizationsubscription osubbillingschedule dashboardservice threatintelligence aivision aispeech stackmonitoring servicemesh adm licensemanager onesubscription governancerulescontrolplane waa networkfirewall vnmonitoring emwarehouse lockbox fusionapps mediaservices opa opensearch cloudmigrations cloudbridge disasterrecovery containerinstances aidocument queue recovery vbsinst identitydomains ##SPECNAME##
NON_GEN_TARGETS = common common/auth objectstorage/transfer example
TARGETS = $(NON_GEN_TARGETS) $(GEN_TARGETS)

TARGETS_WITH_TESTS = common common/auth objectstorage/transfer
TARGETS_WITH_INTEG_TESTS = integtest
TARGETS_BUILD = $(patsubst %,build-%, $(TARGETS))
TARGETS_CLEAN = $(patsubst %,clean-%, $(GEN_TARGETS))
TARGETS_LINT = $(patsubst %,lint-%, $(TARGETS))
TARGETS_TEST = $(patsubst %,test-%, $(TARGETS_WITH_TESTS))
TARGETS_TESTFILTERED = $(patsubst %,testfiltered-%, $(TARGETS_WITH_TESTS))
TARGETS_INTEG_TEST = $(patsubst %,test-%, $(TARGETS_WITH_INTEG_TESTS))
TARGETS_RELEASE= $(patsubst %,release-%, $(TARGETS))
GOLINT=$(GOPATH)/bin/golint
LINT_FLAGS=-min_confidence 0.9 -set_exit_status

AUTOTEST_DIR = autotest

# directories under gen targets which contains hand writen code
EXCLUDED_CLEAN_DIRECTORIES = objectstorage/transfer*

.PHONY: $(TARGETS_BUILD) $(TARGET_TEST)

build: lint $(TARGETS_BUILD)

test: build $(TARGETS_TEST)

test-all: build build-autotest test test-integ

test-integ: 
	@if [ -d $(TARGETS_WITH_INTEG_TESTS) ]; then\
		make build $(TARGETS_INTEG_TEST);\
	fi

lint: $(TARGETS_LINT)

clean: $(TARGETS_CLEAN)

$(TARGETS_LINT): lint-%:%
	@echo "linting and formatting: $<"
	@(cd $< && gofmt -s -w .)
	@if [ \( $< = common \) -o \( $< = common/auth \) ]; then\
		(cd $< && $(GOLINT) -set_exit_status .);\
	else\
		(cd $< && $(GOLINT) $(LINT_FLAGS) .);\
	fi

# for sample code, only build them via 'go test -c'
$(TARGETS_BUILD): build-%:%
	@echo "building: $<"
	@if [ \( $< = example \) ]; then\
		(cd $< && go test -c);\
	else\
		(cd $< && find . -name '*_integ_test.go' | xargs -I{} mv {} ../integtest);\
		(cd $< && go build -v);\
	fi

$(TARGETS_TEST): test-%:%
	@(cd $< && go test -v)

$(TARGETS_TESTFILTERED): testfiltered-%:%
	@(cd $< && go test -v -run $(TEST_NAME))

$(TARGETS_INTEG_TEST): test-%:%
	@(cd $< && go test -v)

$(TARGETS_CLEAN): clean-%:%
	@echo "cleaning $<"
	@-find $< -not -path "$<" | grep -vZ ${EXCLUDED_CLEAN_DIRECTORIES} | xargs rm -rf

# clean all generated code under GEN_TARGETS folder
clean-generate:
	for target in ${GEN_TARGETS}; do \
		echo "cleaning $$target"; \
		find $$target -not -path "$$target" | grep -vZ ${EXCLUDED_CLEAN_DIRECTORIES} | xargs rm -rf; \
	done

pre-doc:
	@echo "Rendering doc server to ${DOC_SERVER_URL}"
	find . -name \*.go |xargs sed -i 's/{{DOC_SERVER_URL}}/${DOC_SERVER_URL}/g'
	# Note: This should stay the old docs URL (with us-phoenix-1), because it
	# processes go files and changes the old URL into the new URL
	find . -name \*.go |xargs sed -i 's/https:\/\/docs.us-phoenix-1.oraclecloud.com/${DOC_SERVER_URL}/g'

pre-doc-local:
	@echo "Rendering doc server to ${DOC_SERVER_URL}"
	find . -name \*.go |xargs sed -i '' 's/{{DOC_SERVER_URL}}/${DOC_SERVER_URL}/g'
	# Note: This should stay the old docs URL (with us-phoenix-1), because it
	# processes go files and changes the old URL into the new URL
	find . -name \*.go |xargs sed -i '' 's/https:\/\/docs.us-phoenix-1.oraclecloud.com/${DOC_SERVER_URL}/g'


gen-version:
	go generate -x

release: gen-version build pre-doc

build-autotest:
	@if [ -d $(AUTOTEST_DIR) ]; then\
		(cd $(AUTOTEST_DIR) && gofmt -s -w . && go test -c);\
	fi

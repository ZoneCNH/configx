.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: vet
vet:
	go vet ./...

.PHONY: test
test:
	go test ./...

.PHONY: race
race:
	go test -race ./...

.PHONY: lint
lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed"; \
		exit 1; \
	fi

.PHONY: integration
integration:
	./scripts/run_integration.sh

.PHONY: security
security:
	@if command -v govulncheck >/dev/null 2>&1; then \
		govulncheck ./...; \
	else \
		echo "govulncheck not installed"; \
		exit 1; \
	fi
	./scripts/check_secrets.sh

.PHONY: boundary
boundary:
	./scripts/check_boundary.sh

.PHONY: contracts
contracts:
	./scripts/check_contracts.sh

.PHONY: property
property:
	go test ./... -run 'Test.*Property|Test.*Invariant'

.PHONY: fuzz-smoke
fuzz-smoke:
	./scripts/run_fuzz_smoke.sh

.PHONY: golden
golden:
	go test ./... -run 'Test.*Golden|Test.*Snapshot'

.PHONY: evidence
evidence:
	./scripts/generate_manifest.sh

.PHONY: release-evidence-check
release-evidence-check:
	RELEASE_EVIDENCE_REQUIRE_PASSED=1 ./scripts/check_release_evidence.sh

.PHONY: ci
ci: fmt vet lint test race boundary security contracts

.PHONY: ci-extended
ci-extended: ci property golden fuzz-smoke

.PHONY: release-check
release-check: ci integration
	CHECK_STATUS=passed $(MAKE) evidence
	$(MAKE) release-evidence-check

.PHONY: release-check-extended
release-check-extended: ci-extended integration
	CHECK_STATUS=passed $(MAKE) evidence
	$(MAKE) release-evidence-check

.PHONY: release-final-check
release-final-check: release-check
	RELEASE_EVIDENCE_REQUIRE_PASSED=1 RELEASE_EVIDENCE_REQUIRE_CLEAN=1 ./scripts/check_release_evidence.sh

.PHONY: release-preflight
release-preflight:
	./scripts/check_release_preflight.sh "$(VERSION)"
	GOWORK=off $(MAKE) release-final-check

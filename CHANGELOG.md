# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed

- Replace `jwt-go` with `golang-jwt/jwt`.

## [1.2.1] - 2021-06-02

### Fixed

- Fix missing new `architect-orb` version.

## [1.2.0] - 2021-06-02

### Changed

- Prepare helm values to configuration management.
- Update architect-orb to v3.0.0.

## [1.1.0] - 2021-01-29

### Added

- Add monitoring labels and add basic labels

### Changed

- Update dependencies for Kubernetes 1.18.
- Added `node-operator.giantswarm.io/version` label **anti**-selector to avoid reconciling future `DrainerConfig`s which use it and reserve it for use by future `node-operator` versions.

## [1.0.2]

### Changed

- Updated errors package to better handle k8s api errors.

## [1.0.1]

### Changed

- Fix registry value for docker image.

## [1.0.0] - 2020-04-29

### Changed

- Push `node-operator` chart into `control-plane` catalog instead of quay.io.
- Push `node-operator` app CRs into `<provider>-app-collection` repository.

[Unreleased]: https://github.com/giantswarm/node-operator/compare/v1.2.1...HEAD
[1.2.1]: https://github.com/giantswarm/node-operator/compare/v1.2.0...v1.2.1
[1.2.0]: https://github.com/giantswarm/node-operator/compare/v1.1.0...v1.2.0
[1.1.0]: https://github.com/giantswarm/node-operator/compare/v1.0.1...v1.1.0
[1.0.2]: https://github.com/giantswarm/node-operator/compare/v1.0.1...v1.0.2
[1.0.1]: https://github.com/giantswarm/node-operator/compare/v1.0.0...v1.0.1
[1.0.0]: https://github.com/giantswarm/node-operator/tag/v1.0.0

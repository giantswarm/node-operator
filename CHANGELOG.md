# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [2.0.3] - 2023-03-02

## [2.0.2] - 2023-03-02

- Handling of Spot instances which have already been deleted from the cluster, but the DrainingConfigs are still present

### Added

- Use Kubernetes Events on `AWSCluster` CR for draining status updates.
- Added the use the runtime/default seccomp profile.
- Concurrent execution of drainining to make sure the operator can handle multiple nodes going down for multiple clusters running in the same MC

### Changed

- Replace current logic with Kubernetes internal cordon/drain logic.

## [1.4.2] - 2022-12-06

### Fixed

- Ignore `cert-exporter-deployment.`

## [1.4.1] - 2022-10-10

### Fixed

- Change RBAC apiVersion to `rbac.authorization.k8s.io/v1`.

## [1.4.0] - 2022-10-10

### Changed

- Change draining timeout from 10 to 60 minutes.
- Bump dependencies and go to 1.18.

## [1.3.0] - 2022-03-09

### Changed

- Replace `jwt-go` with `golang-jwt/jwt`.
- Drop `apiextensions` dependency and move `DrainerConfig` API into this project.

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

[Unreleased]: https://github.com/giantswarm/node-operator/compare/v2.0.2...HEAD
[2.0.2]: https://github.com/giantswarm/node-operator/compare/v2.0.1...v2.0.2
[2.0.1]: https://github.com/giantswarm/node-operator/compare/v1.4.2...v2.0.1
[1.4.2]: https://github.com/giantswarm/node-operator/compare/v1.4.1...v1.4.2
[1.4.1]: https://github.com/giantswarm/node-operator/compare/v1.4.0...v1.4.1
[1.4.0]: https://github.com/giantswarm/node-operator/compare/v1.3.0...v1.4.0
[1.3.0]: https://github.com/giantswarm/node-operator/compare/v1.2.1...v1.3.0
[1.2.1]: https://github.com/giantswarm/node-operator/compare/v1.2.0...v1.2.1
[1.2.0]: https://github.com/giantswarm/node-operator/compare/v1.1.0...v1.2.0
[1.1.0]: https://github.com/giantswarm/node-operator/compare/v1.0.1...v1.1.0
[1.0.2]: https://github.com/giantswarm/node-operator/compare/v1.0.1...v1.0.2
[1.0.1]: https://github.com/giantswarm/node-operator/compare/v1.0.0...v1.0.1
[1.0.0]: https://github.com/giantswarm/node-operator/tag/v1.0.0

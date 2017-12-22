# node operator

[![Build Status](https://api.travis-ci.org/giantswarm/node-operator.svg)](https://travis-ci.org/giantswarm/node-operator) [![Go Report Card](https://goreportcard.com/badge/github.com/giantswarm/node-operator)](https://goreportcard.com/report/github.com/giantswarm/node-operator) [![](https://godoc.org/github.com/giantswarm/node-operator?status.svg)](http://godoc.org/github.com/giantswarm/node-operator) [![](https://img.shields.io/docker/pulls/giantswarm/node-operator.svg)](http://hub.docker.com/giantswarm/node-operator) [![IRC Channel](https://img.shields.io/badge/irc-%23giantswarm-blue.svg)](https://kiwiirc.com/client/irc.freenode.net/#giantswarm)


Short description what it is and for what you can use it and why.

Mention if there's some tools (internal or external) that it works especially well together with.

## Prerequisites

### How to build

#### Dependencies

Dependencies are managed using [`glide`](https://github.com/Masterminds/glide) and contained in the `vendor` directory. See `glide.yaml` for a list of libraries this project directly depends on and `glide.lock` for complete information on all external libraries and their versions used.

**Note:** The `vendor` directory is **flattened**. Always use the `--strip-vendor` (or `-v`) flag when working with `glide`.

#### Building the standard way

```nohighlight
go build
```

#### Cross-compiling in a container

Here goes the documentation on compiling for different architectures from inside a Docker container.

## Running PROJECT

- How to use
- What does it do exactly

## Further Steps

Links to documentation
Links to godoc

## Future Development

- Future directions/vision

## Contact

- Mailing list: [giantswarm](https://groups.google.com/forum/!forum/giantswarm)
- IRC: #[giantswarm](irc://irc.freenode.org:6667/#giantswarm) on freenode.org
- Bugs: [issues](https://github.com/giantswarm/PROJECT/issues)

## Contributing & Reporting Bugs

See [.github/CONTRIBUTING.md](/giantswarm/node-operator/blob/master/.github/CONTRIBUTING.md) for details on submitting patches, the contribution workflow as well as reporting bugs.

## License

PROJECT is under the Apache 2.0 license. See the [LICENSE](/giantswarm/node-operator/blob/master/LICENSE) file for details.

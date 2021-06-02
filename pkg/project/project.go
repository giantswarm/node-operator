package project

var (
	description = "Project node-operator drains Kubernetes nodes on behalf of watched CRDs."
	gitSHA      = "n/a"
	name        = "node-operator"
	source      = "https://github.com/giantswarm/node-operator"
	version     = "1.2.0"
)

func Description() string {
	return description
}

func GitSHA() string {
	return gitSHA
}

func Name() string {
	return name
}

func Source() string {
	return source
}

func Version() string {
	return version
}

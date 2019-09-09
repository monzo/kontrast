# kontrast

`kubectl diff` with pretty colours. Currently alpha - use at your own risk (`kontrast` doesn't do any write operations, so should be okay).

## Installation

`go install github.com/monzo/kontrast`

Note: this takes a while [1]

## Usage

`kontrast my-manifest.yaml`


## Release process

1. Create a new semver version tag on master in "v0.1.2" format
2. Push the tag
3. Run `make clean dist` to build the binary releases
4. Create a new release on Github and upload all files from `bin/`

---
###### [1] Why does it take so long to build/install? Why is the binary 150MB?

Kubernetes applies defaults to each created resource. The average manifest doesn't contain every possible option, so a lot of these defaults will show up as deltas. `kontrast` gets around this by applying `scheme.Scheme.SetDefaults(obj)`, which essentially does the same operation as the Kubernetes apiserver when creating an object. Unfortunately this only works if the full apiserver code is imported (i.e. `k8s.io/kubernetes/pkg/master` which imports https://github.com/kubernetes/kubernetes/blob/master/pkg/master/import_known_versions.go)

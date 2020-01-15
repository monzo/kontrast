# kontrast

`kubectl diff` with pretty colours. Currently alpha - use at your own risk (`kontrast` doesn't do any write operations, so should be okay).

## Installation

`go install github.com/monzo/kontrast`

Note: this takes a while [1]

## Usage

`kontrast my-manifest.yaml`

## Note on Developing

If you are running `dep` to introduce a new scheme from a custom Kubernetes resource type, we are aware of at least one upstream repository hosted in BitBucket and expecting [mercurial](https://www.mercurial-scm.org/) to access. Without it installed, `dep` will likely hang and not provide any clues even under verbose mode. 

On macOS, you can install it with Homebrew:

```
brew install mercurial
```

---
###### [1] Why does it take so long to build/install? Why is the binary 150MB?

Kubernetes applies defaults to each created resource. The average manifest doesn't contain every possible option, so a lot of these defaults will show up as deltas. `kontrast` gets around this by applying `scheme.Scheme.SetDefaults(obj)`, which essentially does the same operation as the Kubernetes apiserver when creating an object. Unfortunately this only works if the full apiserver code is imported (i.e. `k8s.io/kubernetes/pkg/master` which imports https://github.com/kubernetes/kubernetes/blob/master/pkg/master/import_known_versions.go)

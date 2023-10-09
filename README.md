# kube-lock
## Now Available as a [Krew](https://krew.sigs.k8s.io/) Plugin!

<p align="center">
  <img src="./logo/kube-lock.png" width="350" />
</p>

A pane of glass between you and your Kubernetes clusters.
- Sits as an intermediary between you and `kubectl`, allowing you to `lock` and `unlock` contexts.
- Prevents misfires to production / high-value Kubernetes clusters that you might have strong IAM privileges on.
- Supports custom 'Profiles', allowing you to restrict certain verbs from being passed to high-value clusters.  

If you wish to build it and try it out though, you can do one of the following:

## install as a krew plugin!
1. kubectl krew index add kube-lock https://github.com/chaosinthecrd/kube-lock.git (this is a temporary step while the plugin is getting accepted to the upstream index)
2. kubectl krew install kube-lock/lock
3. create an alias in your `.bashrc` or `.zshrc` like: `alias kubectl="kubectl-lock kubectl --"`
4. From here, you can use `kube-lock` by calling `kubectl lock` followed by the subcommand you wish to use (e.g., `kubectl lock lock`)

## install manually
1. run `go build -o="kubectl-lock" .` in the root of the repo, or download from the [Github Releases page](https://github.com/ChaosInTheCRD/kube-lock/releases).
2. copy the produced binary into somewhere within your path
3. create an alias in your `.bashrc` or `.zshrc` like: `alias kubectl="kubectl-lock kubectl --"`
4. From here, you can use `kube-lock` by calling `kubectl-lock` followed by the subcommand you wish to use (e.g., `kubectl-lock lock`)

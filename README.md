# kube-lock
A pain of glass between you and your Kubernetes clusters.
- Sits as a middle-man between you and `kubectl`, allowing you to `lock` and `unlock` contexts.
- Prevents misfires to production / high-value Kubernetes clusters that you might have strong IAM privileges on.
- Supports custom 'Profiles', allowing you to restrict certain verbs from being passed to high-value clusters.  

☢️ Still a work in progress ☢️

If you wish to build it and try it out though, simply:
- run `go build .` in the root of the repo
- copy the produced binary into somewhere within your path
- create an alias in your `.bashrc` or `.zshrc` like: `alias kubectl="kube-lock kubectl --"

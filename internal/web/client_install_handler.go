package web

import (
	"fmt"
	"net/http"
	"strings"
)

// DefaultSidecarProxyAddr is the default location where a client can
// download the sidecar proxy container image.
var DefaultSidecarProxyAddr = "127.0.0.1:5000/sidecar-proxy:1.0.0"

var (
	// SidecarProxyAddr is the docker registry address of the sidecar-proxy image
	// Set via "web.sidecarproxyaddr"
	SidecarProxyAddr = DefaultSidecarProxyAddr
)

// Guest is used for the Guest tenant and role name.
const Guest = "Guest"

// ClientInstallHandler returns a handler that will serve up an installer
// script to requesting clients.
func ClientInstallHandler(rootCA string, insecure bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host := r.Host
		var sb strings.Builder

		q := r.URL.Query()
		pps := ""

		for _, pp := range q["proxy-port"] {
			t := strings.Split(pp, ":")
			pps += fmt.Sprintf(" --proxy-port %s=%s", t[0], t[1])
		}

		inject := fmt.Sprintf("karavictl inject --image-addr %s --proxy-host %s --insecure=%v %s", SidecarProxyAddr, host, insecure, pps)
		if rootCA != "" {
			inject += fmt.Sprintf(" --root-certificate %s", rootCA)
		}

		checkDrivers := fmt.Sprintf(`
export DRIVERS="%s"				
if [ "${DRIVERS}" == "" ]; then
    export DRIVERS=$(kubectl get namespace)
fi
`, strings.Join(q["namespace"], ","))

		vxflexos := fmt.Sprintf(`
if [[ $DRIVERS =~ "vxflexos" ]]; then
    kubectl get secrets,deployments,daemonsets -n vxflexos -o yaml | %s | kubectl apply -f -
    kubectl rollout restart -n vxflexos deploy/vxflexos-controller
    kubectl rollout restart -n vxflexos ds/vxflexos-node
    kubectl rollout status -n vxflexos deploy/vxflexos-controller
    kubectl rollout status -n vxflexos ds/vxflexos-node
fi
`, inject)

		powermax := fmt.Sprintf(`
if [[ $DRIVERS =~ "powermax" ]]; then
    kubectl get secrets,deployments,daemonsets,configmap -n powermax -o yaml | %s | kubectl apply -f -
    kubectl rollout restart -n powermax deploy/powermax-controller
    kubectl rollout restart -n powermax ds/powermax-node	
    kubectl rollout status -n powermax deploy/powermax-controller
    kubectl rollout status -n powermax ds/powermax-node
fi
`, inject)

		powerscale := fmt.Sprintf(`
if [[ $DRIVERS =~ "isilon" || $DRIVERS =~ "powerscale" ]]; then
	kubectl get secrets,deployments,daemonsets,configmap -n isilon -o yaml | %s | kubectl apply -f -
	kubectl rollout restart -n isilon deploy/isilon-controller
	kubectl rollout restart -n isilon ds/isilon-node	
	kubectl rollout status -n isilon deploy/isilon-controller
	kubectl rollout status -n isilon ds/isilon-node
fi
`, inject)

		fmt.Fprintln(&sb, checkDrivers)
		fmt.Fprintln(&sb, powermax)
		fmt.Fprintln(&sb, vxflexos)
		fmt.Fprintln(&sb, powerscale)

		fmt.Fprintln(w, sb.String())
	})
}

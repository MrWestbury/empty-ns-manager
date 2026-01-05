package internal

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/mrwestbury/empty-ns-manager/api/v1alpha1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/utils/strings/slices"
)

var (
	EMPTYNS_SAFE_NAMESPACES = []string{"default", "kube-system", "kube-public", "kube-node-lease"}
)

type reconciliationResult struct {
	Namespace     string
	IsEmpty       bool
	PolicyApplied string
	ActionTaken   string
	Timestamp     time.Time
}

type EmptyNSController struct {
	policyWatcher  *PolicyWatcher
	NSManager      *NamespaceManager
	WebServer      *WebServer
	refreshTicker  *time.Ticker
	activePolicies map[string]v1alpha1.EmptyNsPolicy
}

func NewEmptyNSController(config *rest.Config, appVersion string) *EmptyNSController {
	controller := &EmptyNSController{}
	controller.activePolicies = make(map[string]v1alpha1.EmptyNsPolicy)

	policyWatcher := NewPolicyWatcher(config)
	policyWatcher.PolicyAdd = controller.addPolicy
	policyWatcher.PolicyDelete = controller.removePolicy
	policyWatcher.PolicyUpdate = controller.updatePolicy
	nsManager := NewNamespaceManager(config)
	webServer := NewWebServer(&controller.activePolicies, appVersion)

	controller.policyWatcher = policyWatcher
	controller.NSManager = nsManager
	controller.WebServer = webServer
	controller.activePolicies = make(map[string]v1alpha1.EmptyNsPolicy)

	return controller
}

func (app *EmptyNSController) Run(stopCh <-chan struct{}) {
	go app.policyWatcher.Run(stopCh)
	go app.WebServer.Start()
	app.refreshTicker = time.NewTicker(30 * time.Second)
	go app.RefreshTimer(stopCh)
}

func (app *EmptyNSController) Stop() {
	if app.refreshTicker != nil {
		app.refreshTicker.Stop()
	}
}

func (app *EmptyNSController) RefreshTimer(stopCh <-chan struct{}) {
	for {
		select {
		case <-app.refreshTicker.C:
			// Call the refresh function here
			app.scan()

		case <-stopCh:
			app.Stop()
			return
		}
	}
}

func (app *EmptyNSController) addPolicy(policy v1alpha1.EmptyNsPolicy) {
	app.activePolicies[string(policy.UID)] = policy
	fmt.Printf("Policy added: %s\n", string(policy.UID))
}

func (app *EmptyNSController) removePolicy(policyUid string) {
	delete(app.activePolicies, policyUid)
	fmt.Printf("Policy removed: %s\n", policyUid)
}

func (app *EmptyNSController) updatePolicy(oldPolicy, newPolicy v1alpha1.EmptyNsPolicy) {
	app.activePolicies[string(newPolicy.UID)] = newPolicy
	fmt.Printf("Policy updated: %s\n", newPolicy.Name)
}

func (app *EmptyNSController) scan() {
	ctx := context.TODO()
	nsList, err := app.NSManager.ScanNamespaces(ctx)
	if err != nil {
		// Handle error
		return
	}

	// Process the list of namespaces as needed
	for _, ns := range nsList {
		if slices.Contains(EMPTYNS_SAFE_NAMESPACES, ns.Name) {
			continue
		}
		fmt.Printf("Namespace %s is not in safe list\n", ns.Name)
		result, err := app.reconcileNamespace(ctx, ns)
		if err != nil {
			fmt.Printf("Error reconciling namespace %s: %v\n", ns.Name, err)
			continue
		}
		fmt.Printf("Reconciliation result for namespace %s: %+v\n", ns.Name, result)
	}
}

func (app *EmptyNSController) reconcileNamespace(ctx context.Context, ns v1.Namespace) (*reconciliationResult, error) {
	result := &reconciliationResult{
		Namespace: ns.Name,
		Timestamp: time.Now(),
	}

	// Check previous scan results
	policyLabels := app.getPolicyLabels(ns)

	// Check if namespace is empty
	nsEmpty := app.NSManager.isNSEmpty(ns.Name)

	for _, policy := range app.activePolicies {
		fmt.Printf("Checking %s against policy %s\n", ns.Name, policy.Name)
		if app.nsMatchPolicy(ns, policy) {
			fmt.Printf("Namespace %s matches policy %s\n", ns.Name, policy.Name)

			// Check if namespace is empty
			fmt.Printf("Namespace %s empty status: %v\n", ns.Name, nsEmpty)

			if val, exists := policyLabels["policy-scan-date"]; exists {
				timestamp, err := strconv.ParseInt(val, 10, 64)
				if err != nil {
					fmt.Printf("Error parsing policy-scan-date for namespace %s: %v\n", ns.Name, err)
					continue
				}

				timeSinceLastScan := time.Since(time.UnixMilli(timestamp))
				if timeSinceLastScan < time.Minute*5 {
					fmt.Printf("Skipping namespace %s, recently scanned\n", ns.Name)
					continue
				}
			}

			if nsEmpty {
				// Take action based on policy
				newLabels := map[string]string{
					fmt.Sprintf("%s/policy", EMPTYNS_NAMESPACE):           policy.Name,
					fmt.Sprintf("%s/policy-scan-date", EMPTYNS_NAMESPACE): fmt.Sprintf("%d", time.Now().UnixMilli()),
					fmt.Sprintf("%s/policy-action", EMPTYNS_NAMESPACE):    policy.Spec.PolicyAction,
				}
				err := app.NSManager.AddLabelToNamespace(ctx, ns.Name, newLabels)
				if err != nil {
					fmt.Printf("Error labeling namespace %s: %v\n", ns.Name, err)
				}
			}
		}
	}

	return result, nil
}

func (app *EmptyNSController) getPolicyLabels(namespace v1.Namespace) map[string]string {
	labels := make(map[string]string)
	for key, value := range namespace.Labels {
		if len(key) > len(EMPTYNS_NAMESPACE)+1 && key[:len(EMPTYNS_NAMESPACE)+1] == EMPTYNS_NAMESPACE+"/" {
			labels[key[len(EMPTYNS_NAMESPACE)+1:]] = value
		}
	}
	return labels
}

func (app *EmptyNSController) nsMatchPolicy(namespace v1.Namespace, policy v1alpha1.EmptyNsPolicy) bool {
	// Implement logic to check if a namespace matches a given policy
	for key, value := range policy.Spec.NamespaceSelector.MatchLabels {
		if nsValue, exists := namespace.Labels[key]; !exists || nsValue != value {
			return false
		}
	}
	return true
}

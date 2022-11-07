package controllers

import (
	"context"
	"fmt"
	"github.com/waigel/gopass-operator/controllers/gp"
	"github.com/waigel/gopass-operator/controllers/gp_kubernetes"
	"github.com/waigel/gopass-operator/controllers/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	waigelcomv1alpha1 "github.com/waigel/gopass-operator/api/v1alpha1"
)

var log = logf.Log.WithName("controller_gopass")
var finalizer = "gopass/finalizer.secret"

// GopassItemReconciler reconciles a GopassItem object
type GopassItemReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *GopassItemReconciler) HandleGopassItem(ctx context.Context, gopassItem *waigelcomv1alpha1.GopassItem) error {
	secretName := gopassItem.Name
	secretNamespace := gopassItem.Namespace
	secretLabels := gopassItem.Labels
	secretType := gopassItem.Type

	sec, err := gp.GetGopassSecret(ctx, gopassItem.Spec.ItemPath)
	if err != nil {
		return fmt.Errorf("failed to retrieve secret (%s) item from gopass store: %v", gopassItem.Spec.ItemPath, err)
	}

	//Create owner reference
	ownerRef := metav1.OwnerReference{
		APIVersion: gopassItem.APIVersion,
		Kind:       gopassItem.Kind,
		Name:       gopassItem.Name,
		UID:        gopassItem.UID,
	}

	gp_kubernetes.CreateKubernetesSecretFromGopassSecret(r.Client, secretName, secretNamespace, sec, "autoRestart", secretLabels, secretType, &ownerRef)

	return nil
}

//+kubebuilder:rbac:groups=waigel.com,resources=gopassitems,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=waigel.com,resources=gopassitems/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=waigel.com,resources=gopassitems/finalizers,verbs=update

// Reconcile is part of the main gp_kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the GopassItem object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *GopassItemReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = logf.FromContext(ctx)
	reqLogger := log.WithValues("Request.Namespace", req.Namespace, "Request.Name", req.Name)
	reqLogger.Info("Reconciling GopassItem")

	//first we try to fetch the GopassItem instance
	gopassItem := &waigelcomv1alpha1.GopassItem{}
	err := r.Get(ctx, req.NamespacedName, gopassItem)
	if err != nil {
		reqLogger.Error(err, "Failed to get GopassItem")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	//check if the deployment is not deleted
	if gopassItem.ObjectMeta.DeletionTimestamp.IsZero() {
		//Add finalizer if not present
		if !utils.ContainsString(gopassItem.ObjectMeta.Finalizers, finalizer) {
			gopassItem.ObjectMeta.Finalizers = append(gopassItem.ObjectMeta.Finalizers, finalizer)
			if err := r.Update(ctx, gopassItem); err != nil {
				return ctrl.Result{}, err
			}
		}
		//Handle creation and update of the deployment if needed
		err = r.HandleGopassItem(ctx, gopassItem)
		return reconcile.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *GopassItemReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&waigelcomv1alpha1.GopassItem{}).
		Complete(r)
}

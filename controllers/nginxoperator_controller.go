/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	operatorv1alpha1 "github.com/gildub/nginx-operator/api/v1alpha1"
	myassets "github.com/gildub/nginx-operator/assets"

	appsv1 "k8s.io/api/apps/v1"
)

// NginxOperatorReconciler reconciles a NginxOperator object
type NginxOperatorReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=operator.example.com,resources=nginxoperators,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=operator.example.com,resources=nginxoperators/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=operator.example.com,resources=nginxoperators/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the NginxOperator object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *NginxOperatorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	operatorCR := &operatorv1alpha1.NginxOperator{}
	err := r.Get(ctx, req.NamespacedName, operatorCR)

	if err != nil && errors.IsNotFound(err) {
		logger.Info("Operator resource object not found.")
		return ctrl.Result{}, nil
	} else if err != nil {
		logger.Error(err, "Error getting operator resource object")
		return ctrl.Result{}, err
	}

	// // Check if this Deployment already exists
	// found := &appsv1.Deployment{}
	// err = r.Get(context.TODO(), types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, found)
	// var result *reconcile.Result
	// result, err = r.ensureDeployment(req, instance, r.backendDeployment(instance))
	// if result != nil {
	// 	log.Error(err, "Deployment Not ready")
	// 	return *result, err
	// }

	// // Check if this Service already exists
	// result, err = r.ensureService(req, instance, r.backendService(instance))
	// if result != nil {
	// 	log.Error(err, "Service Not ready")
	// 	return *result, err
	// }

	// // Deployment and Service already exists - don't requeue
	// log.Info("Skip reconcile: Deployment and service already exists",
	// 	"Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)

	deployment := &appsv1.Deployment{}
	create := false
	err = r.Get(ctx, req.NamespacedName, deployment)

	if err != nil && errors.IsNotFound(err) {
		create = true
		// deployment = assets.GetDeploymentFromFile("./manifests/nginx_deployment.yaml")
		deployment = myassets.GetDeploymentFromFile("/home/gildub/github.com/gildub/nginx-operator/assets/manifests/nginx_deployment.yaml")
	} else if err != nil {
		logger.Error(err, "Error getting existing Nginx deployment.")
		return ctrl.Result{}, err
	}

	deployment.Namespace = req.Namespace
	deployment.Name = req.Name

	if operatorCR.Spec.Replicas != nil {
		deployment.Spec.Replicas = operatorCR.Spec.Replicas
	}

	if operatorCR.Spec.Port != nil {
		deployment.Spec.Template.Spec.Containers[0].Ports[0].ContainerPort = *operatorCR.Spec.Port
	}

	ctrl.SetControllerReference(operatorCR, deployment, r.Scheme)

	if create {
		err = r.Create(ctx, deployment)
	} else {
		err = r.Update(ctx, deployment)
	}

	return ctrl.Result{}, err
}

// SetupWithManager sets up the controller with the Manager.
func (r *NginxOperatorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&operatorv1alpha1.NginxOperator{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}

/*
Copyright 2020 Sung Kang.

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package controllers

import (
	"context"
	"fmt"
	"github.com/zorkian/go-datadog-api"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	datadogv1 "github.com/skang0601/datadog-k8s-operator/api/v1alpha1"
)

// MonitorReconciler reconciles a Monitor object
type MonitorReconciler struct {
	client.Client
	Log           logr.Logger
	Scheme        *runtime.Scheme
	DataDogClient *datadog.Client
}

// +kubebuilder:rbac:groups=datadog.github.com/skang0601/datadog-k8s-operator,resources=monitors,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=datadog.github.com/skang0601/datadog-k8s-operator,resources=monitors/status,verbs=get;update;patch

func (r *MonitorReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("datadogmonitor", req.NamespacedName)

	datadogMonitor := &datadogv1.Monitor{}

	if err := r.Get(ctx, req.NamespacedName, datadogMonitor); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if !datadogMonitor.HasFinalizer(datadogv1.MonitorFinalizerName) {
		log.Info("Attaching finalizer to the CRD")
		if err := r.addFinalizer(datadogMonitor); err != nil {
			return ctrl.Result{}, err
		}
	}

	if datadogMonitor.IsBeingDeleted() {
		// Case 1: Deletion of the Monitor
		log.Info(fmt.Sprintf("Deleting monitor ID:(%d) from Datadog", datadogMonitor.Status.Id))
		log.Info("Removing Finalizers")
		if err := r.handleFinalizer(datadogMonitor); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: false}, nil
	}

	if !datadogMonitor.IsSubmitted() {
		// Case 2: Creation of the monitor
		log.Info("Submitting monitor to Datadog")
		if err := r.Submit(datadogMonitor); err != nil {
			return ctrl.Result{}, err
		}
		log.Info(fmt.Sprintf("Created a Monitor: %d", datadogMonitor.Status.Id))
		return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
	}

	if datadogMonitor.IsSubmitted() {
		// Case 3: Updates of the monitor
		log.Info("Submitting monitor to Datadog")
		if err := r.Submit(datadogMonitor); err != nil {
			return ctrl.Result{}, err
		}
		log.Info(fmt.Sprintf("Created a Monitor: %d", datadogMonitor.Status.Id))
		return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
	}

	return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
}

// We wanna add a method at the resource layer to add finalizers
func (r *MonitorReconciler) addFinalizer(instance *datadogv1.Monitor) error {
	instance.AddFinalizer()
	return r.Update(context.Background(), instance)
}

// This handle deletion of the object
func (r *MonitorReconciler) handleFinalizer(instance *datadogv1.Monitor) error {
	if !instance.HasFinalizer(datadogv1.MonitorFinalizerName) {
		return nil
	}

	if err := r.Delete(instance); err != nil {
		return err
	}

	instance.RemoveFinalizer()

	if err := r.Update(context.Background(), instance); err != nil {
		return err
	}
	return nil
}

/*
* This defines all of the interaction between the DataDog API with the SDK and our controller
 */
func (r *MonitorReconciler) Submit(instance *datadogv1.Monitor) error {
	ddMonitor, err := r.DataDogClient.CreateMonitor(instance.ToApi())
	if err != nil {
		instance.Status.Error = err.Error()
		return err
	}

	now := metav1.Now()

	instance.Status = datadogv1.MonitorStatus{
		Id:      int32(*ddMonitor.Id),
		Active:  true,
		Url:     fmt.Sprintf("https://app.datadoghq.com/monitors/%d", int32(*ddMonitor.Id)),
		Created: &now,
	}

	if err := r.Update(context.Background(), instance); err != nil {
		return err
	}
	return nil
}

func (r *MonitorReconciler) UpdateMonitor(instance *datadogv1.Monitor) error {
	err := r.DataDogClient.UpdateMonitor(instance.ToApi())
	if err != nil {
		instance.Status.Error = err.Error()
		return err
	}

	now := metav1.Now()
	instance.Status.LastUpdated = &now

	if err := r.Update(context.Background(), instance); err != nil {
		return err
	}
	return nil
}

func (r *MonitorReconciler) Delete(instance *datadogv1.Monitor) error {
	if err := r.DataDogClient.DeleteMonitor(int(instance.Status.Id)); err != nil {
		instance.Status.Error = err.Error()
		// TODO: Only in the cases of 404 monitor not found should I not care
		return nil
	}
	return nil
}

func (r *MonitorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&datadogv1.Monitor{}).
		Complete(r)
}

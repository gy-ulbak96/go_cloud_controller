/*
Copyright 2017 The Kubernetes Authors.

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

package main

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"

	cloudcontrollerv1alpha1 "github.com/gy-ulbak96/go_cloud_controller/pkg/apis/cloudcontroller/v1alpha1"
	clientset "github.com/gy-ulbak96/go_cloud_controller/pkg/generated/clientset/versioned"
	samplescheme "github.com/gy-ulbak96/go_cloud_controller/pkg/generated/clientset/versioned/scheme"
	informers "github.com/gy-ulbak96/go_cloud_controller/pkg/generated/informers/externalversions/cloudcontroller/v1alpha1"
	listers "github.com/gy-ulbak96/go_cloud_controller/pkg/generated/listers/cloudcontroller/v1alpha1"
	cloudclient "github.com/gy-ulbak96/go_cloud_controller/cloudclient"
)

const controllerAgentName = "cloudcontroller"

const (
	SuccessSynced = "Synced"
	ErrResourceExists = "ErrResourceExists"
	MessageResourceExists = "Resource %q already exists and is not managed by Server"
	MessageResourceSynced = "Server synced successfully"
	cloudUrl = "http://127.0.0.1:8080"
)

type Controller struct {
	kubeclientset kubernetes.Interface
	sampleclientset clientset.Interface
	serversLister        listers.ServerLister
	serversSynced        cache.InformerSynced
	workqueue workqueue.RateLimitingInterface
	recorder record.EventRecorder
}

func NewController(
	kubeclientset kubernetes.Interface,
	sampleclientset clientset.Interface,
	serverInformer informers.ServerInformer) *Controller {

	utilruntime.Must(samplescheme.AddToScheme(scheme.Scheme))
	klog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartStructuredLogging(0)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})

	controller := &Controller{
		kubeclientset:     kubeclientset,
		sampleclientset:   sampleclientset,
		serversLister:        serverInformer.Lister(),
		serversSynced:        serverInformer.Informer().HasSynced,
		workqueue:         workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "Servers"),
		recorder:          recorder,
	}

	klog.Info("Setting up event handlers")

	serverInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueueServer,
		UpdateFunc: func(old, new interface{}) {
			controller.enqueueServer(new)
		},
		DeleteFunc: controller.enqueueServer,
	})
	
	return controller
}


func (c *Controller) Run(workers int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.workqueue.ShutDown()

	klog.Info("Starting Server controller")

	klog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.serversSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	klog.Info("Starting workers")
	for i := 0; i < workers; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	klog.Info("Started workers")
	<-stopCh
	klog.Info("Shutting down workers")

	return nil
}


func (c *Controller) runWorker() {
	for c.processNextWorkItem() {
	}
}


func (c *Controller) processNextWorkItem() bool {
	obj, shutdown := c.workqueue.Get()

	if shutdown {
		return false
	}

	err := func(obj interface{}) error {
		defer c.workqueue.Done(obj)
		var key string
		var ok bool
		if key, ok = obj.(string); !ok {
			c.workqueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}

		if err := c.syncHandler(key); err != nil {
			c.workqueue.AddRateLimited(key)
			return fmt.Errorf("error syncing '%s': %s, requeuing", key, err.Error())
		}
		c.workqueue.Forget(obj)
		klog.Infof("Successfully synced '%s'", key)
		return nil
	}(obj)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}

	return true
}


func (c *Controller) syncHandler(key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	server, err := c.serversLister.Servers(namespace).Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("server '%s' in work queue no longer exists", key))
			// C1 := cloudclient.CloudClient{"http://127.0.0.1:8080"}
			// C1.DeleteServer(Realserver)
			return nil
		}
		return err
	}

	serverName := server.Spec.ServerName
	if serverName == "" {
		utilruntime.HandleError(fmt.Errorf("%s: servername must be specified", key))
		return nil
	}

  client, err := cloudclient.CreateClient(cloudUrl)  
  if err != nil {                                                                                                          
      utilruntime.HandleError(fmt.Errorf("%+v", err.Error()))
      return nil
  }

   if server.Status.ServerId != "" {     

        // Server object의 deletion timestamp와 finalizer field를 가져옵니다.                                                                          
        deletionTimestamp := server.ObjectMeta.GetDeletionTimestamp()
        finalizer := server.ObjectMeta.GetFinalizers()
        

        // finalizer와 deletion timestamp가 있으면 실행합니다. (사용자가 Server object를 삭제한 경우)
        if len(finalizer) > 0  && deletionTimestamp != nil {
            // Cloud API를 호출하여 Cloud에 Server가 실제 있는지 체크합니다.
            serverId := server.Status.ServerId
            _, errGetServer := client.GetServer(serverId)
            if errGetServer != nil {
                utilruntime.HandleError(fmt.Errorf("%+v", errGetServer.Error()))
                return nil
            }

            // Cloud API를 호출하여 server를 삭제합니다.

            errDeleteServer := deleteServer(serverId)
            if errDeleteServer != nil {
                utilruntime.HandleError(fmt.Errorf("%+v", errDeleteServer.Error()))
                return nil
            }
            fmt.Println("Success to delete on " + serverName)

            // Server  object에서 finalizer를 제거합니다. 기존 server object를 직접 수정하지 않고 DeepCopy를 통해 복사한 object를 수정한 뒤 업데이트합니다.

            serverCopy := server.DeepCopy()
            serverCopy.ObjectMeta.SetFinalizers([]string{})
            _, errUpdate := c.sampleclientset.CloudcontrollerV1alpha1().Servers(server.Namespace).Update(context.TODO(), serverCopy, metav1.UpdateOptions{})
            if errUpdate != nil {
                utilruntime.HandleError(fmt.Errorf("%+v", errUpdate.Error()))
                return nil
            }
            fmt.Println("Success to remove a finalizer on " + serverName)
            return nil
        }

        // finalizer와 deletion timestamp가 없으면 실행합니다.(Server가 생성된 직후 최초 한번 실행. finalizer를 추가해주어 실제 서버가 삭제되기 전까지

        // server object가 제거되지 않도록 유지하기 위한 목적)

        if len(finalizer) == 0 && deletionTimestamp == nil {

            // Server object에 finalizer를 추가합니다.(위와 마찬가지로 deepcopy를 통해 업데이트)
            serverCopy := server.DeepCopy()
            serverCopy.ObjectMeta.SetFinalizers([]string{"cloudcontroller.k8s.io"})
            _, err := c.sampleclientset.CloudcontrollerV1alpha1().Servers(server.Namespace).Update(context.TODO(), serverCopy, metav1.UpdateOptions{})
            if err != nil {
                utilruntime.HandleError(fmt.Errorf("%+v", err.Error()))
                return nil
            }
            fmt.Println("Success to add a finalizer " + serverName)
        }

        // Cloud API를 호출하여 server가 생성되어 있는지 체크합니다. 생성되어 있으면 nil을 return하여 handler를 종료합니다.

        res, _ := client.GetServer(server.Status.ServerId)
        if res.Name != "" {
            fmt.Println(res.Name + " has already been created. Skip to create the Server")
            serverStatus := cloudcontrollerv1alpha1.ServerStatus{
                ServerId: server.Status.ServerId,
            }
            err = c.updateServerStatus(server, serverStatus)
            if err != nil {
                return err
            }
            return nil
        }
    }

	serverSpec := cloudclient.ServerSpec{
		Name: serverName,
	}

  res,err := client.CreateServer(&serverSpec)
	if err != nil{
		utilruntime.HandleError(fmt.Errorf("%+v", err.Error()))
    return nil
	}
	
	serverId := res.Id
  serverStatus := cloudcontrollerv1alpha1.ServerStatus{
		ServerId: serverId,
	}
	

	klog.Infof("30sec return")
	err = c.updateServerStatus(server, serverStatus)
	if err != nil {
		return err
	}

	c.recorder.Event(server, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
	return nil
}



func (c *Controller) updateServerStatus(server *cloudcontrollerv1alpha1.Server, serverStatus cloudcontrollerv1alpha1.ServerStatus) error {
	serverCopy := server.DeepCopy()
	//serverid가 없을 경우에 serverid주입
  serverCopy.Status.ServerId = serverStatus.ServerId
	_, err := c.sampleclientset.CloudcontrollerV1alpha1().Servers(server.Namespace).UpdateStatus(context.TODO(), serverCopy, metav1.UpdateOptions{})
	return err
}


func (c *Controller) enqueueServer(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}
	c.workqueue.Add(key)
}


func deleteServer(id string) error{
	  client, err := cloudclient.CreateClient(cloudUrl)
    if err != nil {
        utilruntime.HandleError(fmt.Errorf("%+v", err.Error()))
        return err
    }

    errDeleteServer := client.DeleteServer(id)
    if errDeleteServer != nil {
        utilruntime.HandleError(fmt.Errorf("%+v", errDeleteServer.Error()))
        return errDeleteServer
    }

    return nil
}
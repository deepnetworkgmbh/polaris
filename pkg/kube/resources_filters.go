package kube

import (
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	)

func filterNamespaces(rp *ResourceProvider, namespaces []string) {
	filtered := rp.Namespaces[:0]
	for _, x := range rp.Namespaces {
		for _, ns := range namespaces {
			if x.Name == ns {
				filtered = append(filtered, x)
			}
		}
	}
	for i := len(filtered); i < len(rp.Namespaces); i++ {
		rp.Namespaces[i] = corev1.Namespace{}
	}
	rp.Namespaces = filtered
}

func filterDeployments(rp *ResourceProvider, namespaces []string) {
	filtered := rp.Deployments[:0]
	for _, x := range rp.Deployments {
		for _, ns := range namespaces {
			if x.Namespace == ns {
				filtered = append(filtered, x)
			}
		}
	}
	for i := len(filtered); i < len(rp.Deployments); i++ {
		rp.Deployments[i] = appsv1.Deployment{}
	}
	rp.Deployments = filtered
}

func filterStatefulSets(rp *ResourceProvider, namespaces []string) {
	filtered := rp.StatefulSets[:0]
	for _, x := range rp.StatefulSets {
		for _, ns := range namespaces {
			if x.Namespace == ns {
				filtered = append(filtered, x)
			}
		}
	}
	for i := len(filtered); i < len(rp.StatefulSets); i++ {
		rp.StatefulSets[i] = appsv1.StatefulSet{}
	}
	rp.StatefulSets = filtered
}

func filterDaemonSets(rp *ResourceProvider, namespaces []string) {
	filtered := rp.DaemonSets[:0]
	for _, x := range rp.DaemonSets {
		for _, ns := range namespaces {
			if x.Namespace == ns {
				filtered = append(filtered, x)
			}
		}
	}
	for i := len(filtered); i < len(rp.DaemonSets); i++ {
		rp.DaemonSets[i] = appsv1.DaemonSet{}
	}
	rp.DaemonSets = filtered
}

func filterJobs(rp *ResourceProvider, namespaces []string) {
	filtered := rp.Jobs[:0]
	for _, x := range rp.Jobs {
		for _, ns := range namespaces {
			if x.Namespace == ns {
				filtered = append(filtered, x)
			}
		}
	}
	for i := len(filtered); i < len(rp.Jobs); i++ {
		rp.Jobs[i] = batchv1.Job{}
	}
	rp.Jobs = filtered
}

func filterCronJobs(rp *ResourceProvider, namespaces []string) {
	filtered := rp.CronJobs[:0]
	for _, x := range rp.CronJobs {
		for _, ns := range namespaces {
			if x.Namespace == ns {
				filtered = append(filtered, x)
			}
		}
	}
	for i := len(filtered); i < len(rp.CronJobs); i++ {
		rp.CronJobs[i] = batchv1beta1.CronJob{}
	}
	rp.CronJobs = filtered
}

func filterReplicationControllers(rp *ResourceProvider, namespaces []string) {
	filtered := rp.ReplicationControllers[:0]
	for _, x := range rp.ReplicationControllers {
		for _, ns := range namespaces {
			if x.Namespace == ns {
				filtered = append(filtered, x)
			}
		}
	}
	for i := len(filtered); i < len(rp.ReplicationControllers); i++ {
		rp.ReplicationControllers[i] = corev1.ReplicationController{}
	}
	rp.ReplicationControllers = filtered
}

func filterPods(rp *ResourceProvider, namespaces []string) {
	filtered := rp.Pods[:0]
	for _, x := range rp.Pods {
		for _, ns := range namespaces {
			if x.Namespace == ns {
				filtered = append(filtered, x)
			}
		}
	}
	for i := len(filtered); i < len(rp.Pods); i++ {
		rp.Pods[i] = corev1.Pod{}
	}
	rp.Pods = filtered
}

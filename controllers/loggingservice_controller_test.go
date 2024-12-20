package controllers

import (
	"testing"

	loggingService "github.com/Netcracker/qubership-logging-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var containerRuntimeTests = []struct {
	description          string
	cr                   loggingService.LoggingService
	nodes                []*corev1.Node
	containerRuntimeType string
}{
	{
		"Container runtime type is not set and there is no nodes to get",
		loggingService.LoggingService{
			Spec: loggingService.LoggingServiceSpec{
				ContainerRuntimeType: "",
			},
		},
		[]*corev1.Node{},
		DefaultContainerRuntimeType,
	},
	{
		"Container runtime type is set and there is no nodes to get",
		loggingService.LoggingService{
			Spec: loggingService.LoggingServiceSpec{
				ContainerRuntimeType: "cri-o",
			},
		},
		[]*corev1.Node{},
		"cri-o",
	},
	{
		"Container runtime type is set and there is one node",
		loggingService.LoggingService{
			Spec: loggingService.LoggingServiceSpec{
				ContainerRuntimeType: "cri-o",
			},
		},
		[]*corev1.Node{
			&corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-1",
				},
				Status: corev1.NodeStatus{
					NodeInfo: corev1.NodeSystemInfo{
						ContainerRuntimeVersion: "containerd://1.4.2",
					},
				},
			},
		},
		"cri-o",
	},
	{
		"Container runtime type is set and there are two nodes",
		loggingService.LoggingService{
			Spec: loggingService.LoggingServiceSpec{
				ContainerRuntimeType: "docker",
			},
		},
		[]*corev1.Node{
			&corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-1",
				},
				Status: corev1.NodeStatus{
					NodeInfo: corev1.NodeSystemInfo{
						ContainerRuntimeVersion: "containerd://1.4.2",
					},
				},
			},
			&corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-2",
				},
				Status: corev1.NodeStatus{
					NodeInfo: corev1.NodeSystemInfo{
						ContainerRuntimeVersion: "cri-o://1.29.6-3.rhaos4.16.gitfd433b7.el9",
					},
				},
			},
		},
		"docker",
	},
	{
		"Container runtime type is not set and there is one node",
		loggingService.LoggingService{},
		[]*corev1.Node{
			&corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-1",
				},
				Status: corev1.NodeStatus{
					NodeInfo: corev1.NodeSystemInfo{
						ContainerRuntimeVersion: "cri-o://1.29.6-3.rhaos4.16.gitfd433b7.el9",
					},
				},
			},
		},
		"cri-o",
	},
	{
		"Container runtime type is not set and there are two nodes with the same container runtime type",
		loggingService.LoggingService{},
		[]*corev1.Node{
			&corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-1",
				},
				Status: corev1.NodeStatus{
					NodeInfo: corev1.NodeSystemInfo{
						ContainerRuntimeVersion: "cri-o://1.29.6-3.rhaos4.16.gitfd433b7.el9",
					},
				},
			},
			&corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-2",
				},
				Status: corev1.NodeStatus{
					NodeInfo: corev1.NodeSystemInfo{
						ContainerRuntimeVersion: "cri-o://1.29.6-3.rhaos4.16.gitfd433b7.el9",
					},
				},
			},
		},
		"cri-o",
	},
	{
		"Container runtime type is not set and there are two nodes with different container runtime types",
		loggingService.LoggingService{},
		[]*corev1.Node{
			&corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-1",
				},
				Status: corev1.NodeStatus{
					NodeInfo: corev1.NodeSystemInfo{
						ContainerRuntimeVersion: "cri-o://1.29.6-3.rhaos4.16.gitfd433b7.el9",
					},
				},
			},
			&corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-2",
				},
				Status: corev1.NodeStatus{
					NodeInfo: corev1.NodeSystemInfo{
						ContainerRuntimeVersion: "docker://24.0.8",
					},
				},
			},
		},
		DefaultContainerRuntimeType,
	},
}

func Test_updateDynamicParameters(t *testing.T) {
	testScheme := runtime.NewScheme()
	err := loggingService.AddToScheme(testScheme)
	if err != nil {
		t.Error("can't add test schema in arrays of schemas")
	}
	err = corev1.AddToScheme(testScheme)
	if err != nil {
		t.Error("can't add test schema in arrays of schemas")
	}
	for _, tt := range containerRuntimeTests {
		t.Run(tt.description, func(t *testing.T) {
			fakeClientBuilder := fake.NewClientBuilder().WithScheme(testScheme)
			for _, node := range tt.nodes {
				fakeClientBuilder = fakeClientBuilder.WithObjects(node)
			}
			fakeClient := fakeClientBuilder.Build()
			reconciler := &LoggingServiceReconciler{
				Client: fakeClient,
				Scheme: testScheme,
			}
			reconciler.updateDynamicParameters(&tt.cr)
			if reconciler.DynamicParameters.ContainerRuntimeType != tt.containerRuntimeType {
				t.Errorf("expected %q, got %q", tt.containerRuntimeType, reconciler.DynamicParameters.ContainerRuntimeType)
			}
		})
	}
}

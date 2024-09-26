package e2e

import (
	"context"
	"fmt"

	"github.com/VictoriaMetrics/operator/api/operator/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1beta1vm "github.com/VictoriaMetrics/operator/api/operator/v1beta1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
)

//nolint:dupl,lll
var _ = Describe("e2e vmcluster", func() {
	namespace := "default"
	var ctx context.Context
	namespacedName := types.NamespacedName{
		Namespace: namespace,
	}
	It("must clean up previous test resutls", func() {
		ctx = context.Background()
		// clean up before tests
		Expect(k8sClient.DeleteAllOf(ctx, &v1beta1vm.VMCluster{}, &client.DeleteAllOfOptions{
			ListOptions: client.ListOptions{
				Namespace: namespace,
			},
		})).To(Succeed())
		Eventually(func() bool {
			var unDeletedObjects v1beta1vm.VMClusterList
			Expect(k8sClient.List(ctx, &unDeletedObjects, &client.ListOptions{
				Namespace: namespace,
			})).To(Succeed())
			return len(unDeletedObjects.Items) == 0
		}, eventualDeletionTimeout).Should(BeTrue())

	})

	Context("create", func() {
		JustBeforeEach(func() {
			ctx = context.Background()
		})
		AfterEach(func() {
			Expect(k8sClient.Delete(ctx, &v1beta1vm.VMCluster{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: namespace,
					Name:      namespacedName.Name,
				},
			})).To(Succeed(), "must delete vmcluster after test")
			Eventually(func() error {
				err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      namespacedName.Name,
					Namespace: namespace,
				}, &v1beta1vm.VMCluster{})
				if errors.IsNotFound(err) {
					return nil
				}
				return fmt.Errorf("want NotFound error, got: %w", err)
			}, eventualDeletionTimeout, 1).Should(BeNil())
		})

		DescribeTable("should create vmcluster", func(name string, cr *v1beta1vm.VMCluster) {
			namespacedName.Name = name
			cr.Name = name
			Expect(k8sClient.Create(ctx, cr)).To(Succeed())
			Eventually(func() error {
				return expectObjectStatusOperational(ctx, k8sClient, &v1beta1vm.VMCluster{}, namespacedName)
			}, eventualStatefulsetAppReadyTimeout).Should(Succeed())

		},
			Entry("without any componets", "empty", &v1beta1vm.VMCluster{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: namespace,
					Name:      namespacedName.Name,
				},
				Spec: v1beta1vm.VMClusterSpec{RetentionPeriod: "1"},
			}),
			Entry("with all components", "all-services",
				&v1beta1vm.VMCluster{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: namespace,
						Name:      namespacedName.Name,
					},
					Spec: v1beta1vm.VMClusterSpec{
						RetentionPeriod: "1",
						VMStorage: &v1beta1vm.VMStorage{
							CommonApplicationDeploymentParams: v1beta1vm.CommonApplicationDeploymentParams{
								ReplicaCount: ptr.To[int32](1),
							},
						},
						VMSelect: &v1beta1vm.VMSelect{
							CommonApplicationDeploymentParams: v1beta1vm.CommonApplicationDeploymentParams{
								ReplicaCount: ptr.To[int32](1),
							},
						},
						VMInsert: &v1beta1vm.VMInsert{CommonApplicationDeploymentParams: v1beta1vm.CommonApplicationDeploymentParams{
							ReplicaCount: ptr.To[int32](1),
						},
						},
					},
				},
			),
			Entry("with vmstorage and vmselect", "with-select", &v1beta1vm.VMCluster{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: namespace,
					Name:      namespacedName.Name,
				},
				Spec: v1beta1vm.VMClusterSpec{
					RetentionPeriod: "1",
					VMStorage: &v1beta1vm.VMStorage{
						CommonApplicationDeploymentParams: v1beta1vm.CommonApplicationDeploymentParams{
							ReplicaCount: ptr.To[int32](1),
						},
					},
					VMSelect: &v1beta1vm.VMSelect{
						CommonApplicationDeploymentParams: v1beta1vm.CommonApplicationDeploymentParams{
							ReplicaCount: ptr.To[int32](1),
						},
					},
				},
			}),
			Entry("with vmstorage and vminsert", "with-insert", &v1beta1vm.VMCluster{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: namespace,
					Name:      namespacedName.Name,
				},
				Spec: v1beta1vm.VMClusterSpec{
					RetentionPeriod: "1",
					VMStorage: &v1beta1vm.VMStorage{
						CommonApplicationDeploymentParams: v1beta1vm.CommonApplicationDeploymentParams{
							ReplicaCount: ptr.To[int32](1),
						},
					},
					VMInsert: &v1beta1vm.VMInsert{CommonApplicationDeploymentParams: v1beta1vm.CommonApplicationDeploymentParams{
						ReplicaCount: ptr.To[int32](1),
					},
					},
				},
			}),
			Entry("with security enable and without default resources", "all-secure",
				&v1beta1vm.VMCluster{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: namespace,
						Name:      namespacedName.Name,
					},
					Spec: v1beta1vm.VMClusterSpec{
						RetentionPeriod: "1",
						VMStorage: &v1beta1vm.VMStorage{
							CommonDefaultableParams: v1beta1vm.CommonDefaultableParams{
								//								UseStrictSecurity:   ptr.To(true),
								UseDefaultResources: ptr.To(false),
							},
							CommonApplicationDeploymentParams: v1beta1vm.CommonApplicationDeploymentParams{
								ReplicaCount: ptr.To[int32](1),
							},
						},
						VMSelect: &v1beta1vm.VMSelect{
							CommonDefaultableParams: v1beta1vm.CommonDefaultableParams{
								//							UseStrictSecurity:   ptr.To(true),
								UseDefaultResources: ptr.To(false),
							},

							CommonApplicationDeploymentParams: v1beta1vm.CommonApplicationDeploymentParams{
								ReplicaCount: ptr.To[int32](1),
							},
						},
						VMInsert: &v1beta1vm.VMInsert{
							CommonDefaultableParams: v1beta1vm.CommonDefaultableParams{
								//						UseStrictSecurity:   ptr.To(true),
								UseDefaultResources: ptr.To(false),
							},
							CommonApplicationDeploymentParams: v1beta1vm.CommonApplicationDeploymentParams{
								ReplicaCount: ptr.To[int32](1),
							},
						},
					},
				},
			),
		)
	})
	Context("update", func() {
		AfterEach(func() {
			Expect(k8sClient.Delete(ctx, &v1beta1vm.VMCluster{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: namespace,
					Name:      namespacedName.Name,
				},
			})).To(Succeed())
			Eventually(func() error {
				return k8sClient.Get(context.Background(), types.NamespacedName{
					Name:      namespacedName.Name,
					Namespace: namespace,
				}, &v1beta1vm.VMCluster{})

			}, eventualDeletionTimeout).Should(MatchError(errors.IsNotFound, "want not found error"))
		})

		type testStep struct {
			setup  func(*v1beta1vm.VMCluster)
			modify func(*v1beta1vm.VMCluster)
			verify func(*v1beta1vm.VMCluster)
		}

		DescribeTable("should update exist cluster",
			func(name string, initCR *v1beta1vm.VMCluster, steps ...testStep) {
				namespacedName.Name = name
				initCR.Namespace = namespace
				initCR.Name = name
				ctx = context.Background()
				Expect(k8sClient.Create(ctx, initCR)).To(Succeed())
				Eventually(func() error {
					return expectObjectStatusOperational(ctx, k8sClient, initCR, namespacedName)
				}, eventualStatefulsetAppReadyTimeout).
					Should(Succeed())
				for _, step := range steps {
					if step.setup != nil {
						step.setup(initCR)
					}
					// update and wait ready
					Eventually(func() error {
						var toUpdate v1beta1.VMCluster
						if err := k8sClient.Get(ctx, namespacedName, &toUpdate); err != nil {
							return err
						}
						step.modify(&toUpdate)
						return k8sClient.Update(ctx, &toUpdate)
					}, eventualExpandingTimeout).Should(Succeed())
					Eventually(func() error {
						return expectObjectStatusExpanding(ctx, k8sClient, initCR, namespacedName)
					}, eventualStatefulsetAppReadyTimeout).
						Should(Succeed())
					Eventually(func() error {
						return expectObjectStatusOperational(ctx, k8sClient, initCR, namespacedName)
					}, eventualStatefulsetAppReadyTimeout).
						Should(Succeed())
					var updated v1beta1vm.VMCluster
					Expect(k8sClient.Get(ctx, namespacedName, &updated)).To(Succeed())
					// verify results
					step.verify(&updated)
				}
			},
			Entry("by scaling select and storage replicas to 2", "storage-select-r-2",
				&v1beta1vm.VMCluster{
					Spec: v1beta1vm.VMClusterSpec{
						RetentionPeriod: "1",
						VMStorage: &v1beta1vm.VMStorage{
							CommonApplicationDeploymentParams: v1beta1vm.CommonApplicationDeploymentParams{
								ReplicaCount: ptr.To[int32](1),
							},
						},
						VMSelect: &v1beta1vm.VMSelect{
							CommonApplicationDeploymentParams: v1beta1vm.CommonApplicationDeploymentParams{
								ReplicaCount: ptr.To[int32](1),
							},
						},
						VMInsert: &v1beta1vm.VMInsert{CommonApplicationDeploymentParams: v1beta1vm.CommonApplicationDeploymentParams{
							ReplicaCount: ptr.To[int32](1),
						},
						},
					},
				},
				testStep{
					modify: func(cr *v1beta1vm.VMCluster) {
						cr.Spec.VMStorage.ReplicaCount = ptr.To[int32](2)
						cr.Spec.VMSelect.ReplicaCount = ptr.To[int32](2)
					},
					verify: func(cr *v1beta1vm.VMCluster) {
						Expect(expectPodCount(k8sClient, 2, namespace, cr.VMStorageSelectorLabels())).To(BeEmpty())
						Expect(expectPodCount(k8sClient, 2, namespace, cr.VMSelectSelectorLabels())).To(BeEmpty())
					},
				},
			),
			Entry("by scaling storage and insert replicas to 2", "storage-insert-r-2",
				&v1beta1vm.VMCluster{
					Spec: v1beta1vm.VMClusterSpec{
						RetentionPeriod: "1",
						VMStorage: &v1beta1vm.VMStorage{
							CommonApplicationDeploymentParams: v1beta1vm.CommonApplicationDeploymentParams{
								ReplicaCount: ptr.To[int32](1),
							},
						},
						VMSelect: &v1beta1vm.VMSelect{
							CommonApplicationDeploymentParams: v1beta1vm.CommonApplicationDeploymentParams{
								ReplicaCount: ptr.To[int32](1),
							},
						},
						VMInsert: &v1beta1vm.VMInsert{CommonApplicationDeploymentParams: v1beta1vm.CommonApplicationDeploymentParams{
							ReplicaCount: ptr.To[int32](1),
						},
						},
					},
				},
				testStep{
					modify: func(cr *v1beta1vm.VMCluster) {
						cr.Spec.VMStorage.ReplicaCount = ptr.To[int32](2)
						cr.Spec.VMInsert.ReplicaCount = ptr.To[int32](2)
					},
					verify: func(cr *v1beta1vm.VMCluster) {
						Expect(expectPodCount(k8sClient, 2, namespace, cr.VMStorageSelectorLabels())).To(BeEmpty())
						Eventually(func() string {
							return expectPodCount(k8sClient, 2, namespace, cr.VMInsertSelectorLabels())
						}, eventualDeploymentPodTimeout).Should(BeEmpty())
					},
				},
			),
			Entry("by changing storage revisionHistoryLimit to 2", "storage-revision-2",
				&v1beta1vm.VMCluster{
					Spec: v1beta1vm.VMClusterSpec{
						RetentionPeriod: "1",
						VMStorage: &v1beta1vm.VMStorage{
							CommonApplicationDeploymentParams: v1beta1vm.CommonApplicationDeploymentParams{
								ReplicaCount: ptr.To[int32](1),
							},
						},
						VMSelect: &v1beta1vm.VMSelect{
							CommonApplicationDeploymentParams: v1beta1vm.CommonApplicationDeploymentParams{
								ReplicaCount: ptr.To[int32](1),
							},
						},
						VMInsert: &v1beta1vm.VMInsert{CommonApplicationDeploymentParams: v1beta1vm.CommonApplicationDeploymentParams{
							ReplicaCount: ptr.To[int32](1),
						},
						},
					},
				},
				testStep{
					modify: func(cr *v1beta1vm.VMCluster) {
						cr.Spec.VMStorage.RevisionHistoryLimitCount = ptr.To[int32](2)
					},
					verify: func(cr *v1beta1vm.VMCluster) {
						var updatedCluster v1beta1vm.VMCluster
						Expect(k8sClient.Get(ctx, namespacedName, &updatedCluster)).To(Succeed())
						Expect(*updatedCluster.Spec.VMStorage.RevisionHistoryLimitCount).To(Equal(int32(2)))
					},
				},
			),
			Entry("by adding clusterNative ports", "storage-native-r-2",
				&v1beta1vm.VMCluster{
					Spec: v1beta1vm.VMClusterSpec{
						RetentionPeriod: "1",
						VMStorage: &v1beta1vm.VMStorage{
							CommonApplicationDeploymentParams: v1beta1vm.CommonApplicationDeploymentParams{
								ReplicaCount: ptr.To[int32](1),
							},
						},
						VMSelect: &v1beta1vm.VMSelect{
							CommonApplicationDeploymentParams: v1beta1vm.CommonApplicationDeploymentParams{
								ReplicaCount: ptr.To[int32](1),
							},
						},
						VMInsert: &v1beta1vm.VMInsert{CommonApplicationDeploymentParams: v1beta1vm.CommonApplicationDeploymentParams{
							ReplicaCount: ptr.To[int32](1),
						},
						},
					},
				},
				testStep{
					modify: func(cr *v1beta1vm.VMCluster) {
						cr.Spec.VMInsert.ClusterNativePort = "8035"
						cr.Spec.VMSelect.ClusterNativePort = "8036"
					},
					verify: func(cr *v1beta1vm.VMCluster) {
						var updatedSvc corev1.Service
						Expect(k8sClient.Get(ctx,
							types.NamespacedName{Namespace: namespace, Name: "vmselect-" + cr.Name},
							&updatedSvc)).
							To(Succeed())
						Expect(updatedSvc.Spec.Ports).To(HaveLen(2))
						Expect(updatedSvc.Spec.Ports[1].Port).To(Equal(int32(8036)))
						Expect(k8sClient.Get(ctx,
							types.NamespacedName{Namespace: namespace, Name: "vminsert-" + cr.Name},
							&updatedSvc)).
							To(Succeed())
						Expect(updatedSvc.Spec.Ports).To(HaveLen(2))
						Expect(updatedSvc.Spec.Ports[1].Port).To(Equal(int32(8035)))
					},
				},
			),
			Entry("by deleting select component", "select-delete",
				&v1beta1vm.VMCluster{
					Spec: v1beta1vm.VMClusterSpec{
						RetentionPeriod: "1",
						VMStorage: &v1beta1vm.VMStorage{
							CommonApplicationDeploymentParams: v1beta1vm.CommonApplicationDeploymentParams{
								ReplicaCount: ptr.To[int32](1),
							},
						},
						VMSelect: &v1beta1vm.VMSelect{
							CommonApplicationDeploymentParams: v1beta1vm.CommonApplicationDeploymentParams{
								ReplicaCount: ptr.To[int32](1),
							},
						},
						VMInsert: &v1beta1vm.VMInsert{CommonApplicationDeploymentParams: v1beta1vm.CommonApplicationDeploymentParams{
							ReplicaCount: ptr.To[int32](1),
						},
						},
					},
				},
				testStep{
					setup: func(cr *v1beta1vm.VMCluster) {
						Expect(k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: "vmselect-" + cr.Name}, &appsv1.StatefulSet{})).To(Succeed())
						Expect(k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: "vmselect-" + cr.Name}, &corev1.Service{})).To(Succeed())
						Expect(k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: "vmselect-" + cr.Name}, &v1beta1vm.VMServiceScrape{})).To(Succeed())
					},
					modify: func(cr *v1beta1vm.VMCluster) {
						cr.Spec.VMSelect = nil
					},
					verify: func(cr *v1beta1vm.VMCluster) {
						Eventually(func() error {
							return k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: "vmselect-" + cr.Name}, &appsv1.StatefulSet{})
						}, eventualDeletionTimeout).Should(MatchError(errors.IsNotFound, "IsNotFound"))
						Eventually(func() error {
							return k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: "vmselect-" + cr.Name}, &corev1.Service{})
						}, eventualDeletionTimeout).Should(MatchError(errors.IsNotFound, "IsNotFound"))
						Eventually(func() error {
							return k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: "vmselect-" + cr.Name}, &v1beta1vm.VMServiceScrape{})
						}, eventualDeletionTimeout).Should(MatchError(errors.IsNotFound, "IsNotFound"))

					},
				},
				testStep{
					modify: func(cr *v1beta1vm.VMCluster) {
						cr.Spec.VMSelect = &v1beta1vm.VMSelect{
							CommonDefaultableParams: v1beta1vm.CommonDefaultableParams{
								UseDefaultResources: ptr.To(false),
							},
							CommonApplicationDeploymentParams: v1beta1vm.CommonApplicationDeploymentParams{
								ReplicaCount: ptr.To[int32](1),
							},
						}
					},
					verify: func(cr *v1beta1vm.VMCluster) {
						Expect(k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: "vmselect-" + cr.Name}, &appsv1.StatefulSet{})).To(Succeed())
						Expect(k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: "vmselect-" + cr.Name}, &corev1.Service{})).To(Succeed())
						Expect(k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: "vmselect-" + cr.Name}, &v1beta1vm.VMServiceScrape{})).To(Succeed())
					},
				},
			),
			Entry("by deleting storage and insert components", "storage-insert-delete",
				&v1beta1vm.VMCluster{
					Spec: v1beta1vm.VMClusterSpec{
						RetentionPeriod: "1",
						VMStorage: &v1beta1vm.VMStorage{
							CommonApplicationDeploymentParams: v1beta1vm.CommonApplicationDeploymentParams{
								ReplicaCount: ptr.To[int32](1),
							},
						},
						VMSelect: &v1beta1vm.VMSelect{
							CommonApplicationDeploymentParams: v1beta1vm.CommonApplicationDeploymentParams{
								ReplicaCount: ptr.To[int32](1),
							},
						},
						VMInsert: &v1beta1vm.VMInsert{CommonApplicationDeploymentParams: v1beta1vm.CommonApplicationDeploymentParams{
							ReplicaCount: ptr.To[int32](1),
						},
						},
					},
				},
				testStep{
					setup: func(cr *v1beta1vm.VMCluster) {
						Expect(k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: "vmstorage-" + cr.Name}, &appsv1.StatefulSet{})).To(Succeed())
						Expect(k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: "vmstorage-" + cr.Name}, &corev1.Service{})).To(Succeed())
						Expect(k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: "vmstorage-" + cr.Name}, &v1beta1vm.VMServiceScrape{})).To(Succeed())
						Expect(k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: "vminsert-" + cr.Name}, &appsv1.Deployment{})).To(Succeed())
						Expect(k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: "vminsert-" + cr.Name}, &corev1.Service{})).To(Succeed())
						Expect(k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: "vminsert-" + cr.Name}, &v1beta1vm.VMServiceScrape{})).To(Succeed())
					},
					modify: func(cr *v1beta1vm.VMCluster) {
						cr.Spec.VMStorage = nil
						cr.Spec.VMInsert = nil
						cr.Spec.VMSelect.ExtraArgs = map[string]string{
							"storageNode": "http://non-exist-vmstorage:8402",
						}
					},
					verify: func(cr *v1beta1vm.VMCluster) {
						Eventually(func() error {
							return k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: "vmstorage-" + cr.Name}, &appsv1.StatefulSet{})
						}, eventualDeletionTimeout).Should(MatchError(errors.IsNotFound, "IsNotFound"))
						Eventually(func() error {
							return k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: "vmstorage-" + cr.Name}, &corev1.Service{})
						}, eventualDeletionTimeout).Should(MatchError(errors.IsNotFound, "IsNotFound"))
						Eventually(func() error {
							return k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: "vmstorage-" + cr.Name}, &v1beta1vm.VMServiceScrape{})
						}, eventualDeletionTimeout).Should(MatchError(errors.IsNotFound, "IsNotFound"))
						Eventually(func() error {
							return k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: "vminsert-" + cr.Name}, &appsv1.Deployment{})
						}, eventualDeletionTimeout).Should(MatchError(errors.IsNotFound, "IsNotFound"))

					},
				},
				testStep{
					modify: func(cr *v1beta1vm.VMCluster) {
						cr.Spec.VMStorage = &v1beta1vm.VMStorage{
							CommonDefaultableParams: v1beta1vm.CommonDefaultableParams{
								UseDefaultResources: ptr.To(false),
							},
							CommonApplicationDeploymentParams: v1beta1vm.CommonApplicationDeploymentParams{
								ReplicaCount: ptr.To[int32](1),
							},
						}
						cr.Spec.VMInsert = &v1beta1vm.VMInsert{
							CommonDefaultableParams: v1beta1vm.CommonDefaultableParams{
								UseDefaultResources: ptr.To(false),
							},
							CommonApplicationDeploymentParams: v1beta1vm.CommonApplicationDeploymentParams{
								ReplicaCount: ptr.To[int32](1),
							},
						}
					},
					verify: func(cr *v1beta1vm.VMCluster) {
						Expect(k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: "vmstorage-" + cr.Name}, &appsv1.StatefulSet{})).To(Succeed())
						Expect(k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: "vmstorage-" + cr.Name}, &corev1.Service{})).To(Succeed())
						Expect(k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: "vmstorage-" + cr.Name}, &v1beta1vm.VMServiceScrape{})).To(Succeed())
						Expect(k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: "vminsert-" + cr.Name}, &appsv1.Deployment{})).To(Succeed())
						Expect(k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: "vminsert-" + cr.Name}, &corev1.Service{})).To(Succeed())
						Expect(k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: "vminsert-" + cr.Name}, &v1beta1vm.VMServiceScrape{})).To(Succeed())
					},
				},
			),
			Entry("by deleting deleting and renaming additional services", "select-additional-svc",
				&v1beta1vm.VMCluster{
					Spec: v1beta1vm.VMClusterSpec{
						RetentionPeriod: "1",
						VMStorage: &v1beta1vm.VMStorage{
							CommonApplicationDeploymentParams: v1beta1vm.CommonApplicationDeploymentParams{
								ReplicaCount: ptr.To[int32](1),
							},
							ServiceSpec: &v1beta1vm.AdditionalServiceSpec{
								EmbeddedObjectMetadata: v1beta1vm.EmbeddedObjectMetadata{
									Name: "my-service-name",
								},
								Spec: corev1.ServiceSpec{
									Type: corev1.ServiceTypeClusterIP,
									Ports: []corev1.ServicePort{
										{
											TargetPort: intstr.FromInt(8431),
											Name:       "web-port",
											Port:       8435,
										},
									},
								},
							},
						},
						VMSelect: &v1beta1vm.VMSelect{
							CommonApplicationDeploymentParams: v1beta1vm.CommonApplicationDeploymentParams{
								ReplicaCount: ptr.To[int32](1),
							},
							ServiceSpec: &v1beta1vm.AdditionalServiceSpec{
								Spec: corev1.ServiceSpec{
									Type: corev1.ServiceTypeClusterIP,
									Ports: []corev1.ServicePort{
										{
											TargetPort: intstr.FromInt(8431),
											Name:       "web-port",
											Port:       8435,
										},
									},
								},
							},
						},
					},
				},
				testStep{
					setup: func(cr *v1beta1vm.VMCluster) {
						Expect(k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: "vmselect-" + cr.Name + "-additional-service"}, &corev1.Service{})).To(Succeed())
						Expect(k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: "my-service-name"}, &corev1.Service{})).To(Succeed())
					},
					modify: func(cr *v1beta1vm.VMCluster) {
						cr.Spec.VMSelect.ServiceSpec = nil
						cr.Spec.VMStorage.ServiceSpec.Name = ""
					},
					verify: func(cr *v1beta1vm.VMCluster) {
						Eventually(func() error {
							return k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: "vmselect-" + cr.Name + "-additional-service"}, &corev1.Service{})
						}, eventualDeletionTimeout).Should(MatchError(errors.IsNotFound, "IsNotFound"))
						Expect(k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: "vmstorage-" + cr.Name + "-additional-service"}, &corev1.Service{})).To(Succeed())
						Expect(k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: "my-service-name"}, &corev1.Service{})).To(MatchError(errors.IsNotFound, "IsNotFound"))
					},
				},
				testStep{
					modify: func(cr *v1beta1vm.VMCluster) {
						cr.Spec.VMStorage.ServiceSpec.UseAsDefault = true
						cr.Spec.VMSelect.ServiceSpec = &v1beta1vm.AdditionalServiceSpec{
							EmbeddedObjectMetadata: v1beta1vm.EmbeddedObjectMetadata{
								Name: "my-service-name-v2",
							},
							Spec: corev1.ServiceSpec{
								Type: corev1.ServiceTypeClusterIP,
								Ports: []corev1.ServicePort{
									{
										TargetPort: intstr.FromInt(8431),
										Name:       "web-port",
										Port:       8436,
									},
								},
							},
						}
					},
					verify: func(cr *v1beta1vm.VMCluster) {
						Eventually(func() error {
							return k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: "vmstorage-" + cr.Name + "-additional-service"}, &corev1.Service{})
						}, eventualDeletionTimeout).Should(MatchError(errors.IsNotFound, "IsNotFound"))
						Expect(k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: "my-service-name-v2"}, &corev1.Service{})).To(Succeed())
						var stSvc corev1.Service
						Expect(k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: "vmstorage-" + cr.Name}, &stSvc)).To(Succeed())
						Expect(stSvc.Spec.Ports).To(HaveLen(4))
					},
				},
			),
			Entry("by adding imagePullSecret", "storage-image-pull-secret",
				&v1beta1vm.VMCluster{
					Spec: v1beta1vm.VMClusterSpec{
						RetentionPeriod:  "1",
						ImagePullSecrets: nil,
						VMStorage: &v1beta1vm.VMStorage{
							CommonApplicationDeploymentParams: v1beta1vm.CommonApplicationDeploymentParams{
								ReplicaCount: ptr.To[int32](1),
							},
						},
						VMSelect: &v1beta1vm.VMSelect{
							CommonApplicationDeploymentParams: v1beta1vm.CommonApplicationDeploymentParams{
								ReplicaCount: ptr.To[int32](1),
							},
						},
						VMInsert: &v1beta1vm.VMInsert{CommonApplicationDeploymentParams: v1beta1vm.CommonApplicationDeploymentParams{
							ReplicaCount: ptr.To[int32](1),
						},
						},
					},
				},
				testStep{
					setup: func(v *v1beta1vm.VMCluster) {
						pullSecret := corev1.Secret{
							ObjectMeta: metav1.ObjectMeta{Name: "test-pull-secret", Namespace: namespace},
							Data: map[string][]byte{
								".dockerconfigjson": []byte(`{"auths":{"test.example.com":{"username":"test","password":"12345","email":"test@example.com","auth":"dGVzdDoxMjM0NQ=="}}}`),
							},
							Type: corev1.SecretTypeDockerConfigJson,
						}
						Expect(k8sClient.Create(ctx, &pullSecret)).To(Succeed())
					},
					modify: func(cr *v1beta1vm.VMCluster) {
						cr.Spec.ImagePullSecrets = []corev1.LocalObjectReference{
							{Name: "test-pull-secret"},
						}
					},
					verify: func(cr *v1beta1vm.VMCluster) {
						var sts appsv1.StatefulSet
						nss := types.NamespacedName{Namespace: namespace, Name: cr.Spec.VMStorage.GetNameWithPrefix(cr.Name)}
						Expect(k8sClient.Get(ctx, nss, &sts)).To(Succeed())
						Expect(sts.Spec.Template.Spec.ImagePullSecrets).To(HaveLen(1))
						Expect(k8sClient.Delete(ctx,
							&corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "test-pull-secret", Namespace: namespace}})).
							To(Succeed())
					},
				},
			),
		)
	})
})

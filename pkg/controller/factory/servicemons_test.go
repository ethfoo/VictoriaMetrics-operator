package factory

import (
	"context"
	monitoringv1 "github.com/VictoriaMetrics/operator/pkg/apis/monitoring/v1"
	victoriametricsv1beta1 "github.com/VictoriaMetrics/operator/pkg/apis/victoriametrics/v1beta1"
	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sort"
	"testing"
)

func testGetScheme() *runtime.Scheme {
	s := scheme.Scheme
	s.AddKnownTypes(victoriametricsv1beta1.SchemeGroupVersion,
		&victoriametricsv1beta1.VmAgent{},
		&victoriametricsv1beta1.VmAgentList{},
		&victoriametricsv1beta1.VmAlert{},
		&victoriametricsv1beta1.VmAlertList{},
		&victoriametricsv1beta1.VmSingle{},
		&victoriametricsv1beta1.VmSingleList{},
		&victoriametricsv1beta1.VmAlertmanager{},
		&victoriametricsv1beta1.VmAlertmanagerList{},
	)
	s.AddKnownTypes(monitoringv1.SchemeGroupVersion,
		&monitoringv1.PodMonitor{},
		&monitoringv1.PodMonitorList{},
		&monitoringv1.ServiceMonitor{},
		&monitoringv1.ServiceMonitorList{},
		&monitoringv1.PrometheusRule{},
		&monitoringv1.PrometheusRuleList{},
	)
	return s
}

func TestSelectServiceMonitors(t *testing.T) {
	type args struct {
		p *victoriametricsv1beta1.VmAgent
		l logr.Logger
	}
	tests := []struct {
		name             string
		args             args
		want             []string
		wantErr          bool
		predefinedObjest []runtime.Object
	}{
		{
			name: "select service monitor inside vmagent namespace",
			args: args{
				p: &victoriametricsv1beta1.VmAgent{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "default-agent",
						Namespace: "default",
					},
					Spec: victoriametricsv1beta1.VmAgentSpec{
						ServiceMonitorSelector: &metav1.LabelSelector{},
					},
				},
				l: logf.Log.WithName("unit-test"),
			},
			predefinedObjest: []runtime.Object{
				&monitoringv1.ServiceMonitor{
					ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "default-monitor"},
					Spec:       monitoringv1.ServiceMonitorSpec{},
				},
			},
			want:    []string{"default/default-monitor"},
			wantErr: false,
		},
		{
			name: "select service monitor from namespace with filter",
			args: args{
				p: &victoriametricsv1beta1.VmAgent{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "default-agent",
						Namespace: "default",
					},
					Spec: victoriametricsv1beta1.VmAgentSpec{
						ServiceMonitorNamespaceSelector: &metav1.LabelSelector{MatchLabels: map[string]string{"name": "stage"}},
						ServiceMonitorSelector:          &metav1.LabelSelector{},
					},
				},
				l: logf.Log.WithName("unit-test"),
			},
			predefinedObjest: []runtime.Object{
				&v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "default"}},
				&monitoringv1.ServiceMonitor{
					ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "default-monitor"},
					Spec:       monitoringv1.ServiceMonitorSpec{},
				},
				&v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "stg", Labels: map[string]string{"name": "stage"}}},
				&monitoringv1.ServiceMonitor{
					ObjectMeta: metav1.ObjectMeta{Namespace: "stg", Name: "default-monitor"},
					Spec:       monitoringv1.ServiceMonitorSpec{},
				},
			},
			want:    []string{"stg/default-monitor"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := []runtime.Object{}
			obj = append(obj, tt.predefinedObjest...)
			fclient := fake.NewFakeClientWithScheme(testGetScheme(), obj...)
			got, err := SelectServiceMonitors(context.TODO(), tt.args.p, fclient)
			if (err != nil) != tt.wantErr {
				t.Errorf("SelectServiceMonitors() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			gotNames := []string{}
			for monitorName := range got {
				gotNames = append(gotNames, monitorName)
			}
			sort.Strings(gotNames)
			if !reflect.DeepEqual(gotNames, tt.want) {
				t.Errorf("SelectServiceMonitors() got = %v, want %v", gotNames, tt.want)
			}
		})
	}
}

func TestSelectPodMonitors(t *testing.T) {
	type args struct {
		p *victoriametricsv1beta1.VmAgent
		l logr.Logger
	}
	tests := []struct {
		name              string
		args              args
		want              []string
		wantErr           bool
		predefinedObjects []runtime.Object
	}{
		{
			name: "selector pod monitor at vmagent ns",
			args: args{
				p: &victoriametricsv1beta1.VmAgent{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "example-vmagent",
						Namespace: "default",
					},
					Spec: victoriametricsv1beta1.VmAgentSpec{
						PodMonitorSelector: &metav1.LabelSelector{},
					},
				},
				l: logf.Log.WithName("unit-test"),
			},
			predefinedObjects: []runtime.Object{
				&monitoringv1.PodMonitor{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pod1",
						Namespace: "default",
					},
					Spec: monitoringv1.PodMonitorSpec{},
				},
			},
			wantErr: false,
			want:    []string{"default/pod1"},
		},
		{
			name: "selector pod monitor at different ns with ns selector",
			args: args{
				p: &victoriametricsv1beta1.VmAgent{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "example-vmagent",
						Namespace: "default",
					},
					Spec: victoriametricsv1beta1.VmAgentSpec{
						PodMonitorNamespaceSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{"name": "monitoring"},
						},
					},
				},
				l: logf.Log.WithName("unit-test"),
			},
			predefinedObjects: []runtime.Object{
				&v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "monitor", Labels: map[string]string{"name": "monitoring"}}},
				&monitoringv1.PodMonitor{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pod2",
						Namespace: "monitor",
					},
					Spec: monitoringv1.PodMonitorSpec{},
				},
				&v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "default"}},
				&monitoringv1.PodMonitor{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pod1",
						Namespace: "default",
					},
					Spec: monitoringv1.PodMonitorSpec{},
				},
			},
			wantErr: false,
			want:    []string{"monitor/pod2"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := []runtime.Object{}
			obj = append(obj, tt.predefinedObjects...)
			fclient := fake.NewFakeClientWithScheme(testGetScheme(), obj...)
			got, err := SelectPodMonitors(context.TODO(), tt.args.p, fclient)
			if (err != nil) != tt.wantErr {
				t.Errorf("SelectPodMonitors() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			gotNames := []string{}
			for podName := range got {
				gotNames = append(gotNames, podName)
			}
			sort.Strings(gotNames)
			if !reflect.DeepEqual(gotNames, tt.want) {
				t.Errorf("SelectPodMonitors() got = %v, want %v", gotNames, tt.want)
			}
		})
	}
}

func Test_getCredFromConfigMap(t *testing.T) {
	type args struct {
		ns       string
		sel      v1.ConfigMapKeySelector
		cacheKey string
		cache    map[string]*v1.ConfigMap
	}
	tests := []struct {
		name              string
		args              args
		want              string
		wantErr           bool
		predefinedObjects []runtime.Object
	}{
		{
			name: "extract key from cm",
			args: args{
				ns:    "default",
				sel:   v1.ConfigMapKeySelector{Key: "tls-conf", LocalObjectReference: v1.LocalObjectReference{Name: "tls-cm"}},
				cache: map[string]*v1.ConfigMap{},
			},
			predefinedObjects: []runtime.Object{
				&v1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{Name: "tls-cm", Namespace: "default"},
					Data:       map[string]string{"tls-conf": "secret-data"},
				},
			},
			want:    "secret-data",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := []runtime.Object{}
			obj = append(obj, tt.predefinedObjects...)
			fclient := fake.NewFakeClientWithScheme(testGetScheme(), obj...)

			got, err := getCredFromConfigMap(context.TODO(), fclient, tt.args.ns, tt.args.sel, tt.args.cacheKey, tt.args.cache)
			if (err != nil) != tt.wantErr {
				t.Errorf("getCredFromConfigMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getCredFromConfigMap() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getCredFromSecret(t *testing.T) {
	type args struct {
		ns       string
		sel      v1.SecretKeySelector
		cacheKey string
		cache    map[string]*v1.Secret
	}
	tests := []struct {
		name              string
		args              args
		want              string
		wantErr           bool
		predefinedObjects []runtime.Object
	}{
		{
			name: "extract tls key data from secret",
			args: args{
				ns: "default",
				sel: v1.SecretKeySelector{LocalObjectReference: v1.LocalObjectReference{
					Name: "tls-secret"},
					Key: "key.pem"},
				cacheKey: "tls-secret",
				cache:    map[string]*v1.Secret{},
			},
			want:    "tls-key-data",
			wantErr: false,
			predefinedObjects: []runtime.Object{
				&v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "tls-secret",
						Namespace: "default",
					},
					Data: map[string][]byte{"ca.crt": []byte(`ca-data`), "key.pem": []byte(`tls-key-data`)},
				},
			},
		},
		{
			name: "fail extract missing tls cert data from secret",
			args: args{
				ns: "default",
				sel: v1.SecretKeySelector{LocalObjectReference: v1.LocalObjectReference{
					Name: "tls-secret"},
					Key: "cert.pem"},
				cacheKey: "tls-secret",
				cache:    map[string]*v1.Secret{},
			},
			want:    "",
			wantErr: true,
			predefinedObjects: []runtime.Object{
				&v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "tls-secret",
						Namespace: "default",
					},
					Data: map[string][]byte{"ca.crt": []byte(`ca-data`), "key.pem": []byte(`tls-key-data`)},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := []runtime.Object{}
			obj = append(obj, tt.predefinedObjects...)
			fclient := fake.NewFakeClientWithScheme(testGetScheme(), obj...)

			got, err := getCredFromSecret(context.TODO(), fclient, tt.args.ns, tt.args.sel, tt.args.cacheKey, tt.args.cache)
			if (err != nil) != tt.wantErr {
				t.Errorf("getCredFromSecret() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getCredFromSecret() got = %v, want %v", got, tt.want)
			}
		})
	}
}

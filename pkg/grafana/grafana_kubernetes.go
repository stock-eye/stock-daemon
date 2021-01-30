package grafana

import (
	"bytes"
	"encoding/json"
	"log"
	"strings"
	"text/template"

	grafanav1 "github.com/linclaus/grafana-operator/api/v1"
	"github.com/linclaus/stock-daemon/pkg/kubernetes"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var GrafanaDashboardGroupVersionResource = schema.GroupVersionResource{
	Group:    "grafana.monitoring.io",
	Version:  "v1",
	Resource: "grafanadashboards",
}

var GrafanaDashboardGroupVersionKind = schema.GroupVersionKind{
	Group:   "grafana.monitoring.io",
	Version: "v1",
	Kind:    "GrafanaDashboard",
}

func MakeAggerationGrafanaDashboardResource(name, title, increaseExpr, increaseExpr10m, decreaseExpr, decreaseExpr10m string) grafanav1.GrafanaDashboard {
	t, err := template.ParseFiles(viper.GetString("GRAFANA_AGG_TEMPLATE"))
	if err != nil {
		panic(err)
	}
	buf := new(bytes.Buffer)
	t.Execute(buf, map[string]string{
		"Name":            name,
		"Title":           title,
		"IncreaseExpr":    increaseExpr,
		"IncreaseExpr10m": increaseExpr10m,
		"DecreaseExpr":    decreaseExpr,
		"DecreaseExpr10m": decreaseExpr10m,
	})
	gd := grafanav1.GrafanaDashboard{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "Aggregation",
			Namespace: "default",
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "GrafanaDashboard",
			APIVersion: "grafana.monitoring.io/v1",
		},
	}
	err = yaml.Unmarshal(buf.Bytes(), &gd)
	if err != nil {
		log.Println(err.Error())
	}
	return gd
}

func GetGrafanaDashboard(namespace, name string) (grafanav1.GrafanaDashboard, error) {
	gd := grafanav1.GrafanaDashboard{}
	gdb, err := kubernetes.GetCustomObject(namespace, GrafanaDashboardGroupVersionResource, name)
	if err != nil {
		return gd, err
	}
	err = json.Unmarshal(gdb, &gd)
	return gd, err
}

func DeleteGrafanaDashboard(namespace, name string) error {
	err := kubernetes.DeleteCustomObject(namespace, GrafanaDashboardGroupVersionResource, name)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil
		} else {
			return err
		}
	}
	return nil
}

func CreateOrUpdateGrafanaDashboard(gd grafanav1.GrafanaDashboard) error {
	ogdb, err := kubernetes.GetCustomObject(gd.Namespace, GrafanaDashboardGroupVersionResource, gd.Name)
	ogd := grafanav1.GrafanaDashboard{}
	json.Unmarshal(ogdb, &ogd)
	gdb, _ := json.Marshal(gd)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return kubernetes.CreateCustomObject(gd.Namespace, GrafanaDashboardGroupVersionResource, GrafanaDashboardGroupVersionKind, string(gdb))
		} else {
			return err
		}
	} else {
		return kubernetes.UpdateCustomObject(gd.Namespace, GrafanaDashboardGroupVersionResource, GrafanaDashboardGroupVersionKind, string(gdb), ogd.GetResourceVersion())
	}
}

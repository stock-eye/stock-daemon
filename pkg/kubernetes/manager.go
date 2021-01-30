package kubernetes

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"

	"github.com/sirupsen/logrus"
)

var k8sClient dynamic.Interface

func init() {
	config, err := clientcmd.BuildConfigFromFlags("http://localhost:8001", "")
	if err != nil {
		logrus.Error("Build k8s config failed, err: ", err.Error())
	}

	k8sClient, err = dynamic.NewForConfig(config)
	if err != nil {
		logrus.Error("Get k8s client failed, err: ", err.Error())
	}
}

func GetCustomObject(namespace string, resource schema.GroupVersionResource, croName string) ([]byte, error) {
	data, err := k8sClient.Resource(resource).Namespace(namespace).Get(context.TODO(), croName, metav1.GetOptions{})

	if err != nil {
		logrus.Error(err)

		return nil, fmt.Errorf("get %s custom object with name [%s] in namespace [%s] failed, err info %s",
			resource, namespace, croName, err.Error())
	}

	result, err := data.MarshalJSON()

	return result, nil
}

func DeleteCustomObject(namespace string, resource schema.GroupVersionResource, croName string) error {
	return k8sClient.Resource(resource).Namespace(namespace).Delete(context.TODO(), croName, metav1.DeleteOptions{})
}

func UpdateCustomObject(namespace string, resource schema.GroupVersionResource, kind schema.GroupVersionKind, cro string, resourceVersion string) error {
	decoder := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	obj := &unstructured.Unstructured{}
	if _, _, err := decoder.Decode([]byte(cro), &kind, obj); err != nil {
		return err
	}

	obj.SetResourceVersion(resourceVersion)
	_, err := k8sClient.Resource(resource).Namespace(namespace).Update(context.TODO(), obj, metav1.UpdateOptions{})
	return err
}

func CreateCustomObject(namespace string, resource schema.GroupVersionResource, kind schema.GroupVersionKind, cro string) error {
	decoder := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	obj := &unstructured.Unstructured{}
	if _, _, err := decoder.Decode([]byte(cro), &kind, obj); err != nil {
		return err
	}

	_, err := k8sClient.Resource(resource).Namespace(namespace).Create(context.TODO(), obj, metav1.CreateOptions{})
	return err
}

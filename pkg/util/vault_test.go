package util

import (
	"fmt"
	"testing"

	"k8s.io/api/core/v1"
)

func TestFunc(t *testing.T) {

	data := NewConfigFormConfigMap("", &v1.ConfigMap{
		Data: map[string]string{
			"name": "inmem",
		},
	})

	fmt.Println(data)
	fmt.Println("----------------------------")

	res := NewConfigWithDefaultParams()
	fmt.Println(res)

	fmt.Println("-----------------")

	res = NewConfigWithEtcd(res, "124.0.0.1")
	fmt.Println(res)
}

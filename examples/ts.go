package main

import "fmt"

func main() {
	data := map[string]interface{}{
		"name": "app",
		"info": map[string]string{"k0": "v0"},
		"sub": map[string]interface{}{
			"subK1": "val1",
			"subK2": "val2",
			"subK3": "val2",
			"subK4": []string{"v1", "v2"},
		},
	}

	fmt.Println(data)
}

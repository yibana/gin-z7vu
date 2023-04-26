package task

import (
	"encoding/json"
	"gin/amazon"
	"io/ioutil"
	"path/filepath"
	"strings"
)

var TaskInstance = NewTask()
var CategoryPaths = get_category_paths()

type CategoryPath struct {
	Path string
	Url  string
}

func get_category_paths() []CategoryPath {
	filepath := filepath.Join(".", "category.json")
	categorys, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil
	}
	var category amazon.Category
	err = json.Unmarshal(categorys, &category)
	if err != nil {
		return nil
	}
	// 遍历category，获取所有末尾节点的path
	var paths []CategoryPath
	var f func(path []string, categorys *amazon.Category)
	f = func(path []string, categorys *amazon.Category) {
		if len(categorys.Sub) == 0 {
			paths = append(paths, CategoryPath{Path: strings.Join(path, " > "), Url: categorys.Url})
		} else {
			for _, sub := range categorys.Sub {
				f(append(path, sub.Name), sub)
			}
		}
	}
	f([]string{category.Name}, &category)
	return paths
}

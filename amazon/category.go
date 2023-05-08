package amazon

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"sync"
)

type Category struct {
	Url       string
	Name      string
	Sub       []*Category
	Maybe_end bool      `json:"-"` // 有可能是最后一级
	Parent    *Category `json:"-"`
}

type CategoryManger struct {
	root       *Category
	lock       sync.Mutex
	Uniqueness map[string]bool
	cansave    bool
}

func (c *CategoryManger) SaveRootToFile(fn string) error {
	if !c.cansave {
		return nil
	}
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.root != nil {
		marshal, err := json.Marshal(c.root)
		if err != nil {
			return err
		}
		return ioutil.WriteFile(fn, marshal, 0644)
	}
	return nil
}

func (c *CategoryManger) Add(parent interface{}, category *Category) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.root == nil {
		c.root = category
		return
	}

	c.cansave = true
	var p *Category
	if parent == nil {
		p = c.root
	} else {
		p = parent.(*Category)
	}
	category.Parent = p
	p.Sub = append(p.Sub, category)
}

type TreeNode struct {
	Value    string      `json:"value"`
	Label    string      `json:"label"`
	Children []*TreeNode `json:"children"`
}

type TreeSelect struct {
	Key      string        `json:"key"`
	Label    string        `json:"label"`
	Data     string        `json:"data"`
	Children []*TreeSelect `json:"children,omitempty"`
}

func ConvertToTreeNode(Path string, category *Category) *TreeNode {
	value := category.Name
	if len(Path) > 0 {
		value = fmt.Sprintf("%s > %s", Path, category.Name)
	}
	treeNode := &TreeNode{
		Value: value,
		Label: category.Name,
	}
	if len(category.Sub) > 0 {
		treeNode.Children = make([]*TreeNode, len(category.Sub))
		for i, sub := range category.Sub {
			treeNode.Children[i] = ConvertToTreeNode(value, sub)
		}
		treeNode.Value = fmt.Sprintf("%s(%d)", value, rand.Intn(10000))
	} else {
		treeNode.Value = fmt.Sprintf("[End%d]%s", rand.Intn(10000), value)
	}
	return treeNode
}

func ConvertToTreeSelect(Key, Path string, category *Category) *TreeSelect {
	value := category.Name
	if len(Path) > 0 {
		value = fmt.Sprintf("%s > %s", Path, category.Name)
	}
	treeNode := &TreeSelect{
		Data:  value,
		Label: category.Name,
		Key:   Key,
	}
	if len(category.Sub) > 0 {
		treeNode.Children = make([]*TreeSelect, len(category.Sub))
		for i, sub := range category.Sub {
			treeNode.Children[i] = ConvertToTreeSelect(fmt.Sprintf("%s-%d", Key, i), value, sub)
		}
		//treeNode.Data = fmt.Sprintf("%s(%d)", value, rand.Intn(10000))
	} else {
		//treeNode.Data = fmt.Sprintf("[End%d]%s", rand.Intn(10000), value)
	}
	return treeNode
}

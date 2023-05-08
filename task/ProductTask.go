package task

type ProductTask struct {
	*BaseTask
}

func NewProductTask() *ProductTask {
	return &ProductTask{NewBaseTask("www.amazon.ca")}
}

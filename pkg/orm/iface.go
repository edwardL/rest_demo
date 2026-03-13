package orm

type Model interface {
	Mapping() []*Mapping
}

// type set（类型集）约束
// 定义一个约束：PModel 表示 “某个类型 T 的指针，并且实现了 Model”
type PModel[T any] interface {
	*T
	Model
}

type emptyModel struct{}

func (*emptyModel) Mapping() []*Mapping {
	return nil
}

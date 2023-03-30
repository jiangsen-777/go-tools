package main

type Node struct {
	Id       int64   `json:"id"`                 // id
	Name     string  `json:"name"`               // 名称
	ParentId int64   `json:"parent_id,optional"` // 父id
	Children []*Node `json:"children"`           // 子节点
}

func BuildTree(partsTypes []*Node, typeParentTree map[int64][]*Node) []*Node {
	if partsTypes == nil {
		return make([]*Node, 0)
	}
	for _, partsType := range partsTypes {
		partsType.Children = BuildTree(typeParentTree[partsType.Id], typeParentTree)
	}
	return partsTypes
}

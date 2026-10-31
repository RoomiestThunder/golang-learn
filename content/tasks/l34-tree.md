---
title: Бинарное дерево поиска
level: medium
---
@@ statement
Тип `TreeNode` уже объявлен. Реализуй:

- `Insert(root *TreeNode, v int) *TreeNode` — вставить значение в BST (дубликаты игнорировать), вернуть корень (для `nil` — новый узел).
- `Inorder(root *TreeNode) []int` — обход «лево–узел–право» (для BST — отсортированный порядок). Пустое дерево → пустой (не nil) срез.
- `MaxDepth(root *TreeNode) int` — высота (пустое — 0).
- `LevelOrder(root *TreeNode) [][]int` — обход по уровням (BFS), пустое дерево → пустой результат.
- `IsValidBST(root *TreeNode) bool` — настоящая проверка свойства BST (не только «левый < узел < правый» у соседей!).
- `LowestCommonAncestor(root *TreeNode, p, q int) *TreeNode` — LCA двух значений **в BST** за O(h), без обхода всего дерева; значения гарантированно есть в дереве.
- `InvertTree(root *TreeNode) *TreeNode` — зеркально поменять местами детей у всех узлов.
@@ starter
package solution

type TreeNode struct {
	Val         int
	Left, Right *TreeNode
}

func Insert(root *TreeNode, v int) *TreeNode { return root }

func Inorder(root *TreeNode) []int { return nil }

func MaxDepth(root *TreeNode) int { return 0 }

func LevelOrder(root *TreeNode) [][]int { return nil }

func IsValidBST(root *TreeNode) bool { return false }

func LowestCommonAncestor(root *TreeNode, p, q int) *TreeNode { return nil }

func InvertTree(root *TreeNode) *TreeNode { return root }
@@ tests
package solution

import (
	"reflect"
	"slices"
	"testing"
)

func bst(vals ...int) *TreeNode {
	var root *TreeNode
	for _, v := range vals {
		root = Insert(root, v)
	}
	return root
}

func TestInsertInorder(t *testing.T) {
	root := bst(5, 3, 8, 1, 4, 7, 9, 3, 5)
	if got := Inorder(root); !slices.Equal(got, []int{1, 3, 4, 5, 7, 8, 9}) {
		t.Errorf("Inorder = %v (дубликаты игнорируются)", got)
	}
	if root.Val != 5 || root.Left.Val != 3 || root.Right.Val != 8 {
		t.Error("структура дерева")
	}
	if got := Inorder(nil); got == nil || len(got) != 0 {
		t.Errorf("Inorder(nil) = %#v, нужен пустой не-nil срез", got)
	}
}

func TestDepthAndLevels(t *testing.T) {
	root := bst(5, 3, 8, 1, 4, 7, 9)
	if MaxDepth(root) != 3 || MaxDepth(nil) != 0 || MaxDepth(bst(1, 2, 3, 4)) != 4 {
		t.Error("MaxDepth")
	}
	want := [][]int{{5}, {3, 8}, {1, 4, 7, 9}}
	if got := LevelOrder(root); !reflect.DeepEqual(got, want) {
		t.Errorf("LevelOrder = %v", got)
	}
	if len(LevelOrder(nil)) != 0 {
		t.Error("LevelOrder(nil)")
	}
}

func TestIsValidBST(t *testing.T) {
	if !IsValidBST(bst(5, 3, 8, 1, 4)) || !IsValidBST(nil) {
		t.Error("валидное дерево")
	}
	// дерево, где у соседей порядок верный, но 6 в левом поддереве 5 нарушает свойство
	bad := &TreeNode{5, &TreeNode{3, nil, &TreeNode{6, nil, nil}}, &TreeNode{8, nil, nil}}
	if IsValidBST(bad) {
		t.Error("6 в левом поддереве корня 5 — не BST")
	}
	dup := &TreeNode{2, &TreeNode{2, nil, nil}, nil}
	if IsValidBST(dup) {
		t.Error("дубликаты недопустимы")
	}
}

func TestLCA(t *testing.T) {
	root := bst(6, 2, 8, 0, 4, 7, 9, 3, 5)
	tests := []struct{ p, q, want int }{{2, 8, 6}, {2, 4, 2}, {3, 5, 4}, {0, 5, 2}, {7, 9, 8}, {6, 9, 6}}
	for _, tt := range tests {
		if got := LowestCommonAncestor(root, tt.p, tt.q); got == nil || got.Val != tt.want {
			t.Errorf("LCA(%d, %d) = %v, want %d", tt.p, tt.q, got, tt.want)
		}
	}
}

func TestInvert(t *testing.T) {
	root := InvertTree(bst(4, 2, 7, 1, 3, 6, 9))
	if got := LevelOrder(root); !reflect.DeepEqual(got, [][]int{{4}, {7, 2}, {9, 6, 3, 1}}) {
		t.Errorf("после инверсии: %v", got)
	}
	if InvertTree(nil) != nil {
		t.Error("InvertTree(nil)")
	}
}
@@ solution
package solution

type TreeNode struct {
	Val         int
	Left, Right *TreeNode
}

func Insert(root *TreeNode, v int) *TreeNode {
	if root == nil {
		return &TreeNode{Val: v}
	}
	switch {
	case v < root.Val:
		root.Left = Insert(root.Left, v)
	case v > root.Val:
		root.Right = Insert(root.Right, v)
	}
	return root
}

func Inorder(root *TreeNode) []int {
	out := []int{}
	var walk func(n *TreeNode)
	walk = func(n *TreeNode) {
		if n == nil {
			return
		}
		walk(n.Left)
		out = append(out, n.Val)
		walk(n.Right)
	}
	walk(root)
	return out
}

func MaxDepth(root *TreeNode) int {
	if root == nil {
		return 0
	}
	return 1 + max(MaxDepth(root.Left), MaxDepth(root.Right))
}

func LevelOrder(root *TreeNode) [][]int {
	if root == nil {
		return nil
	}
	var res [][]int
	queue := []*TreeNode{root}
	for len(queue) > 0 {
		size := len(queue)
		level := make([]int, 0, size)
		for i := 0; i < size; i++ {
			n := queue[0]
			queue = queue[1:]
			level = append(level, n.Val)
			if n.Left != nil {
				queue = append(queue, n.Left)
			}
			if n.Right != nil {
				queue = append(queue, n.Right)
			}
		}
		res = append(res, level)
	}
	return res
}

func IsValidBST(root *TreeNode) bool {
	var check func(n *TreeNode, lo, hi *int) bool
	check = func(n *TreeNode, lo, hi *int) bool {
		if n == nil {
			return true
		}
		if (lo != nil && n.Val <= *lo) || (hi != nil && n.Val >= *hi) {
			return false
		}
		return check(n.Left, lo, &n.Val) && check(n.Right, &n.Val, hi)
	}
	return check(root, nil, nil)
}

func LowestCommonAncestor(root *TreeNode, p, q int) *TreeNode {
	for root != nil {
		switch {
		case p < root.Val && q < root.Val:
			root = root.Left
		case p > root.Val && q > root.Val:
			root = root.Right
		default:
			return root
		}
	}
	return nil
}

func InvertTree(root *TreeNode) *TreeNode {
	if root == nil {
		return nil
	}
	root.Left, root.Right = InvertTree(root.Right), InvertTree(root.Left)
	return root
}
@@ hint
`IsValidBST`: передавай вниз допустимый диапазон `(lo, hi)`. LCA в BST: если оба значения меньше корня — иди влево, если оба больше — вправо, иначе корень и есть LCA.

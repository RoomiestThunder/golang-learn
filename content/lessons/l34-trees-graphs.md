---
id: l34
week: 7
day: 4
minutes: 120
title: Алгоритмы IV: деревья и графы
goals: Рекурсия по деревьям, BFS/DFS, обход сетки, топологическая сортировка, union-find, Дейкстра
lc: invert-binary-tree | Invert Binary Tree | Easy | Рекурсия: поменяй детей местами и инвертируй поддеревья
lc: maximum-depth-of-binary-tree | Maximum Depth of Binary Tree | Easy | 1 + max(глубина левого, глубина правого)
lc: binary-tree-level-order-traversal | Binary Tree Level Order Traversal | Medium | BFS с обработкой уровня по размеру очереди
lc: validate-binary-search-tree | Validate Binary Search Tree | Medium | Передавай допустимый диапазон (lo, hi) в рекурсию
lc: lowest-common-ancestor-of-a-binary-search-tree | Lowest Common Ancestor of a Binary Search Tree | Medium | Свойство BST: спускайся, пока p и q по одну сторону
lc: number-of-islands | Number of Islands | Medium | DFS/BFS по сетке с пометкой посещённых
lc: course-schedule | Course Schedule | Medium | Поиск цикла: топологическая сортировка (Kahn)
lc: rotting-oranges | Rotting Oranges | Medium | Многоисточниковый BFS по сетке
lc: clone-graph | Clone Graph | Medium | DFS/BFS + map старый узел → клон
lc: network-delay-time | Network Delay Time | Medium | Дейкстра на куче
lc: redundant-connection | Redundant Connection | Medium | Union-Find: ребро, соединяющее уже связанные вершины
---

## Бинарные деревья

```go
type TreeNode struct {
	Val         int
	Left, Right *TreeNode
}
```

Почти все задачи — **рекурсия по трём шагам**: (1) базовый случай (`nil`), (2) решение для левого и правого поддерева, (3) комбинация.

```go run
package main

import "fmt"

type TreeNode struct {
	Val         int
	Left, Right *TreeNode
}

func maxDepth(n *TreeNode) int {
	if n == nil {
		return 0
	}
	return 1 + max(maxDepth(n.Left), maxDepth(n.Right))
}

func invert(n *TreeNode) *TreeNode {
	if n == nil {
		return nil
	}
	n.Left, n.Right = invert(n.Right), invert(n.Left)
	return n
}

// BST-свойство: все слева < узел < все справа, поэтому передаём границы вниз
func isValidBST(n *TreeNode, lo, hi *int) bool {
	if n == nil {
		return true
	}
	if (lo != nil && n.Val <= *lo) || (hi != nil && n.Val >= *hi) {
		return false
	}
	return isValidBST(n.Left, lo, &n.Val) && isValidBST(n.Right, &n.Val, hi)
}

func inorder(n *TreeNode, visit func(int)) { // для BST даёт отсортированный порядок
	if n == nil {
		return
	}
	inorder(n.Left, visit)
	visit(n.Val)
	inorder(n.Right, visit)
}

// BFS по уровням
func levelOrder(root *TreeNode) [][]int {
	if root == nil {
		return nil
	}
	var res [][]int
	queue := []*TreeNode{root}
	for len(queue) > 0 {
		size := len(queue) // сколько узлов на текущем уровне
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

func main() {
	root := &TreeNode{4, &TreeNode{2, &TreeNode{1, nil, nil}, &TreeNode{3, nil, nil}}, &TreeNode{7, nil, nil}}
	fmt.Println(maxDepth(root), isValidBST(root, nil, nil), levelOrder(root))
	var vals []int
	inorder(root, func(v int) { vals = append(vals, v) })
	fmt.Println(vals)
	invert(root)
	fmt.Println(levelOrder(root))
}
```

Обходы: **preorder** (узел, лево, право), **inorder** (лево, узел, право — для BST отсортировано), **postorder** (лево, право, узел — когда ответ зависит от детей), **level-order (BFS)** — очередь.

Рекурсия с накоплением результата: передавай аккумулятор (слайс или замыкание); глубина рекурсии для вырожденного дерева (цепочка) может быть большой — в Go стек растёт, но для миллионов уровней лучше итеративный обход со своим стеком.

Частые задачи: диаметр дерева, баланс, симметрия, LCA, путь с максимальной суммой, сериализация, k-й наименьший в BST, вид справа.

## Графы

Представление:

```go
// список смежности — стандарт для разреженных графов
adj := make([][]int, n)
for _, e := range edges {
	adj[e[0]] = append(adj[e[0]], e[1])
	adj[e[1]] = append(adj[e[1]], e[0]) // для неориентированного
}
// взвешенный: type Edge struct{ to, w int }; adj [][]Edge
// сетка: соседи через массив смещений dirs = [][2]int{{1,0},{-1,0},{0,1},{0,-1}}
```

### BFS — кратчайший путь по числу рёбер и обход «волной»

```go run
package main

import "fmt"

// длина кратчайшего пути в сетке (0 — свободно, 1 — стена) из (0,0) в правый нижний угол; -1, если нет
func shortestPath(grid [][]int) int {
	n, m := len(grid), len(grid[0])
	if grid[0][0] == 1 || grid[n-1][m-1] == 1 {
		return -1
	}
	dirs := [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}
	dist := make([][]int, n)
	for i := range dist {
		dist[i] = make([]int, m)
		for j := range dist[i] {
			dist[i][j] = -1
		}
	}
	dist[0][0] = 0
	queue := [][2]int{{0, 0}}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		if cur == [2]int{n - 1, m - 1} {
			return dist[cur[0]][cur[1]]
		}
		for _, d := range dirs {
			r, c := cur[0]+d[0], cur[1]+d[1]
			if r >= 0 && r < n && c >= 0 && c < m && grid[r][c] == 0 && dist[r][c] == -1 {
				dist[r][c] = dist[cur[0]][cur[1]] + 1
				queue = append(queue, [2]int{r, c})
			}
		}
	}
	return -1
}

func main() {
	grid := [][]int{
		{0, 0, 0},
		{1, 1, 0},
		{0, 0, 0},
	}
	fmt.Println(shortestPath(grid))
}
```

**Правило:** отмечай вершину посещённой **при добавлении** в очередь (а не при извлечении) — иначе будут дубликаты.

### DFS — компоненты, поиск цикла, перебор путей

```go
func numIslands(grid [][]byte) int {
	count := 0
	var dfs func(r, c int)
	dfs = func(r, c int) {
		if r < 0 || r >= len(grid) || c < 0 || c >= len(grid[0]) || grid[r][c] != '1' {
			return
		}
		grid[r][c] = '0' // «затопили» — отметка посещения на месте
		dfs(r+1, c); dfs(r-1, c); dfs(r, c+1); dfs(r, c-1)
	}
	for r := range grid {
		for c := range grid[r] {
			if grid[r][c] == '1' {
				count++
				dfs(r, c)
			}
		}
	}
	return count
}
```

Рекурсивные замыкания в Go требуют предварительного объявления: `var dfs func(...)`, затем присваивание.

### Топологическая сортировка (Kahn)

Для ориентированного ациклического графа: порядок вершин, где каждое ребро идёт «вперёд». Считаем входящие степени, кладём в очередь вершины с нулевой степенью, «удаляем» их, уменьшая степени соседей. Если обработано меньше `n` вершин — в графе **цикл**. Применения: порядок сборки, зависимости курсов, планирование.

### Union-Find (DSU)

Множества с операциями `Find` (представитель) и `Union` за почти O(1) (сжатие путей + ранги). Применения: компоненты связности, поиск цикла в неориентированном графе, алгоритм Краскала.

```go
type DSU struct{ parent, rank []int }

func NewDSU(n int) *DSU {
	p := make([]int, n)
	for i := range p { p[i] = i }
	return &DSU{p, make([]int, n)}
}

func (d *DSU) Find(x int) int {
	if d.parent[x] != x {
		d.parent[x] = d.Find(d.parent[x]) // сжатие путей
	}
	return d.parent[x]
}

func (d *DSU) Union(a, b int) bool { // false, если уже в одном множестве
	ra, rb := d.Find(a), d.Find(b)
	if ra == rb { return false }
	if d.rank[ra] < d.rank[rb] { ra, rb = rb, ra }
	d.parent[rb] = ra
	if d.rank[ra] == d.rank[rb] { d.rank[ra]++ }
	return true
}
```

### Дейкстра — кратчайшие пути во взвешенном графе (веса ≥ 0)

Куча `(расстояние, вершина)`; извлекаем ближайшую, «релаксируем» рёбра. Сложность O((V+E) log V). Для отрицательных весов — Беллман–Форд; для всех пар — Флойд–Уоршелл; для сетки без весов хватает BFS.

## Как выбрать

| Задача | Метод |
|---|---|
| Кратчайший путь без весов, ближайшее, «волна» | BFS |
| Компоненты, все пути, проверка достижимости, цикл | DFS |
| Порядок с зависимостями | топосортировка |
| Динамическая связность, цикл в неориентированном | Union-Find |
| Кратчайший путь с весами ≥ 0 | Дейкстра |
| Дерево | рекурсия (DFS) / BFS по уровням |

## Проверь себя

<details><summary>Какой обход BST даёт отсортированную последовательность?</summary>Inorder (левое поддерево, узел, правое).</details>

<details><summary>Когда помечать вершину посещённой в BFS?</summary>При добавлении в очередь — чтобы одна вершина не попала в очередь несколько раз.</details>

<details><summary>Как обнаружить цикл в ориентированном графе?</summary>Топологическая сортировка (Kahn): если обработаны не все вершины — есть цикл; либо DFS с тремя цветами.</details>

<details><summary>Какова сложность Дейкстры на куче?</summary>O((V + E) log V).</details>

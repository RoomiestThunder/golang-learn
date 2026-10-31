---
title: Графы: BFS, топосортировка, Дейкстра, DSU
level: hard
---
@@ statement
Четыре задачи на графы.

1. `ShortestPathGrid(grid [][]int) int` — длина кратчайшего пути (в шагах) из левого верхнего угла в правый нижний по клеткам `0` (4 направления; `1` — стена). Нет пути (или старт/финиш — стена) → `-1`. Старт совпадает с финишем (сетка 1×1, клетка свободна) → `0`.
2. `TopoSort(n int, edges [][2]int) ([]int, bool)` — вершины `0..n-1`, ребро `[a, b]` значит «a раньше b». Вернуть любой корректный порядок и `true`; при наличии цикла — `nil, false`.
3. `ShortestPaths(n int, edges [][3]int, src int) []int` — алгоритм Дейкстры на куче (`container/heap`): ребро `[from, to, weight]` (ориентированное, вес ≥ 0). Результат — расстояния от `src`, недостижимые вершины — `-1`.
4. `CountComponents(n int, edges [][2]int) int` — число компонент связности неориентированного графа (используй Union-Find со сжатием путей).
@@ starter
package solution

func ShortestPathGrid(grid [][]int) int { return 0 }

func TopoSort(n int, edges [][2]int) ([]int, bool) { return nil, false }

func ShortestPaths(n int, edges [][3]int, src int) []int { return nil }

func CountComponents(n int, edges [][2]int) int { return 0 }
@@ tests
package solution

import (
	"math/rand/v2"
	"slices"
	"testing"
	"time"
)

func TestShortestPathGrid(t *testing.T) {
	tests := []struct {
		grid [][]int
		want int
	}{
		{[][]int{{0, 0, 0}, {1, 1, 0}, {0, 0, 0}}, 4},
		{[][]int{{0}}, 0},
		{[][]int{{1}}, -1},
		{[][]int{{0, 1}, {1, 0}}, -1},
		{[][]int{{0, 0, 0}, {0, 0, 0}, {0, 0, 0}}, 4},
		{[][]int{{0, 1, 0, 0, 0}, {0, 1, 0, 1, 0}, {0, 0, 0, 1, 0}, {1, 1, 1, 1, 0}, {0, 0, 0, 0, 0}}, 12},
	}
	for i, tt := range tests {
		if got := ShortestPathGrid(tt.grid); got != tt.want {
			t.Errorf("случай %d: %d, want %d", i, got, tt.want)
		}
	}
	big := make([][]int, 1500)
	for i := range big {
		big[i] = make([]int, 1500)
	}
	start := time.Now()
	if got := ShortestPathGrid(big); got != 2998 {
		t.Errorf("пустая сетка 1500×1500: %d", got)
	}
	if time.Since(start) > 3*time.Second {
		t.Errorf("слишком медленно: %v", time.Since(start))
	}
}

func validTopo(order []int, n int, edges [][2]int) bool {
	if len(order) != n {
		return false
	}
	pos := make([]int, n)
	seen := make([]bool, n)
	for i, v := range order {
		if v < 0 || v >= n || seen[v] {
			return false
		}
		seen[v] = true
		pos[v] = i
	}
	for _, e := range edges {
		if pos[e[0]] >= pos[e[1]] {
			return false
		}
	}
	return true
}

func TestTopoSort(t *testing.T) {
	edges := [][2]int{{5, 2}, {5, 0}, {4, 0}, {4, 1}, {2, 3}, {3, 1}}
	order, ok := TopoSort(6, edges)
	if !ok || !validTopo(order, 6, edges) {
		t.Errorf("order = %v, ok = %v", order, ok)
	}
	if _, ok := TopoSort(3, [][2]int{{0, 1}, {1, 2}, {2, 0}}); ok {
		t.Error("цикл должен давать false")
	}
	if order, ok := TopoSort(3, nil); !ok || len(order) != 3 {
		t.Errorf("без рёбер: %v %v", order, ok)
	}
	if order, ok := TopoSort(0, nil); !ok || len(order) != 0 {
		t.Errorf("пустой граф: %v %v", order, ok)
	}
	if _, ok := TopoSort(2, [][2]int{{1, 1}}); ok {
		t.Error("петля — это цикл")
	}
}

func TestShortestPaths(t *testing.T) {
	edges := [][3]int{{0, 1, 4}, {0, 2, 1}, {2, 1, 2}, {1, 3, 1}, {2, 3, 5}}
	got := ShortestPaths(5, edges, 0)
	if !slices.Equal(got, []int{0, 3, 1, 4, -1}) {
		t.Errorf("got %v", got)
	}
	if got := ShortestPaths(3, [][3]int{{0, 1, 0}, {1, 2, 0}}, 1); !slices.Equal(got, []int{-1, 0, 0}) {
		t.Errorf("нулевые веса, ориентированность: %v", got)
	}
	// большой разреженный граф
	n := 100_000
	var big [][3]int
	for i := 0; i+1 < n; i++ {
		big = append(big, [3]int{i, i + 1, 1 + rand.IntN(5)})
		big = append(big, [3]int{i, (i * 7) % n, 1 + rand.IntN(50)})
	}
	start := time.Now()
	d := ShortestPaths(n, big, 0)
	if time.Since(start) > 3*time.Second || d[0] != 0 || d[n-1] <= 0 {
		t.Errorf("большой граф: %v, d[n-1]=%d", time.Since(start), d[n-1])
	}
}

func TestCountComponents(t *testing.T) {
	tests := []struct {
		n     int
		edges [][2]int
		want  int
	}{
		{5, [][2]int{{0, 1}, {1, 2}, {3, 4}}, 2}, {5, nil, 5}, {4, [][2]int{{0, 1}, {1, 2}, {2, 3}, {3, 0}}, 1}, {0, nil, 0}, {3, [][2]int{{0, 0}, {1, 2}}, 2},
	}
	for _, tt := range tests {
		if got := CountComponents(tt.n, tt.edges); got != tt.want {
			t.Errorf("CountComponents(%d, %v) = %d, want %d", tt.n, tt.edges, got, tt.want)
		}
	}
}
@@ solution
package solution

import "container/heap"

func ShortestPathGrid(grid [][]int) int {
	n := len(grid)
	if n == 0 || len(grid[0]) == 0 {
		return -1
	}
	m := len(grid[0])
	if grid[0][0] == 1 || grid[n-1][m-1] == 1 {
		return -1
	}
	dist := make([][]int32, n)
	for i := range dist {
		dist[i] = make([]int32, m)
		for j := range dist[i] {
			dist[i][j] = -1
		}
	}
	dist[0][0] = 0
	queue := [][2]int{{0, 0}}
	dirs := [4][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}
	for head := 0; head < len(queue); head++ {
		cur := queue[head]
		if cur[0] == n-1 && cur[1] == m-1 {
			return int(dist[cur[0]][cur[1]])
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

func TopoSort(n int, edges [][2]int) ([]int, bool) {
	adj := make([][]int, n)
	indeg := make([]int, n)
	for _, e := range edges {
		adj[e[0]] = append(adj[e[0]], e[1])
		indeg[e[1]]++
	}
	queue := []int{}
	for v := 0; v < n; v++ {
		if indeg[v] == 0 {
			queue = append(queue, v)
		}
	}
	order := make([]int, 0, n)
	for head := 0; head < len(queue); head++ {
		v := queue[head]
		order = append(order, v)
		for _, w := range adj[v] {
			if indeg[w]--; indeg[w] == 0 {
				queue = append(queue, w)
			}
		}
	}
	if len(order) != n {
		return nil, false
	}
	return order, true
}

type item struct{ node, dist int }
type pq []item

func (p pq) Len() int           { return len(p) }
func (p pq) Less(i, j int) bool { return p[i].dist < p[j].dist }
func (p pq) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p *pq) Push(x any)        { *p = append(*p, x.(item)) }
func (p *pq) Pop() any {
	old := *p
	x := old[len(old)-1]
	*p = old[:len(old)-1]
	return x
}

func ShortestPaths(n int, edges [][3]int, src int) []int {
	type edge struct{ to, w int }
	adj := make([][]edge, n)
	for _, e := range edges {
		adj[e[0]] = append(adj[e[0]], edge{e[1], e[2]})
	}
	dist := make([]int, n)
	for i := range dist {
		dist[i] = -1
	}
	dist[src] = 0
	h := &pq{{src, 0}}
	for h.Len() > 0 {
		cur := heap.Pop(h).(item)
		if cur.dist > dist[cur.node] {
			continue // устаревшая запись
		}
		for _, e := range adj[cur.node] {
			nd := cur.dist + e.w
			if dist[e.to] == -1 || nd < dist[e.to] {
				dist[e.to] = nd
				heap.Push(h, item{e.to, nd})
			}
		}
	}
	return dist
}

func CountComponents(n int, edges [][2]int) int {
	parent := make([]int, n)
	for i := range parent {
		parent[i] = i
	}
	var find func(x int) int
	find = func(x int) int {
		if parent[x] != x {
			parent[x] = find(parent[x])
		}
		return parent[x]
	}
	comps := n
	for _, e := range edges {
		if a, b := find(e[0]), find(e[1]); a != b {
			parent[a] = b
			comps--
		}
	}
	return comps
}
@@ hint
BFS: очередь + массив расстояний (-1 = не посещено); отмечай при добавлении в очередь. Kahn: считай входящие степени, обрабатывай вершины с нулевой степенью. Дейкстра: пропускай устаревшие записи кучи (`cur.dist > dist[cur.node]`).

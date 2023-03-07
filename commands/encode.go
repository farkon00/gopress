package commands

import (
	"encoding/binary"
	"fmt"
	"gopress/shared"
	"io/fs"
	"os"
	"sort"
	"strings"
)

type TreeNode struct {
	freq        int
	value       byte
	left, right *TreeNode
}

func left_pad(_s fmt.Stringer) (res string) {
	s := _s.String()
	for _, line := range strings.Split(s, "\n") {
		res = fmt.Sprintf("%s\n  %s", res, line)
	}
	return
}

func (node TreeNode) String() string {
	children := ""
	if node.left != nil {
		children = fmt.Sprintf("%s%s", left_pad(node.left), left_pad(node.right))
	}
	return fmt.Sprintf("%c(freq %d)%s", node.value, node.freq, children)
}

type InfBitSeq struct {
	length int
	data   []byte
}

func (bs *InfBitSeq) add_bit(bit bool) {
	if bs.length%8 == 0 {
		bs.data = append(bs.data, 0)
	}

	if bit {
		bs.data[bs.length/8] |= (1 << (bs.length % 8))
	}
	bs.length += 1
}

func (bs *InfBitSeq) concat(other shared.BitSeq) (res InfBitSeq) {
	for i := 0; i < other.Length; i++ {
		bs.add_bit((other.Data[i/8] & (1 << (i % 8))) != 0)
	}
	return
}

func get_freqs(input []byte) (res map[byte]int) {
	res = make(map[byte]int)
	for _, char := range input {
		prev, ok := res[char]
		if ok {
			res[char] = prev + 1
		} else {
			res[char] = 1
		}
	}
	return
}

func make_queue(freqs map[byte]int) (res []TreeNode) {
	for k, v := range freqs {
		res = append(res, TreeNode{freq: v, value: k})
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].freq > res[j].freq
	})
	return
}

func put_node(queue *[]TreeNode, node TreeNode) {
	inserted := false

	for i := range *queue {
		curr := (*queue)[len(*queue)-1-i]
		if curr.freq > node.freq {
			inserted = true
			*queue = append((*queue)[:len(*queue)-i], (*queue)[len(*queue)-i-1:]...)
			(*queue)[len(*queue)-i-1] = node
			break
		}
	}

	if !inserted {
		*queue = append([]TreeNode{node}, *queue...)
	}
}

func make_tree(queue []TreeNode) TreeNode {
	for len(queue) > 1 {
		left, right := queue[len(queue)-1], queue[len(queue)-2]
		queue = queue[:len(queue)-2]
		put_node(&queue, TreeNode{freq: left.freq + right.freq, left: &left, right: &right})
	}

	return queue[0]
}

func make_map(node TreeNode, path shared.BitSeq, encoding_map *map[byte]shared.BitSeq) {
	if node.left == nil {
		(*encoding_map)[node.value] = path
	} else {
		make_map(*node.left, path.Add_bit(false), encoding_map)
		make_map(*node.right, path.Add_bit(true), encoding_map)
	}
}

// Tree debug
// func generate_dot(node *TreeNode) string {
// 	if node.left == nil {
// 		return fmt.Sprintf("Node%p[label=%#v]", node, fmt.Sprintf("%#v\n%d", fmt.Sprintf("%c", node.value), node.freq))
// 	} else {
// 		node_repr := fmt.Sprintf("Node%p[label=\"%d\"]", node, node.freq)
// 		connections := fmt.Sprintf("Node%[1]p -> Node%[2]p[label=\"0\"]\nNode%[1]p -> Node%[3]p[label=\"1\"]",
// 			node, node.left, node.right)
// 		return fmt.Sprintf("%s\n%s\n%s\n%s", node_repr, connections, generate_dot(node.left), generate_dot(node.right))
// 	}
// }

func generate_content(bits InfBitSeq) (res []byte) {
	res = append(res, 0, 0, 0, 0)
	binary.LittleEndian.PutUint32(res, uint32(bits.length))
	res = append(res, bits.data...)
	return
}

func generate_map(m map[byte]shared.BitSeq) (res []byte) {
	for k, v := range m {
		res = append(res, 0, 0, 0, 0)
		binary.LittleEndian.PutUint32(res[len(res)-4:], uint32(v.Length))
		res = append(res, v.Data[:shared.Ceil_div(v.Length, 8)]...)
		res = append(res, k)
		// fmt.Println(res, v.length)
	}
	res = append(res, 0, 0, 0, 0)
	return
}

func Encode(config shared.Config, input []byte) {
	encoding_map := make(map[byte]shared.BitSeq)
	tree := make_tree(make_queue(get_freqs(input)))
	make_map(tree, shared.BitSeq{}, &encoding_map)

	res_content := InfBitSeq{}
	for _, char := range input {
		res_content.concat(encoding_map[char])
	}
	content := append(generate_map(encoding_map), generate_content(res_content)...)
	var filename string
	if config.Output == "" {
		filename = config.Filename + ".gprs"
	} else {
		filename = config.Output
	}
	err := os.WriteFile(filename, content, fs.FileMode(os.ModePerm))
	if err != nil {
		fmt.Println("Failed to write to file\n Traceback:")
		panic(err)
	}
	// Tree debug
	// os.WriteFile("graph.dot",
	// 	[]byte(fmt.Sprintf("digraph Tree {\n%s\n}", generate_dot(&tree))),
	// 	fs.FileMode(os.ModePerm))

	// Map debug
	// for k, v := range encoding_map {
	// 	fmt.Printf("%c: %08b(%d)\n", k, v.data, v.length)
	// }

	// Result debug
	// fmt.Printf("%d %08b\n", res_content.length, res_content.data)
}
